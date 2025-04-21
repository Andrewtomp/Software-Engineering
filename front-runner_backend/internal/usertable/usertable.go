package usertable

import (
	"errors"
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/validemail"
	"log"
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents an application user in the database.
// It includes fields for both local password-based authentication
// and OAuth provider-based authentication.
// swagger:model User
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Name         string
	BusinessName string
	Provider     string `gorm:"not null;default:'local';index"`
	ProviderID   string `gorm:"not null;index"`
}

var (
	// db will hold the GORM DB instance
	db        *gorm.DB
	setupOnce sync.Once
)

// Setup initializes the database connection for the usertable package.
// It ensures the connection is obtained only once using sync.Once.
func Setup() {
	setupOnce.Do(func() {
		coredbutils.LoadEnv()
		db, _ = coredbutils.GetDB()
	})
}

// MigrateUserDB runs the GORM auto-migration for the User model.
// It ensures the users table schema matches the User struct definition.
func MigrateUserDB() {
	if db == nil {
		log.Fatal("Database connection is not initialized")
	}
	log.Println("Running user database migrations...")
	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("User database migration complete")
}

// ClearUserTable deletes all records from the users table.
// USE WITH EXTREME CAUTION, especially in production environments.
// Primarily intended for testing or complete resets.
func ClearUserTable(db *gorm.DB) error {
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
		return fmt.Errorf("error clearing users table: %w", err)
	}
	return nil
}

// RegisterUser handles the HTTP request for creating a new 'local' user account.
// It expects email, password, name, and optionally businessName via form data.
//
// @Summary      Register a new local user
// @Description  Registers a new user account using email and password for local authentication.
// @Tags         Authentication
// @Accept       application/x-www-form-urlencoded
// @Produce      text/plain
// @Param        email        formData string true  "User's Email Address" example("user@example.com")
// @Param        password     formData string true  "User's Password (min length recommended)" example("password123")
// @Param        name         formData string true  "User's Full Name" example("John Doe")
// @Param        businessName formData string false "User's Business Name (Optional)" example("JD Enterprises")
// @Success      200 {string} string "User registered successfully"
// @Failure      400 {string} string "Bad Request: Missing required fields (email, password, name), or invalid email format"
// @Failure      409 {string} string "Conflict: Email address is already registered"
// @Failure      500 {string} string "Internal Server Error: Failed to hash password or save user to database"
// @Router       /api/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	businessName := r.FormValue("businessName")
	name := r.FormValue("name")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Validate the email format using the standard library.
	if !validemail.Valid(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	var existingUser User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email already in use", http.StatusConflict)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Database error", http.StatusInternalServerError)
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
		Name:         name,
		BusinessName: businessName,
		Provider:     "local",
		ProviderID:   "",
	}

	// Use GORM to insert the new user into the database.
	if err := db.Create(&user).Error; err != nil {
		log.Printf("Database error creating user %s: %v", email, err)
		http.Error(w, "Failed to register user due to database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Local user created successfully: %s", email)
	fmt.Fprintf(w, "User registered successfully")
}

// GetUserByProviderID finds a user based on their OAuth provider and provider-specific ID.
// Returns the user pointer or nil if not found. Returns an error for database issues.
func GetUserByProviderID(provider, providerID string) (*User, error) {
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}
	var user User
	err := db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail finds a user based on their email address.
// Returns the user pointer or nil if not found. Returns an error for database issues.
func GetUserByEmail(email string) (*User, error) {
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("Error fetching user by email (%s): %v", email, err)
		return nil, fmt.Errorf("database error fetching user by email: %w", err)
	}
	return &user, nil
}

// CreateUser saves a new user record to the database.
// Performs basic validation before attempting to save.
// Returns an error if validation fails or the database operation fails.
func CreateUser(user *User) error {
	if db == nil {
		return errors.New("database connection not initialized")
	}
	if user.Email == "" || !validemail.Valid(user.Email) {
		return errors.New("invalid or missing email")
	}
	if user.Provider == "" {
		return errors.New("provider cannot be empty")
	}
	if user.Provider != "local" && user.ProviderID == "" {
		return errors.New("provider ID is required for non-local providers")
	}
	if user.Provider == "local" && user.PasswordHash == "" {
		return errors.New("password hash is required for local provider")
	}

	err := db.Create(user).Error
	if err != nil {
		log.Printf("Error creating user (%s, %s): %v", user.Provider, user.Email, err)
		return fmt.Errorf("database error creating user: %w", err)
	}
	log.Printf("User created successfully: Provider=%s, Email=%s, ID=%d", user.Provider, user.Email, user.ID)
	return nil
}

// GetUserByID finds a user based on their primary key ID.
// Returns the user pointer or nil if not found. Returns an error for database issues.
func GetUserByID(userID uint) (*User, error) {
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}
	var user User
	err := db.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("Error fetching user by ID (%d): %v", userID, err)
		return nil, fmt.Errorf("database error fetching user by ID: %w", err)
	}
	return &user, nil
}

// UpdateUser updates an existing user record in the database using all fields from the provided User struct.
// It requires the User struct to have a valid ID.
// Returns an error if the database operation fails.
func UpdateUser(user *User) error {
	if db == nil {
		return errors.New("database connection not initialized")
	}
	if user.ID == 0 {
		return errors.New("cannot update user without ID")
	}

	log.Printf("DEBUG: UpdateUser attempting to Save user ID: %d", user.ID)

	// Use Save to update all fields, or Updates to update specific fields/non-zero fields
	result := db.Updates(user)
	err := result.Error

	if err != nil {
		log.Printf("Error updating user (ID: %d): %v", user.ID, err)
		return fmt.Errorf("database error updating user: %w", err)
	}

	if result.RowsAffected == 0 {
		log.Printf("DEBUG: UpdateUser(ID: %d) - RowsAffected is 0. Returning gorm.ErrRecordNotFound.", user.ID) // Added log
		log.Printf("Attempted to update non-existent user (ID: %d), no rows affected.", user.ID)
		// Return ErrRecordNotFound to indicate the record wasn't found for updating.
		return gorm.ErrRecordNotFound
	}

	log.Printf("DEBUG: UpdateUser(ID: %d) - RowsAffected is %d (> 0). Returning nil.", user.ID, result.RowsAffected)
	return nil
}
