package usertable

import (
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/validemail"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @title Authentication Endpoints
// @description Endpoints for user registration, login, and logout.
//
// User represents an authenticated user.
// swagger:model User
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	BusinessName string
}

var (
	// db will hold the GORM DB instance
	db *gorm.DB
)

func init() {
	db = coredbutils.GetDB()
}

// ClearUserTable deletes all records from the users table.
// The AllowGlobalUpdate flag is required for global deletes in GORM v2.
//
// @Summary     Clear user table
// @Description Deletes all records from the user table in the database. Useful for testing and reset purposes.
//
// @Tags        dbtable, user, housekeeping
// @Success     200 {string} string "User table cleared successfully"
// @Failure     500 {string} string "Error clearing users table"
func ClearUserTable(db *gorm.DB) error {
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
		return fmt.Errorf("error clearing users table: %w", err)
	}
	return nil
}

// RegisterUser creates a new user record.
//
// @Summary      Register a new user
// @Description  Registers a new user using email, password, and an optional business name.
//
// @Tags         authentication, user, dbtable
// @Accept       application/x-www-form-urlencoded
// @Produce      plain
// @Param        email formData string true "User email"
// @Param        password formData string true "User password"
// @Param        business_name formData string false "Business name"
// @Success      200 {string} string "User registered successfully"
// @Failure      400 {string} string "Email and password are required or invalid email format"
// @Failure      409 {string} string "Email already in use or database error"
// @Router       /register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	businessName := r.FormValue("business_name")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Validate the email format using the standard library.
	if !validemail.Valid(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Hash the password before saving.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Create a new user record.
	user := User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		BusinessName: businessName,
	}

	// Use GORM to insert the new user into the database.
	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Email already in use or database error", http.StatusConflict)
		return
	}

	fmt.Fprintf(w, "User registered successfully")
}
