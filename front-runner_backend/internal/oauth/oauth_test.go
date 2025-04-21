package oauth

import (
	"errors"
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/usertable" // Import usertable
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const projectDirName = "front-runner_backend" // Adjust if your project dir name is different

// Global vars for testing
var testDB *gorm.DB
var testStore *sessions.CookieStore
var testSessionAuthKey = "test-auth-key-32-bytes-long-000" // Must be 32 or 64 bytes for AES-128/256
var testSessionEncKey = "test-enc-key-needs-to-be-32-byte" // Must be 16 or 32 bytes for AES-128/256

// init loads environment variables and sets up the database and session store.
func init() {
	// --- Environment Variable Loading ---
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	rootPath := re.FindString(cwd)
	if rootPath == "" {
		log.Fatalf("Could not find project root directory '%s' from cwd '%s'", projectDirName, cwd)
	}

	envPath := rootPath + `/.env`
	err = godotenv.Load(envPath)
	if err != nil {
		log.Printf("Warning: Could not load .env file from %s: %v. Assuming env vars are set.", envPath, err)
	}

	// --- Set Dummy Env Vars for OAuth Testing ---
	// These are needed by oauth.Setup()
	os.Setenv("GOOGLE_CLIENT_ID", "test_client_id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test_client_secret")
	os.Setenv("GOOGLE_CALLBACK_URL", "http://localhost:8080/auth/google/callback") // Must match setup
	// os.Setenv("SESSION_AUTH_KEY", testSessionAuthKey) // oauth.Setup doesn't use these directly anymore
	// os.Setenv("SESSION_ENC_KEY", testSessionEncKey)

	// --- Database Setup (using usertable's setup) ---
	usertable.Setup()               // Initialize usertable DB connection
	testDB, _ = coredbutils.GetDB() // Get the DB instance used by usertable
	if testDB == nil {
		log.Fatal("Database connection failed in init")
	}
	usertable.MigrateUserDB() // Ensure user table exists

	// --- Session Store Setup ---
	testStore = sessions.NewCookieStore([]byte(testSessionAuthKey), []byte(testSessionEncKey))
	testStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		// Secure: true, // Set Secure=false for httptest unless it's configured for HTTPS
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	gothic.Store = testStore // IMPORTANT: Tell gothic to use our test store

	// --- OAuth Setup ---
	Setup(testStore) // Call the package's Setup function AFTER setting env vars and store
}

// TestMain manages the test database state.
func TestMain(m *testing.M) {
	if testDB == nil {
		fmt.Println("Test database connection is nil, cannot proceed.")
		os.Exit(1)
	}
	if testStore == nil {
		fmt.Println("Test session store is nil, cannot proceed.")
		os.Exit(1)
	}

	// Clear the user table before running tests.
	fmt.Println("Clearing test user table before tests...")
	if err := usertable.ClearUserTable(testDB); err != nil {
		fmt.Printf("Failed to clear test user table before tests: %v\n", err)
		os.Exit(1)
	}

	// Run the tests.
	code := m.Run()

	// Clear the user table after tests.
	fmt.Println("Clearing test user table after tests...")
	if err := usertable.ClearUserTable(testDB); err != nil {
		fmt.Printf("Failed to clear test user table after tests: %v\n", err)
		// Don't exit fatally here, just report the error
	}

	os.Exit(code)
}

// Helper to create a user directly for testing dependencies
func createTestUserDirectly(t *testing.T, email, name, provider, providerID string) *usertable.User {
	t.Helper()
	user := &usertable.User{
		Email:      email,
		Name:       name,
		Provider:   provider,
		ProviderID: providerID,
	}

	// *** FIX: Add hash generation for local provider ***
	if provider == "local" {
		// Use a dummy password for testing purposes
		dummyPassword := "testpassword"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dummyPassword), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash dummy password for %s: %v", email, err)
		}
		user.PasswordHash = string(hashedPassword)
	}
	// *** End Fix ***

	err := usertable.CreateUser(user) // Call CreateUser AFTER potentially setting hash
	if err != nil {
		t.Fatalf("Failed to create test user %s directly: %v", email, err)
	}
	if user.ID == 0 {
		t.Fatalf("Created test user %s directly but ID is zero", email)
	}
	return user
}

// Helper to create a request with a valid session cookie for a given user ID
func createRequestWithSession(t *testing.T, userID uint) *http.Request {
	t.Helper()
	req := httptest.NewRequest("GET", "/", nil)
	session, err := testStore.New(req, sessionName)
	if err != nil {
		t.Fatalf("Failed to create new session: %v", err)
	}
	session.Values[userSessionKey] = userID
	// Need a ResponseRecorder to actually save the cookie header
	rec := httptest.NewRecorder()
	err = session.Save(req, rec)
	if err != nil {
		t.Fatalf("Failed to save session to recorder: %v", err)
	}
	// Get the cookie from the recorder and add it to a new request
	cookieHeader := rec.Header().Get("Set-Cookie")
	if cookieHeader == "" {
		t.Fatal("Failed to get Set-Cookie header from recorder")
	}
	// Create a new request and add the cookie
	finalReq := httptest.NewRequest("GET", "/", nil)
	finalReq.Header.Set("Cookie", cookieHeader)
	return finalReq
}

// --- Test Cases ---

func TestHandleGoogleLogin(t *testing.T) {
	// This test is limited because BeginAuthHandler relies heavily on gothic's
	// internal state and redirection. We mainly check that it doesn't panic
	// and returns a redirect status, assuming gothic handles the actual redirect URL.
	testStore = sessions.NewCookieStore([]byte(testSessionAuthKey), []byte(testSessionEncKey))
	testStore.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7, // 7 days
	}
	gothic.Store = testStore
	Setup(testStore)
	// Secure
	t.Run("initiates_redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/google", nil)
		rec := httptest.NewRecorder()

		// We expect BeginAuthHandler to be called. Since we can't easily mock it,
		// we call our handler and check for a redirect status code.
		// The actual redirect target is determined by goth/gothic.
		HandleGoogleLogin(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d; got %d", http.StatusTemporaryRedirect, rec.Code)
		}
		// We cannot easily verify the exact Location header here without complex mocking.
		// log.Printf("Redirect Location: %s", rec.Header().Get("Location"))
	})
}

func TestHandleGoogleCallback(t *testing.T) {
	originalCompleteUserAuth := gothic.CompleteUserAuth

	t.Cleanup(func() {
		gothic.CompleteUserAuth = originalCompleteUserAuth
	})
	// --- Test Case: New User Registration ---
	t.Run("new_user_success", func(t *testing.T) {
		// Simulate gothUser returned by gothic.CompleteUserAuth
		mockGothUser := &goth.User{
			Provider:    "google",
			UserID:      "google_user_id_123",
			Email:       "new_google_user@example.com",
			Name:        "New Google User",
			AccessToken: "test-token",
		}

		gothic.CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {
			// Return the desired mock user and no error
			return *mockGothUser, nil
		}

		req := httptest.NewRequest("GET", "/auth/google/callback?state=teststate", nil) // State might be needed by gothic internally
		rec := httptest.NewRecorder()

		// // Inject the simulated gothUser into the request context
		// ctx := simulateCompleteUserAuth(req, mockGothUser, nil)
		// req = req.WithContext(ctx)

		HandleGoogleCallback(rec, req)

		// 1. Check for redirect
		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d; got %d. Body: %s", http.StatusTemporaryRedirect, rec.Code, rec.Body.String())
		}
		if loc := rec.Header().Get("Location"); loc != "/" {
			t.Errorf("Expected redirect location '/'; got %q", loc)
		}

		// 2. Verify user was created in DB
		dbUser, err := usertable.GetUserByProviderID("google", mockGothUser.UserID)
		if err != nil {
			t.Fatalf("Error fetching user from DB: %v", err)
		}
		if dbUser == nil {
			t.Fatalf("User with provider ID %s was not found in DB after callback", mockGothUser.UserID)
		}
		if dbUser.Email != mockGothUser.Email || dbUser.Name != mockGothUser.Name {
			t.Errorf("DB user data mismatch. Expected Email: %s, Name: %s. Got Email: %s, Name: %s",
				mockGothUser.Email, mockGothUser.Name, dbUser.Email, dbUser.Name)
		}

		// 3. Verify session cookie was set with the correct user ID
		cookieHeader := rec.Header().Get("Set-Cookie")
		if cookieHeader == "" {
			t.Fatal("Expected Set-Cookie header, but none found")
		}
		// Parse the cookie to verify content (simplified check)
		tempReq := httptest.NewRequest("GET", "/", nil) // Need a request to read the cookie
		tempReq.Header.Set("Cookie", cookieHeader)
		session, err := testStore.Get(tempReq, sessionName)
		if err != nil {
			t.Fatalf("Failed to get session from response cookie: %v", err)
		}
		sessionUserID, ok := session.Values[userSessionKey].(uint)
		if !ok || sessionUserID != dbUser.ID {
			t.Errorf("Expected session user ID %d; got %v (type %T)", dbUser.ID, session.Values[userSessionKey], session.Values[userSessionKey])
		}
	})

	// --- Test Case: Existing User Login ---
	t.Run("existing_user_success", func(t *testing.T) {
		// 1. Create the user in the DB first
		provider := "google"
		providerID := "existing_google_user_456"
		email := "existing@example.com"
		name := "Existing User"
		existingUser := createTestUserDirectly(t, email, name, provider, providerID)

		// 2. Simulate gothUser returned by gothic.CompleteUserAuth
		mockGothUser := &goth.User{
			Provider:    provider,
			UserID:      providerID,
			Email:       email,
			Name:        name, // Name might be updated by Google, test this if needed
			AccessToken: "test-token-existing",
		}

		gothic.CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {
			return *mockGothUser, nil
		}

		req := httptest.NewRequest("GET", "/auth/google/callback?state=teststate2", nil)
		rec := httptest.NewRecorder()

		// 3. Inject the simulated gothUser
		// ctx := simulateCompleteUserAuth(req, mockGothUser, nil)
		// req = req.WithContext(ctx)

		HandleGoogleCallback(rec, req)

		// 4. Check redirect
		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d; got %d. Body: %s", http.StatusTemporaryRedirect, rec.Code, rec.Body.String())
		}
		if loc := rec.Header().Get("Location"); loc != "/" {
			t.Errorf("Expected redirect location '/'; got %q", loc)
		}

		// 5. Verify session cookie was set with the existing user ID
		cookieHeader := rec.Header().Get("Set-Cookie")
		if cookieHeader == "" {
			t.Fatal("Expected Set-Cookie header, but none found")
		}
		tempReq := httptest.NewRequest("GET", "/", nil)
		tempReq.Header.Set("Cookie", cookieHeader)
		session, err := testStore.Get(tempReq, sessionName)
		if err != nil {
			t.Fatalf("Failed to get session from response cookie: %v", err)
		}
		sessionUserID, ok := session.Values[userSessionKey].(uint)
		if !ok || sessionUserID != existingUser.ID {
			t.Errorf("Expected session user ID %d; got %v (type %T)", existingUser.ID, session.Values[userSessionKey], session.Values[userSessionKey])
		}

		// 6. Optional: Verify user count hasn't increased
		var count int64
		testDB.Model(&usertable.User{}).Count(&count)
		// Adjust expected count based on users created in this test run
		// This is fragile, better to check specific user wasn't re-created
	})

	// --- Test Case: Error from CompleteUserAuth ---
	t.Run("complete_user_auth_error", func(t *testing.T) {
		authError := errors.New("simulated auth error from gothic")

		gothic.CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {
			// Return an empty user and the mock error
			return goth.User{}, authError
		}

		req := httptest.NewRequest("GET", "/auth/google/callback?state=teststate3", nil)
		rec := httptest.NewRecorder()

		// Inject the simulated error
		// ctx := simulateCompleteUserAuth(req, nil, authError)
		// req = req.WithContext(ctx)

		HandleGoogleCallback(rec, req)

		// Check for appropriate error response (not a redirect)
		// The handler currently writes the error message directly.
		if rec.Code != http.StatusOK { // Or potentially 500 depending on error type
			t.Errorf("Expected status %d; got %d", http.StatusOK, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), authError.Error()) {
			t.Errorf("Expected response body to contain %q, got %q", authError.Error(), rec.Body.String())
		}
	})

	// --- Test Case: Database Error on User Creation ---
	// This is harder to test reliably without DB mocking.
	// We can simulate it by making the email invalid *after* CompleteUserAuth
	// returns it, causing CreateUser to fail.
	t.Run("db_create_error", func(t *testing.T) {
		mockGothUser := &goth.User{
			Provider:    "google",
			UserID:      "google_user_id_db_error",
			Email:       "invalid-email-after-auth", // This will fail CreateUser validation
			Name:        "DB Error User",
			AccessToken: "test-token-db-error",
		}

		gothic.CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {
			return *mockGothUser, nil
		}

		req := httptest.NewRequest("GET", "/auth/google/callback?state=teststatedberror", nil)
		rec := httptest.NewRecorder()

		// ctx := simulateCompleteUserAuth(req, mockGothUser, nil)
		// req = req.WithContext(ctx)

		HandleGoogleCallback(rec, req)

		// Check for internal server error
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d; got %d. Body: %s", http.StatusInternalServerError, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "Failed to create user account") {
			t.Errorf("Expected error message 'Failed to create user account', got %q", rec.Body.String())
		}
	})
}

func TestHandleLogout(t *testing.T) {
	t.Run("logout_with_session", func(t *testing.T) {
		// 1. Create a user and a request with their session
		user := createTestUserDirectly(t, "logout@example.com", "Logout User", "local", "") // Provider doesn't matter here
		req := createRequestWithSession(t, user.ID)
		rec := httptest.NewRecorder()

		// 2. Call HandleLogout
		HandleLogout(rec, req)

		// 3. Check for redirect
		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d; got %d", http.StatusTemporaryRedirect, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/" {
			t.Errorf("Expected redirect location '/'; got %q", loc)
		}

		// 4. Check that the session cookie is expired
		cookieHeader := rec.Header().Get("Set-Cookie")
		if cookieHeader == "" {
			t.Fatal("Expected Set-Cookie header, but none found")
		}
		// Parse the cookie to verify Max-Age or Expires
		resp := http.Response{Header: http.Header{"Set-Cookie": {cookieHeader}}}
		cookies := resp.Cookies()
		foundCookie := false
		for _, cookie := range cookies {
			if cookie.Name == sessionName {
				foundCookie = true
				// Check if expired (MaxAge<0 or Expires in the past)
				if cookie.MaxAge >= 0 && !cookie.Expires.Before(time.Now()) {
					t.Errorf("Expected cookie MaxAge < 0 or Expires in the past, got MaxAge=%d, Expires=%s", cookie.MaxAge, cookie.Expires)
				}
				break
			}
		}
		if !foundCookie {
			t.Errorf("Expected cookie named %q in Set-Cookie header", sessionName)
		}
	})

	t.Run("logout_without_session", func(t *testing.T) {
		// 1. Create a request without a session
		req := httptest.NewRequest("GET", "/logout", nil)
		rec := httptest.NewRecorder()

		// 2. Call HandleLogout
		HandleLogout(rec, req)

		// 3. Check for redirect
		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d; got %d", http.StatusTemporaryRedirect, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/" {
			t.Errorf("Expected redirect location '/'; got %q", loc)
		}

		// 4. Check that it still tries to expire the cookie (best effort)
		cookieHeader := rec.Header().Get("Set-Cookie")
		if cookieHeader == "" {
			t.Fatal("Expected Set-Cookie header even without initial session, but none found")
		}
		resp := http.Response{Header: http.Header{"Set-Cookie": {cookieHeader}}}
		cookies := resp.Cookies()
		foundCookie := false
		for _, cookie := range cookies {
			if cookie.Name == sessionName {
				foundCookie = true
				if cookie.MaxAge >= 0 && !cookie.Expires.Before(time.Now()) {
					t.Errorf("Expected cookie MaxAge < 0 or Expires in the past, got MaxAge=%d, Expires=%s", cookie.MaxAge, cookie.Expires)
				}
				break
			}
		}
		if !foundCookie {
			t.Errorf("Expected cookie named %q in Set-Cookie header", sessionName)
		}
	})
}

func TestGetCurrentUser(t *testing.T) {
	t.Run("valid_session", func(t *testing.T) {
		// 1. Create user and request with session
		user := createTestUserDirectly(t, "current@example.com", "Current User", "local", "")
		req := createRequestWithSession(t, user.ID)

		// 2. Call GetCurrentUser
		currentUser, err := GetCurrentUser(req)

		// 3. Verify result
		if err != nil {
			t.Fatalf("GetCurrentUser failed: %v", err)
		}
		if currentUser == nil {
			t.Fatal("Expected current user, but got nil")
		}
		if currentUser.ID != user.ID || currentUser.Email != user.Email {
			t.Errorf("Expected user ID %d, Email %s; got ID %d, Email %s",
				user.ID, user.Email, currentUser.ID, currentUser.Email)
		}
	})

	t.Run("no_session", func(t *testing.T) {
		// 1. Create request without session
		req := httptest.NewRequest("GET", "/", nil)

		// 2. Call GetCurrentUser
		currentUser, err := GetCurrentUser(req)

		// 3. Verify result
		if err != nil {
			// Check if the error is specifically about securecookie validation if keys mismatch
			if !strings.Contains(err.Error(), "failed to get session") {
				t.Fatalf("GetCurrentUser failed with unexpected error: %v", err)
			}
			// If it's a securecookie error, it might mean the store keys are bad,
			// but for a request with *no* cookie, Get should ideally not error here.
			// Depending on gorilla/sessions behavior, this might sometimes error.
			log.Printf("Note: GetCurrentUser errored on no_session request: %v", err)
		}
		if currentUser != nil {
			t.Errorf("Expected nil user when no session exists, got user ID %d", currentUser.ID)
		}
	})

	t.Run("session_without_user_id_key", func(t *testing.T) {
		// 1. Create request and session, but don't add the user ID
		req := httptest.NewRequest("GET", "/", nil)
		session, _ := testStore.New(req, sessionName)
		rec := httptest.NewRecorder()
		session.Save(req, rec) // Save empty session
		cookieHeader := rec.Header().Get("Set-Cookie")
		finalReq := httptest.NewRequest("GET", "/", nil)
		finalReq.Header.Set("Cookie", cookieHeader)

		// 2. Call GetCurrentUser
		currentUser, err := GetCurrentUser(finalReq)

		// 3. Verify result
		if err != nil {
			t.Fatalf("GetCurrentUser failed: %v", err)
		}
		if currentUser != nil {
			t.Errorf("Expected nil user when session lacks user ID key, got user ID %d", currentUser.ID)
		}
	})

	t.Run("session_with_invalid_user_id_type", func(t *testing.T) {
		// 1. Create request and session, add wrong type for user ID
		req := httptest.NewRequest("GET", "/", nil)
		session, _ := testStore.New(req, sessionName)
		session.Values[userSessionKey] = "not-a-uint" // Store a string
		rec := httptest.NewRecorder()
		session.Save(req, rec)
		cookieHeader := rec.Header().Get("Set-Cookie")
		finalReq := httptest.NewRequest("GET", "/", nil)
		finalReq.Header.Set("Cookie", cookieHeader)

		// 2. Call GetCurrentUser
		currentUser, err := GetCurrentUser(finalReq)

		// 3. Verify result
		if err == nil {
			t.Fatal("Expected an error due to invalid user ID type, but got nil")
		}
		if !strings.Contains(err.Error(), "invalid user ID type in session") {
			t.Errorf("Expected error message containing 'invalid user ID type in session', got: %v", err)
		}
		if currentUser != nil {
			t.Errorf("Expected nil user when session has invalid user ID type, got user ID %d", currentUser.ID)
		}
	})

	t.Run("session_with_valid_id_but_user_deleted", func(t *testing.T) {
		// 1. Create user, get ID, then delete user
		user := createTestUserDirectly(t, "deleted@example.com", "Deleted User", "local", "")
		deletedUserID := user.ID
		err := testDB.Delete(&usertable.User{}, deletedUserID).Error
		if err != nil {
			t.Fatalf("Failed to delete user for test: %v", err)
		}

		// 2. Create request with session for the deleted user ID
		req := createRequestWithSession(t, deletedUserID)

		// 3. Call GetCurrentUser
		currentUser, err := GetCurrentUser(req)

		// 4. Verify result (should return nil user, potentially an error or nil error)
		// The current implementation of GetUserByID returns nil, nil if not found.
		// GetCurrentUser wraps the GetUserByID error, so we expect an error here.
		if err != nil {
			t.Fatalf("Expected nil error when user ID %d not found in DB, but got: %v", deletedUserID, err)
		}
		if currentUser != nil {
			t.Errorf("Expected nil user when user ID %d not found in DB, got user ID %d", deletedUserID, currentUser.ID)
		}
	})
}
