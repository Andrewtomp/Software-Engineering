package login

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// db will hold the GORM DB instance
	db *gorm.DB

	// sessionStore is used to manage sessions
	sessionStore *sessions.CookieStore
)

// User represents the user model in the database.
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
}

// Init sets up the session store and connects to the PostgreSQL database using GORM.
func init() {
	// Initialize the session store with a random key.
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	sessionStore = sessions.NewCookieStore(key)

	// Build the DSN (Data Source Name) for PostgreSQL.
	// Adjust the parameters (host, user, password, dbname, port, sslmode, TimeZone) as needed.
	dsn := "host=localhost port=5432 user=johnny dbname=users sslmode=disable TimeZone=UTC"

	// Open a connection to the PostgreSQL database using GORM.
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Automatically migrate the schema, creating the table if it doesn't exist.
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(fmt.Sprintf("failed to migrate database schema: %v", err))
	}
}

// RegisterUser creates a new user record.
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
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
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// Use GORM to insert the new user into the database.
	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Username already exists or database error", http.StatusConflict)
		return
	}

	fmt.Fprintf(w, "User registered successfully")
}

// LoginUser authenticates a user and creates a session.
func LoginUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	var user User
	// Look up the user by username.
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Compare the provided password with the stored hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create a session.
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	// Optionally, generate a token.
	token := generateToken()
	fmt.Fprintf(w, "Logged in successfully. Token: %s", token)
}

// LogoutUser clears the user's session.
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		http.Error(w, "Error getting session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1 // Mark the session for deletion.
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Logged out successfully")
}

// generateToken creates a random token.
func generateToken() string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(token)
}
