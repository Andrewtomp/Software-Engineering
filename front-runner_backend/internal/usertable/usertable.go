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

// @title Authentication Endpoints
// @description Endpoints for user registration, login, and logout.
//
// User represents an authenticated user.
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

func Setup() {
	setupOnce.Do(func() {
		coredbutils.LoadEnv()
		db = coredbutils.GetDB()
	})
}

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
// @Tags         authentication
// @Accept       application/x-www-form-urlencoded
// @Produce      plain
// @Param        email formData string true "User email"
// @Param        password formData string true "User password"
// @Param        business_name formData string false "Business name"
// @Success      200 {string} string "User registered successfully"
// @Failure      400 {string} string "Email and password are required or invalid email format"
// @Failure      409 {string} string "Email already in use or database error"
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

func UpdateUser(user *User) error {
	if db == nil {
		return errors.New("database connection not initialized")
	}
	if user.ID == 0 {
		return errors.New("cannot update user without ID")
	}
	// Use Save to update all fields, or Updates to update specific fields/non-zero fields
	err := db.Save(user).Error
	if err != nil {
		log.Printf("Error updating user (ID: %d): %v", user.ID, err)
		return fmt.Errorf("database error updating user: %w", err)
	}
	return nil
}
