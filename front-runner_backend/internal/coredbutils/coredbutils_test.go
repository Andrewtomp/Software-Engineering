package coredbutils

import (
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/joho/godotenv"
)

const projectDirName = "front-runner_backend"

func init() {
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))

	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		log.Fatalf("Problem loading .env file. cwd:%s; cause: %s", cwd, err)
	}
	LoadEnv()
}

// TestGetDB tests the GetDB function for initializing a database connection and
// verifies that the connection can be pinged successfully.
//
// @Summary      Test database connection initialization
// @Description  Calls GetDB to obtain a DB instance, retrieves the underlying sql.DB, and attempts to ping the database to ensure the connection is valid.
//
// @Tags         database, testing
func TestGetDB(t *testing.T) {
	// Call the function to get the DB connection.
	db := GetDB()
	if db == nil {
		t.Fatal("expected non-nil DB, got nil")
	}

	// Retrieve the underlying sql.DB to perform a ping.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	// Attempt to ping the database.
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}
