// main.go
// @title Front Runner API
// @version 1.0
// @description API documentation for the Front Runner application.
// @contact.name API Support
// @contact.email jonathan.bravo@ufl.edu
// @license.name MIT
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	_ "front-runner/docs" // This is important for swagger to find your docs!
	"front-runner/internal/routes"

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

	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "verbose", false, "Enable logging of incomming HTTP requests")
	flag.Parse()

	router := mux.NewRouter()

	routeHandler := routes.RegisterRoutes(router, verboseFlag)

	server := &http.Server{
		Addr:      port,
		Handler:   routeHandler,
		TLSConfig: config,
	}

	log.Printf("Listening on %s...", port)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
