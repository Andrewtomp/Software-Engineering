package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

// TestRegisterRoutes_Match verifies that the router matches expected routes and methods.
//
// @Summary      Verify route matching
// @Description  Creates a new router with registered routes and checks that expected HTTP methods and paths are correctly matched.
//
// @Tags         testing, routes, router
func TestRegisterRoutes_Match(t *testing.T) {
	// Create a new router and register the routes.
	router := mux.NewRouter()
	RegisterRoutes(router, false)

	// Define a list of test cases with the expected HTTP method and path.
	testCases := []struct {
		method      string
		path        string
		shouldMatch bool
	}{
		{"POST", "/api/register", true},
		{"POST", "/api/login", true},
		{"GET", "/api/logout", true},
		{"GET", "/swagger/index.html", true},
		{"GET", "/static/somefile.js", true},
		{"GET", "/", true},
	}

	// For each test case, create a new request and verify that the router can match it.
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
// @Description  Creates a temporary directory with a dummy index.html file to simulate a static build directory.
//
// @Tags         testing, helper, routes, router
func createDummyIndex(t *testing.T) string {
	// Create a temporary directory to simulate the static build directory.
	tempDir, err := os.MkdirTemp("", "build")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	// defer os.RemoveAll(tempDir)

	// Create a dummy index.html file.
	dummyIndexPath := filepath.Join(tempDir, "index.html")
	dummyContent := []byte("dummy index")
	if err := os.WriteFile(dummyIndexPath, dummyContent, 0644); err != nil {
		t.Fatalf("failed to write dummy index.html: %v", err)
	}

	return tempDir
}

// registerDummyRoutes registers dummy API endpoints and static file routes using the provided temporary directory.
//
// @Summary      Register dummy routes
// @Description  Registers dummy API endpoints and static file serving routes for testing purposes using a temporary directory.
//
// @Tags         testing, routes, router, helper
func registerDummyRoutes(router *mux.Router, tempDir string) {
	// Register your API routes as usual.
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")
	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	// Swagger can remain as is or be stubbed if needed.
	router.PathPrefix("/swagger/").Handler(http.NotFoundHandler())
	// Static files: use our temporary directory.
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(tempDir, "static"))))
	router.PathPrefix("/static/").Handler(staticHandler)
	// For "/" we serve files from tempDir.
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(tempDir)))
}

// TestRegisterRoutes_WithDummyStaticFile verifies that the static file server returns the dummy index file.
//
// @Summary      Test static file serving
// @Description  Sets up a dummy static file server and verifies that a GET request to "/" returns the expected dummy index content.
//
// @Tags         testing, router, routes
func TestRegisterRoutes_WithDummyStaticFile(t *testing.T) {
	tempDir := createDummyIndex(t)

	// Override the static file server in the router.
	router := mux.NewRouter()

	registerDummyRoutes(router, tempDir)

	// Also, register the router on the default mux.
	// To avoid conflicts in tests, reset the default mux.
	http.DefaultServeMux = http.NewServeMux()
	http.Handle("/", router)

	// Now, send a request to "/"
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("failed to create GET request for '/': %v", err)
	}

	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	// Expect a 200 status code and the dummy content.
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "dummy index" {
		t.Errorf("expected body %q, got %q", "dummy index", rr.Body.String())
	}
}
