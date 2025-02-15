package validemail

import "testing"

// TestValidEmail verifies that a properly formatted email address is considered valid.
//
// @Summary      Validate a correct email address
// @Description  Ensures that a correctly formatted email (e.g., "user@tester.com") returns true using the Valid function.
//
// @Tags         validate, email, testing
func TestValidEmail(t *testing.T) {
	email := "user@tester.com"

	if err := Valid(email); err != true {
		t.Fatalf("valid email came back as invalid: %s", email)
	}
}

// TestInvalidEmail verifies that improperly formatted email addresses are considered invalid.
//
// @Summary      Validate incorrect email addresses
// @Description  Ensures that various improperly formatted email addresses return false using the Valid function.
// @Tags         validate, email, testing
func TestInvalidEmail(t *testing.T) {
	email := []string{
		"usertester.com",
		"user",
		"user@.com",
		"@tester.com",
		"@.com",
		"@",
		".com",
	}

	for _, em := range email {
		if err := Valid(em); err != false {
			t.Fatalf("invalid email came back as valid: %s", email)
		}
	}
}
