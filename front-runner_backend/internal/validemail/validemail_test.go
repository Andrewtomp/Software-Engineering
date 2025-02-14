package validemail

import "testing"

func TestValidEmail(t *testing.T) {
	email := "user@tester.com"

	if err := Valid(email); err != true {
		t.Fatalf("valid email came back as invalid: %s", email)
	}
}

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
