// main.go
// @title Front Runner API
// @version 1.0
// @description API documentation for the Front Runner application.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
package main

import (
	"crypto/tls"
	"front-runner/internal/login"
	"log"
	"net/http"

	_ "front-runner/docs" // This is important for swagger to find your docs!

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger" // swagger middleware
)

const (
	port = ":8080"
)

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load X509 key pair: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/register", login.RegisterUser).Methods("POST")
	router.HandleFunc("/login", login.LoginUser).Methods("POST")
	router.HandleFunc("/logout", login.LogoutUser).Methods("GET")

	// Serve Swagger UI on /swagger/*
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Serve static files for your webpage
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("../front-runner/build/static/")))
	router.PathPrefix("/static/").Handler(s)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../front-runner/build")))
	http.Handle("/", router)

	server := &http.Server{
		Addr:      port,
		Handler:   router,
		TLSConfig: config,
	}

	log.Printf("Listening on %s...", port)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
