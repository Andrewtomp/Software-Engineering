// login.go
package login

import (
	"errors"
	"front-runner/internal/usertable"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	sessionName    = "front-runner-session"
	userSessionKey = "userID" // Key to store user ID in session
)

var (
	db                 *gorm.DB
	sharedSessionStore *sessions.CookieStore
)

// Setup initializes the login package with necessary dependencies.
// It requires a database connection and a configured session store.
func Setup(database *gorm.DB, store *sessions.CookieStore) {
	if database == nil || store == nil {
		log.Fatal("Login Setup: Recieved nil database or session store")
	}
	db = database
	sharedSessionStore = store
	usertable.Setup()
	log.Println("login package init: sessionStore initialized")
}

// LoginUser authenticates a user via email and password and establishes a session.
// It expects form data with 'email' and 'password'.
// On success, it redirects the user to the root path ('/').
//
// @Summary      User Login (Email/Password)
// @Description  Authenticates a user using email and password. Creates a session cookie upon successful authentication and redirects to the homepage.
// @Tags         Authentication
// @Accept       application/x-www-form-urlencoded
// @Param        email     formData  string  true  "User's Email Address"
// @Param        password  formData  string  true  "User's Password"
// @Success      303  {string}  string  "Redirects to / on successful login"
// @Failure      400  {string}  string  "Bad Request: Email and password are required"
// @Failure      401  {string}  string  "Unauthorized: Invalid credentials"
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/login [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session first.
	// session, err := sessionStore.Get(r, "auth")
	session, err := sharedSessionStore.Get(r, sessionName)
	if err != nil {
		log.Printf("Login: Error getting session (attempting to clear): %v", err)
		// Clear potentially invalid cookie and try creating a new session
		session = sessions.NewSession(sharedSessionStore, sessionName)
		session.Options.MaxAge = -1 // Expire immediately
		session.Save(r, w)          // Attempt to save the cleared session/cookie
		// Proceed to login attempt with a fresh (empty) session state
	}

	if userID, ok := session.Values[userSessionKey].(uint); ok && userID > 0 {
		http.Error(w, "User is already logged in", http.StatusConflict)
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
	err = db.Where("email = ? AND provider = ?", email, "local").First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			log.Printf("Database error during login for email %s: %v", email, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Compare the provided password with the stored hash.
	if user.PasswordHash == "" {
		log.Printf("Attempt to login with empty hash for local user: %s", email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// session.Values["authenticated"] = true
	// session.Values["user_id"] = user.ID
	session.Values[userSessionKey] = user.ID

	// Save the session.
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	// fmt.Fprintf(w, "Logged in successfully.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutUser clears the user's session information, effectively logging them out.
// It redirects the user to the root path ('/') regardless of initial login state.
//
// @Summary      User Logout
// @Description  Logs out the current user by clearing the session cookie and redirects to the homepage.
// @Tags         Authentication
// @Success      303  {string}  string "Redirects to / after logout"
// @Failure      500  {string}  string "Internal Server Error (while saving cleared session)"
// @Router       /logout [get] // Assuming logout is triggered by a GET request to /logout
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	// session, err := sessionStore.Get(r, "auth")
	session, err := sharedSessionStore.Get(r, sessionName)
	if err != nil {
		log.Printf("Logout: Error getting session (ignoring): %v", err)
		session = sessions.NewSession(sharedSessionStore, sessionName)
		session.Options.MaxAge = -1 // Expire the cookie immediately
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	delete(session.Values, userSessionKey)

	// Clearing the session cookie by marking it for deletion
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session during logout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
