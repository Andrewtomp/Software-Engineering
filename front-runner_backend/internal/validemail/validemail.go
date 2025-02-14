package validemail

import "net/mail"

// valid uses mail.ParseAddress to check whether the provided email is valid.
func Valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
