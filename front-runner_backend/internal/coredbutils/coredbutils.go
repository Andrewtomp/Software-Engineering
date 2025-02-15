package coredbutils

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetDB initializes and returns a connection to the PostgreSQL database.
//
// @Summary      Initialize database connection
// @Description  Build the DSN and connect to the PostgreSQL database using GORM. Returns a *gorm.DB instance that can be used by internal packages to interact with the database.
//
// @Tags         database
func GetDB() *gorm.DB {
	// Build the DSN (Data Source Name) for PostgreSQL.
	// Adjust the parameters (host, user, password, dbname, port, sslmode, TimeZone) as needed.
	dsn := "host=localhost port=5432 user=johnny dbname=frontrunner sslmode=disable TimeZone=UTC"

	var err error

	// Open a connection to the PostgreSQL database using GORM.
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	return db
}
