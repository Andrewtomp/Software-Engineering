package validemail

import "net/mail"

// Checks if the provided email address is in a valid format. Returns true if valid, false otherwise.
func Valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
