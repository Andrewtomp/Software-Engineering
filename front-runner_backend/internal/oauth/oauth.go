package oauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions" // You'll need session management
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	// Assuming you have a user service/package
	"front-runner/internal/usertable"
)

// Store must be initialized somewhere accessible, often in main or setup
// IMPORTANT: The key should be kept secret, perhaps from env vars.
// var Store = sessions.NewCookieStore([]byte("your-very-secret-key")) // Replace with secure key
var sharedStore *sessions.CookieStore // Define it here

const (
	sessionName    = "front-runner-session"
	userSessionKey = "userID" // Key to store user ID in session
)

// Setup initializes the OAuth providers and session store
func Setup(store *sessions.CookieStore) {
	if store == nil {
		log.Fatal("OAuth Setup: Received nil session store")
	}
	sharedStore = store

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	// Ensure callback URL matches what's registered with Google AND your server setup
	// For local dev: "http://localhost:8080/auth/google/callback"
	// For ngrok: Use the ngrok URL + "/auth/google/callback"
	googleCallbackURL := os.Getenv("GOOGLE_CALLBACK_URL") // e.g., http://localhost:8080/auth/google/callback
	ngrokDomain := os.Getenv("NGROK_DOMAIN")
	callbackPath := "/auth/google/callback" // Define your callback path

	if googleCallbackURL == "" { // If the specific redirect URI isn't set...
		if ngrokDomain != "" { // ...and the ngrok domain is set...
			// Ensure ngrokDomain starts with https://
			if !strings.HasPrefix(ngrokDomain, "https://") && !strings.HasPrefix(ngrokDomain, "http://") {
				ngrokDomain = "https://" + ngrokDomain // Default to https
			}
			googleCallbackURL = ngrokDomain + callbackPath // ...construct it using ngrok domain
			log.Printf("GOOGLE_REDIRECT_URI not set, using NGROK_DOMAIN: %s", googleCallbackURL)
		} else {
			// Fallback if neither is set (adjust as needed)
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080" // Default if not set
			}
			googleCallbackURL = "http://localhost:" + port + callbackPath // Or callbackPath // Or your default local setup
			log.Printf("Warning: GOOGLE_REDIRECT_URI and NGROK_DOMAIN not set. Defaulting Google callback to: %s", googleCallbackURL)
		}
	} else {
		log.Printf("Using explicitly set GOOGLE_REDIRECT_URI: %s", googleCallbackURL)
	}

	if googleClientID == "" || googleClientSecret == "" {
		log.Println("OAuth Setup: Warning: Google OAuth environment variables (GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET) not fully set. Google login may fail.")
		// Decide if this should be fatal or just a warning
	}

	goth.UseProviders(
		google.New(googleClientID, googleClientSecret, googleCallbackURL, "email", "profile"),
	)
	log.Println("Goth providers initialized.")

	// // Initialize session store (replace with a secure key!)
	// sessionAuthKey := os.Getenv("SESSION_AUTH_KEY")
	// if sessionAuthKey == "" {
	// 	log.Fatal("SESSION_AUTH_KEY environment variable not set")
	// 	// Or generate a temporary one for dev, but log a strong warning
	// }
	// sessionEncKey := os.Getenv("SESSION_ENC_KEY") // Optional encryption key
	// if sessionEncKey == "" {
	// 	Store = sessions.NewCookieStore([]byte(sessionAuthKey))
	// 	log.Println("Warning: SESSION_ENC_KEY not set. Session data will not be encrypted.")
	// } else {
	// 	Store = sessions.NewCookieStore([]byte(sessionAuthKey), []byte(sessionEncKey))
	// }

	// Store.Options = &sessions.Options{
	// 	Path:     "/",
	// 	MaxAge:   86400 * 7, // 7 days
	// 	HttpOnly: true,
	// 	Secure:   true,                 // Set to true if using HTTPS (which you are)
	// 	SameSite: http.SameSiteLaxMode, // Or StrictMode
	// }
	// gothic.Store = Store // Tell gothic to use this store

	// goth.UseProviders(
	// 	google.New(googleClientID, googleClientSecret, googleCallbackURL, "email", "profile"),
	// 	// Add other providers here if needed
	// )
}

// HandleGoogleLogin initiates the Google OAuth flow
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	providerName := "google"
	ctx := context.WithValue(r.Context(), gothic.ProviderParamKey, providerName)
	r = r.WithContext(ctx)
	gothic.BeginAuthHandler(w, r)
}

// HandleGoogleCallback handles the callback from Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Printf("Error completing Google auth: %v", err)
		if strings.Contains(err.Error(), "securecookie: the value is not valid") {
			// Attempt to clear the potentially bad cookie
			session := sessions.NewSession(sharedStore, sessionName)
			session.Options.MaxAge = -1
			session.Save(r, w)
			http.Error(w, "Session validation failed. Please try logging in again.", http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "Error completing Google auth: %v", err)
		}
		return
	}

	// --- Your Logic Here ---
	// 1. Check if user exists in your database based on gothUser.UserID or gothUser.Email
	user, err := usertable.GetUserByProviderID("google", gothUser.UserID) // Or GetUserByEmail(gothUser.Email)

	if err != nil { // Handle potential DB errors properly
		log.Printf("Error checking for user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// 2. If user doesn't exist, create a new user record
		log.Printf("User not found, creating new user: %s (%s)", gothUser.Name, gothUser.Email)
		newUser := &usertable.User{ // Adapt to your User struct
			Email:      gothUser.Email,
			Name:       gothUser.Name,
			Provider:   "google",
			ProviderID: gothUser.UserID,
			// Add other fields like AvatarURL: gothUser.AvatarURL if needed
		}
		err = usertable.CreateUser(newUser) // Implement CreateUser
		if err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Failed to create user account", http.StatusInternalServerError)
			return
		}
		user = newUser // Use the newly created user
	} else {
		log.Printf("Found existing user: %s (%s)", user.Name, user.Email)
		// Optionally update user info (Name, AvatarURL) from gothUser here
	}

	// 3. Create a session for the user
	session, err := sharedStore.Get(r, sessionName)
	if err != nil {
		// Handle error (though Get usually creates a new session if none exists)
		log.Printf("Error getting session in HandleGoogleCallback: %v", err)
		if strings.Contains(err.Error(), "securecookie: the value is not valid") {
			clearSession := sessions.NewSession(sharedStore, sessionName)
			clearSession.Options.MaxAge = -1
			clearSession.Save(r, w) // Best effort to clear
		}
		http.Error(w, "Session error after login. Please try again.", http.StatusInternalServerError)
		return
	}

	session.Values[userSessionKey] = user.ID // Store your internal user ID
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Session saving error", http.StatusInternalServerError)
		return
	}

	log.Printf("User %d logged in successfully via Google", user.ID)

	// 4. Redirect to a logged-in page (e.g., dashboard)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect) // Or wherever logged-in users should go
}

// HandleLogout clears the user session
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := sharedStore.Get(r, sessionName)
	if err == nil { // Only proceed if session exists
		// Clear session values
		delete(session.Values, userSessionKey)
		session.Options.MaxAge = -1 // Expire the cookie immediately
		err = session.Save(r, w)
		if err != nil {
			log.Printf("Error saving session during logout: %v", err)
		}
	} else {
		// If getting session failed, still try to clear cookie
		clearSession := sessions.NewSession(sharedStore, sessionName)
		clearSession.Options.MaxAge = -1
		clearSession.Save(r, w) // Best effort
		log.Printf("Logout: Error getting session (attempted clear): %v", err)
	}
	// Redirect to home page or login page after logout
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// --- Helper Function (Example) ---
// You might put this in a middleware or context utility

// GetCurrentUser retrieves the logged-in user from the session
func GetCurrentUser(r *http.Request) (*usertable.User, error) {
	session, err := sharedStore.Get(r, sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	userIDVal := session.Values[userSessionKey]
	if userIDVal == nil {
		return nil, nil // No user logged in
	}

	userID, ok := userIDVal.(uint) // Assuming your user ID is uint
	if !ok {
		// This indicates a problem with how the ID was stored
		log.Printf("Error: User ID in session is not of expected type (uint). Value: %v", userIDVal)
		// Clear the invalid session value?
		// session.Values[userSessionKey] = nil
		// Consider saving the session here if you modify it
		return nil, fmt.Errorf("invalid user ID type in session")
	}

	// Fetch user details from your database
	user, err := usertable.GetUserByID(userID) // Implement GetUserByID
	if err != nil {
		// Handle case where user ID in session doesn't match a DB user
		// log.Printf("Error fetching user %d from DB: %v", userID, err)
		// Clear the invalid session value?
		// session.Values[userSessionKey] = nil
		// Consider saving the session here if you modify it
		return nil, fmt.Errorf("failed to retrieve user from database: %w", err)
	}
	if user == nil {
		// User ID was in session but user not found in DB (maybe deleted?)
		log.Printf("Warning: User ID %d found in session but not in database.", userID)
		// Clear the invalid session value?
		// session.Values[userSessionKey] = nil
		// Consider saving the session here if you modify it
		return nil, nil
	}

	return user, nil
}
