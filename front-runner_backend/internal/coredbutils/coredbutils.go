package coredbutils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db            *gorm.DB = nil
	dsn           string
	dbLoadEnvOnce sync.Once
	dbConnOnce    sync.Once
)

func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1"
}

// Grabs the relevant enviroment variables and constructs the DSN for the postgres database
func LoadEnv() {
	dbLoadEnvOnce.Do(func() {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		name := os.Getenv("DB_NAME")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")

		if host == "" {
			log.Fatal("No provided DB_HOST value.")
		}

		if port == "" {
			log.Fatal("No provided DB_PORT value.")
		}

		if name == "" {
			log.Fatal("No provided DB_NAME value.")
		}

		if user == "" {
			log.Fatal("No provided DB_USER value.")
		}

		local := isLocalHost(host)

		if password == "" && !local {
			// Require password *only* if not connecting to localhost/127.0.0.1
			log.Fatal("DB_PASSWORD environment variable must be set for non-local database connections.")
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
	})
}

// GetDB initializes and returns a connection to the PostgreSQL database.
//
// @Summary      Initialize database connection
// @Description  Build the DSN and connect to the PostgreSQL database using GORM. Returns a *gorm.DB instance that can be used by internal packages to interact with the database.
//
// @Tags         database
func GetDB() *gorm.DB {
	dbConnOnce.Do(func() {
		var err error

		// Open a connection to the PostgreSQL database using GORM.
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
	})

	return db
}
