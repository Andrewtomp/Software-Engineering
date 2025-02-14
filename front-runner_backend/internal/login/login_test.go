// internal/login/login_test.go
package login

import (
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/usertable"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

// TestMain sets up the test database environment before tests run.
func TestMain(m *testing.M) {
	// Get the test database instance.
	db = coredbutils.GetDB()

	// Clear the database before running tests.
	if err := usertable.ClearUserTable(db); err != nil {
		fmt.Printf("failed to clear test database: %v\n", err)
		os.Exit(1)
	}

	// Run the tests.
	code := m.Run()

	// Optionally, clear the database after tests.
	if err := usertable.ClearUserTable(db); err != nil {
		fmt.Printf("failed to clear test database after tests: %v\n", err)
		os.Exit(1)
	}

	os.Exit(code)
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
	usertable.RegisterUser(recRegister, reqRegister)
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
	usertable.RegisterUser(rr, req)

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
