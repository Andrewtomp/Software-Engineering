package coredbutils

import "testing"

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
