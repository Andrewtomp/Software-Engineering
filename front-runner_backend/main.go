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
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"

	_ "front-runner/docs" // This is important for swagger to find your docs!
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/prodtable"
	"front-runner/internal/routes"
	"front-runner/internal/storefronttable"
	"front-runner/internal/usertable"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pborman/getopt/v2"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

var (
	port          = "8080"
	local    bool = false
	verbose  bool = false
	envFile  string
	useNgrok bool = false // <-- Add flag for ngrok
)

func setupModules() {
	// Database
	coredbutils.LoadEnv()

	// Users table
	usertable.Setup()
	usertable.MigrateUserDB()

	// Login information
	login.Setup()

	// Product Table
	prodtable.Setup()
	prodtable.MigrateProdDB()

	// Storefront Table
	storefronttable.Setup()
	storefronttable.MigrateStorefrontDB()
}

func main() {
	getopt.FlagLong(&verbose, "verbose", 'v', "Enable logging of incomming HTTP requests")
	getopt.FlagLong(&port, "port", 'p', "Specify port to listen for connnections")
	getopt.FlagLong(&local, "local", 'l', "Only listen for connnections over localhost")
	getopt.FlagLong(&envFile, "env", 0, "Specify enviroment variable file to load from")
	getopt.FlagLong(&useNgrok, "ngrok", 0, "Expose the server via ngrok (requires NGROK_AUTHTOKEN env var)")
	getopt.Parse()

	var err error
	if envFile == "" {
		err = godotenv.Load()
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Error loading default .env file: %v", err)
		}
	} else {
		err = godotenv.Load(envFile)
		if err != nil {
			log.Fatalf("Error loading specified .env file '%s': %v", envFile, err)
		}
	}

	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	setupModules()

	certFile := "server.crt"
	keyFile := "server.key"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load X509 key pair (%s, %s): %v", certFile, keyFile, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	router := mux.NewRouter()
	routeHandler := routes.RegisterRoutes(router, verbose)

	// var address = ""
	// if local {
	// 	address = "localhost"
	// }

	server := &http.Server{
		// Addr:      address + ":" + port,
		Handler:   routeHandler,
		TLSConfig: tlsConfig,
	}

	if useNgrok {
		// Using Ngrok
		log.Println("Ngrok mode enabled. Attempting to start tunnel...")

		// Check for the ngrok auth token environment variable
		if os.Getenv("NGROK_AUTHTOKEN") == "" {
			log.Fatal("Error: NGROK_AUTHTOKEN environment variable is not set. " +
				"Please set it to your ngrok authentication token when using the --ngrok flag. " +
				"Get one from https://dashboard.ngrok.com/get-started/your-authtoken")
		}

		staticDomain := os.Getenv("NGROK_DOMAIN")
		log.Printf("Attempting to use ngrok static domain: %s", staticDomain)

		// Create an ngrok listener
		// It will establish the tunnel using your authtoken from the environment
		// The default endpoint type is http/https
		tun, err := ngrok.Listen(context.Background(),
			config.HTTPEndpoint(
				config.WithDomain(staticDomain),
			// config.WithScheme(config.SchemeHTTPS),
			), // Basic HTTPS endpoint tunnel
			ngrok.WithAuthtokenFromEnv(), // Read token from NGROK_AUTHTOKEN env var
		)
		if err != nil {
			log.Fatalf("Failed to create ngrok listener: %v", err)
		}

		// // --- ADD THIS ---
		// actualNgrokURL := tun.URL() // Get the URL ngrok actually assigned
		// log.Printf(">>> Ngrok tunnel established. Use this EXACT URL: %s <<<", actualNgrokURL)
		// // --- END ADD ---

		log.Println("Successfully connected to ngrok.")
		log.Printf("Public URL: %s", tun.URL()) // Log the public URL ngrok provides
		log.Printf("Server ready to accept connections via ngrok tunnel...")

		// Start the HTTPS server using the ngrok listener and your existing TLS config
		// ServeTLS requires the listener, certFile, and keyFile args are ignored here
		// because the TLSConfig is already set on the server.
		// err = server.ServeTLS(tun, "", "")
		err = http.Serve(tun, router)

	} else {
		// Standard Local Server (Not using Ngrok)
		address := ""
		if local {
			address = "localhost" // Bind only to loopback interface
			log.Printf("Starting local HTTPS server (localhost only) on port %s...", port)
		} else {
			log.Printf("Starting local HTTPS server (all interfaces) on port %s...", port)
		}
		server.Addr = address + ":" + port

		// Start the HTTPS server using the standard library listener creation.
		// ListenAndServeTLS needs the certFile and keyFile paths again here.
		err = server.ListenAndServeTLS(certFile, keyFile)
	}

	// --- Handle Server Exit ---
	// Log error only if it's not the standard server closed signal
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	} else {
		log.Println("Server stopped gracefully.")
	}

	// log.Printf("Listening on %s...", server.Addr)
	// err = server.ListenAndServeTLS("", "")
	// if err != nil {
	// 	log.Fatalf("Failed to start server: %v", err)
	// }
}
