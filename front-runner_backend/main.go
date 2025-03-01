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
	"log"
	"net/http"

	_ "front-runner/docs" // This is important for swagger to find your docs!
	"front-runner/internal/routes"

	"github.com/gorilla/mux"
	"github.com/pborman/getopt/v2"
)

var (
	port         = "8080"
	verbose bool = false
)

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load X509 key pair: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	getopt.FlagLong(&verbose, "verbose", 'v', "Enable logging of incomming HTTP requests")
	getopt.FlagLong(&port, "port", 'p', "Specify port to listen for connnections")

	getopt.Parse()

	router := mux.NewRouter()

	routeHandler := routes.RegisterRoutes(router, verbose)

	server := &http.Server{
		Addr:      ":" + port,
		Handler:   routeHandler,
		TLSConfig: config,
	}

	log.Printf("Listening on %s...", server.Addr)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
