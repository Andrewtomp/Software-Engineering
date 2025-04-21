// front-runner/internal/storefronttable/storefronttable.go
package storefronttable // Correct package name

import (
	"encoding/json"
	"errors"
	"fmt"
	"front-runner/internal/coredbutils" // Use coredbutils for DB access
	"front-runner/internal/oauth"       // Import oauth

	"log"
	"net/http"
	"strconv"
	"strings" // Needed for unique constraint check example
	"sync"
	"time"

	"gorm.io/gorm"
)

// --- Struct Definitions ---

// StorefrontLink represents a linked external storefront in the database.
type StorefrontLink struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null;index:idx_user_store_unique,unique,priority:1"` // Composite unique index
	StoreType   string `gorm:"not null;index:idx_user_store_unique,unique,priority:2"` // Composite unique index
	StoreName   string `gorm:"index:idx_user_store_unique,unique,priority:3"`          // Composite unique index
	Credentials string `gorm:"not null;type:text"`                                     // Store encrypted data (use text type for potentially longer strings)
	StoreID     string `gorm:"index"`                                                  // Index for potential lookups by StoreID
	StoreURL    string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// StorefrontLinkAddPayload is used to decode the JSON body when adding a link.
type StorefrontLinkAddPayload struct {
	StoreType string `json:"storeType"`
	StoreName string `json:"storeName"` // User-defined nickname
	ApiKey    string `json:"apiKey"`    // Example credential field
	ApiSecret string `json:"apiSecret"` // Example credential field
	StoreId   string `json:"storeId"`   // Platform-specific ID
	StoreUrl  string `json:"storeUrl"`  // Storefront URL
	// Add other potential credential fields as needed per platform
}

// StorefrontLinkReturn is the struct returned to the frontend.
// IMPORTANT: It omits the sensitive Credentials field.
type StorefrontLinkReturn struct {
	ID        uint   `json:"id"`
	StoreType string `json:"storeType"`
	StoreName string `json:"storeName"`
	StoreID   string `json:"storeId"` // Match frontend JSON keys
	StoreURL  string `json:"storeUrl"`
}

// StorefrontLinkUpdatePayload defines the fields allowed for updating a storefront link.
type StorefrontLinkUpdatePayload struct {
	StoreName string `json:"storeName"` // User-defined nickname
	StoreId   string `json:"storeId"`   // Platform-specific ID
	StoreUrl  string `json:"storeUrl"`  // Storefront URL
}

// --- Package Variables ---
var (
	db        *gorm.DB
	setupOnce sync.Once
)

// --- Setup and Migration ---

// Setup initializes the database connection for the storefronttable package
// and ensures the encryption key is loaded.
func Setup() {
	setupOnce.Do(func() {

		// Load encryption key first, fail early if not configured
		// This refers to loadEncryptionKey in the same (storefronttable) package

		if err := loadEncryptionKey(); err != nil {
			// Use log.Fatalf to stop execution if security cannot be guaranteed
			log.Fatalf("FATAL: Failed to setup storefronttable package security: %v", err)
		}

		// Get DB connection from coredbutils
		coredbutils.LoadEnv()
		db, _ = coredbutils.GetDB()
		// login.Setup() // REMOVE: Login setup is now centralized
		if db == nil {
			log.Fatal("FATAL: Database connection is nil in storefronttable Setup.")
		}

		log.Println("storefronttable package setup complete.")
	})
}

// MigrateStorefrontDB runs the database migration for the StorefrontLink table.
func MigrateStorefrontDB() {
	// Ensure setup has run and db is initialized
	if db == nil {
		log.Fatal("Storefronttable Database connection is not initialized before migration. Ensure Setup() is called first.")
	}
	log.Println("Running storefront link database migrations...")
	// AutoMigrate will create the table, add missing columns/indexes,
	// but typically won't delete/change existing ones without extra configuration.
	err := db.AutoMigrate(&StorefrontLink{})
	if err != nil {
		log.Fatalf("Storefront link migration failed: %v", err)
	}
	log.Println("Storefront link database migration complete.")
}

// ClearStorefrontTable removes all records from the storefront_links table. USE WITH CAUTION.
func ClearStorefrontTable(db *gorm.DB) error {
	if db == nil {
		return errors.New("storefront db is nil")
	}
	log.Println("DB COMMAND: DELETE FROM storefront_links")
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&StorefrontLink{}).Error; err != nil {
		return fmt.Errorf("error clearing storefront_links table: %w", err)
	}
	log.Println("Cleared storefront_links table successfully.")
	return nil
}

// --- API Handlers ---

// AddStorefront handles linking a new external storefront.
// @Summary      Link a new storefront
// @Description  Links a new external storefront (e.g., Amazon, Pinterest) to the user's account, storing credentials securely. Requires authentication.
// @Tags         Storefronts
// @Accept       json
// @Param        storefrontLink body StorefrontLinkAddPayload true "Storefront Link Details (including credentials like apiKey, apiSecret)"
// @Success      201 {object} StorefrontLinkReturn "Successfully linked storefront (credentials omitted)"
// @Failure      400 {string} string "Bad Request - Invalid input, missing fields, or JSON parsing error"
// @Failure      401 {string} string "Unauthorized - User session invalid or expired"
// @Failure      409 {string} string "Conflict - A link with this name/type already exists for the user"
// @Failure      500 {string} string "Internal Server Error - E.g., failed to encrypt, database error"
// @Security     ApiKeyAuth // Assuming ApiKeyAuth is defined for session/token auth
// @Router       /api/add_storefront [post]
func AddStorefront(w http.ResponseWriter, r *http.Request) {
	userID, ok := checkAuth(w, r) // Check authentication first
	if !ok {
		return // Error response already sent by checkAuth
	}

	var payload StorefrontLinkAddPayload
	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Good practice to close body

	// --- Basic Input Validation ---
	if strings.TrimSpace(payload.StoreType) == "" {
		http.Error(w, "Missing required field: storeType", http.StatusBadRequest)
		return
	}
	// Example validation: Require credentials for certain types if known
	if payload.StoreType == "amazon" && (payload.ApiKey == "" || payload.ApiSecret == "") {
		http.Error(w, "apiKey and apiSecret are required for Amazon links", http.StatusBadRequest)
		return
	}
	// Add more specific validation as needed per store type

	// --- Prepare and Encrypt Credentials ---
	// Structure credentials for consistent encryption (JSON map is flexible)
	credentialsMap := map[string]string{
		// Only include non-empty credentials provided by the user
	}
	if payload.ApiKey != "" {
		credentialsMap["apiKey"] = payload.ApiKey
	}
	if payload.ApiSecret != "" {
		credentialsMap["apiSecret"] = payload.ApiSecret
	}
	// Add other fields from payload if they are considered sensitive

	// Only encrypt if there are actual credentials to store
	var encryptedCredentials string
	var err error
	if len(credentialsMap) > 0 {
		credentialsJSON, errMarshal := json.Marshal(credentialsMap)
		if errMarshal != nil {
			log.Printf("Error marshalling credentials for user %d: %v", userID, errMarshal)
			http.Error(w, "Internal server error while preparing credentials", http.StatusInternalServerError)
			return
		}
		encryptedCredentials, err = encryptCredentials(string(credentialsJSON))
		if err != nil {
			log.Printf("Error encrypting credentials for user %d: %v", userID, err)
			http.Error(w, "Failed to secure credentials", http.StatusInternalServerError)
			return
		}
	} else {
		// Handle cases where no credentials are provided, if allowed by your logic
		// encryptedCredentials = "" // Or perhaps return an error if credentials are required
		log.Printf("Warning: No credentials provided for storefront type %s for user %d", payload.StoreType, userID)
		// Depending on requirements, you might allow linking without credentials
		// For now, let's assume some credentials (even empty ones if allowed) lead to an empty encrypted string
		encryptedCredentials = ""
		// If you *require* credentials, add validation earlier.
	}

	// --- Create Database Record ---
	newLink := StorefrontLink{
		UserID:      userID,
		StoreType:   payload.StoreType,
		StoreName:   payload.StoreName,
		Credentials: encryptedCredentials, // Store the encrypted string
		StoreID:     payload.StoreId,
		StoreURL:    payload.StoreUrl,
	}

	// Provide a default name if the user didn't specify one
	if strings.TrimSpace(newLink.StoreName) == "" {
		newLink.StoreName = fmt.Sprintf("%s Link", payload.StoreType) // e.g., "Amazon Link"
	}

	// Attempt to save to database
	result := db.Create(&newLink)
	if result.Error != nil {
		// Check specifically for unique constraint violation
		// Note: Error message checks are database-dependent and brittle.
		// Using GORM's error types or specific DB driver errors is more reliable if available.
		// Example check for PostgreSQL unique violation:
		if strings.Contains(result.Error.Error(), "unique constraint") || strings.Contains(result.Error.Error(), "idx_user_store_unique") {
			http.Error(w, "A storefront link with this type and name already exists for your account.", http.StatusConflict) // 409 Conflict
		} else {
			// Log the unexpected database error
			log.Printf("Error saving storefront link for user %d: %v", userID, result.Error)
			http.Error(w, "Failed to save storefront link due to a database error", http.StatusInternalServerError)
		}
		return
	}

	// --- Return Success Response (Safe Data Only) ---
	// Create the return object *without* credentials
	returnData := StorefrontLinkReturn{
		ID:        newLink.ID,
		StoreType: newLink.StoreType,
		StoreName: newLink.StoreName,
		StoreID:   newLink.StoreID,
		StoreURL:  newLink.StoreURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(returnData)
}

// GetStorefronts retrieves all linked storefronts for the logged-in user.
// @Summary      Get linked storefronts
// @Description  Retrieves a list of all external storefronts linked by the currently authenticated user. Credentials are *never* included. Requires authentication.
// @Tags         Storefronts
// @Success      200 {array} StorefrontLinkReturn "List of linked storefronts (empty array if none)"
// @Failure      401 {string} string "Unauthorized - User session invalid or expired"
// @Failure      500 {string} string "Internal Server Error - Database query failed"
// @Security     ApiKeyAuth
// @Router       /api/get_storefronts [get]
func GetStorefronts(w http.ResponseWriter, r *http.Request) {
	userID, ok := checkAuth(w, r)
	if !ok {
		return
	}

	var links []StorefrontLink
	// Query database for links belonging to the user, order them consistently
	result := db.Where("user_id = ?", userID).Order("store_type asc, store_name asc").Find(&links)
	if result.Error != nil {
		log.Printf("Error retrieving storefront links for user %d: %v", userID, result.Error)
		http.Error(w, "Failed to retrieve storefront links", http.StatusInternalServerError)
		return
	}

	// Prepare return data (transforming DB model to safe return model)
	returnData := make([]StorefrontLinkReturn, len(links)) // Pre-allocate slice
	for i, link := range links {
		returnData[i] = StorefrontLinkReturn{
			ID:        link.ID,
			StoreType: link.StoreType,
			StoreName: link.StoreName,
			StoreID:   link.StoreID,
			StoreURL:  link.StoreURL,
		}
	}

	// Return the array (it will be `[]` if no links were found, which is correct)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(returnData)
}

// UpdateStorefront handles updating non-sensitive details of an existing storefront link.
// @Summary      Update a storefront link
// @Description  Updates the name, store ID, or store URL of an existing storefront link belonging to the authenticated user. Store type and credentials cannot be updated via this endpoint.
// @Tags         Storefronts
// @Accept       json
// @Param        id query integer true "ID of the Storefront Link to update" Format(uint) example(123)
// @Param        storefrontUpdate body StorefrontLinkUpdatePayload true "Fields to update (storeName, storeId, storeUrl)"
// @Success      200 {object} StorefrontLinkReturn "Successfully updated storefront link details"
// @Failure      400 {string} string "Bad Request - Invalid input, missing ID, or JSON parsing error"
// @Failure      401 {string} string "Unauthorized - User session invalid or expired"
// @Failure      403 {string} string "Forbidden - User does not own this storefront link"
// @Failure      404 {string} string "Not Found - Storefront link with the specified ID not found"
// @Failure      409 {string} string "Conflict - Update would violate a unique constraint (e.g., duplicate name)"
// @Failure      500 {string} string "Internal Server Error - Database update failed"
// @Security     ApiKeyAuth
// @Router       /api/update_storefront [put]
func UpdateStorefront(w http.ResponseWriter, r *http.Request) {
	userID, ok := checkAuth(w, r) // Check authentication
	if !ok {
		return
	}

	// --- Get and Validate ID from Query Parameter ---
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing required query parameter: id", http.StatusBadRequest)
		return
	}
	linkID64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID format: must be a positive integer", http.StatusBadRequest)
		return
	}
	linkID := uint(linkID64)

	// --- Decode Request Body ---
	var payload StorefrontLinkUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// --- Find Existing Record ---
	var link StorefrontLink
	result := db.First(&link, linkID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Storefront link with ID %d not found", linkID), http.StatusNotFound)
		} else {
			log.Printf("Error finding storefront link ID %d for update: %v", linkID, result.Error)
			http.Error(w, "Internal server error while searching for link", http.StatusInternalServerError)
		}
		return
	}

	// --- Verify Ownership ---
	if link.UserID != userID {
		log.Printf("Security violation: User %d attempted to update storefront link ID %d owned by user %d", userID, linkID, link.UserID)
		http.Error(w, "Forbidden: You do not have permission to update this storefront link", http.StatusForbidden)
		return
	}

	// --- Update Fields (Apply changes from payload) ---
	// We update the fields based on the payload.
	// Handle empty StoreName: If user sends empty name, default it like in AddStorefront.
	link.StoreName = strings.TrimSpace(payload.StoreName)
	if link.StoreName == "" {
		link.StoreName = fmt.Sprintf("%s Link", link.StoreType) // Default name based on existing type
	}
	link.StoreID = payload.StoreId   // Allow empty StoreId if desired
	link.StoreURL = payload.StoreUrl // Allow empty StoreUrl if desired

	// Note: StoreType and Credentials are NOT updated here.

	// --- Save Changes to Database ---
	saveResult := db.Save(&link)
	if saveResult.Error != nil {
		// Check for unique constraint violation on update
		if strings.Contains(saveResult.Error.Error(), "unique constraint") || strings.Contains(saveResult.Error.Error(), "idx_user_store_unique") {
			http.Error(w, "Update failed: A storefront link with the new name already exists for this type.", http.StatusConflict) // 409 Conflict
		} else {
			log.Printf("Error updating storefront link ID %d for user %d: %v", linkID, userID, saveResult.Error)
			http.Error(w, "Failed to update storefront link due to a database error", http.StatusInternalServerError)
		}
		return
	}

	// --- Return Success Response (Updated, Safe Data) ---
	returnData := StorefrontLinkReturn{
		ID:        link.ID,        // ID doesn't change
		StoreType: link.StoreType, // Type doesn't change
		StoreName: link.StoreName, // Updated name
		StoreID:   link.StoreID,   // Updated Store ID
		StoreURL:  link.StoreURL,  // Updated URL
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(returnData)
}

// DeleteStorefront removes a linked storefront for the logged-in user by its ID.
// @Summary      Unlink a storefront
// @Description  Removes the link to an external storefront specified by its unique ID. User must own the link. Requires authentication.
// @Tags         Storefronts
// @Param        id query integer true "ID of the Storefront Link to delete" Format(uint) example(123)
// @Success      200 {string} string "Storefront unlinked successfully"
// @Success      204 {string} string "Storefront unlinked successfully (No Content)" // Added 204 as an alternative success
// @Failure      400 {string} string "Bad Request - Invalid or missing 'id' query parameter"
// @Failure      401 {string} string "Unauthorized - User session invalid or expired"
// @Failure      403 {string} string "Forbidden - User does not own this storefront link"
// @Failure      404 {string} string "Not Found - Storefront link with the specified ID not found"
// @Failure      500 {string} string "Internal Server Error - Database deletion failed"
// @Security     ApiKeyAuth
// @Router       /api/delete_storefront [delete]
func DeleteStorefront(w http.ResponseWriter, r *http.Request) {
	userID, ok := checkAuth(w, r)
	if !ok {
		return
	}

	// --- Get and Validate ID from Query Parameter ---
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing required query parameter: id", http.StatusBadRequest)
		return
	}

	// Parse the ID string to an unsigned integer
	linkID64, err := strconv.ParseUint(idStr, 10, 32) // Use 32 or 64 based on expected ID size
	if err != nil {
		http.Error(w, "Invalid ID format: must be a positive integer", http.StatusBadRequest)
		return
	}
	linkID := uint(linkID64) // Convert to uint for GORM matching

	// --- Find the Link in Database ---
	var link StorefrontLink
	// Use GORM's First method, which finds by primary key automatically if given an integer
	result := db.First(&link, linkID)

	// Handle errors during find
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Storefront link with ID %d not found", linkID), http.StatusNotFound) // 404 Not Found
		} else {
			// Log unexpected database error
			log.Printf("Error finding storefront link ID %d for deletion: %v", linkID, result.Error)
			http.Error(w, "Internal server error while searching for link", http.StatusInternalServerError)
		}
		return
	}

	// --- Verify Ownership ---
	// Crucial security check: does the found link belong to the logged-in user?
	if link.UserID != userID {
		// Log potential security violation attempt
		log.Printf("Security violation: User %d attempted to delete storefront link ID %d owned by user %d", userID, linkID, link.UserID)
		http.Error(w, "Forbidden: You do not have permission to delete this storefront link", http.StatusForbidden) // 403 Forbidden
		return
	}

	// --- Delete the Link ---
	// Perform the delete operation using the found link object
	deleteResult := db.Delete(&link)
	if deleteResult.Error != nil {
		// Log error during deletion
		log.Printf("Error deleting storefront link ID %d for user %d: %v", linkID, userID, deleteResult.Error)
		http.Error(w, "Failed to delete storefront link due to a database error", http.StatusInternalServerError)
		return
	}

	// Check if any row was actually deleted (optional, but good practice)
	if deleteResult.RowsAffected == 0 {
		log.Printf("Warning: Delete command for storefront link ID %d affected 0 rows (was it already deleted?)", linkID)
		// Still return success, as the desired state (link doesn't exist) is achieved
	}

	// --- Return Success Response ---
	// Respond with 200 OK or 204 No Content (both indicate success)
	w.WriteHeader(http.StatusOK) // 200 OK with optional message
	fmt.Fprintf(w, "Storefront link (ID: %d) unlinked successfully", linkID)

	// Alternatively, use 204 No Content if no message body is needed:
	// w.WriteHeader(http.StatusNoContent)
}

// --- Helper Functions ---

// checkAuth is a helper to verify login status and retrieve UserID using the unified oauth package.
// It writes appropriate HTTP errors (401, 500) directly to the response writer.
// Returns the UserID and true if authenticated, otherwise 0 and false.
func checkAuth(w http.ResponseWriter, r *http.Request) (userID uint, ok bool) {
	// Use the unified GetCurrentUser function
	user, err := oauth.GetCurrentUser(r)

	if err != nil {
		// Log the internal session/database error
		log.Printf("checkAuth: Error getting current user: %v", err)
		http.Error(w, "Internal Server Error: Could not verify user session.", http.StatusInternalServerError)
		return 0, false
	}

	if user == nil {
		// User is not logged in (no error occurred, just no user found in session)
		http.Error(w, "Unauthorized: Please log in.", http.StatusUnauthorized)
		return 0, false
	}

	// User is authenticated
	return user.ID, true
}
