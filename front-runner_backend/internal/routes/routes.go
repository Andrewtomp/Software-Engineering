package routes

import (
	"front-runner/internal/login"
	"front-runner/internal/usertable"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(router *mux.Router) {
	// API endpoints
	router.HandleFunc("/register", usertable.RegisterUser).Methods("POST")
	router.HandleFunc("/login", login.LoginUser).Methods("POST")
	router.HandleFunc("/logout", login.LogoutUser).Methods("GET")

	// Serve Swagger UI on /swagger/*
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Serve static files for your webpage
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("../front-runner/build/static/")))
	router.PathPrefix("/static/").Handler(s)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../front-runner/build")))
	http.Handle("/", router)
}
