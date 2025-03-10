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
	"front-runner/internal/coredbutils"
	"front-runner/internal/imageStore"
	"front-runner/internal/login"
	"front-runner/internal/prodtable"
	"front-runner/internal/routes"
	"front-runner/internal/usertable"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pborman/getopt/v2"
	"gorm.io/gorm"
)

var (
	port         = "8080"
	local   bool = false
	verbose bool = false
	envFile string
	db      *gorm.DB
)

func setupModules() {
	// Database
	coredbutils.LoadEnv()
	db = coredbutils.GetDB()

	// Users table
	usertable.Setup()
	usertable.MigrateUserDB()

	// Login information
	login.Setup()

	// Product Table
	prodtable.Setup()
	prodtable.MigrateProdDB()

	// Image Store
	imageStore.Setup()
}

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
	getopt.FlagLong(&local, "local", 'l', "Only listen for connnections over localhost")
	getopt.FlagLong(&envFile, "env", 'e', "Specify enviroment variable file to load from")

	getopt.Parse()

	if envFile == "" {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(envFile)
	}

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	setupModules()

	router := mux.NewRouter()

	routeHandler := routes.RegisterRoutes(router, verbose)

	var address = ""
	if local {
		address = "localhost"
	}

	server := &http.Server{
		Addr:      address + ":" + port,
		Handler:   routeHandler,
		TLSConfig: config,
	}

	log.Printf("Listening on %s...", server.Addr)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
