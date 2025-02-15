package validemail

import "net/mail"

// Valid checks whether the provided email address is valid using net/mail's ParseAddress.
//
// @Summary      Validate email address
// @Description  Checks if the provided email address is in a valid format. Returns true if valid, false otherwise.
//
// @Tags         utility, email, validate
// @Param        email query string true "Email address to validate"
// @Success      200 {boolean} boolean "true if email is valid, false otherwise"
// @Router       /validemail [get]
func Valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
