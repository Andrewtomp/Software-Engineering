// login.go
package login

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/usertable"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// File-level annotations (optional):
// @title Authentication Endpoints
// @description Endpoints for user registration, login, and logout.

var (
	// db will hold the GORM DB instance
	db *gorm.DB

	// sessionStore is used to manage sessions
	sessionStore *sessions.CookieStore
)

// Init sets up the session store and connects to the PostgreSQL database using GORM.
func Setup() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		// Initialize the session store with a random key.
		key := make([]byte, 64)
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		sessionStore = sessions.NewCookieStore(key)
	} else {
		// Optionally, hash the secret to ensure it has the desired length.
		hash := sha256.Sum256([]byte(secret))
		key, _ := hex.DecodeString(hex.EncodeToString(hash[:]))
		sessionStore = sessions.NewCookieStore(key)
	}

	// Setting the auth cookie to ba available through the whole domain
	// sessionStore = sessions.NewCookieStore(key)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // e.g. valid for 7 days by default
		HttpOnly: true,
	}

	coredbutils.LoadEnv()
	db = coredbutils.GetDB()
	usertable.Setup()
	log.Println("login package init: sessionStore initialized")
}

// LoginUser authenticates a user and creates a session.
//
// @Summary      User login
// @Description  Authenticates a user and creates a session.
//
// @Tags         authentication
// @Accept       application/x-www-form-urlencoded
// @Produce      plain
// @Param        email formData string true "User email"
// @Param        password formData string true "User password"
// @Success      200 {string} string "Logged in successfully."
// @Failure      400 {string} string "Email and password are required"
// @Failure      401 {string} string "Invalid credentials"
// @Router       /api/login [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session first.
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		if err.Error() == "securecookie: the value is not valid" {
			// Invalidate the current cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "auth",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			// Remove the cookie from the request so that a new session is created without trying to decode it
			r.Header.Set("Cookie", "")
			// Optionally, try to get a fresh session
			session, err = sessionStore.New(r, "auth")
			if err != nil {
				http.Error(w, "Error creating session", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
			return
		}
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

	var user usertable.User
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

	// fmt.Fprintf(w, "Logged in successfully.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutUser clears the user's session.
//
// @Summary      User logout
// @Description  Logs out the current user by clearing the session.
//
// @Tags         authentication
// @Produce      plain
// @Success      200 {string} string "Logged out successfully"
// @Router       /api/logout [post]
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
		http.Error(w, "Error saving session"+err.Error(), http.StatusInternalServerError)
		return
	}

	// Provide a message based on whether the user was logged in.
	if loggedIn {
		fmt.Fprintf(w, "Logged out successfully")
	} else {
		fmt.Fprintf(w, "User is already logged out")
	}
}

// Given a request header, checks if the user is logged in.
func IsLoggedIn(r *http.Request) bool {
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		return false
	}
	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

// Given a request header, retrieves the user's ID.
func GetUserID(r *http.Request) (uint, error) {
	session, err := sessionStore.Get(r, "auth")
	if err != nil {
		return 0, err
	}
	userID, ok := session.Values["user_id"].(uint)
	if !ok {
		return 0, errors.New("no user id associated with session")
	}

	return userID, nil
}
