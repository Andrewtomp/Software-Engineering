package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

// TestRegisterRoutes_AllRoutes verifies that the router correctly matches the expected routes and HTTP methods
// for all registered endpoints including API, data, Swagger, and SPA routes.
//
// @Summary      Verify all route matching
// @Description  Creates a new router with registered routes and tests that expected HTTP methods and paths for API endpoints,
// data endpoints, Swagger UI, and SPA static routes are properly matched.
// @Tags         testing, routes, router
func TestRegisterRoutes_AllRoutes(t *testing.T) {
	// Create a new router and register all routes.
	router := mux.NewRouter()
	RegisterRoutes(router, false)

	// Define test cases for each endpoint.
	testCases := []struct {
		method      string
		path        string
		shouldMatch bool
	}{
		// API endpoints
		{"POST", "/api/register", true},
		{"POST", "/api/login", true},
		{"POST", "/api/logout", true},
		{"POST", "/api/add_product", true},
		{"PUT", "/api/delete_product", true},
		{"PUT", "/api/update_product", true},
		// API catch-all for invalid endpoints
		{"GET", "/api/nonexistent", true},
		// Data endpoints
		{"GET", "/api/data/image/sample.jpg", true},
		{"POST", "/api/data/upload", true},
		// Swagger UI
		{"GET", "/swagger/index.html", true},
		// SPA routes for ReactJS frontend
		{"GET", "/login", true},
		{"GET", "/register", true},
		{"GET", "/static/somefile.js", true},
		{"GET", "/assets/somefile.css", true},
		{"GET", "/manifest.json", true},
		// Catch-all SPA route wrapped with authMiddleware (e.g., for a dashboard)
		{"GET", "/dashboard", true},
	}

	// Iterate through each test case and verify matching.
	for _, tc := range testCases {
		req, err := http.NewRequest(tc.method, tc.path, nil)
		if err != nil {
			t.Fatalf("failed to create %s request for %s: %v", tc.method, tc.path, err)
		}

		var match mux.RouteMatch
		matched := router.Match(req, &match)
		if matched != tc.shouldMatch {
			t.Errorf("expected match=%v for %s %s, got %v", tc.shouldMatch, tc.method, tc.path, matched)
		}
	}
}

// createDummyIndex creates a dummy index.html file in a temporary directory for testing static file serving.
//
// @Summary      Create dummy index file
// @Description  Generates a temporary directory containing a dummy index.html file to simulate a static build directory for testing purposes.
// @Tags         testing, helper, routes, router
func createDummyIndex(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "build")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	// Uncomment the following line to automatically clean up the temporary directory after tests.
	// defer os.RemoveAll(tempDir)

	dummyIndexPath := filepath.Join(tempDir, "index.html")
	dummyContent := []byte("dummy index")
	if err := os.WriteFile(dummyIndexPath, dummyContent, 0644); err != nil {
		t.Fatalf("failed to write dummy index.html: %v", err)
	}

	return tempDir
}

// registerDummyRoutes registers dummy API endpoints and static file routes using the provided temporary directory.
// This helper function is used to simulate API and static file routing for testing purposes.
//
// @Summary      Register dummy routes for testing
// @Description  Sets up dummy API endpoints (for registration, login, and logout) and static file serving routes using a temporary directory.
// @Tags         testing, routes, router, helper
func registerDummyRoutes(router *mux.Router, tempDir string) {
	// Dummy API endpoints.
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")
	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	// Stub Swagger route.
	router.PathPrefix("/swagger/").Handler(http.NotFoundHandler())
	// Static files: use our temporary directory.
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(tempDir, "static"))))
	router.PathPrefix("/static/").Handler(staticHandler)
	// For "/" serve files directly from tempDir.
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(tempDir)))
}

// TestRegisterRoutes_WithDummyStaticFile verifies that the static file server returns the dummy index file.
// This test ensures that a GET request to the root path ("/") serves the expected index.html content.
//
// @Summary      Test static file serving via dummy routes
// @Description  Sets up a dummy static file server with a temporary directory and verifies that a GET request to "/" returns the dummy index file content.
// @Tags         testing, router, routes
func TestRegisterRoutes_WithDummyStaticFile(t *testing.T) {
	tempDir := createDummyIndex(t)
	router := mux.NewRouter()
	registerDummyRoutes(router, tempDir)
	// Reset the default mux to avoid conflicts in tests.
	http.DefaultServeMux = http.NewServeMux()
	http.Handle("/", router)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("failed to create GET request for '/': %v", err)
	}

	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	// Verify that the status code is 200 and the response body matches the dummy content.
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "dummy index" {
		t.Errorf("expected body %q, got %q", "dummy index", rr.Body.String())
	}
}
