package coredbutils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db            *gorm.DB  = nil // Global database connection pool instance.
	dsn           string          // Data Source Name constructed from environment variables.
	dbLoadEnvOnce sync.Once       // Ensures environment variables for DSN are loaded only once.
	dbConnOnce    sync.Once       // Ensures database connection is established only once.
	loadEnvErr    error           // Store error from LoadEnv attempt
	getDBErr      error           // Store error from GetDB attempt
)

// isLocalHost checks if the provided host string refers to localhost.
func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1"
}

// LoadEnv reads database connection details (host, port, name, user, password)
// from environment variables, constructs the Data Source Name (DSN) string,
// and stores it in the package-level 'dsn' variable.
// It uses sync.Once to ensure this operation happens only once per application run.
// It enforces password requirement for non-local connections.
func LoadEnv() error {
	dbLoadEnvOnce.Do(func() {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		name := os.Getenv("DB_NAME")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")

		if host == "" {
			loadEnvErr = errors.New("DB_HOST environment variable is required")
			return
		}
		if port == "" {
			loadEnvErr = errors.New("DB_PORT environment variable is required")
			return
		}
		if name == "" {
			loadEnvErr = errors.New("DB_NAME environment variable is required")
			return
		}
		if user == "" {
			loadEnvErr = errors.New("DB_USER environment variable is required")
			return
		}

		local := isLocalHost(host)

		if password == "" && !local {
			// Require password *only* if not connecting to localhost/127.0.0.1
			loadEnvErr = errors.New("DB_PASSWORD environment variable must be set for non-local database connections")
			return
		}

		dsnParts := []string{
			fmt.Sprintf("host=%s", host),
			fmt.Sprintf("port=%s", port),
			fmt.Sprintf("user=%s", user),
			fmt.Sprintf("dbname=%s", name),
			"sslmode=disable", // Adjust sslmode as needed for your setup
			"TimeZone=UTC",
		}

		// Conditionally add the password part
		if password != "" {
			dsnParts = append(dsnParts, fmt.Sprintf("password=%s", password))
		}
		// If password is empty, we simply don't add the password part.
		// The database driver will attempt connection without a password.

		// Join the parts into the final DSN string
		dsn = strings.Join(dsnParts, " ")
		log.Println("Database DSN constructed.")
	})
	return loadEnvErr
}

// GetDB establishes a connection to the PostgreSQL database using the DSN
// constructed by LoadEnv. It uses FROM as the ORM.
// It leverages sync.Once to ensure the connection is established only once
// and returns the singleton *grom.DB instance.
// It's crucial to call LoadEnv before calling GetDB for the firtst time.
// This function will fatal on connection errors.
func GetDB() (*gorm.DB, error) {
	// Ensure LoadEnv has run and check its potential error first
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment for DB: %w", err)
	}
	// If LoadEnv succeeded but DSN is somehow still empty (shouldn't happen with current logic)
	if dsn == "" && loadEnvErr == nil {
		return nil, errors.New("database DSN is not initialized despite LoadEnv success")
	}

	dbConnOnce.Do(func() {
		var err error
		// Open connection
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			// Log the DSN without password for security
			safeDSN := dsn
			if strings.Contains(safeDSN, "password=") {
				safeDSN = strings.Split(safeDSN, " password=")[0] + " password=***"
			}
			log.Printf("failed to connect to database using DSN '%s': %v", safeDSN, err)
			getDBErr = fmt.Errorf("failed to connect to database: %w", err) // Store the connection error
		} else {
			log.Println("Database connection established successfully.")
		}
	})

	if getDBErr != nil {
		return nil, getDBErr // Return stored connection error
	}
	if db == nil && getDBErr == nil {
		// Safeguard: Should not happen if Do ran without error
		return nil, errors.New("database connection is nil after initialization attempt without error")
	}
	return db, nil // Return the singleton instance and nil error
}

// ResetDBStateForTests resets the internal state (sync.Once, variables)
// This is specifically for allowing tests to run LoadEnv/GetDB multiple times
// with different environment variables or scenarios. **Do not use in production code.**
func ResetDBStateForTests() {
	dbLoadEnvOnce = sync.Once{}
	dbConnOnce = sync.Once{}
	dsn = ""
	db = nil
	loadEnvErr = nil
	getDBErr = nil
	log.Println("--- coredbutils state reset for testing ---")
}
