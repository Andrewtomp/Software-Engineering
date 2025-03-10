package coredbutils

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	dsn string
)

// Grabs the relevant enviroment variables and constructs the DSN for the postgres database
func LoadEnv() {
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

	if password == "" {
		log.Fatal("No provided DB_PASSWORD value.")
	}

	// Build the DSN (Data Source Name) for PostgreSQL.
	// Adjust the parameters (host, user, password, dbname, port, sslmode, TimeZone) as needed.
	dsn = fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, name)
}

// GetDB initializes and returns a connection to the PostgreSQL database.
//
// @Summary      Initialize database connection
// @Description  Build the DSN and connect to the PostgreSQL database using GORM. Returns a *gorm.DB instance that can be used by internal packages to interact with the database.
//
// @Tags         database
func GetDB() *gorm.DB {

	var err error

	// Open a connection to the PostgreSQL database using GORM.
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	return db
}
