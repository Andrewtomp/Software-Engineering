// main.go
// @title Front Runner API
// @version 1.0
// @description API documentation for the Front Runner application.
// @contact.name API Support
// @contact.email jonathan.bravo@ufl.edu
// @license.name MIT
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080 // Update this if using ngrok static domain for Swagger docs
// @BasePath /
package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings" // Import strings

	_ "front-runner/docs" // This is important for swagger to find your docs!
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/oauth" // Import oauth
	"front-runner/internal/prodtable"
	"front-runner/internal/routes"
	"front-runner/internal/storefronttable"
	"front-runner/internal/usertable"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions" // Import sessions
	"github.com/joho/godotenv"
	"github.com/markbates/goth/gothic"
	"github.com/pborman/getopt/v2"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"gorm.io/gorm" // Import gorm
)

var (
	port              = "8080"
	local        bool = false
	verbose      bool = false
	envFile      string
	useNgrok     bool                  = false
	db           *gorm.DB              // Hold DB connection globally in main
	sessionStore *sessions.CookieStore // Hold session store globally in main
	callbackURL  string                // Store the determined callback URL
	isSecure     = false               // Track if session cookie should be secure
)

// setupModules initializes essential components like the database connection,
// session store, and dependent internal packages (auth, tables, etc.).
// It ensures that environment variables are loaded and necessary migrations are run.
// It populates the global db and sessionStore variables.
func setupModules() {
	// --- 1. Load Environment Variables ---
	// Moved env loading to the start of main, but keep LoadEnv for DB utils if needed
	coredbutils.LoadEnv() // Ensures DB utils have env vars if they need them independently

	// --- 2. Initialize Database ---
	db, _ = coredbutils.GetDB()
	if db == nil {
		log.Fatal("Failed to get database connection")
	}

	sessionAuthKeyBase64 := strings.TrimSpace(os.Getenv("SESSION_AUTH_KEY"))
	if sessionAuthKeyBase64 == "" {
		log.Fatal("SESSION_AUTH_KEY environment variable not set")
	}
	sessionEncKeyBase64 := strings.TrimSpace(os.Getenv("SESSION_ENC_KEY")) // Optional

	authKey, err := base64.StdEncoding.DecodeString(sessionAuthKeyBase64)
	if err != nil {
		log.Fatalf("Failed to decode SESSION_AUTH_KEY: %v", err)
	}

	var encKey []byte
	if sessionEncKeyBase64 != "" {
		encKey, err = base64.StdEncoding.DecodeString(sessionEncKeyBase64)
		if err != nil {
			log.Fatalf("Failed to decode SESSION_ENC_KEY: %v", err)
		}
		// AES requires 16, 24, or 32 bytes for encryption key
		if len(encKey) != 16 && len(encKey) != 24 && len(encKey) != 32 {
			log.Fatalf("Decoded SESSION_ENC_KEY must be 16, 24, or 32 bytes long, got %d", len(encKey))
		}
		// log.Printf("Decoded SESSION_ENC_KEY length: %d bytes", len(encKey))
	}

	// Determine callback URL and secure flag *before* initializing store
	callbackURL = os.Getenv("GOOGLE_REDIRECT_URI")
	ngrokDomain := os.Getenv("NGROK_DOMAIN")
	callbackPath := "/auth/google/callback"

	if callbackURL == "" {
		if useNgrok && ngrokDomain != "" {
			if !strings.HasPrefix(ngrokDomain, "https://") && !strings.HasPrefix(ngrokDomain, "http://") {
				ngrokDomain = "https://" + ngrokDomain // Default to https
			}
			callbackURL = ngrokDomain + callbackPath
			log.Printf("GOOGLE_REDIRECT_URI not set, using NGROK_DOMAIN: %s", callbackURL)
		} else {
			// Default for local non-ngrok runs
			callbackURL = "https://localhost:" + port + callbackPath // Assume HTTPS locally
			log.Printf("Warning: GOOGLE_REDIRECT_URI and NGROK_DOMAIN not set or ngrok not used. Defaulting Google callback to: %s", callbackURL)
		}
	} else {
		log.Printf("Using explicitly set GOOGLE_REDIRECT_URI: %s", callbackURL)
	}
	isSecure = strings.HasPrefix(callbackURL, "https://") // Set secure flag based on final URL

	if len(encKey) == 0 {
		sessionStore = sessions.NewCookieStore(authKey) // Use decoded authKey ONLY
		log.Println("Session store initialized (no encryption).")
	} else {
		sessionStore = sessions.NewCookieStore(authKey, encKey) // Use decoded authKey and encKey
		log.Println("Session store initialized (with encryption).")
	}

	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   isSecure, // Use the determined secure flag
		SameSite: http.SameSiteLaxMode,
	}
	gothic.Store = sessionStore
	log.Println("Session store initialized.")

	// --- 4. Setup Dependent Packages (passing DB and Session Store) ---
	// Users table (only needs DB)
	usertable.Setup() // Assumes usertable.Setup only needs coredbutils.GetDB() internally now
	usertable.MigrateUserDB()

	// OAuth (needs Session Store and Callback URL - handled internally via env vars now)
	oauth.Setup(sessionStore) // oauth.Setup reads env vars and initializes goth/store

	// Login (needs DB and Session Store)
	login.Setup(db, sessionStore) // Pass the initialized DB and Store

	// Product Table (only needs DB)
	prodtable.Setup() // Assumes prodtable.Setup only needs coredbutils.GetDB() internally now
	prodtable.MigrateProdDB()

	// Storefront Table (only needs DB, encryption key loaded internally)
	storefronttable.Setup() // Assumes storefronttable.Setup uses coredbutils.GetDB() and loads key internally
	storefronttable.MigrateStorefrontDB()

	log.Println("All modules set up.")
}

// main is the entry point of the application.
// It parses command-line flags, loads environment variables, sets up modules,
// configures the HTTP/S server (potentially with ngrok), registers routes,
// and starts listening for incoming connections.
func main() {
	// --- Argument Parsing ---
	getopt.FlagLong(&verbose, "verbose", 'v', "Enable logging of incomming HTTP requests")
	getopt.FlagLong(&port, "port", 'p', "Specify port to listen for connnections")
	getopt.FlagLong(&local, "local", 'l', "Only listen for connnections over localhost")
	getopt.FlagLong(&envFile, "env", 0, "Specify enviroment variable file to load from")
	getopt.FlagLong(&useNgrok, "ngrok", 0, "Expose the server via ngrok (requires NGROK_AUTHTOKEN env var)")
	getopt.Parse()

	// --- Load Environment Variables ---
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
	log.Println("Environment variables loaded.")

	// --- Setup Database, Session Store, and Modules ---
	setupModules() // This now initializes DB, Session Store, and sets up other packages

	// --- TLS Configuration ---
	certFile := "server.crt"
	keyFile := "server.key"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load X509 key pair (%s, %s): %v", certFile, keyFile, err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// --- Router Setup ---
	router := mux.NewRouter()

	// --- OAuth Routes ---
	// These need to be registered *before* the general API/SPA routes
	// Ensure paths match the callback URL used in oauth.Setup
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/google", oauth.HandleGoogleLogin).Methods("GET")
	authRouter.HandleFunc("/google/callback", oauth.HandleGoogleCallback).Methods("GET")
	// Note: The /logout route might be better placed under /api or kept separate
	// Let's keep it separate for now, matching previous setup
	router.HandleFunc("/logout", oauth.HandleLogout).Methods("GET") // Use unified logout

	// --- Register Other Routes (API, Swagger, SPA) ---
	// routes.RegisterRoutes now handles API, Swagger, and SPA routing including auth middleware
	// The middleware uses oauth.GetCurrentUser which uses the shared sessionStore
	routeHandler := routes.RegisterRoutes(router, verbose)

	// --- Server Configuration ---
	server := &http.Server{
		Handler:   routeHandler, // Use the handler returned by RegisterRoutes (includes logging if enabled)
		TLSConfig: tlsConfig,
	}

	// --- Server Start (Ngrok or Local) ---
	if useNgrok {
		log.Println("Ngrok mode enabled. Attempting to start tunnel...")
		if os.Getenv("NGROK_AUTHTOKEN") == "" {
			log.Fatal("Error: NGROK_AUTHTOKEN environment variable is not set.")
		}

		// Use the static domain from env var for consistency
		staticDomain := os.Getenv("NGROK_DOMAIN")
		log.Printf("Attempting to use ngrok static domain: %s", staticDomain)

		// Create listener with the static domain
		listenerOpts := []config.HTTPEndpointOption{}
		if staticDomain != "" {
			listenerOpts = append(listenerOpts, config.WithDomain(staticDomain))
		}

		tun, err := ngrok.Listen(context.Background(),
			config.HTTPEndpoint(listenerOpts...),
			ngrok.WithAuthtokenFromEnv(),
		)
		if err != nil {
			log.Fatalf("Failed to create ngrok listener: %v", err)
		}

		log.Printf("Successfully connected to ngrok. Public URL: %s", tun.URL())
		log.Printf("Server ready to accept connections via ngrok tunnel...")

		// Serve using the ngrok listener (TLS is handled by ngrok)
		// We pass the main router directly, as TLSConfig is for local serving
		err = http.Serve(tun, router) // Serve plain HTTP to the tunnel

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		} else {
			log.Println("Server stopped gracefully.")
		}
	} else {
		// Standard Local Server
		address := ""
		if local {
			address = "localhost"
			log.Printf("Starting local HTTPS server (localhost only) on port %s...", port)
		} else {
			log.Printf("Starting local HTTPS server (all interfaces) on port %s...", port)
		}
		server.Addr = address + ":" + port

		log.Printf("Server listening on %s", server.Addr)
		// Start local HTTPS server using ListenAndServeTLS
		err = server.ListenAndServeTLS(certFile, keyFile)
	}

	// --- Handle Server Exit ---
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	} else {
		log.Println("Server stopped gracefully.")
	}
}
