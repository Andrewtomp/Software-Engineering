package usertable

import (
	"fmt"
	"front-runner/internal/coredbutils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	// Get the test database instance.
	db = coredbutils.GetDB()

	// Clear the database before running tests.
	if err := ClearUserTable(db); err != nil {
		fmt.Printf("failed to clear test database: %v\n", err)
		os.Exit(1)
	}

	// Run the tests.
	code := m.Run()

	// Optionally, clear the database after tests.
	if err := ClearUserTable(db); err != nil {
		fmt.Printf("failed to clear test database after tests: %v\n", err)
		os.Exit(1)
	}

	os.Exit(code)
}

func TestDirectUserEntry(t *testing.T) {

	t.Logf("db %s", db.Name())

	// Use the same password for all test users (for simplicity).
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("error hashing test password: %v", err)
	}

	testUsers := []User{
		{Email: "alice@example.com", PasswordHash: string(hashedPassword), BusinessName: "Alice Co"},
		{Email: "bob@example.com", PasswordHash: string(hashedPassword), BusinessName: "Bob LLC"},
		{Email: "charlie@example.com", PasswordHash: string(hashedPassword), BusinessName: "Charlie Inc"},
	}

	for _, user := range testUsers {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("error creating test user (%s): %v", user.Email, err)
		}
		t.Logf("User %s created successfuly", user.Email)
	}

	// Optionally verify that users are in the database.
	for _, expected := range testUsers {
		var user User
		if err := db.First(&user, "email = ?", expected.Email).Error; err != nil {
			t.Fatalf("failed to find user %s: %v", expected.Email, err)
		}
		t.Logf("Found user %s !", user.Email)
		// Additional assertions can be added here.
	}

	// Ensure cleanup is properly handled.
	if err := ClearUserTable(db); err != nil {
		t.Fatalf("failed to clear database: %v", err)
	}
}

// TestRegisterUser checks that registering a new user works as expected.
func TestRegisterUser(t *testing.T) {
	// Prepare form data using the correct keys.
	form := url.Values{}
	form.Add("email", "testuser@example.com")
	form.Add("password", "testpassword")
	form.Add("business_name", "TestBusiness")

	// Create a POST request with the form data.
	req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use a ResponseRecorder to record the response.
	rec := httptest.NewRecorder()

	// Call the RegisterUser handler.
	RegisterUser(rec, req)

	// Check the status code.
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	// Check that the response contains the success message.
	if !strings.Contains(rec.Body.String(), "User registered successfully") {
		t.Errorf("unexpected response body: %q", rec.Body.String())
	}

	// Ensure cleanup is properly handled.
	if err := ClearUserTable(db); err != nil {
		t.Fatalf("failed to clear database: %v", err)
	}
}

// TestRegisterUserEmptyFields verifies that missing form values result in an error.
func TestRegisterUserEmptyFields(t *testing.T) {
	// Create a request with no form data.
	req := httptest.NewRequest("POST", "/register", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	RegisterUser(rec, req)

	// Since email and password are missing, we expect a 400 Bad Request.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d; got %d", http.StatusBadRequest, rec.Code)
	}
}
