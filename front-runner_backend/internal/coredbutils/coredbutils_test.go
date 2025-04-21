package coredbutils

import (
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"  // Using testify for assertions
	"github.com/stretchr/testify/require" // Using require for fatal checks
)

const projectDirName = "front-runner_backend" // Adjust if needed

// setup loads .env for integration tests. Runs once for the package.
var setupOnce sync.Once

func setup(t *testing.T) {
	setupOnce.Do(func() {
		re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
		cwd, _ := os.Getwd()
		rootPath := re.Find([]byte(cwd))
		if rootPath == nil {
			t.Fatalf("Could not find project root directory '%s' from '%s'", projectDirName, cwd)
		}

		envPath := string(rootPath) + `/.env`
		err := godotenv.Load(envPath)
		// Don't fatal if .env is missing, integration test will fail later if needed
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Problem loading .env file from %s: %v", envPath, err)
		} else if err == nil {
			log.Printf("Loaded environment variables from %s for integration tests", envPath)
		}
	})
}

// TestIsLocalHost tests the isLocalHost helper function.
func TestIsLocalHost(t *testing.T) {
	testCases := []struct {
		host     string
		expected bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"192.168.1.1", false},
		{"example.com", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.host, func(t *testing.T) {
			assert.Equal(t, tc.expected, isLocalHost(tc.host))
		})
	}
}

// TestLoadEnv tests the LoadEnv function under various conditions.
// Assumes LoadEnv has been refactored to return errors.
func TestLoadEnv(t *testing.T) {
	// Helper to set env vars and reset state for each subtest
	runTestWithEnv := func(t *testing.T, env map[string]string, testFunc func(t *testing.T)) {
		ResetDBStateForTests() // Reset state before setting env
		originalEnv := map[string]string{}
		for k, v := range env {
			originalEnv[k] = os.Getenv(k) // Store original value
			t.Setenv(k, v)                // Set test value
		}

		testFunc(t) // Run the actual test logic

		// Restore original environment variables (optional, t.Setenv handles cleanup)
		// for k, v := range originalEnv {
		// 	os.Setenv(k, v)
		// }
		ResetDBStateForTests() // Reset state after test
	}

	// --- Test Cases ---

	t.Run("Success_Local_NoPassword", func(t *testing.T) {
		env := map[string]string{
			"DB_HOST": "localhost",
			"DB_PORT": "5432",
			"DB_NAME": "testdb",
			"DB_USER": "testuser",
			// "DB_PASSWORD": "", // Explicitly missing
		}
		runTestWithEnv(t, env, func(t *testing.T) {
			err := LoadEnv()
			assert.NoError(t, err)
			assert.NotEmpty(t, dsn) // Internal check: DSN should be set
			assert.NotContains(t, dsn, "password=")
		})
	})

	t.Run("Success_Local_WithPassword", func(t *testing.T) {
		env := map[string]string{
			"DB_HOST":     "127.0.0.1",
			"DB_PORT":     "5432",
			"DB_NAME":     "testdb",
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpassword",
		}
		runTestWithEnv(t, env, func(t *testing.T) {
			err := LoadEnv()
			assert.NoError(t, err)
			assert.NotEmpty(t, dsn)
			assert.Contains(t, dsn, "password=testpassword")
		})
	})

	t.Run("Success_Remote_WithPassword", func(t *testing.T) {
		env := map[string]string{
			"DB_HOST":     "remote.host.com",
			"DB_PORT":     "5432",
			"DB_NAME":     "testdb",
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpassword",
		}
		runTestWithEnv(t, env, func(t *testing.T) {
			err := LoadEnv()
			assert.NoError(t, err)
			assert.NotEmpty(t, dsn)
			assert.Contains(t, dsn, "password=testpassword")
		})
	})

	t.Run("Failure_Remote_NoPassword", func(t *testing.T) {
		env := map[string]string{
			"DB_HOST": "remote.host.com",
			"DB_PORT": "5432",
			"DB_NAME": "testdb",
			"DB_USER": "testuser",
			// "DB_PASSWORD": "", // Explicitly missing
		}
		runTestWithEnv(t, env, func(t *testing.T) {
			err := LoadEnv()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "DB_PASSWORD environment variable must be set")
			assert.Empty(t, dsn) // DSN should not be set on error
		})
	})

	// Test missing required variables
	requiredVars := []string{"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER"}
	baseEnv := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_NAME":     "testdb",
		"DB_USER":     "testuser",
		"DB_PASSWORD": "pw", // Include password to avoid that error
	}

	for _, missingVar := range requiredVars {
		t.Run("Failure_Missing_"+missingVar, func(t *testing.T) {
			testEnv := make(map[string]string)
			for k, v := range baseEnv { // Copy base env
				testEnv[k] = v
			}
			delete(testEnv, missingVar) // Remove the variable to test

			runTestWithEnv(t, testEnv, func(t *testing.T) {
				err := LoadEnv()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), missingVar+" environment variable is required")
				assert.Empty(t, dsn)
			})
		})
	}

	t.Run("LoadEnv_RunsOnce", func(t *testing.T) {
		ResetDBStateForTests()
		env := map[string]string{
			"DB_HOST": "localhost", "DB_PORT": "5432", "DB_NAME": "db1", "DB_USER": "user1",
		}
		for k, v := range env {
			t.Setenv(k, v)
		}
		err1 := LoadEnv() // First call
		require.NoError(t, err1)
		dsn1 := dsn

		// Change env var and call again - DSN should NOT change due to sync.Once
		t.Setenv("DB_NAME", "db2")
		err2 := LoadEnv()
		require.NoError(t, err2) // Should still return nil error from first run
		dsn2 := dsn

		assert.Equal(t, dsn1, dsn2)
		assert.Contains(t, dsn1, "dbname=db1") // Ensure it used the first value
		ResetDBStateForTests()
	})
}

// TestGetDB_Integration performs an integration test requiring a live database
// based on the .env file configuration.
// Assumes GetDB has been refactored to return errors.
func TestGetDB_Integration(t *testing.T) {
	setup(t) // Load .env file if available

	// Reset state in case other tests ran LoadEnv with different settings
	ResetDBStateForTests()

	// Call GetDB - this will implicitly call the refactored LoadEnv
	dbInstance, err := GetDB()

	// Check if LoadEnv failed based on .env content
	if err != nil && strings.Contains(err.Error(), "failed to load environment") {
		t.Skipf("Skipping DB integration test: Failed to load environment from .env: %v", err)
	}
	// Check if connection failed
	if err != nil && strings.Contains(err.Error(), "failed to connect to database") {
		t.Skipf("Skipping DB integration test: Failed to connect to database specified in .env: %v", err)
	}
	// Any other error from GetDB
	require.NoError(t, err, "GetDB() failed unexpectedly during integration test")
	require.NotNil(t, dbInstance, "GetDB() returned nil DB instance unexpectedly")

	// Retrieve the underlying sql.DB to perform a ping.
	sqlDB, err := dbInstance.DB()
	require.NoError(t, err, "Failed to get underlying sql.DB")
	require.NotNil(t, sqlDB, "Underlying sql.DB is nil")

	// Attempt to ping the database.
	err = sqlDB.Ping()
	assert.NoError(t, err, "Failed to ping database using connection from .env")

	ResetDBStateForTests() // Clean up state
}

// TestGetDB_Singleton verifies that GetDB returns the same instance on multiple calls.
// Assumes GetDB has been refactored to return errors.
func TestGetDB_Singleton(t *testing.T) {
	setup(t) // Load .env for a potentially successful connection
	ResetDBStateForTests()

	// First call
	db1, err1 := GetDB()
	if err1 != nil {
		t.Skipf("Skipping singleton test: Initial GetDB() failed: %v", err1)
	}
	require.NotNil(t, db1)

	// Second call
	db2, err2 := GetDB()
	require.NoError(t, err2, "Second call to GetDB() returned an unexpected error")
	require.NotNil(t, db2)

	// Check if the instances are the same (pointer comparison)
	assert.Same(t, db1, db2, "GetDB() should return the same instance on subsequent calls")

	ResetDBStateForTests()
}

// TestGetDB_Errors tests error conditions for GetDB.
// Assumes GetDB has been refactored to return errors.
func TestGetDB_Errors(t *testing.T) {

	t.Run("Failure_Before_LoadEnv_Success", func(t *testing.T) {
		// This test relies on the internal check within the refactored GetDB
		// It ensures GetDB calls LoadEnv and handles its error.
		t.Log("Calling ResetDBStateForTests()")
		ResetDBStateForTests()
		t.Log("Finished ResetDBStateForTests()")
		// Set env vars that cause LoadEnv to fail
		t.Log("Setting environment variables...")
		t.Setenv("DB_HOST", "remote.host")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_NAME", "remotedb")
		t.Setenv("DB_USER", "remoteuser")
		t.Log("Finished setting environment variables.")
		// Missing DB_PASSWORD for remote host

		t.Log("Calling GetDB()...")
		dbInstance, err := GetDB()
		t.Logf("GetDB() returned. Error: %v, Instance: %p", err, dbInstance)

		t.Log("Starting assertions...")
		assert.Error(t, err)
		assert.Nil(t, dbInstance)
		assert.Contains(t, err.Error(), "failed to load environment") // Check error comes from LoadEnv
		assert.Contains(t, err.Error(), "DB_PASSWORD environment variable must be set")
		t.Log("Finished assertions.")

		t.Log("Calling final ResetDBStateForTests()")
		ResetDBStateForTests()
		t.Log("--- Exiting Failure_Before_LoadEnv_Success ---") // Log exit
	})

	t.Run("Failure_Connection_Error", func(t *testing.T) {
		ResetDBStateForTests()
		// Set env vars that LoadEnv will accept, but lead to connection failure
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "9999") // Assuming nothing runs on this port
		t.Setenv("DB_NAME", "fakedb")
		t.Setenv("DB_USER", "fakeuser")
		t.Setenv("DB_PASSWORD", "fakepw")
		t.Setenv("DB_EXTRA_PARAMS_FOR_TEST", "connect_timeout=1")

		dbInstance, err := GetDB()
		assert.Error(t, err)
		assert.Nil(t, dbInstance)
		assert.Contains(t, err.Error(), "failed to connect to database") // Check error comes from gorm.Open

		ResetDBStateForTests()
	})
}
