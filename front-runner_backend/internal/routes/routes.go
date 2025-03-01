package routes

import (
	"front-runner/internal/login"
	"front-runner/internal/prodtable"
	"front-runner/internal/usertable"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

// Serves the static ReactJS site. Continually serves index.html as a result of ReactJS internal routing.
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Join internally call path.Clean to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)

	// check whether a file exists or is a directory at the given path
	fi, err := os.Stat(path)
	if os.IsNotExist(err) || fi.IsDir() {
		// file does not exist or path is a directory, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func InvalidAPI(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Invalid API Endpoint", http.StatusNotFound)
}

// RegisterRoutes sets up all the application routes including API endpoints, Swagger UI, and static file serving.
//
// @Summary      Register application routes
// @Description  Registers API endpoints for user registration, login, and logout. REgisters API endpoints for adding, removing, and updateing a product Also sets up the Swagger UI on /swagger/* and serves static files for the frontend.
//
// @Tags         routes, router
// @Accept       json
// @Produce      html
func RegisterRoutes(router *mux.Router, logging bool) http.Handler {
	// API subrouter
	api := router.PathPrefix("/api").Subrouter()

	// API endpoints
	// User Table
	api.HandleFunc("/register", usertable.RegisterUser).Methods("POST")
	// Login
	api.HandleFunc("/login", login.LoginUser).Methods("POST")
	api.HandleFunc("/logout", login.LogoutUser).Methods("GET")
	// Product Table
	api.HandleFunc("/add_product", prodtable.AddProduct)
	api.HandleFunc("/delete_product", prodtable.DeleteProduct)
	api.HandleFunc("/update_product", prodtable.UpdateProduct)
	// Serve Swagger UI on /swagger/*
	api.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Serve static files for webpage
	spa := spaHandler{staticPath: "../front-runner/build", indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	// Logging
	if logging {
		return handlers.LoggingHandler(os.Stdout, router)
	}
	return router
}
