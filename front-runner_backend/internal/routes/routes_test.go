// internal/routes/routes_test.go
package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	// Need these for setup even if not directly used in every test
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/oauth"
	"front-runner/internal/usertable"
)

const projectDirName = "front-runner_backend"

// Global test variables needed for middleware test setup
var (
	testDB           *gorm.DB // May not be needed directly, but setup requires it
	testSessionStore *sessions.CookieStore
	setupEnvOnce     sync.Once
)

// setupTestEnvironment loads environment variables, initializes DB and session store for tests.
// Minimal setup needed just for authMiddleware test.
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
		}

		// Initialize DB connection (needed for dependent setups)
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

		// Setup dependent packages needed for oauth.GetCurrentUser used in middleware
		// We don't need full DB setup/migrations for *these* specific tests,
		// but oauth.Setup requires the session store.
		usertable.Setup() // Needs DB
		oauth.Setup(testSessionStore)
		login.Setup(testDB, testSessionStore) // Needs DB and Store
		// prodtable.Setup() // Not strictly needed for routes tests
		// storefronttable.Setup() // Not strictly needed for routes tests

		usertable.MigrateUserDB()
	})

	// No table clearing needed for these route/middleware tests
	err := usertable.ClearUserTable(testDB)
	require.NoError(t, err, "Failed to clear user table for tests")
}

// TestRouteExistenceAndBasicHandling checks if routes are registered and return expected basic status codes.
func TestRouteExistenceAndBasicHandling(t *testing.T) {
	setupTestEnvironment(t) // Needed for oauth setup used by middleware

	// Create a router and register actual routes
	router := mux.NewRouter()
	testHandler := RegisterRoutes(router, false)
	server := httptest.NewServer(testHandler)
	defer server.Close()

	testCases := []struct {
		method         string
		path           string
		expectedStatus int
		body           string // Field exists
		contentType    string // Field exists
	}{
		// --- API Routes ---
		{"POST", "/api/register", http.StatusOK, // Expect 400 now with incomplete data
			"email=test@test.com&password=pw&name=Test", // Provide required fields
			"application/x-www-form-urlencoded",
		},
		{"POST", "/api/login", http.StatusBadRequest, // Login also needs data, expect 400 without it
			"", // Empty body
			"", // No specific content type needed for this basic check
		},
		{"POST", "/api/logout", http.StatusSeeOther, "", ""}, // Added "" for body and contentType
		{"POST", "/api/add_product", http.StatusUnauthorized, "", ""},
		{"DELETE", "/api/delete_product?id=1", http.StatusUnauthorized, "", ""},
		{"PUT", "/api/update_product?id=1", http.StatusUnauthorized, "", ""},
		{"GET", "/api/get_product?id=1", http.StatusUnauthorized, "", ""},
		{"GET", "/api/get_products", http.StatusUnauthorized, "", ""},
		{"GET", "/api/get_product_image?image=test.jpg", http.StatusUnauthorized, "", ""},
		{"POST", "/api/add_storefront", http.StatusUnauthorized, "", ""},
		{"GET", "/api/get_storefronts", http.StatusUnauthorized, "", ""},
		{"PUT", "/api/update_storefront?id=1", http.StatusUnauthorized, "", ""},
		{"DELETE", "/api/delete_storefront?id=1", http.StatusUnauthorized, "", ""},

		// --- Invalid API Route ---
		{"GET", "/api/nonexistent/route", http.StatusNotFound, "", ""},

		// --- Auth Routes ---
		{"GET", "/auth/google", http.StatusSeeOther, "", ""},
		{"GET", "/auth/google/callback", http.StatusSeeOther, "", ""},
		{"GET", "/logout", http.StatusSeeOther, "", ""},

		// --- Swagger ---
		{"GET", "/swagger/index.html", http.StatusOK, "", ""},

		// --- SPA Routes ---
		{"GET", "/login", http.StatusNotFound, "", ""},
		{"GET", "/register", http.StatusNotFound, "", ""},
		{"GET", "/static/some.js", http.StatusNotFound, "", ""},
		{"GET", "/assets/some.css", http.StatusNotFound, "", ""},
		{"GET", "/", http.StatusSeeOther, "", ""},
		{"GET", "/dashboard", http.StatusSeeOther, "", ""},
		{"GET", "/some/other/path", http.StatusSeeOther, "", ""},
	}

	client := server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.method, tc.path), func(t *testing.T) {
			var reqBody io.Reader
			if tc.body != "" {
				reqBody = strings.NewReader(tc.body)
			}
			req, err := http.NewRequest(tc.method, server.URL+tc.path, reqBody)
			require.NoError(t, err)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Expected status %d for %s %s, got %d", tc.expectedStatus, tc.method, tc.path, resp.StatusCode)

			if tc.expectedStatus == http.StatusSeeOther {
				location := resp.Header.Get("Location")
				// For SPA redirects, expect /login. For auth callback failure, it might redirect elsewhere (e.g., / or /login)
				if !strings.HasPrefix(tc.path, "/auth/") && tc.path != "/logout" && !strings.HasPrefix(tc.path, "/api") {
					assert.Contains(t, location, "/login", "Expected SPA redirect location to contain '/login' for %s %s", tc.method, tc.path)
				} else if tc.path == "/auth/google/callback" {
					// Check it redirects somewhere, maybe root or login
					assert.NotEmpty(t, location, "Expected redirect location for failed callback %s %s", tc.method, tc.path)
				}
			}
		})
	}
}

// TestSPAHandler tests the spaHandler's logic directly.
func TestSPAHandler(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.html")
	staticDir := filepath.Join(tempDir, "static")
	staticFilePath := filepath.Join(staticDir, "app.js")

	err := os.WriteFile(indexPath, []byte("<html>Index</html>"), 0644)
	require.NoError(t, err)
	err = os.MkdirAll(staticDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(staticFilePath, []byte("console.log('app');"), 0644)
	require.NoError(t, err)

	// Create the handler instance
	handler := spaHandler{staticPath: tempDir, indexPath: "index.html"}

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"Root", "/", http.StatusOK, "<html>Index</html>"},
		{"Index explicitly", "/index.html", http.StatusOK, "<html>Index</html>"},
		{"Subdirectory path", "/some/other/path", http.StatusOK, "<html>Index</html>"}, // Should serve index
		{"Existing static file", "/static/app.js", http.StatusOK, "console.log('app');"},
		{"Non-existent static file", "/static/style.css", http.StatusOK, "<html>Index</html>"}, // Should serve index
		{"Request for directory", "/static/", http.StatusOK, "<html>Index</html>"},             // Should serve index
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, tc.expectedBody, rr.Body.String())
		})
	}
}

// Helper to create a test user directly in the DB
func createTestUser(t *testing.T, email, password string) *usertable.User {
	t.Helper()
	// Use a dummy password hash for middleware tests if bcrypt is slow/not needed
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// require.NoError(t, err, "Failed to hash password for test user")
	hashedPassword := "$2a$10$..." // Or generate once

	user := &usertable.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User " + email,
		Provider:     "local",
	}
	err := usertable.CreateUser(user)
	require.NoError(t, err, "Failed to create test user using usertable.CreateUser")
	createdUser, err := usertable.GetUserByEmail(email)
	require.NoError(t, err, "Failed to fetch created test user by email")
	require.NotNil(t, createdUser, "Fetched created test user should not be nil")
	return createdUser
}

// TestAuthMiddleware tests the authMiddleware logic directly.
func TestAuthMiddleware(t *testing.T) {
	setupTestEnvironment(t) // Need session store and oauth setup

	// Dummy handler that the middleware should call if auth succeeds
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Auth OK")
	})

	// Wrap the dummy handler with the middleware
	authHandler := authMiddleware(nextHandler)

	// --- Define constants locally for the test ---
	const sessionName = "front-runner-session"
	const userSessionKey = "userID"
	// --- End constant definition ---

	// --- Test Case 1: No Session Cookie ---
	t.Run("NoSession", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected/resource", nil)
		rr := httptest.NewRecorder()

		authHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusSeeOther, rr.Code, "Expected redirect status")
		location := rr.Header().Get("Location")
		assert.Equal(t, "/login", location, "Expected redirect to /login")
	})

	// --- Test Case 2: Valid Session Cookie ---
	t.Run("ValidSession", func(t *testing.T) {
		testUser := createTestUser(t, "auth@test.com", "password")

		req := httptest.NewRequest("GET", "/protected/resource", nil)

		// Create a valid session cookie using the local constants
		session, err := testSessionStore.New(req, sessionName) // Use local constant
		require.NoError(t, err)
		// Simulate a logged-in user (doesn't have to exist in DB for this middleware test)
		session.Values[userSessionKey] = testUser.ID // Use local constant

		rrCookieSetter := httptest.NewRecorder()
		err = testSessionStore.Save(req, rrCookieSetter, session)
		require.NoError(t, err)
		cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
		require.NotEmpty(t, cookieHeader)
		req.Header.Set("Cookie", cookieHeader) // Add the valid cookie

		// Make request to the middleware-wrapped handler
		rr := httptest.NewRecorder()
		authHandler.ServeHTTP(rr, req)

		// Assert that the next handler was called
		assert.Equal(t, http.StatusOK, rr.Code, "Expected OK status from next handler")
		assert.Equal(t, "Auth OK", rr.Body.String(), "Expected body from next handler")
	})

	// --- Test Case 3: Invalid/Expired Session Cookie (Optional but good) ---
	t.Run("InvalidSessionValue", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected/resource", nil)

		// Create a session with an invalid value type for user ID using local constants
		session, err := testSessionStore.New(req, sessionName) // Use local constant
		require.NoError(t, err)
		session.Values[userSessionKey] = "not-a-uint" // Invalid type using local constant

		rrCookieSetter := httptest.NewRecorder()
		err = testSessionStore.Save(req, rrCookieSetter, session)
		require.NoError(t, err)
		cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
		require.NotEmpty(t, cookieHeader)
		req.Header.Set("Cookie", cookieHeader)

		// Make request
		rr := httptest.NewRecorder()
		authHandler.ServeHTTP(rr, req)

		// oauth.GetCurrentUser should return an error, leading to redirect
		assert.Equal(t, http.StatusSeeOther, rr.Code, "Expected redirect status on invalid session value")
		location := rr.Header().Get("Location")
		assert.Equal(t, "/login", location, "Expected redirect to /login on invalid session value")
	})
}

// TestInvalidAPI tests the handler for undefined API routes.
func TestInvalidAPI(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/this/does/not/exist", nil)
	rr := httptest.NewRecorder()

	InvalidAPI(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid API Endpoint")
}

// package routes

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"github.com/gorilla/mux"
// )

// // TestRegisterRoutes_AllRoutes verifies that the router correctly matches the expected routes and HTTP methods
// // for all registered endpoints including API, data, Swagger, and SPA routes.
// //
// // @Summary      Verify all route matching
// // @Description  Creates a new router with registered routes and tests that expected HTTP methods and paths for API endpoints,
// // data endpoints, Swagger UI, and SPA static routes are properly matched.
// // @Tags         testing, routes, router
// func TestRegisterRoutes_AllRoutes(t *testing.T) {
// 	// Create a new router and register all routes.
// 	router := mux.NewRouter()
// 	RegisterRoutes(router, false)

// 	// Define test cases for each endpoint.
// 	testCases := []struct {
// 		method      string
// 		path        string
// 		shouldMatch bool
// 	}{
// 		// API endpoints
// 		{"POST", "/api/register", true},
// 		{"POST", "/api/login", true},
// 		{"POST", "/api/logout", true},
// 		{"POST", "/api/add_product", true},
// 		{"PUT", "/api/delete_product", true},
// 		{"PUT", "/api/update_product", true},
// 		// API catch-all for invalid endpoints
// 		{"GET", "/api/nonexistent", true},
// 		// Data endpoints
// 		{"GET", "/api/data/image/sample.jpg", true},
// 		{"POST", "/api/data/upload", true},
// 		// Swagger UI
// 		{"GET", "/swagger/index.html", true},
// 		// SPA routes for ReactJS frontend
// 		{"GET", "/login", true},
// 		{"GET", "/register", true},
// 		{"GET", "/static/somefile.js", true},
// 		{"GET", "/assets/somefile.css", true},
// 		{"GET", "/manifest.json", true},
// 		// Catch-all SPA route wrapped with authMiddleware (e.g., for a dashboard)
// 		{"GET", "/dashboard", true},
// 	}

// 	// Iterate through each test case and verify matching.
// 	for _, tc := range testCases {
// 		req, err := http.NewRequest(tc.method, tc.path, nil)
// 		if err != nil {
// 			t.Fatalf("failed to create %s request for %s: %v", tc.method, tc.path, err)
// 		}

// 		var match mux.RouteMatch
// 		matched := router.Match(req, &match)
// 		if matched != tc.shouldMatch {
// 			t.Errorf("expected match=%v for %s %s, got %v", tc.shouldMatch, tc.method, tc.path, matched)
// 		}
// 	}
// }

// // createDummyIndex creates a dummy index.html file in a temporary directory for testing static file serving.
// //
// // @Summary      Create dummy index file
// // @Description  Generates a temporary directory containing a dummy index.html file to simulate a static build directory for testing purposes.
// // @Tags         testing, helper, routes, router
// func createDummyIndex(t *testing.T) string {
// 	tempDir, err := os.MkdirTemp("", "build")
// 	if err != nil {
// 		t.Fatalf("failed to create temporary directory: %v", err)
// 	}
// 	// Uncomment the following line to automatically clean up the temporary directory after tests.
// 	// defer os.RemoveAll(tempDir)

// 	dummyIndexPath := filepath.Join(tempDir, "index.html")
// 	dummyContent := []byte("dummy index")
// 	if err := os.WriteFile(dummyIndexPath, dummyContent, 0644); err != nil {
// 		t.Fatalf("failed to write dummy index.html: %v", err)
// 	}

// 	return tempDir
// }

// // registerDummyRoutes registers dummy API endpoints and static file routes using the provided temporary directory.
// // This helper function is used to simulate API and static file routing for testing purposes.
// //
// // @Summary      Register dummy routes for testing
// // @Description  Sets up dummy API endpoints (for registration, login, and logout) and static file serving routes using a temporary directory.
// // @Tags         testing, routes, router, helper
// func registerDummyRoutes(router *mux.Router, tempDir string) {
// 	// Dummy API endpoints.
// 	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")
// 	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")
// 	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
// 	// Stub Swagger route.
// 	router.PathPrefix("/swagger/").Handler(http.NotFoundHandler())
// 	// Static files: use our temporary directory.
// 	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(tempDir, "static"))))
// 	router.PathPrefix("/static/").Handler(staticHandler)
// 	// For "/" serve files directly from tempDir.
// 	router.PathPrefix("/").Handler(http.FileServer(http.Dir(tempDir)))
// }

// // TestRegisterRoutes_WithDummyStaticFile verifies that the static file server returns the dummy index file.
// // This test ensures that a GET request to the root path ("/") serves the expected index.html content.
// //
// // @Summary      Test static file serving via dummy routes
// // @Description  Sets up a dummy static file server with a temporary directory and verifies that a GET request to "/" returns the dummy index file content.
// // @Tags         testing, router, routes
// func TestRegisterRoutes_WithDummyStaticFile(t *testing.T) {
// 	tempDir := createDummyIndex(t)
// 	router := mux.NewRouter()
// 	registerDummyRoutes(router, tempDir)
// 	// Reset the default mux to avoid conflicts in tests.
// 	http.DefaultServeMux = http.NewServeMux()
// 	http.Handle("/", router)

// 	req, err := http.NewRequest("GET", "/", nil)
// 	if err != nil {
// 		t.Fatalf("failed to create GET request for '/': %v", err)
// 	}

// 	rr := httptest.NewRecorder()
// 	http.DefaultServeMux.ServeHTTP(rr, req)

// 	// Verify that the status code is 200 and the response body matches the dummy content.
// 	if rr.Code != http.StatusOK {
// 		t.Errorf("expected status 200, got %d", rr.Code)
// 	}
// 	if rr.Body.String() != "dummy index" {
// 		t.Errorf("expected body %q, got %q", "dummy index", rr.Body.String())
// 	}
// }
