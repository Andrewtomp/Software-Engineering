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

	router := mux.NewRouter()

	routes.RegisterRoutes(router)

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
