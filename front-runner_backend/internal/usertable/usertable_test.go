package usertable

import (
	"errors"
	"fmt"
	"front-runner/internal/coredbutils"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const projectDirName = "front-runner_backend" // Adjust if your project dir name is different

// Global db instance for tests
var testDB *gorm.DB

// init loads environment variables and sets up the database connection once.
func init() {
	// Navigate up to the project root to find the .env file
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	rootPath := re.FindString(cwd)
	if rootPath == "" {
		log.Fatalf("Could not find project root directory '%s' from cwd '%s'", projectDirName, cwd)
	}

	envPath := rootPath + `/.env`
	err = godotenv.Load(envPath)
	if err != nil {
		// It's okay if .env doesn't exist in CI/CD, variables might be set directly
		log.Printf("Warning: Could not load .env file from %s: %v. Assuming env vars are set.", envPath, err)
	}

	// Use coredbutils to load env vars it needs (like DB DSN)
	coredbutils.LoadEnv()
	// Use the package's Setup function
	Setup()
	// Assign the global db instance from the package to our testDB variable
	testDB = db
	if testDB == nil {
		log.Fatal("Database connection failed in init")
	}
	// Ensure migration runs for the test database
	MigrateUserDB()
}

// TestMain manages the test database state, clearing it before and after tests.
func TestMain(m *testing.M) {
	if testDB == nil {
		fmt.Println("Test database connection is nil, cannot proceed.")
		os.Exit(1)
	}

	// Clear the database before running tests.
	fmt.Println("Clearing test database before tests...")
	if err := ClearUserTable(testDB); err != nil {
		fmt.Printf("Failed to clear test database before tests: %v\n", err)
		os.Exit(1)
	}

	// Run the tests.
	code := m.Run()

	// Clear the database after tests.
	fmt.Println("Clearing test database after tests...")
	if err := ClearUserTable(testDB); err != nil {
		fmt.Printf("Failed to clear test database after tests: %v\n", err)
		// Don't exit fatally here, just report the error
	}

	os.Exit(code)
}

// Helper function to create a user for testing purposes
func createTestUser(t *testing.T, email, password, name, businessName, provider, providerID string) *User {
	t.Helper() // Marks this as a test helper
	var hash string
	var err error
	if provider == "local" && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password for %s: %v", email, err)
		}
		hash = string(hashedPassword)
	}

	user := &User{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		BusinessName: businessName,
		Provider:     provider,
		ProviderID:   providerID,
	}

	if err = CreateUser(user); err != nil { // Use the package's CreateUser function
		t.Fatalf("Failed to create test user %s: %v", email, err)
	}
	if user.ID == 0 {
		t.Fatalf("Created user %s but ID is zero", email)
	}
	return user
}

// TestCreateAndGetUser tests the basic CreateUser and GetUserByEmail functions.
func TestCreateAndGetUser(t *testing.T) {
	email := "direct_entry@example.com"
	password := "password123"
	name := "Direct Entry User"
	businessName := "Direct Co"

	// Create user using the package function
	createdUser := createTestUser(t, email, password, name, businessName, "local", "")

	// Retrieve user using the package function
	retrievedUser, err := GetUserByEmail(email)
	if err != nil {
		t.Fatalf("Error retrieving user %s: %v", email, err)
	}
	if retrievedUser == nil {
		t.Fatalf("Failed to find user %s after creation", email)
	}

	// Verify details
	if retrievedUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, retrievedUser.ID)
	}
	if retrievedUser.Email != email {
		t.Errorf("Expected user email %s, got %s", email, retrievedUser.Email)
	}
	if retrievedUser.Name != name {
		t.Errorf("Expected user name %s, got %s", name, retrievedUser.Name)
	}
	if retrievedUser.BusinessName != businessName {
		t.Errorf("Expected business name %s, got %s", businessName, retrievedUser.BusinessName)
	}
	if retrievedUser.Provider != "local" {
		t.Errorf("Expected provider 'local', got %s", retrievedUser.Provider)
	}
}

// TestRegisterUserHandler tests the RegisterUser HTTP handler thoroughly.
func TestRegisterUserHandler(t *testing.T) {
	// --- Subtest: Successful Registration ---
	t.Run("success", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "register_success@example.com")
		form.Add("password", "testpassword")
		form.Add("name", "Register Success User")     // Added name field
		form.Add("businessName", "TestBusiness Inc.") // Corrected field name

		req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		RegisterUser(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d; got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "User registered successfully") {
			t.Errorf("Unexpected response body: %q", rec.Body.String())
		}

		// Verify user exists in DB
		user, err := GetUserByEmail("register_success@example.com")
		if err != nil || user == nil {
			t.Fatalf("Failed to find user register_success@example.com in DB after registration: %v", err)
		}
		if user.Name != "Register Success User" {
			t.Errorf("Expected name 'Register Success User', got '%s'", user.Name)
		}
		if user.BusinessName != "TestBusiness Inc." {
			t.Errorf("Expected business name 'TestBusiness Inc.', got '%s'", user.BusinessName)
		}
	})

	// --- Subtest: Missing Email/Password ---
	t.Run("missing_fields", func(t *testing.T) {
		form := url.Values{}
		// form.Add("email", "missing@example.com") // Missing email
		form.Add("password", "testpassword")
		form.Add("name", "Missing Email")

		req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		RegisterUser(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing email; got %d", http.StatusBadRequest, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Email and password are required") {
			t.Errorf("Expected error message for missing fields, got: %q", rec.Body.String())
		}
	})

	// --- Subtest: Invalid Email Format ---
	t.Run("invalid_email", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "invalid-email-format")
		form.Add("password", "testpassword")
		form.Add("name", "Invalid Email User")

		req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		RegisterUser(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid email; got %d", http.StatusBadRequest, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Invalid email format") {
			t.Errorf("Expected error message for invalid email, got: %q", rec.Body.String())
		}
	})

	// --- Subtest: Duplicate Email ---
	t.Run("duplicate_email", func(t *testing.T) {
		// First, ensure the user exists (create if necessary, though 'success' test might have)
		_ = createTestUser(t, "duplicate@example.com", "pass1", "Dupe User", "Dupe Co", "local", "")

		// Now, try to register again with the same email
		form := url.Values{}
		form.Add("email", "duplicate@example.com")
		form.Add("password", "pass2")
		form.Add("name", "Another Dupe User")

		req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		RegisterUser(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("Expected status %d for duplicate email; got %d", http.StatusConflict, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Email already in use") {
			t.Errorf("Expected error message for duplicate email, got: %q", rec.Body.String())
		}
	})
}

// TestCreateUserFunction tests the CreateUser function directly.
func TestCreateUserFunction(t *testing.T) {
	validHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	testCases := []struct {
		name        string
		user        *User
		expectError bool
		errorMsg    string // Substring to check in error message
	}{
		{
			name:        "success_local",
			user:        &User{Email: "create_local@example.com", PasswordHash: string(validHash), Name: "Local Create", Provider: "local"},
			expectError: false,
		},
		{
			name:        "success_oauth",
			user:        &User{Email: "create_oauth@example.com", Name: "OAuth Create", Provider: "google", ProviderID: "google123"},
			expectError: false,
		},
		{
			name:        "fail_missing_email",
			user:        &User{PasswordHash: string(validHash), Name: "No Email", Provider: "local"},
			expectError: true,
			errorMsg:    "invalid or missing email",
		},
		{
			name:        "fail_invalid_email",
			user:        &User{Email: "bad-email", PasswordHash: string(validHash), Name: "Bad Email", Provider: "local"},
			expectError: true,
			errorMsg:    "invalid or missing email",
		},
		{
			name:        "fail_missing_provider",
			user:        &User{Email: "no_provider@example.com", PasswordHash: string(validHash), Name: "No Provider"},
			expectError: true,
			errorMsg:    "provider cannot be empty",
		},
		{
			name:        "fail_missing_provider_id_oauth",
			user:        &User{Email: "oauth_no_id@example.com", Name: "OAuth No ID", Provider: "github"}, // Missing ProviderID
			expectError: true,
			errorMsg:    "provider ID is required",
		},
		{
			name:        "fail_missing_password_local",
			user:        &User{Email: "local_no_pass@example.com", Name: "Local No Pass", Provider: "local"}, // Missing PasswordHash
			expectError: true,
			errorMsg:    "password hash is required",
		},
		// Duplicate email test requires pre-existing user, handled separately or rely on DB constraints
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CreateUser(tc.user)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.errorMsg != "" && !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error message containing %q, got: %v", tc.errorMsg, err)
				}
				// Clean up potentially partially created invalid user if necessary, though CreateUser should prevent it
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				// Verify user exists if creation was expected to succeed
				var foundUser User
				dbErr := testDB.Where("email = ?", tc.user.Email).First(&foundUser).Error
				if dbErr != nil {
					t.Fatalf("Failed to find user %s after successful CreateUser call: %v", tc.user.Email, dbErr)
				}
				if foundUser.ID == 0 {
					t.Errorf("User %s created but ID is zero", tc.user.Email)
				}
			}
		})
	}

	// Test duplicate email constraint specifically
	t.Run("fail_duplicate_email", func(t *testing.T) {
		email := "duplicate_create@example.com"
		user1 := &User{Email: email, PasswordHash: string(validHash), Name: "Dupe 1", Provider: "local"}
		user2 := &User{Email: email, PasswordHash: string(validHash), Name: "Dupe 2", Provider: "local"}

		err1 := CreateUser(user1)
		if err1 != nil {
			t.Fatalf("Failed to create first user for duplicate test: %v", err1)
		}

		err2 := CreateUser(user2)
		if err2 == nil {
			t.Errorf("Expected an error when creating user with duplicate email, but got nil")
		} else if !strings.Contains(err2.Error(), "database error creating user") { // Check for the wrapped error
			// Note: The exact underlying DB error message (like "UNIQUE constraint failed") might vary.
			// Checking for the wrapped message from CreateUser is more robust.
			t.Errorf("Expected database error for duplicate email, got: %v", err2)
		}
	})
}

// TestGetUserByID tests retrieving users by their primary key ID.
func TestGetUserByID(t *testing.T) {
	// Create a user to test with
	user := createTestUser(t, "getbyid@example.com", "pass", "Get By ID", "ID Co", "local", "")
	if user.ID == 0 {
		t.Fatal("Test user for GetUserByID has zero ID")
	}

	t.Run("found", func(t *testing.T) {
		foundUser, err := GetUserByID(user.ID)
		if err != nil {
			t.Fatalf("Error getting user by ID %d: %v", user.ID, err)
		}
		if foundUser == nil {
			t.Fatalf("GetUserByID(%d) returned nil, expected user", user.ID)
		}
		if foundUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		nonExistentID := uint(999999) // Assume this ID won't exist
		foundUser, err := GetUserByID(nonExistentID)
		if err != nil {
			// We expect gorm.ErrRecordNotFound, which GetUserByID should handle
			if !errors.Is(err, gorm.ErrRecordNotFound) && !strings.Contains(err.Error(), "database error fetching user by ID") {
				t.Fatalf("Expected RecordNotFound or wrapped error for non-existent ID %d, got: %v", nonExistentID, err)
			}
			// If it's the wrapped error, the user should still be nil
			if foundUser != nil {
				t.Errorf("Expected nil user when error occurred for non-existent ID %d, but got user", nonExistentID)
			}
		} else if foundUser != nil {
			// If no error, user must be nil
			t.Fatalf("Expected nil user for non-existent ID %d, but got a user", nonExistentID)
		}
		// If err is nil and foundUser is nil, this is the expected outcome (handled gorm.ErrRecordNotFound internally)
	})
}

// TestGetUserByProviderID tests retrieving users by OAuth provider and ID.
func TestGetUserByProviderID(t *testing.T) {
	provider := "testprovider"
	providerID := "testprovider123"
	email := "getbyprovider@example.com"
	// Create an OAuth-style user
	user := createTestUser(t, email, "", "Get By Provider", "Provider Co", provider, providerID)

	t.Run("found", func(t *testing.T) {
		foundUser, err := GetUserByProviderID(provider, providerID)
		if err != nil {
			t.Fatalf("Error getting user by provider %s/%s: %v", provider, providerID, err)
		}
		if foundUser == nil {
			t.Fatalf("GetUserByProviderID(%s, %s) returned nil, expected user", provider, providerID)
		}
		if foundUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
		}
		if foundUser.Provider != provider || foundUser.ProviderID != providerID {
			t.Errorf("Expected provider/ID %s/%s, got %s/%s", provider, providerID, foundUser.Provider, foundUser.ProviderID)
		}
	})

	t.Run("not_found_provider", func(t *testing.T) {
		foundUser, err := GetUserByProviderID("wrongprovider", providerID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { // Allow nil error or RecordNotFound
			t.Fatalf("Expected nil or RecordNotFound error for wrong provider, got: %v", err)
		}
		if foundUser != nil {
			t.Fatalf("Expected nil user for wrong provider, but got a user")
		}
	})

	t.Run("not_found_id", func(t *testing.T) {
		foundUser, err := GetUserByProviderID(provider, "wrongid")
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { // Allow nil error or RecordNotFound
			t.Fatalf("Expected nil or RecordNotFound error for wrong provider ID, got: %v", err)
		}
		if foundUser != nil {
			t.Fatalf("Expected nil user for wrong provider ID, but got a user")
		}
	})
}

// TestUpdateUser tests updating existing user records.
func TestUpdateUser(t *testing.T) {
	// Create a user to update
	user := createTestUser(t, "update@example.com", "pass", "Update Me", "Update Co", "local", "")
	originalID := user.ID
	originalEmail := user.Email // Email shouldn't change typically, but test other fields

	t.Run("success", func(t *testing.T) {
		// Modify fields
		user.Name = "Updated Name"
		user.BusinessName = "Updated Business Name"
		// You could potentially update other fields like ProviderID if logic allows

		err := UpdateUser(user)
		if err != nil {
			t.Fatalf("UpdateUser failed: %v", err)
		}

		// Retrieve the user again to verify changes
		updatedUser, getErr := GetUserByID(originalID)
		if getErr != nil || updatedUser == nil {
			t.Fatalf("Failed to retrieve user %d after update: %v", originalID, getErr)
		}

		if updatedUser.Name != "Updated Name" {
			t.Errorf("Expected updated name 'Updated Name', got '%s'", updatedUser.Name)
		}
		if updatedUser.BusinessName != "Updated Business Name" {
			t.Errorf("Expected updated business name 'Updated Business Name', got '%s'", updatedUser.BusinessName)
		}
		if updatedUser.Email != originalEmail { // Ensure email didn't change unexpectedly
			t.Errorf("Email changed during update, expected %s, got %s", originalEmail, updatedUser.Email)
		}
	})

	t.Run("fail_no_id", func(t *testing.T) {
		userWithoutID := &User{
			// ID: 0, // Explicitly zero or just omitted
			Email: "noid@example.com",
			Name:  "No ID User",
		}
		err := UpdateUser(userWithoutID)
		if err == nil {
			t.Errorf("Expected error when updating user without ID, but got nil")
		} else if !strings.Contains(err.Error(), "cannot update user without ID") {
			t.Errorf("Expected error message about missing ID, got: %v", err)
		}
	})

	t.Run("fail_update_nonexistent", func(t *testing.T) {
		nonExistentUser := &User{
			ID:    999998, // Assume this ID doesn't exist
			Email: "nonexistent@example.com",
			Name:  "Non Existent",
		}

		var checkUser User
		checkErr := testDB.First(&checkUser, nonExistentUser.ID).Error
		if checkErr == nil {
			t.Logf("DEBUG: User with ID %d unexpectedly exists before update attempt!", nonExistentUser.ID)
		} else if !errors.Is(checkErr, gorm.ErrRecordNotFound) {
			t.Logf("DEBUG: Error checking for user %d before update: %v", nonExistentUser.ID, checkErr)
		} else {
			t.Logf("DEBUG: Confirmed user with ID %d does not exist before update attempt.", nonExistentUser.ID)
		}

		// GORM's Save behaves like "update or insert" if the primary key isn't found,
		// but our UpdateUser function uses Save which requires the record to exist
		// based on the primary key. If the record doesn't exist, Save returns ErrRecordNotFound.
		err := UpdateUser(nonExistentUser)
		if err == nil {
			t.Errorf("Expected an error when trying to update a non-existent user ID, but got nil")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) && !strings.Contains(err.Error(), "database error updating user") {
			// Check for either the raw GORM error or our wrapped error
			t.Errorf("Expected RecordNotFound or wrapped DB error for non-existent user update, got: %v", err)
		}
	})
}

// package usertable

// import (
// 	"fmt"
// 	"front-runner/internal/coredbutils"
// 	"log"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"os"
// 	"regexp"
// 	"strings"
// 	"testing"

// 	"github.com/joho/godotenv"
// 	"golang.org/x/crypto/bcrypt"
// )

// const projectDirName = "front-runner_backend"

// func init() {
// 	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
// 	cwd, _ := os.Getwd()
// 	rootPath := re.Find([]byte(cwd))

// 	err := godotenv.Load(string(rootPath) + `/.env`)
// 	if err != nil {
// 		log.Fatalf("Problem loading .env file. cwd:%s; cause: %s", cwd, err)
// 	}
// 	coredbutils.LoadEnv()
// 	Setup()
// }

// // TestMain sets up the test database environment for user table tests.
// //
// // @Summary      Setup user table test environment
// // @Description  Initializes the database connection and clears the user table before and after running tests.
// //
// // @Tags         testing, user, dbtable
// func TestMain(m *testing.M) {
// 	// Get the test database instance.
// 	db = coredbutils.GetDB()

// 	// Clear the database before running tests.
// 	if err := ClearUserTable(db); err != nil {
// 		fmt.Printf("failed to clear test database: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Run the tests.
// 	code := m.Run()

// 	// Optionally, clear the database after tests.
// 	if err := ClearUserTable(db); err != nil {
// 		fmt.Printf("failed to clear test database after tests: %v\n", err)
// 		os.Exit(1)
// 	}

// 	os.Exit(code)
// }

// // TestDirectUserEntry tests direct insertion of user records into the database.
// //
// // @Summary      Direct user entry test
// // @Description  Inserts a set of test users directly into the database, then verifies that each user can be retrieved successfully.
// //
// // @Tags         testing, user, dbtable
// func TestDirectUserEntry(t *testing.T) {

// 	t.Logf("db %s", db.Name())

// 	// Use the same password for all test users (for simplicity).
// 	password := "password123"
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		t.Fatalf("error hashing test password: %v", err)
// 	}

// 	testUsers := []User{
// 		{Email: "alice@example.com", PasswordHash: string(hashedPassword), BusinessName: "Alice Co"},
// 		{Email: "bob@example.com", PasswordHash: string(hashedPassword), BusinessName: "Bob LLC"},
// 		{Email: "charlie@example.com", PasswordHash: string(hashedPassword), BusinessName: "Charlie Inc"},
// 	}

// 	for _, user := range testUsers {
// 		if err := db.Create(&user).Error; err != nil {
// 			t.Fatalf("error creating test user (%s): %v", user.Email, err)
// 		}
// 		t.Logf("User %s created successfuly", user.Email)
// 	}

// 	// Optionally verify that users are in the database.
// 	for _, expected := range testUsers {
// 		var user User
// 		if err := db.First(&user, "email = ?", expected.Email).Error; err != nil {
// 			t.Fatalf("failed to find user %s: %v", expected.Email, err)
// 		}
// 		t.Logf("Found user %s !", user.Email)
// 		// Additional assertions can be added here.
// 	}

// 	// Ensure cleanup is properly handled.
// 	if err := ClearUserTable(db); err != nil {
// 		t.Fatalf("failed to clear database: %v", err)
// 	}
// }

// // TestRegisterUser tests the RegisterUser HTTP handler for successful user registration.
// //
// // @Summary      Test user registration
// // @Description  Simulates a POST request to the /register endpoint with valid form data and verifies that the user is registered successfully.
// //
// // @Tags         testing, user, dbtable
// func TestRegisterUser(t *testing.T) {
// 	// Prepare form data using the correct keys.
// 	form := url.Values{}
// 	form.Add("email", "testuser@example.com")
// 	form.Add("password", "testpassword")
// 	form.Add("business_name", "TestBusiness")

// 	// Create a POST request with the form data.
// 	req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 	// Use a ResponseRecorder to record the response.
// 	rec := httptest.NewRecorder()

// 	// Call the RegisterUser handler.
// 	RegisterUser(rec, req)

// 	// Check the status code.
// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("expected status %d; got %d", http.StatusOK, rec.Code)
// 	}

// 	// Check that the response contains the success message.
// 	if !strings.Contains(rec.Body.String(), "User registered successfully") {
// 		t.Errorf("unexpected response body: %q", rec.Body.String())
// 	}

// 	// Ensure cleanup is properly handled.
// 	if err := ClearUserTable(db); err != nil {
// 		t.Fatalf("failed to clear database: %v", err)
// 	}
// }

// // TestRegisterUserEmptyFields verifies that the registration endpoint returns an error when required fields are missing.
// //
// // @Summary      Test registration with empty fields
// // @Description  Simulates a POST request to /register without form data and expects a 400 Bad Request response due to missing email and password.
// //
// // @Tags         testing, user, dbtable
// func TestRegisterUserEmptyFields(t *testing.T) {
// 	// Create a request with no form data.
// 	req := httptest.NewRequest("POST", "/register", nil)
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	rec := httptest.NewRecorder()

// 	RegisterUser(rec, req)

// 	// Since email and password are missing, we expect a 400 Bad Request.
// 	if rec.Code != http.StatusBadRequest {
// 		t.Fatalf("expected status %d; got %d", http.StatusBadRequest, rec.Code)
// 	}
// }
