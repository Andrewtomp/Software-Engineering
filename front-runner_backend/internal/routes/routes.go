package routes

import (
	"fmt"
	"front-runner/internal/login"
	"front-runner/internal/prodtable"
	"front-runner/internal/usertable"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterRoutes sets up all the application routes including API endpoints, Swagger UI, and static file serving.
//
// @Summary      Register application routes
// @Description  Registers API endpoints for user registration, login, and logout. REgisters API endpoints for adding, removing, and updateing a product Also sets up the Swagger UI on /swagger/* and serves static files for the frontend.
//
// @Tags         routes, router
// @Accept       json
// @Produce      html
func RegisterRoutes(router *mux.Router) {
	// API endpoints
	// User Table
	router.HandleFunc("/register", usertable.RegisterUser).Methods("POST")
	// Login
	router.HandleFunc("/login", login.LoginUser).Methods("POST")
	router.HandleFunc("/logout", login.LogoutUser).Methods("GET")
	// Product Table
	router.HandleFunc("/add_product", prodtable.AddProduct)
	router.HandleFunc("/delete_product", prodtable.DeleteProduct)
	router.HandleFunc("/update_product", prodtable.UpdateProduct)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)

	// Serve Swagger UI on /swagger/*
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Serve static files for your webpage
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("../front-runner/build/static/")))
	router.PathPrefix("/static/").Handler(s)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../front-runner/build")))
	http.Handle("/", router)
}
