// internal/login/login_test.go
package login

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

// DSN for your PostgreSQL database. Adjust these values as needed.
const DSN = "host=localhost port=5432 user=johnny dbname=users sslmode=disable TimeZone=UTC"

// GetTestDB opens a connection to the PostgreSQL database and auto-migrates the schema.
func GetTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Auto-migrate the schema for the login.User model.
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	return db, nil
}

// CreateTestUsers inserts sample test users into the database.
func CreateTestUsers(db *gorm.DB) error {
	// Use the same password for all test users (for simplicity).
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing test password: %w", err)
	}

	testUsers := []User{
		{Email: "alice@example.com", PasswordHash: string(hashedPassword), BusinessName: "Alice Co"},
		{Email: "bob@example.com", PasswordHash: string(hashedPassword), BusinessName: "Bob LLC"},
		{Email: "charlie@example.com", PasswordHash: string(hashedPassword), BusinessName: "Charlie Inc"},
	}

	for _, user := range testUsers {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("error creating test user (%s): %w", user.Email, err)
		}
	}

	return nil
}

// ClearDatabase deletes all records from the users table.
// The AllowGlobalUpdate flag is required for global deletes in GORM v2.
func ClearDatabase(db *gorm.DB) error {
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
		return fmt.Errorf("error clearing users table: %w", err)
	}
	return nil
}

// TestMain sets up the test database environment before tests run.
func TestMain(m *testing.M) {
	var err error
	// Get the test database instance.
	testDB, err = GetTestDB()
	if err != nil {
		fmt.Printf("failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Override the global db in the login package with our test DB.
	db = testDB

	// Clear the database before running tests.
	if err := ClearDatabase(testDB); err != nil {
		fmt.Printf("failed to clear test database: %v\n", err)
		os.Exit(1)
	}

	// Run the tests.
	code := m.Run()

	// Optionally, clear the database after tests.
	if err := ClearDatabase(testDB); err != nil {
		fmt.Printf("failed to clear test database after tests: %v\n", err)
		os.Exit(1)
	}

	os.Exit(code)
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

// TestLoginUser checks that logging in with valid credentials works.
func TestLoginUser(t *testing.T) {
	// First, register a user to log in.
	form := url.Values{}
	form.Add("email", "loginuser@example.com")
	form.Add("password", "loginpassword")
	form.Add("business_name", "LoginBusiness")

	reqRegister := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	reqRegister.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recRegister := httptest.NewRecorder()
	RegisterUser(recRegister, reqRegister)
	if recRegister.Code != http.StatusOK {
		t.Fatalf("failed to register user: %v", recRegister.Body.String())
	}

	// Now, attempt to log in with the registered credentials.
	reqLogin := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	reqLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recLogin := httptest.NewRecorder()
	LoginUser(recLogin, reqLogin)

	if recLogin.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, recLogin.Code)
	}

	// Check that the response contains "Logged in successfully".
	body := recLogin.Body.String()
	if !strings.Contains(body, "Logged in successfully") {
		t.Errorf("unexpected response body: %q", body)
	}

	// Check that a session cookie is set.
	cookies := recLogin.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected session cookie 'auth' to be set")
	}
}

// TestLoginUserInvalid checks that an invalid login attempt returns an error.
func TestLoginUserInvalid(t *testing.T) {
	form := url.Values{}
	form.Add("email", "nonexistent@example.com")
	form.Add("password", "badpassword")
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	LoginUser(rec, req)

	// Expect 401 Unauthorized.
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d; got %d", http.StatusUnauthorized, rec.Code)
	}
}

func createLogoutTestUser(t *testing.T) {
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "secret")
	form.Add("business_name", "TestBusiness")

	req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	RegisterUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("failed to register test user: %s", rr.Body.String())
	}
}

// TestLogoutUser verifies that logging out clears the session.
func TestLogoutUser(t *testing.T) {
	// Creating a test user for logout test
	createLogoutTestUser(t)

	// Step 1: Simulate a valid login to generate a session cookie.
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "secret")
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	LoginUser(rr, req)

	// Log all headers
	for key, values := range rr.Header() {
		for _, value := range values {
			t.Logf("Header %s: %s", key, value)
		}
	}

	// Extract the session cookie from the login response.
	resp := rr.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected at least one cookie, but got none")
	}
	for _, c := range cookies {
		t.Logf("Cookie: %s = %s", c.Name, c.Value)
	}

	cookie := rr.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("Expected session cookie, but got none")
	}

	// Step 2: Use the valid session cookie for the logout request.
	logoutReq := httptest.NewRequest("GET", "/logout", nil)
	logoutReq.Header.Set("Cookie", cookie)

	logoutRR := httptest.NewRecorder()
	LogoutUser(logoutRR, logoutReq)

	// Now you can assert that logout was successful.
	if logoutRR.Body.String() != "Logged out successfully" {
		t.Errorf("Unexpected logout response: %s", logoutRR.Body.String())
	}
}
