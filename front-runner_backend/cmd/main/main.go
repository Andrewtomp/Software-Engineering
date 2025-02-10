// main.go
package main

import (
	"crypto/tls"
	"front-runner/internal/login"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
	router.HandleFunc("/register", login.RegisterUser).Methods("POST")
	router.HandleFunc("/login", login.LoginUser).Methods("POST")
	router.HandleFunc("/logout", login.LogoutUser).Methods("GET")

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
