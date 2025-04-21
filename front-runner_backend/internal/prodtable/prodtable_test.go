// internal/prodtable/prodtable_test.go
package prodtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"front-runner/internal/coredbutils"
	"front-runner/internal/login" // Needed for session constants/setup
	"front-runner/internal/oauth" // Needed for oauth.Setup
	"front-runner/internal/usertable"
)

const projectDirName = "front-runner_backend"

// Global test variables
var (
	testDB           *gorm.DB
	testSessionStore *sessions.CookieStore
	setupEnvOnce     sync.Once
	uploadsTestDir   string // Will hold path from t.TempDir()
)

// setupTestEnvironment loads environment variables, initializes DB and session store for tests.
// It also clears relevant tables before each test run.
func setupTestEnvironment(t *testing.T) {
	t.Helper()
	log.Println("--- Starting setupTestEnvironment ---") // Add this

	setupEnvOnce.Do(func() {
		log.Println("--- Running setupEnvOnce.Do ---") // Add this
		// Find project root
		re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
		cwd, _ := os.Getwd()
		rootPath := re.Find([]byte(cwd))
		if rootPath == nil {
			t.Fatalf("Could not find project root directory '%s' from '%s'", projectDirName, cwd)
		}

		// Load .env file
		envPath := string(rootPath) + `/.env`
		err := godotenv.Load(envPath)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Problem loading .env file from %s: %v", envPath, err)
		} else if err == nil {
			log.Printf("Loaded environment variables from %s for tests", envPath)
		}

		log.Println("--- Initializing DB ---") // Add this
		// Initialize DB connection
		coredbutils.ResetDBStateForTests()
		err = coredbutils.LoadEnv()
		require.NoError(t, err, "Failed to load core DB environment")
		var dbErr error
		testDB, dbErr = coredbutils.GetDB()
		require.NoError(t, dbErr, "Failed to get DB connection for tests")
		log.Println("--- DB Initialized ---") // Add this

		// Initialize Session Store for tests
		authKey := []byte("test-auth-key-32-bytes-long-000")
		encKey := []byte("test-enc-key-needs-to-be-32-byte") // 32 bytes
		require.True(t, len(encKey) == 16 || len(encKey) == 32, "Test encryption key must be 16 or 32 bytes")
		testSessionStore = sessions.NewCookieStore(authKey, encKey)
		testSessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}

		// Setup dependent packages
		usertable.Setup()                     // Uses coredbutils.GetDB()
		oauth.Setup(testSessionStore)         // Uses session store
		login.Setup(testDB, testSessionStore) // Uses DB and session store
		Setup()                               // Setup prodtable package (uses coredbutils.GetDB())

		// Run migrations once after setup
		usertable.MigrateUserDB()
		MigrateProdDB() // Migrates Product and Image tables
	})

	// Create a temporary directory for uploads for this test run
	// This replaces the manual createUploadsDir and os.Remove("uploads")
	uploadsTestDir = t.TempDir()
	// Override the default "uploads" path used in the main code *if necessary*
	// This is tricky. A better approach is to make the uploads path configurable.
	// For now, we assume the handlers write to "uploads" relative to CWD,
	// and we'll manage files within uploadsTestDir manually in tests.
	// If handlers used an absolute path or configurable path, we'd set it here.

	// Clear tables before each test function
	require.NoError(t, usertable.ClearUserTable(testDB), "Failed to clear user table")
	// ClearProdTable needs modification to handle potential errors better
	// Let's clear manually for now for better control
	require.NoError(t, testDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Product{}).Error, "Failed to clear product table")
	require.NoError(t, testDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Image{}).Error, "Failed to clear image table")

}

// Helper to create a test user directly in the DB
func createTestUser(t *testing.T, email, password string) *usertable.User {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password for test user")

	user := &usertable.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User " + email,
		Provider:     "local",
	}
	err = usertable.CreateUser(user)
	require.NoError(t, err, "Failed to create test user using usertable.CreateUser")
	createdUser, err := usertable.GetUserByEmail(email)
	require.NoError(t, err, "Failed to fetch created test user by email")
	require.NotNil(t, createdUser, "Fetched created test user should not be nil")
	return createdUser
}

// Helper to create an authenticated request
func createAuthenticatedRequest(t *testing.T, user *usertable.User, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, body)

	const sessionName = "front-runner-session"
	const userSessionKey = "userID"

	// Create and save session to get cookie header
	session, err := testSessionStore.New(req, sessionName) // Use constant from login pkg
	require.NoError(t, err, "Failed to create new session for auth request")
	session.Values[userSessionKey] = user.ID // Use constant from login pkg
	rrCookieSetter := httptest.NewRecorder()
	err = testSessionStore.Save(req, rrCookieSetter, session)
	require.NoError(t, err, "Failed to save session to get cookie header")
	cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookieHeader, "Set-Cookie header should not be empty after saving session")

	// Add cookie to the actual request
	req.Header.Set("Cookie", cookieHeader)
	return req
}

// Helper to create a dummy file for upload tests
func createDummyFile(t *testing.T, filename string, content string) string {
	t.Helper()
	// Use the temporary uploads directory for this test
	filePath := filepath.Join(uploadsTestDir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create dummy file")
	return filePath
}

// TestAddProduct tests the AddProduct endpoint.
func TestAddProduct(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "addprod@example.com", "password")

	// Prepare multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("productName", "Test Widget")
	_ = writer.WriteField("description", "A wonderful test widget.")
	_ = writer.WriteField("price", "19.95")
	_ = writer.WriteField("count", "100")
	_ = writer.WriteField("tags", "widget,test")

	// Create a dummy file part
	part, err := writer.CreateFormFile("image", "widget.png")
	require.NoError(t, err, "Failed to create form file")
	_, err = io.Copy(part, strings.NewReader("dummy png content"))
	require.NoError(t, err, "Failed to copy dummy content to form file")
	writer.Close() // Close writer to finalize multipart form

	// Create authenticated request
	req := createAuthenticatedRequest(t, user, "POST", "/api/add_product", &buf) // Assuming path from routes.go
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Call the handler
	AddProduct(rr, req)

	// Assertions
	require.Equal(t, http.StatusCreated, rr.Code, "Expected status 201 Created, got %d. Body: %s", rr.Code, rr.Body.String())
	assert.Equal(t, "Product added successfully", rr.Body.String())

	// Verify database insertion
	var product Product
	err = testDB.Preload("Img").Where("user_id = ? AND prod_name = ?", user.ID, "Test Widget").First(&product).Error
	require.NoError(t, err, "Failed to find created product in DB")
	assert.Equal(t, "A wonderful test widget.", product.ProdDescription)
	assert.Equal(t, 19.95, product.ProdPrice)
	assert.Equal(t, uint(100), product.ProdCount)
	assert.Equal(t, "widget,test", product.ProdTags)
	require.NotNil(t, product.Img, "Product should have an associated image")
	assert.NotEmpty(t, product.Img.URL, "Image URL should not be empty")
	assert.True(t, strings.HasSuffix(product.Img.URL, ".png"), "Image URL should have .png extension")

	// Verify file existence (using the actual uploads directory, assuming relative path)
	// This part is brittle if the handler doesn't use a predictable path.
	// Ideally, the handler would use uploadsTestDir if configurable.
	// For now, check relative to CWD.
	imagePath := filepath.Join("uploads", product.Img.URL) // Assuming handler writes to "uploads" relative to CWD
	_, err = os.Stat(imagePath)
	assert.NoError(t, err, "Expected image file '%s' to exist, but stat failed: %v", imagePath, err)
	// Cleanup the created file
	_ = os.Remove(imagePath)
}

// TestDeleteProduct tests the DeleteProduct endpoint.
func TestDeleteProduct(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "delprod@example.com", "password")

	// Create dummy image and product directly in DB
	imageFilename := uuid.NewString() + ".jpg"
	// Create the dummy file in the actual uploads dir for the handler hook to find
	imagePath := filepath.Join("uploads", imageFilename) // Assuming relative path
	err := os.MkdirAll("uploads", 0755)                  // Ensure dir exists
	require.NoError(t, err)
	err = os.WriteFile(imagePath, []byte("dummy jpg"), 0644)
	require.NoError(t, err)

	image := Image{URL: imageFilename, UserID: user.ID}
	err = testDB.Create(&image).Error
	require.NoError(t, err, "Failed to create test image record")

	product := Product{
		UserID:          user.ID,
		ProdName:        "Product To Delete",
		ProdDescription: "Delete me",
		ImgID:           image.ID,
		ProdPrice:       1.00,
		ProdCount:       1,
	}
	err = testDB.Create(&product).Error
	require.NoError(t, err, "Failed to create test product record")

	// Create authenticated request
	targetURL := fmt.Sprintf("/api/delete_product?id=%d", product.ID) // Assuming path from routes.go
	req := createAuthenticatedRequest(t, user, "DELETE", targetURL, nil)
	rr := httptest.NewRecorder()

	// Call the handler
	DeleteProduct(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	assert.Equal(t, "Product deleted successfully", rr.Body.String())

	// Verify product deletion from DB
	var count int64
	err = testDB.Model(&Product{}).Where("id = ?", product.ID).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Product should have been deleted from DB")

	// Verify image record deletion from DB
	err = testDB.Model(&Image{}).Where("id = ?", image.ID).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Image record should have been deleted from DB")

	// Verify file deletion
	_, err = os.Stat(imagePath)
	assert.True(t, os.IsNotExist(err), "Expected image file '%s' to be deleted, but it still exists", imagePath)
	_ = os.RemoveAll("uploads") // Clean up dir just in case
}

// TestUpdateProduct tests the UpdateProduct endpoint.
func TestUpdateProduct(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "updprod@example.com", "password")

	// Create initial image and product
	initialImageFilename := uuid.NewString() + ".gif"
	initialImagePath := filepath.Join("uploads", initialImageFilename) // Assuming relative path
	err := os.MkdirAll("uploads", 0755)
	require.NoError(t, err)
	err = os.WriteFile(initialImagePath, []byte("old gif"), 0644)
	require.NoError(t, err)

	image := Image{URL: initialImageFilename, UserID: user.ID}
	err = testDB.Create(&image).Error
	require.NoError(t, err)

	product := Product{
		UserID:          user.ID,
		ProdName:        "Product To Update",
		ProdDescription: "Old Description",
		ImgID:           image.ID,
		ProdPrice:       5.00,
		ProdCount:       5,
		ProdTags:        "old,tag",
	}
	err = testDB.Create(&product).Error
	require.NoError(t, err)

	// Prepare update data (including a new image)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("description", "New Updated Description") // Match form field name used in handler
	_ = writer.WriteField("price", "9.99")
	_ = writer.WriteField("count", "50")
	_ = writer.WriteField("tags", "new,updated")
	part, err := writer.CreateFormFile("image", "new_image.webp")
	require.NoError(t, err)
	_, err = io.Copy(part, strings.NewReader("new webp content"))
	require.NoError(t, err)
	writer.Close()

	// Create authenticated request
	targetURL := fmt.Sprintf("/api/update_product?id=%d", product.ID) // Assuming path from routes.go
	req := createAuthenticatedRequest(t, user, "PUT", targetURL, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Call the handler
	UpdateProduct(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	assert.Equal(t, "Product updated successfully", rr.Body.String())

	// Verify database update
	var updatedProduct Product
	err = testDB.Preload("Img").Where("id = ?", product.ID).First(&updatedProduct).Error
	require.NoError(t, err, "Failed to find updated product in DB")
	assert.Equal(t, "New Updated Description", updatedProduct.ProdDescription)
	assert.Equal(t, 9.99, updatedProduct.ProdPrice)
	assert.Equal(t, uint(50), updatedProduct.ProdCount)
	assert.Equal(t, "new,updated", updatedProduct.ProdTags)
	assert.Equal(t, "Product To Update", updatedProduct.ProdName) // Name wasn't updated
	require.NotNil(t, updatedProduct.Img, "Updated product should still have an image")
	assert.True(t, strings.HasSuffix(updatedProduct.Img.URL, ".webp"), "Image URL should have been updated to .webp")
	assert.NotEqual(t, initialImageFilename, updatedProduct.Img.URL, "Image URL should have changed")

	// Verify file changes
	_, err = os.Stat(initialImagePath)
	assert.True(t, os.IsNotExist(err), "Expected old image file '%s' to be deleted", initialImagePath)
	newImagePath := filepath.Join("uploads", updatedProduct.Img.URL)
	_, err = os.Stat(newImagePath)
	assert.NoError(t, err, "Expected new image file '%s' to exist", newImagePath)
	_ = os.RemoveAll("uploads") // Cleanup
}

// TestGetProduct tests the GetProduct endpoint.
func TestGetProduct(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "getprod@example.com", "password")

	// Create dummy image and product
	imageFilename := uuid.NewString() + ".png"
	image := Image{URL: imageFilename, UserID: user.ID}
	err := testDB.Create(&image).Error
	require.NoError(t, err)

	product := Product{
		UserID:          user.ID,
		ProdName:        "Specific Product",
		ProdDescription: "Details here",
		ImgID:           image.ID,
		ProdPrice:       12.34,
		ProdCount:       12,
		ProdTags:        "get,specific",
	}
	err = testDB.Create(&product).Error
	require.NoError(t, err)

	// Create authenticated request
	// Use the /api/products/details path based on previous Swagger doc refinement
	targetURL := fmt.Sprintf("/api/get_product?id=%d", product.ID) // Assuming path from routes.go
	req := createAuthenticatedRequest(t, user, "GET", targetURL, nil)
	rr := httptest.NewRecorder()

	// Call the handler
	GetProduct(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Decode JSON response
	var returnedProduct ProductReturn
	err = json.Unmarshal(rr.Body.Bytes(), &returnedProduct)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	// Verify content
	assert.Equal(t, product.ID, returnedProduct.ProdID)
	assert.Equal(t, "Specific Product", returnedProduct.ProdName)
	assert.Equal(t, "Details here", returnedProduct.ProdDescription)
	assert.Equal(t, 12.34, returnedProduct.ProdPrice)
	assert.Equal(t, uint(12), returnedProduct.ProdCount)
	assert.Equal(t, "get,specific", returnedProduct.ProdTags)
	assert.Equal(t, imageFilename, returnedProduct.ImgPath) // Check filename matches
}

// TestGetProducts tests the GetProducts endpoint.
func TestGetProducts(t *testing.T) {
	setupTestEnvironment(t)
	user1 := createTestUser(t, "getprods1@example.com", "password")
	user2 := createTestUser(t, "getprods2@example.com", "password") // Another user

	// Create products for user1
	img1 := Image{URL: uuid.NewString() + ".tga", UserID: user1.ID}
	require.NoError(t, testDB.Create(&img1).Error)
	prod1 := Product{UserID: user1.ID, ProdName: "User1 Prod A", ImgID: img1.ID, ProdPrice: 1.00}
	require.NoError(t, testDB.Create(&prod1).Error)

	img2 := Image{URL: uuid.NewString() + ".bmp", UserID: user1.ID}
	require.NoError(t, testDB.Create(&img2).Error)
	prod2 := Product{UserID: user1.ID, ProdName: "User1 Prod B", ImgID: img2.ID, ProdPrice: 2.00}
	require.NoError(t, testDB.Create(&prod2).Error)

	// Create product for user2 (should not be returned)
	img3 := Image{URL: uuid.NewString() + ".pcx", UserID: user2.ID}
	require.NoError(t, testDB.Create(&img3).Error)
	prod3 := Product{UserID: user2.ID, ProdName: "User2 Prod C", ImgID: img3.ID, ProdPrice: 3.00}
	require.NoError(t, testDB.Create(&prod3).Error)

	// Create authenticated request for user1
	req := createAuthenticatedRequest(t, user1, "GET", "/api/get_products", nil) // Assuming path from routes.go
	rr := httptest.NewRecorder()

	// Call the handler
	GetProducts(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Decode JSON response
	var returnedProducts []ProductReturn
	err := json.Unmarshal(rr.Body.Bytes(), &returnedProducts)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	// Verify content
	require.Len(t, returnedProducts, 2, "Expected 2 products for user1")
	// Check if the correct products were returned (order might vary, check names)
	foundA := false
	foundB := false
	for _, p := range returnedProducts {
		if p.ProdName == "User1 Prod A" {
			foundA = true
			assert.Equal(t, prod1.ID, p.ProdID)
			assert.Equal(t, img1.URL, p.ImgPath)
		}
		if p.ProdName == "User1 Prod B" {
			foundB = true
			assert.Equal(t, prod2.ID, p.ProdID)
			assert.Equal(t, img2.URL, p.ImgPath)
		}
	}
	assert.True(t, foundA, "Product A not found in response")
	assert.True(t, foundB, "Product B not found in response")
}

// TestGetProductImage tests the GetProductImage endpoint.
func TestGetProductImage(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "getimg@example.com", "password")

	// Create dummy image file and record
	imageFilename := uuid.NewString() + ".txt"
	imageContent := "This is the image content."
	// Create the dummy file in the actual uploads dir for the handler to find
	imagePath := filepath.Join("uploads", imageFilename) // Assuming relative path
	err := os.MkdirAll("uploads", 0755)
	require.NoError(t, err)
	err = os.WriteFile(imagePath, []byte(imageContent), 0644)
	require.NoError(t, err)

	image := Image{URL: imageFilename, UserID: user.ID}
	err = testDB.Create(&image).Error
	require.NoError(t, err)

	// Create associated product (not strictly necessary for this handler, but good practice)
	product := Product{UserID: user.ID, ProdName: "Image Test Prod", ImgID: image.ID}
	err = testDB.Create(&product).Error
	require.NoError(t, err)

	// Create authenticated request
	targetURL := fmt.Sprintf("/api/get_product_image?image=%s", imageFilename) // Assuming path from routes.go
	req := createAuthenticatedRequest(t, user, "GET", targetURL, nil)
	rr := httptest.NewRecorder()

	// Call the handler
	GetProductImage(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK, got %d", rr.Code)
	// http.ServeFile should set Content-Type based on extension (or sniffing)
	// For .txt it might be text/plain;charset=utf-8
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/plain", "Expected Content-Type for .txt file")
	assert.Equal(t, imageContent, rr.Body.String(), "Response body should match file content")

	_ = os.RemoveAll("uploads") // Cleanup
}

// TestGetProductImage_NotFound tests image not found scenarios.
func TestGetProductImage_NotFound(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUser(t, "imgnotfound@example.com", "password")

	t.Run("DBRecordNotFound", func(t *testing.T) {
		targetURL := "/api/get_product_image?image=nonexistent.jpg"
		req := createAuthenticatedRequest(t, user, "GET", targetURL, nil)
		rr := httptest.NewRecorder()
		GetProductImage(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Image metadata not found")
	})

	t.Run("FileNotOnDisk", func(t *testing.T) {
		// Create DB record but no file
		imageFilename := uuid.NewString() + ".dat"
		image := Image{URL: imageFilename, UserID: user.ID}
		err := testDB.Create(&image).Error
		require.NoError(t, err)

		targetURL := fmt.Sprintf("/api/get_product_image?image=%s", imageFilename)
		req := createAuthenticatedRequest(t, user, "GET", targetURL, nil)
		rr := httptest.NewRecorder()
		GetProductImage(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Image file not found")
	})
}

// TestProduct_Auth checks authentication/authorization failures.
func TestProduct_Auth(t *testing.T) {
	setupTestEnvironment(t)
	user1 := createTestUser(t, "user1@example.com", "password")
	user2 := createTestUser(t, "user2@example.com", "password")

	// Create product owned by user1
	img1 := Image{URL: uuid.NewString() + ".png", UserID: user1.ID}
	require.NoError(t, testDB.Create(&img1).Error)
	prod1 := Product{UserID: user1.ID, ProdName: "User1 Secret Prod", ImgID: img1.ID}
	require.NoError(t, testDB.Create(&prod1).Error)

	// --- Test Cases ---
	testCases := []struct {
		name           string
		method         string
		urlFunc        func() string // Function to generate URL
		body           io.Reader
		contentType    string
		requestingUser *usertable.User // nil for unauthenticated
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "AddProduct Unauthenticated",
			method:         "POST",
			urlFunc:        func() string { return "/api/add_product" },
			body:           strings.NewReader("productName=fail"), // Minimal body
			contentType:    "application/x-www-form-urlencoded",   // Incorrect type, but auth fails first
			requestingUser: nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User not authenticated",
		},
		{
			name:           "DeleteProduct Unauthenticated",
			method:         "DELETE",
			urlFunc:        func() string { return fmt.Sprintf("/api/delete_product?id=%d", prod1.ID) },
			requestingUser: nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User not authenticated",
		},
		{
			name:           "DeleteProduct Wrong User",
			method:         "DELETE",
			urlFunc:        func() string { return fmt.Sprintf("/api/delete_product?id=%d", prod1.ID) },
			requestingUser: user2, // User2 tries to delete User1's product
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Unauthorized: You do not own this product",
		},
		{
			name:           "GetProduct Unauthenticated",
			method:         "GET",
			urlFunc:        func() string { return fmt.Sprintf("/api/get_product?id=%d", prod1.ID) },
			requestingUser: nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User not authenticated",
		},
		{
			name:           "GetProduct Wrong User",
			method:         "GET",
			urlFunc:        func() string { return fmt.Sprintf("/api/get_product?id=%d", prod1.ID) },
			requestingUser: user2,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Permission denied: You do not own this product",
		},
		// Add more for Update, GetProducts, GetProductImage if needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.requestingUser != nil {
				req = createAuthenticatedRequest(t, tc.requestingUser, tc.method, tc.urlFunc(), tc.body)
			} else {
				req = httptest.NewRequest(tc.method, tc.urlFunc(), tc.body)
			}
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			rr := httptest.NewRecorder()

			// Need to map URL/Method to the correct handler
			// This is slightly awkward without a router in the test.
			// Consider creating a test router if this becomes complex.
			switch {
			case strings.HasPrefix(req.URL.Path, "/api/add_product") && req.Method == "POST":
				AddProduct(rr, req)
			case strings.HasPrefix(req.URL.Path, "/api/delete_product") && req.Method == "DELETE":
				DeleteProduct(rr, req)
			case strings.HasPrefix(req.URL.Path, "/api/get_product") && req.Method == "GET":
				GetProduct(rr, req)
				// Add other handlers here...
			default:
				t.Fatalf("No handler mapped for test case: %s %s", tc.method, tc.urlFunc())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code, "Expected status %d, got %d", tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tc.expectedBody, "Response body mismatch")
			}
		})
	}
	_ = os.RemoveAll("uploads") // Cleanup
}
