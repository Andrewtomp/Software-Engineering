// front-runner/internal/storefronttable/storefronttable_test.go
package storefronttable

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync"
	"testing"

	// Needed for unique email generation
	"front-runner/internal/coredbutils"
	"front-runner/internal/login" // Needed for session constants/setup
	"front-runner/internal/oauth" // Needed for oauth.Setup
	"front-runner/internal/usertable"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const projectDirName = "front-runner_backend"

// Global test variables
var (
	testDB           *gorm.DB
	testSessionStore *sessions.CookieStore
	setupEnvOnce     sync.Once
)

// setupTestEnvironment loads environment variables, initializes DB and session store for tests.
// It also clears relevant tables before each test run.
func setupTestEnvironment(t *testing.T) {
	t.Helper()

	setupEnvOnce.Do(func() {
		// Find project root
		re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
		cwd, _ := os.Getwd()
		rootPath := re.Find([]byte(cwd))
		if rootPath == nil {
			t.Fatalf("Could not find project root directory '%s' from '%s'", projectDirName, cwd)
		}

		// Load .env file
		envPath := string(rootPath) + `/.env`
		err := godotenv.Load(envPath)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Problem loading .env file from %s: %v", envPath, err)
		} else if err == nil {
			log.Printf("Loaded environment variables from %s for tests", envPath)
		}

		// Initialize DB connection
		coredbutils.ResetDBStateForTests()
		err = coredbutils.LoadEnv()
		require.NoError(t, err, "Failed to load core DB environment")
		var dbErr error
		testDB, dbErr = coredbutils.GetDB()
		require.NoError(t, dbErr, "Failed to get DB connection for tests")

		// Initialize Session Store for tests
		authKey := []byte("test-auth-key-32-bytes-long-000")
		encKey := []byte("test-enc-key-needs-to-be-32-byte") // 32 bytes
		require.True(t, len(encKey) == 16 || len(encKey) == 32, "Test encryption key must be 16 or 32 bytes")
		testSessionStore = sessions.NewCookieStore(authKey, encKey)
		testSessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}

		// Setup dependent packages
		usertable.Setup()                     // Uses coredbutils.GetDB()
		oauth.Setup(testSessionStore)         // Uses session store
		login.Setup(testDB, testSessionStore) // Uses DB and session store
		Setup()                               // Setup storefronttable package (uses coredbutils.GetDB() and loads key)

		// Run migrations once after setup
		usertable.MigrateUserDB()
		MigrateStorefrontDB() // Migrates StorefrontLink table
	})

	// Clear tables before each test function
	require.NoError(t, usertable.ClearUserTable(testDB), "Failed to clear user table")
	require.NoError(t, ClearStorefrontTable(testDB), "Failed to clear storefront table") // Use the package's Clear function
}

// Helper to create a test user directly in the DB
func createTestUser(t *testing.T, email, password string) *usertable.User {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password for test user")

	user := &usertable.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User " + email,
		Provider:     "local",
	}
	err = usertable.CreateUser(user)
	require.NoError(t, err, "Failed to create test user using usertable.CreateUser")
	createdUser, err := usertable.GetUserByEmail(email)
	require.NoError(t, err, "Failed to fetch created test user by email")
	require.NotNil(t, createdUser, "Fetched created test user should not be nil")
	return createdUser
}

// Helper to create an authenticated request
func createAuthenticatedRequest(t *testing.T, user *usertable.User, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, body)

	// Use literal strings as constants are not exported from login pkg
	const sessionName = "front-runner-session"
	const userSessionKey = "userID"

	// Create and save session to get cookie header
	session, err := testSessionStore.New(req, sessionName)
	require.NoError(t, err, "Failed to create new session for auth request")
	session.Values[userSessionKey] = user.ID
	rrCookieSetter := httptest.NewRecorder()
	err = testSessionStore.Save(req, rrCookieSetter, session)
	require.NoError(t, err, "Failed to save session to get cookie header")
	cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookieHeader, "Set-Cookie header should not be empty after saving session")

	// Add cookie to the actual request
	req.Header.Set("Cookie", cookieHeader)
	return req
}

// TestAddStorefront tests adding a storefront link.
func TestAddStorefront(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "addsf@example.com", "password123")

	t.Run("Success", func(t *testing.T) {
		payload := StorefrontLinkAddPayload{
			StoreType: "amazon_test",
			StoreName: "My Test Amazon Store",
			ApiKey:    "key123",
			ApiSecret: "secretABC",
			StoreId:   "storeXYZ",
			StoreUrl:  "http://amazon.test/storeXYZ",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		req := createAuthenticatedRequest(t, user, "POST", "/api/add_storefront", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		AddStorefront(rr, req) // Call the handler

		require.Equal(t, http.StatusCreated, rr.Code, "Expected status 201 Created, body: %s", rr.Body.String())

		var responseData StorefrontLinkReturn
		err = json.Unmarshal(rr.Body.Bytes(), &responseData)
		require.NoError(t, err, "Failed to unmarshal response")
		assert.Equal(t, payload.StoreType, responseData.StoreType)
		assert.Equal(t, payload.StoreName, responseData.StoreName)
		assert.Equal(t, payload.StoreId, responseData.StoreID)
		assert.Equal(t, payload.StoreUrl, responseData.StoreURL)
		assert.NotEmpty(t, responseData.ID)

		// Verify in DB
		var savedLink StorefrontLink
		dbResult := testDB.First(&savedLink, responseData.ID)
		require.NoError(t, dbResult.Error)
		assert.Equal(t, user.ID, savedLink.UserID)
		assert.Equal(t, payload.StoreType, savedLink.StoreType)
		assert.Equal(t, payload.StoreName, savedLink.StoreName)
		assert.Equal(t, payload.StoreId, savedLink.StoreID)
		assert.Equal(t, payload.StoreUrl, savedLink.StoreURL)

		// Verify credentials encryption
		require.NotEmpty(t, savedLink.Credentials)
		decryptedCreds, decryptErr := decryptCredentials(savedLink.Credentials)
		require.NoError(t, decryptErr)
		expectedCredsMap := map[string]string{"apiKey": payload.ApiKey, "apiSecret": payload.ApiSecret}
		expectedCredsJSON, _ := json.Marshal(expectedCredsMap)
		assert.JSONEq(t, string(expectedCredsJSON), decryptedCreds)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		payload := StorefrontLinkAddPayload{StoreType: "unauth_test"}
		bodyBytes, _ := json.Marshal(payload)

		// Create request WITHOUT authentication
		req := httptest.NewRequest("POST", "/api/add_storefront", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		AddStorefront(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized")
	})

	t.Run("MissingStoreType", func(t *testing.T) {
		payload := StorefrontLinkAddPayload{StoreName: "No Type Store"} // Missing StoreType
		bodyBytes, _ := json.Marshal(payload)

		req := createAuthenticatedRequest(t, user, "POST", "/api/add_storefront", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		AddStorefront(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing required field: storeType")
	})

	t.Run("DuplicateStoreNameAndType", func(t *testing.T) {
		// Add first link
		payload1 := StorefrontLinkAddPayload{StoreType: "duplicate_test", StoreName: "My Duplicate Store"}
		body1, _ := json.Marshal(payload1)
		req1 := createAuthenticatedRequest(t, user, "POST", "/api/add_storefront", bytes.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		rr1 := httptest.NewRecorder()
		AddStorefront(rr1, req1)
		require.Equal(t, http.StatusCreated, rr1.Code)

		// Attempt to add second link with same type and name
		payload2 := StorefrontLinkAddPayload{StoreType: "duplicate_test", StoreName: "My Duplicate Store"}
		body2, _ := json.Marshal(payload2)
		req2 := createAuthenticatedRequest(t, user, "POST", "/api/add_storefront", bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		rr2 := httptest.NewRecorder()
		AddStorefront(rr2, req2)

		assert.Equal(t, http.StatusConflict, rr2.Code)
		assert.Contains(t, rr2.Body.String(), "already exists")
	})
}

// TestGetUpdateDeleteFlow tests the full lifecycle: Get, Update, Delete.
func TestGetUpdateDeleteFlow(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "flow@example.com", "flowpassword")

	// 1. Add a link to work with
	addPayload := StorefrontLinkAddPayload{
		StoreType: "flow_test",
		StoreName: "Flow Store Original",
		ApiKey:    "flow_key",
		StoreId:   "flow_id_1",
		StoreUrl:  "http://flow.test/1",
	}
	addBodyBytes, _ := json.Marshal(addPayload)
	addReq := createAuthenticatedRequest(t, user, "POST", "/api/add_storefront", bytes.NewReader(addBodyBytes))
	addReq.Header.Set("Content-Type", "application/json")
	addRR := httptest.NewRecorder()
	AddStorefront(addRR, addReq)
	require.Equal(t, http.StatusCreated, addRR.Code, "Flow Setup: Failed to add initial link")
	var addedLinkResp StorefrontLinkReturn
	require.NoError(t, json.Unmarshal(addRR.Body.Bytes(), &addedLinkResp), "Flow Setup: Failed to parse add response")
	linkID := addedLinkResp.ID
	require.NotZero(t, linkID, "Added link ID should not be zero")

	// 2. Get Storefronts and verify the added link
	t.Run("GetAfterAdd", func(t *testing.T) {
		getReq := createAuthenticatedRequest(t, user, "GET", "/api/get_storefronts", nil)
		getRR := httptest.NewRecorder()
		GetStorefronts(getRR, getReq)
		require.Equal(t, http.StatusOK, getRR.Code, "Flow Get: Failed status")

		var links []StorefrontLinkReturn
		require.NoError(t, json.Unmarshal(getRR.Body.Bytes(), &links), "Flow Get: Failed to parse list")
		require.Len(t, links, 1, "Expected 1 link after adding")
		assert.Equal(t, linkID, links[0].ID)
		assert.Equal(t, addPayload.StoreName, links[0].StoreName)
		assert.Equal(t, addPayload.StoreType, links[0].StoreType)
		assert.Equal(t, addPayload.StoreId, links[0].StoreID)
		assert.Equal(t, addPayload.StoreUrl, links[0].StoreURL)
	})

	// 3. Update the link
	t.Run("UpdateLink", func(t *testing.T) {
		updatePayload := StorefrontLinkUpdatePayload{
			StoreName: "Flow Store UPDATED",
			StoreId:   "flow_id_2",
			StoreUrl:  "http://flow.test/2",
		}
		updateBodyBytes, _ := json.Marshal(updatePayload)
		updateURL := fmt.Sprintf("/api/update_storefront?id=%d", linkID)
		updateReq := createAuthenticatedRequest(t, user, "PUT", updateURL, bytes.NewReader(updateBodyBytes))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRR := httptest.NewRecorder()
		UpdateStorefront(updateRR, updateReq)
		require.Equal(t, http.StatusOK, updateRR.Code, "Flow Update: Failed status, body: %s", updateRR.Body.String())

		var updatedLinkResp StorefrontLinkReturn
		require.NoError(t, json.Unmarshal(updateRR.Body.Bytes(), &updatedLinkResp), "Flow Update: Failed to parse response")
		assert.Equal(t, linkID, updatedLinkResp.ID)
		assert.Equal(t, updatePayload.StoreName, updatedLinkResp.StoreName)
		assert.Equal(t, updatePayload.StoreId, updatedLinkResp.StoreID)
		assert.Equal(t, updatePayload.StoreUrl, updatedLinkResp.StoreURL)
		assert.Equal(t, addPayload.StoreType, updatedLinkResp.StoreType) // Type should not change

		// Verify update in DB
		var updatedLinkDB StorefrontLink
		dbResult := testDB.First(&updatedLinkDB, linkID)
		require.NoError(t, dbResult.Error)
		assert.Equal(t, updatePayload.StoreName, updatedLinkDB.StoreName)
		assert.Equal(t, updatePayload.StoreId, updatedLinkDB.StoreID)
		assert.Equal(t, updatePayload.StoreUrl, updatedLinkDB.StoreURL)
		assert.Equal(t, addPayload.StoreType, updatedLinkDB.StoreType) // Type should not change

		// Verify credentials didn't change
		decryptedCreds, decryptErr := decryptCredentials(updatedLinkDB.Credentials)
		require.NoError(t, decryptErr)
		origCredsMap := map[string]string{"apiKey": addPayload.ApiKey} // Reconstruct original creds map
		origCredsJSON, _ := json.Marshal(origCredsMap)
		assert.JSONEq(t, string(origCredsJSON), decryptedCreds)
	})

	// 4. Delete the link
	t.Run("DeleteLink", func(t *testing.T) {
		deleteURL := fmt.Sprintf("/api/delete_storefront?id=%d", linkID)
		deleteReq := createAuthenticatedRequest(t, user, "DELETE", deleteURL, nil)
		deleteRR := httptest.NewRecorder()
		DeleteStorefront(deleteRR, deleteReq)
		// Allow 200 or 204 for successful deletion
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, deleteRR.Code, "Flow Delete: Failed status")

		// Verify deletion in DB
		var deletedLink StorefrontLink
		err := testDB.First(&deletedLink, linkID).Error
		require.Error(t, err, "Expected error when fetching deleted link")
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Expected Gorm ErrRecordNotFound")
	})

	// 5. Get Storefronts again and verify it's empty
	t.Run("GetAfterDelete", func(t *testing.T) {
		getReq := createAuthenticatedRequest(t, user, "GET", "/api/get_storefronts", nil)
		getRR := httptest.NewRecorder()
		GetStorefronts(getRR, getReq)
		require.Equal(t, http.StatusOK, getRR.Code, "Flow Get After Delete: Failed status")

		var links []StorefrontLinkReturn
		require.NoError(t, json.Unmarshal(getRR.Body.Bytes(), &links), "Flow Get After Delete: Failed to parse list")
		assert.Len(t, links, 0, "Expected 0 links after deleting")
	})
}

// TestSpecificErrors tests various error conditions like forbidden, not found.
func TestSpecificErrors(t *testing.T) {
	setupTestEnvironment(t)
	user1 := createTestUser(t, "user1_errors@example.com", "password")
	user2 := createTestUser(t, "user2_errors@example.com", "password")

	// Add a link belonging to User 1
	addPayload := StorefrontLinkAddPayload{StoreType: "errors_test", StoreName: "Link For User 1", ApiKey: "key"}
	addBodyBytes, _ := json.Marshal(addPayload)
	addReq := createAuthenticatedRequest(t, user1, "POST", "/api/add_storefront", bytes.NewReader(addBodyBytes))
	addReq.Header.Set("Content-Type", "application/json")
	addRR := httptest.NewRecorder()
	AddStorefront(addRR, addReq)
	require.Equal(t, http.StatusCreated, addRR.Code)
	var addedResp StorefrontLinkReturn
	require.NoError(t, json.Unmarshal(addRR.Body.Bytes(), &addedResp))
	linkIDUser1 := addedResp.ID

	t.Run("UpdateForbidden", func(t *testing.T) {
		updatePayload := StorefrontLinkUpdatePayload{StoreName: "Attempt Update by User 2"}
		bodyBytes, _ := json.Marshal(updatePayload)
		url := fmt.Sprintf("/api/update_storefront?id=%d", linkIDUser1)                     // Target User 1's link
		req := createAuthenticatedRequest(t, user2, "PUT", url, bytes.NewReader(bodyBytes)) // Logged in as User 2
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		UpdateStorefront(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")
	})

	t.Run("DeleteForbidden", func(t *testing.T) {
		url := fmt.Sprintf("/api/delete_storefront?id=%d", linkIDUser1) // Target User 1's link
		req := createAuthenticatedRequest(t, user2, "DELETE", url, nil) // Logged in as User 2
		rr := httptest.NewRecorder()
		DeleteStorefront(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")

		// Verify User 1's link still exists
		var link StorefrontLink
		err := testDB.First(&link, linkIDUser1).Error
		assert.NoError(t, err, "Link owned by User 1 should not have been deleted")
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		nonExistentID := uint(999999)
		updatePayload := StorefrontLinkUpdatePayload{StoreName: "Update Non Existent"}
		bodyBytes, _ := json.Marshal(updatePayload)
		url := fmt.Sprintf("/api/update_storefront?id=%d", nonExistentID)
		req := createAuthenticatedRequest(t, user1, "PUT", url, bytes.NewReader(bodyBytes)) // Logged in as User 1
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		UpdateStorefront(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "not found")
	})

	t.Run("DeleteNotFound", func(t *testing.T) {
		nonExistentID := uint(999998)
		url := fmt.Sprintf("/api/delete_storefront?id=%d", nonExistentID)
		req := createAuthenticatedRequest(t, user1, "DELETE", url, nil) // Logged in as User 1
		rr := httptest.NewRecorder()
		DeleteStorefront(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "not found")
	})

	t.Run("UpdateMissingID", func(t *testing.T) {
		updatePayload := StorefrontLinkUpdatePayload{StoreName: "Update Missing ID"}
		bodyBytes, _ := json.Marshal(updatePayload)
		url := "/api/update_storefront" // Missing ?id=...
		req := createAuthenticatedRequest(t, user1, "PUT", url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		UpdateStorefront(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing required query parameter: id")
	})

	t.Run("DeleteMissingID", func(t *testing.T) {
		url := "/api/delete_storefront" // Missing ?id=...
		req := createAuthenticatedRequest(t, user1, "DELETE", url, nil)
		rr := httptest.NewRecorder()
		DeleteStorefront(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing required query parameter: id")
	})
}
