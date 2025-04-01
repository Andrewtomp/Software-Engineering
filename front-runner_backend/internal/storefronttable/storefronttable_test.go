// front-runner/internal/storefronttable/storefronttable_test.go
package storefronttable

import (
	"bytes"
	// "context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"     // <-- Import filepath again
	"regexp" // <-- Import regexp again
	"strings"
	"testing"
	"time"

	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/usertable"

	"github.com/joho/godotenv" // <-- Import godotenv again
	asrt "github.com/stretchr/testify/assert"
	req "github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const projectDirName = "front-runner_backend"

func init() {
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))

	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		log.Fatalf("Problem loading .env file. cwd:%s; cause: %s", cwd, err)
	}
	coredbutils.LoadEnv()
	usertable.Setup()
	login.Setup()
	Setup()
}

// --- Constants ---
// Adjust this if your project root identifier is different
// const projectDirName = "front-runner_backend" // Or "front-runner_backend"

// --- Test Utility Functions ---
// (registerAndLoginTestUser function remains the same)
// ... registerAndLoginTestUser definition ...
func registerAndLoginTestUser(t *testing.T, email, password, businessName string) (uint, *http.Cookie, error) {
	// 1. Register User
	regForm := url.Values{}
	regForm.Add("email", email)
	regForm.Add("password", password)
	regForm.Add("business_name", businessName)

	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regForm.Encode()))
	regReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	regRec := httptest.NewRecorder()

	usertable.RegisterUser(regRec, regReq) // Call actual handler

	if regRec.Code != http.StatusOK {
		return 0, nil, fmt.Errorf("failed to register test user '%s', status %d, body: %s", email, regRec.Code, regRec.Body.String())
	}
	t.Logf("Successfully registered test user: %s", email)

	// Retrieve UserID after registration (essential for linking storefronts)
	var registeredUser usertable.User
	// Use the main 'db' connection initialized in TestMain
	err := db.Where("email = ?", email).First(&registeredUser).Error
	if err != nil {
		return 0, nil, fmt.Errorf("failed to retrieve registered user '%s' from DB: %w", email, err)
	}
	if registeredUser.ID == 0 {
		return 0, nil, fmt.Errorf("retrieved user '%s' has ID 0", email)
	}
	userID := registeredUser.ID
	t.Logf("Retrieved UserID %d for %s", userID, email)

	// 2. Login User
	loginForm := url.Values{}
	loginForm.Add("email", email)
	loginForm.Add("password", password) // Use the same password

	loginReq := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()

	login.LoginUser(loginRec, loginReq) // Call actual handler

	// Login might redirect (303) or return OK (200) depending on implementation
	if loginRec.Code != http.StatusSeeOther && loginRec.Code != http.StatusOK {
		return 0, nil, fmt.Errorf("failed to log in test user '%s', expected 303 or 200, got status %d, body: %s", email, loginRec.Code, loginRec.Body.String())
	}

	// 3. Extract Session Cookie
	cookies := loginRec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		// Adjust cookie name if different
		if c.Name == "auth" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		return 0, nil, fmt.Errorf("session cookie 'auth' not found after login for user '%s'", email)
	}
	t.Logf("Successfully logged in test user %s and obtained session cookie", email)

	return userID, sessionCookie, nil
}

func makeTestUser(t *testing.T, password string) (uint, *http.Cookie, error) {
	require := req.New(t)
	uniqueEmail := fmt.Sprintf("addsf_%d@example.com", time.Now().UnixNano())
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("error hashing test password: %v", err)
	}
	userID, sessionCookie, err := registerAndLoginTestUser(t, uniqueEmail, string(hashedPassword), "Add SF Test Biz")
	require.NoError(err, "Setup failed: Could not register and login test user")
	require.NotNil(sessionCookie, "Setup failed: Session cookie is nil")
	require.NotZero(userID, "Setup failed: UserID is zero")

	return userID, sessionCookie, nil
}

// --- TestMain for Simplified Setup/Teardown (No Mocking Version) ---

func TestMain(m *testing.M) {
	log.Println("--- TestMain Start (No Mocking) ---")

	// --- Verify DB Connection and Key Loading ---
	if db == nil {
		log.Fatal("FATAL: Database connection (db) is nil after Setup(). Check DB config in .env and DB service.")
	}
	if len(encryptionKey) == 0 {
		log.Fatal("FATAL: Encryption key was not loaded after Setup(). Ensure 'STOREFRONT_KEY' ennviroment vairiable is set.")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("FATAL: Failed to get underlying SQL DB interface: %v", err)
	}
	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("FATAL: Failed to ping local database after Setup(): %v. Check connection details in .env.", err)
	}
	log.Println("TEST: Successfully connected to database and loaded encryption key.")

	// --- Run Migrations ---
	log.Println("TEST: Running database migrations...")
	usertable.MigrateUserDB()
	MigrateStorefrontDB()
	log.Println("TEST: Database migrations complete.")

	// --- Clear Tables Before Running Tests ---
	log.Println("TEST: Clearing tables before running tests...")
	// Use require inside TestMain - don't pass 't' (nil here)
	if err := usertable.ClearUserTable(db); err != nil {
		log.Fatalf("FATAL: Failed to clear users table before tests: %v", err)
	}
	if err := ClearStorefrontTable(db); err != nil {
		log.Fatalf("FATAL: Failed to clear storefront table before tests: %v", err)
	}

	// --- Run Tests ---
	log.Println("TEST: Starting test execution...")
	exitCode := m.Run()
	log.Println("TEST: Finished test execution.")

	// --- Clear Tables After Tests ---
	log.Println("TEST: Clearing tables after running tests...")
	if err := usertable.ClearUserTable(db); err != nil {
		log.Printf("ERROR: Failed to clear users table after tests: %v", err)
	}
	if err := ClearStorefrontTable(db); err != nil {
		log.Printf("ERROR: Failed to clear storefront table after tests: %v", err)
	}

	log.Println("--- TestMain End (No Mocking) ---")
	os.Exit(exitCode)
}

// --- Test Functions ---
// (Test functions remain the same as the previous "No Mocking" version)
// ... TestAddStorefront_RealLogin definition ...
func TestAddStorefront_RealLogin(t *testing.T) {
	// assert := asrt.New(t)
	// require := req.New(t)
	// TestMain clears tables

	// --- Setup: Register and Login User ---
	password := "testpassword123"
	userID, sessionCookie, _ := makeTestUser(t, password)

	// --- Test Case 1: Success ---
	t.Run("Success", func(t *testing.T) {
		assert := asrt.New(t) // Use subtest 't'
		require := req.New(t) // Use subtest 't'

		payload := StorefrontLinkAddPayload{
			StoreType: "amazon_real_login",
			StoreName: "My Amazon Store (Real Login)",
			ApiKey:    "real_key_456",
			ApiSecret: "real_secret_abc",
			StoreId:   "real_store789",
			StoreUrl:  "http://amazon.com/real_store789",
		}
		body, _ := json.Marshal(payload)

		// Create request and ADD THE COOKIE
		req := httptest.NewRequest(http.MethodPost, "/api/add_storefront", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie) // Add the obtained session cookie

		rr := httptest.NewRecorder()
		AddStorefront(rr, req) // Call the handler

		assert.Equal(http.StatusCreated, rr.Code, "Expected status 201 Created")

		// ... (rest of assertions for success case, verifying response and DB) ...
		var responseData StorefrontLinkReturn
		err := json.Unmarshal(rr.Body.Bytes(), &responseData)
		require.NoError(err, "Failed to unmarshal response") // No 't' needed
		assert.Equal(payload.StoreType, responseData.StoreType)
		assert.Equal(payload.StoreName, responseData.StoreName)
		assert.NotEmpty(responseData.ID)

		// Verify in DB
		var savedLink StorefrontLink
		dbResult := db.First(&savedLink, responseData.ID)
		require.NoError(dbResult.Error)        // No 't' needed
		assert.Equal(userID, savedLink.UserID) // IMPORTANT: Check against the real UserID
		assert.Equal(payload.StoreType, savedLink.StoreType)
		// ... (verify credentials via decryption as before)
		decryptedCreds, decryptErr := decryptCredentials(savedLink.Credentials)
		require.NoError(decryptErr) // No 't' needed
		expectedCredsMap := map[string]string{"apiKey": payload.ApiKey, "apiSecret": payload.ApiSecret}
		expectedCredsJSON, _ := json.Marshal(expectedCredsMap)
		assert.JSONEq(string(expectedCredsJSON), decryptedCreds)
	})

	// --- Test Case 2: Unauthorized (No Cookie) ---
	t.Run("Unauthorized", func(t *testing.T) {
		assert := asrt.New(t) // Use subtest 't'

		payload := StorefrontLinkAddPayload{StoreType: "test_unauth_real"}
		body, _ := json.Marshal(payload)

		// Create request WITHOUT the cookie
		req := httptest.NewRequest(http.MethodPost, "/api/add_storefront", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// NO req.AddCookie(sessionCookie)

		rr := httptest.NewRecorder()
		AddStorefront(rr, req)

		assert.Equal(http.StatusUnauthorized, rr.Code)
		// You might want to check the specific error message from your checkAuth helper
		assert.Contains(rr.Body.String(), "Unauthorized")
	})
}

// ... TestGetUpdateDeleteFlow_RealLogin definition ...
func TestGetUpdateDeleteFlow_RealLogin(t *testing.T) {
	assert := asrt.New(t)
	require := req.New(t)
	// TestMain clears tables

	// --- Setup: Register and Login User ---
	// uniqueEmail := fmt.Sprintf("flow_%d@example.com", time.Now().UnixNano())
	password := "flowpassword"
	_, sessionCookie, err := makeTestUser(t, password)
	// userID, sessionCookie, err := registerAndLoginTestUser(t, uniqueEmail, password, "Flow Test Biz")
	// require.NoError(err, "Setup failed: Could not register and login test user")
	// require.NotNil(sessionCookie, "Setup failed: Session cookie is nil")
	// require.NotZero(userID, "Setup failed: UserID is zero")

	// 1. Add a link to work with
	addPayload := StorefrontLinkAddPayload{StoreType: "demo_flow_real", StoreName: "Real Flow Store", ApiKey: "flow_key_real", StoreId: "flow_real", StoreUrl: "http://example.com/flow_real"}
	addBody, _ := json.Marshal(addPayload)
	addReq := httptest.NewRequest(http.MethodPost, "/api/add_storefront", bytes.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.AddCookie(sessionCookie) // Authenticate
	addRR := httptest.NewRecorder()
	AddStorefront(addRR, addReq)
	require.Equal(http.StatusCreated, addRR.Code, "Flow Setup: Failed to add initial link")
	var addedLinkResp StorefrontLinkReturn
	require.NoError(json.Unmarshal(addRR.Body.Bytes(), &addedLinkResp), "Flow Setup: Failed to parse add response") // No 't'
	linkID := addedLinkResp.ID

	// 2. Get Storefronts
	getReq := httptest.NewRequest(http.MethodGet, "/api/get_storefronts", nil)
	getReq.AddCookie(sessionCookie) // Authenticate
	getRR := httptest.NewRecorder()
	GetStorefronts(getRR, getReq)
	require.Equal(http.StatusOK, getRR.Code, "Flow Get: Failed status")
	var links []StorefrontLinkReturn
	require.NoError(json.Unmarshal(getRR.Body.Bytes(), &links), "Flow Get: Failed to parse list") // No 't'
	// ... (Verify link presence as before) ...
	found := false
	for _, link := range links {
		if link.ID == linkID {
			assert.Equal(addPayload.StoreName, link.StoreName)
			found = true
			break
		}
	}
	assert.True(found, "Flow Get: Newly added link not found")

	// 3. Update the link
	updatePayload := StorefrontLinkUpdatePayload{StoreName: "Real Flow Store UPDATED", StoreId: "flow_real_upd", StoreUrl: "http://example.com/flow_real_upd"}
	updateBody, _ := json.Marshal(updatePayload)
	updateURL := fmt.Sprintf("/api/update_storefront?id=%d", linkID)
	updateReq := httptest.NewRequest(http.MethodPut, updateURL, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(sessionCookie) // Authenticate
	updateRR := httptest.NewRecorder()
	UpdateStorefront(updateRR, updateReq)
	require.Equal(http.StatusOK, updateRR.Code, "Flow Update: Failed status")
	var updatedLinkResp StorefrontLinkReturn                                                                          // Need this if checking response
	require.NoError(json.Unmarshal(updateRR.Body.Bytes(), &updatedLinkResp), "Flow Update: Failed to parse response") // No 't'

	// Verify update in DB
	var updatedLinkDB StorefrontLink
	dbResult := db.First(&updatedLinkDB, linkID)
	require.NoError(dbResult.Error) // No 't'
	assert.Equal(updatePayload.StoreName, updatedLinkDB.StoreName)
	// ... (Check credentials didn't change)
	decryptedCreds, decryptErr := decryptCredentials(updatedLinkDB.Credentials)
	require.NoError(decryptErr)                                    // No 't'
	origCredsMap := map[string]string{"apiKey": addPayload.ApiKey} // Reconstruct original creds map
	origCredsJSON, _ := json.Marshal(origCredsMap)
	assert.JSONEq(string(origCredsJSON), decryptedCreds)

	// 4. Delete the link
	deleteURL := fmt.Sprintf("/api/delete_storefront?id=%d", linkID)
	deleteReq := httptest.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteReq.AddCookie(sessionCookie) // Authenticate
	deleteRR := httptest.NewRecorder()
	DeleteStorefront(deleteRR, deleteReq)
	assert.Contains([]int{http.StatusOK, http.StatusNoContent}, deleteRR.Code, "Flow Delete: Failed status")
	// ... (Verify deletion in DB as before) ...
	var deletedLink StorefrontLink
	err = db.First(&deletedLink, linkID).Error
	assert.Error(err) // No 't' needed
	assert.True(errors.Is(err, gorm.ErrRecordNotFound))
}

// ... TestSpecificErrors_RealLogin definition ...
func TestSpecificErrors_RealLogin(t *testing.T) {
	// assert := asrt.New(t)
	require := req.New(t)

	// Define Helper struct at the beginning of the function scope
	type CheckLink struct {
		ID uint `gorm:"primaryKey"`
	}

	// --- Setup users ---
	password := "password123"
	_, sessionCookie1, _ := makeTestUser(t, password)
	_, sessionCookie2, _ := makeTestUser(t, password)

	// Add a link belonging to User 1
	addPayload := StorefrontLinkAddPayload{StoreType: "errors_real", StoreName: "Link For User 1", ApiKey: "key"}
	addBody, _ := json.Marshal(addPayload)
	addReq := httptest.NewRequest(http.MethodPost, "/api/add_storefront", bytes.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.AddCookie(sessionCookie1) // Use User 1's cookie
	addRR := httptest.NewRecorder()
	AddStorefront(addRR, addReq)
	require.Equal(http.StatusCreated, addRR.Code)
	var addedResp StorefrontLinkReturn
	require.NoError(json.Unmarshal(addRR.Body.Bytes(), &addedResp))
	linkIDUser1 := addedResp.ID // ID of the link owned by User 1

	t.Run("UpdateForbidden", func(t *testing.T) {
		assert := asrt.New(t) // Use subtest 't'

		updatePayload := StorefrontLinkUpdatePayload{StoreName: "Attempt Update by User 2"}
		body, _ := json.Marshal(updatePayload)
		url := fmt.Sprintf("/api/update_storefront?id=%d", linkIDUser1) // Target User 1's link
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie2) // Logged in as User 2
		rr := httptest.NewRecorder()
		UpdateStorefront(rr, req)
		assert.Equal(http.StatusForbidden, rr.Code)
	})

	t.Run("DeleteForbidden", func(t *testing.T) {
		assert := asrt.New(t) // Use subtest 't'

		url := fmt.Sprintf("/api/delete_storefront?id=%d", linkIDUser1) // Target User 1's link
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		req.AddCookie(sessionCookie2) // Logged in as User 2
		rr := httptest.NewRecorder()
		DeleteStorefront(rr, req)
		assert.Equal(http.StatusForbidden, rr.Code)

		// Verify User 1's link still exists
		var link CheckLink
		err := db.Model(&StorefrontLink{}).Select("id").First(&link, linkIDUser1).Error
		assert.NoError(err, "Link owned by User 1 should not have been deleted") // No 't' needed
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		assert := asrt.New(t) // Use subtest 't'

		nonExistentID := uint(999999)
		updatePayload := StorefrontLinkUpdatePayload{StoreName: "Update Non Existent"}
		body, _ := json.Marshal(updatePayload)
		url := fmt.Sprintf("/api/update_storefront?id=%d", nonExistentID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie1) // Logged in as User 1 (doesn't matter for not found)
		rr := httptest.NewRecorder()
		UpdateStorefront(rr, req)
		assert.Equal(http.StatusNotFound, rr.Code)
	})
}
