package routes

import (
	"front-runner/internal/imageStore"
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

// var (
// 	sessionStore *sessions.CookieStore
// )

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

// authMiddleware redirects to /login if the user is not authenticated.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !login.IsLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
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
	// Subrouters
	api := router.PathPrefix("/api").Subrouter()
	data := api.PathPrefix("/data").Subrouter()

	// API endpoints
	// User Table
	api.HandleFunc("/register", usertable.RegisterUser).Methods("POST")
	// Login
	api.HandleFunc("/login", login.LoginUser).Methods("POST")
	api.HandleFunc("/logout", login.LogoutUser).Methods("POST")
	// Product Table
	api.HandleFunc("/add_product", prodtable.AddProduct).Methods("POST")
	api.HandleFunc("/delete_product", prodtable.DeleteProduct).Methods("PUT")
	api.HandleFunc("/update_product", prodtable.UpdateProduct).Methods("PUT")
	api.HandleFunc("/get_product", prodtable.GetProduct).Methods("GET")
	api.HandleFunc("/get_product_image", prodtable.GetProductImage).Methods("GET")
	api.HandleFunc("/get_product_list", prodtable.GetProductList).Methods("GET")

	api.PathPrefix("/").HandlerFunc(InvalidAPI)

	data.HandleFunc("/image/{imagePath}", imageStore.LoadImage).Methods("GET")
	data.HandleFunc("/upload", imageStore.UploadImage).Methods("POST")

	// Serve Swagger UI on /swagger/*
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Page routes.
	spa := spaHandler{staticPath: "../front-runner/build", indexPath: "index.html"}
	// /login always serves the SPA. React will show the login UI.
	router.Handle("/login", spa).Methods("GET")
	router.Handle("/register", spa).Methods("GET")
	// node-specified routes will be routed directly to the spa
	router.PathPrefix("/static").Handler(spa).Methods("GET")
	router.PathPrefix("/assets").Handler(spa).Methods("GET")
	router.PathPrefix("/manifest.json").Handler(spa).Methods("GET")
	// Serve static files for webpage
	// All other pages are wrapped with authMiddleware.
	router.PathPrefix("/").Handler(authMiddleware(spa)).Methods("GET")

	// Logging
	if logging {
		return handlers.LoggingHandler(os.Stdout, router)
	}
	return router
}
