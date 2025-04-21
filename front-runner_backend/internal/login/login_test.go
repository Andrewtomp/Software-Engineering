// internal/login/login_test.go
package login

import (
	"front-runner/internal/coredbutils"
	"front-runner/internal/usertable"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	// "github.com/gorilla/mux" // Not strictly needed for direct handler tests
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
// It also clears the user table before each run via this function.
func setupTestEnvironment(t *testing.T) {
	// Use t.Helper() to mark this as a test helper function
	t.Helper()

	setupEnvOnce.Do(func() {
		// Find project root based on a known directory name
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

		// Initialize DB connection (using error handling version)
		coredbutils.ResetDBStateForTests() // Reset state before loading/getting
		err = coredbutils.LoadEnv()
		if err != nil {
			t.Fatalf("Failed to load core DB environment: %v", err)
		}
		var dbErr error
		testDB, dbErr = coredbutils.GetDB()
		if dbErr != nil {
			t.Fatalf("Failed to get DB connection for tests: %v", dbErr)
		}

		// Initialize Session Store for tests (use dummy keys for testing)
		// IMPORTANT: Use different keys than production!
		authKey := []byte("test-auth-key-32-bytes-long-000")
		// --- MODIFIED: Ensure encKey is 32 bytes ---
		encKey := []byte("test-enc-key-needs-to-be-32-byte") // This string is 32 bytes long
		// --- End Modification ---

		if len(encKey) != 16 && len(encKey) != 32 {
			// This check should now pass
			t.Fatal("Test encryption key must be 16 or 32 bytes")
		}
		testSessionStore = sessions.NewCookieStore(authKey, encKey)
		testSessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 1, // 1 day for tests
			HttpOnly: true,
			Secure:   false, // Usually false for httptest unless configured otherwise
			SameSite: http.SameSiteLaxMode,
		}

		// Setup dependent packages
		usertable.Setup()               // Assumes it uses coredbutils.GetDB() internally
		Setup(testDB, testSessionStore) // Setup the login package with test DB and store

		// Ensure migrations are run (optional if TestMain handles it)
		// usertable.MigrateUserDB()
	})

	// Clear user table before each test function that calls this setup
	err := usertable.ClearUserTable(testDB)
	require.NoError(t, err, "Failed to clear user table before test")
}

// Helper to create a test user directly in the DB
func createTestUser(t *testing.T, email, password string) *usertable.User {
	t.Helper() // Mark as helper
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password for test user")

	user := &usertable.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Provider:     "local",
	}
	// Use the CreateUser function from the usertable package for consistency
	err = usertable.CreateUser(user)
	require.NoError(t, err, "Failed to create test user using usertable.CreateUser")
	// We need the ID, so fetch the user back (CreateUser doesn't return the full user with ID)
	createdUser, err := usertable.GetUserByEmail(email)
	require.NoError(t, err, "Failed to fetch created test user by email")
	require.NotNil(t, createdUser, "Fetched created test user should not be nil")
	return createdUser
}

// TestLoginUser tests successful login.
func TestLoginUser(t *testing.T) {
	setupTestEnvironment(t)
	userEmail := "logintest@example.com"
	userPassword := "password123"
	_ = createTestUser(t, userEmail, userPassword) // Create user directly

	// Create request body
	form := url.Values{}
	form.Add("email", userEmail)
	form.Add("password", userPassword)

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// Call the handler
	LoginUser(rr, req)

	// Assertions
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Expected status code 303 See Other")

	// --- MODIFIED: Check Location header directly ---
	locationHeader := rr.Header().Get("Location")
	require.NotEmpty(t, locationHeader, "Expected Location header to be set")
	assert.Equal(t, "/", locationHeader, "Expected redirect location to be '/'")
	// --- End Modification ---

	// Check if the session cookie was set
	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == sessionName { // Use the constant
			foundCookie = true
			assert.NotEmpty(t, cookie.Value, "Session cookie should not be empty")

			// Optional: Decode cookie to verify content (more involved)
			// session, err := testSessionStore.Get(&http.Request{Header: http.Header{"Cookie": {cookie.String()}}}, sessionName)
			// require.NoError(t, err, "Failed to decode session cookie")
			// assert.Equal(t, createdUser.ID, session.Values[userSessionKey], "User ID in session mismatch")

			break
		}
	}
	assert.True(t, foundCookie, "Expected session cookie '%s' to be set", sessionName)
}

// TestLoginUserInvalid tests login with incorrect password.
func TestLoginUserInvalid(t *testing.T) {
	setupTestEnvironment(t)
	userEmail := "invalidlogin@example.com"
	userPassword := "password123"
	_ = createTestUser(t, userEmail, userPassword) // Create user

	// Create request body with wrong password
	form := url.Values{}
	form.Add("email", userEmail)
	form.Add("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	LoginUser(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status code 401 Unauthorized")
	assert.Contains(t, rr.Body.String(), "Invalid credentials", "Expected error message")
}

// TestLoginUserNotFound tests login with non-existent email.
func TestLoginUserNotFound(t *testing.T) {
	setupTestEnvironment(t)

	form := url.Values{}
	form.Add("email", "nosuchuser@example.com")
	form.Add("password", "password123")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	LoginUser(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status code 401 Unauthorized")
	assert.Contains(t, rr.Body.String(), "Invalid credentials", "Expected error message")
}

// TestLoginUserMissingFields tests login with missing email or password.
func TestLoginUserMissingFields(t *testing.T) {
	setupTestEnvironment(t)

	testCases := []struct {
		name string
		form url.Values
	}{
		{
			name: "Missing Password",
			form: url.Values{"email": {"test@example.com"}},
		},
		{
			name: "Missing Email",
			form: url.Values{"password": {"password123"}},
		},
		{
			name: "Missing Both",
			form: url.Values{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/login", strings.NewReader(tc.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			LoginUser(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400 Bad Request")
			assert.Contains(t, rr.Body.String(), "Email and password are required", "Expected error message")
		})
	}
}

// TestLoginUserAlreadyLoggedIn tests attempting to log in when already logged in.
func TestLoginUserAlreadyLoggedIn(t *testing.T) {
	setupTestEnvironment(t)
	userEmail := "alreadyin@example.com"
	userPassword := "password123"
	user := createTestUser(t, userEmail, userPassword)

	// Simulate an existing valid session
	req := httptest.NewRequest("POST", "/api/login", nil) // Body doesn't matter here
	session, err := testSessionStore.New(req, sessionName)
	require.NoError(t, err, "Failed to create new session for test setup")
	session.Values[userSessionKey] = user.ID // Set the user ID in the session

	// Manually add the session cookie to the request
	rrCookieSetter := httptest.NewRecorder() // Use a temporary recorder to get the cookie header
	err = testSessionStore.Save(req, rrCookieSetter, session)
	require.NoError(t, err, "Failed to save session to get cookie header")
	cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookieHeader, "Set-Cookie header should not be empty after saving session")
	req.Header.Set("Cookie", cookieHeader) // Set the cookie on the actual request

	// Now make the login request with the existing session cookie
	rrLogin := httptest.NewRecorder()
	LoginUser(rrLogin, req) // Call the handler

	// Assertions
	assert.Equal(t, http.StatusConflict, rrLogin.Code, "Expected status code 409 Conflict")
	assert.Contains(t, rrLogin.Body.String(), "User is already logged in", "Expected error message")
}

// TestLogoutUser tests successful logout when logged in.
func TestLogoutUser(t *testing.T) {
	setupTestEnvironment(t)
	userEmail := "logouttest@example.com"
	userPassword := "password123"
	user := createTestUser(t, userEmail, userPassword)

	// Simulate a logged-in state by creating a request with a valid session cookie
	req := httptest.NewRequest("GET", "/logout", nil)
	session, err := testSessionStore.New(req, sessionName)
	require.NoError(t, err, "Failed to create new session for test setup")
	session.Values[userSessionKey] = user.ID

	rrCookieSetter := httptest.NewRecorder()
	err = testSessionStore.Save(req, rrCookieSetter, session)
	require.NoError(t, err, "Failed to save session to get cookie header")
	cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookieHeader, "Set-Cookie header should not be empty after saving session")
	req.Header.Set("Cookie", cookieHeader) // Add the valid cookie

	// Perform the logout request
	rr := httptest.NewRecorder()
	LogoutUser(rr, req)

	// Assertions
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Expected status code 303 See Other")

	// Check Location header directly
	locationHeader := rr.Header().Get("Location")
	require.NotEmpty(t, locationHeader, "Expected Location header to be set")
	assert.Equal(t, "/", locationHeader, "Expected redirect location to be '/'")

	// Check that the session cookie was cleared (MaxAge=0 or -1, or Expires in the past)
	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == sessionName {
			foundCookie = true
			// MaxAge check is generally more reliable than Expires across systems/clocks
			assert.True(t, cookie.MaxAge <= 0, "Expected session cookie MaxAge to be <= 0, got %d", cookie.MaxAge)
			// Check Expires just in case MaxAge isn't set correctly by the library
			if cookie.MaxAge == 0 { // Only check Expires if MaxAge isn't explicitly negative
				assert.True(t, !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now().Add(time.Second*5)), "Expected session cookie Expires to be in the past or very close to it")
			}
			break
		}
	}
	assert.True(t, foundCookie, "Expected session cookie '%s' to be set for clearing", sessionName)
}

// TestLogoutUserNotLoggedIn tests logout when not logged in.
func TestLogoutUserNotLoggedIn(t *testing.T) {
	setupTestEnvironment(t)

	// Perform the logout request without any session cookie
	req := httptest.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()
	LogoutUser(rr, req)

	// Assertions - Should still redirect and attempt to clear cookie
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Expected status code 303 See Other")

	// Check Location header directly
	locationHeader := rr.Header().Get("Location")
	require.NotEmpty(t, locationHeader, "Expected Location header to be set")
	assert.Equal(t, "/", locationHeader, "Expected redirect location to be '/'")

	// Check that a clearing cookie was attempted
	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == sessionName {
			foundCookie = true
			assert.True(t, cookie.MaxAge <= 0, "Expected session cookie MaxAge to be <= 0, got %d", cookie.MaxAge)
			if cookie.MaxAge == 0 {
				assert.True(t, !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now().Add(time.Second*5)), "Expected session cookie Expires to be in the past or very close to it")
			}
			break
		}
	}
	assert.True(t, foundCookie, "Expected session cookie '%s' to be set for clearing even when not logged in", sessionName)
}
