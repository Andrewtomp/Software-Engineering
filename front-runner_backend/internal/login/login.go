package login

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/mail"

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
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	BusinessName string
}

// valid uses mail.ParseAddress to check whether the provided email is valid.
func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
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

	// Setting the auth cookie to ba available through the whole domain
	sessionStore = sessions.NewCookieStore(key)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // e.g. valid for 7 days by default
		HttpOnly: true,
	}
}

// RegisterUser creates a new user record.
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	businessName := r.FormValue("business_name")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Validate the email format using the standard library.
	if !valid(email) {
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

// LoginUser authenticates a user and creates a session.
func LoginUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session first.
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		http.Error(w, "Error retrieving session", http.StatusInternalServerError)
		return
	}

	// Check if the user is already logged in.
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		fmt.Fprintf(w, "User is already logged in")
		return
	}

	// Read login credentials from the request.
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var user User
	// Look up the user by username.
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Compare the provided password with the stored hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Logged in successfully.")
}

// LogoutUser clears the user's session.
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		http.Error(w, "Error getting session", http.StatusInternalServerError)
		return
	}

	// You might also want to explicitly check the authentication flag
	loggedIn := false
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		loggedIn = true
	}

	// Clearing the session cookie by marking it for deletion
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	// Provide a message based on whether the user was logged in.
	if loggedIn {
		fmt.Fprintf(w, "Logged out successfully")
	} else {
		fmt.Fprintf(w, "User is already logged out")
	}
}
