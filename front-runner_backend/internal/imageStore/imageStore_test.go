package imageStore

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/usertable"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupTestDB connects to the PostgreSQL test database using coredbutils and runs migrations
// for the user table and the imageStore's ProductImage table.
// It assumes that your test environment is configured to use a dedicated PostgreSQL test database.
//
// @Summary      Setup PostgreSQL test database for imageStore
// @Description  Connects to the PostgreSQL test database and runs migrations for the user and ProductImage tables.
// @Tags         testing, database, imageStore
func setupTestDB(t *testing.T) *gorm.DB {
	db := coredbutils.GetDB()
	// Run migration for ProductImage.
	if err := db.AutoMigrate(&ProductImage{}); err != nil {
		t.Fatalf("failed to migrate ProductImage table: %v", err)
	}
	// Also migrate the user table.
	if err := db.AutoMigrate(&usertable.User{}); err != nil {
		t.Fatalf("failed to migrate user table: %v", err)
	}
	return db
}

// createFakeUser inserts a fake user into the user table for testing.
// The password is "testpassword". It clears the user table first to avoid duplicates.
//
// @Summary      Create fake user for imageStore tests
// @Description  Inserts a fake user into the users table with known credentials (email and password).
// @Tags         testing, authentication, imageStore
func createFakeUser(t *testing.T) *usertable.User {
	db := coredbutils.GetDB()
	// Clear the user table.
	db.Exec("DELETE FROM users")
	hash, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate password hash: %v", err)
	}
	user := &usertable.User{
		Email:        "fake@example.com",
		PasswordHash: string(hash),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create fake user: %v", err)
	}
	return user
}

// loginFakeUser simulates a login request for the fake user using the existing login endpoint,
// and returns the session cookie from the response.
// This cookie is required for authorized imageStore requests.
//
// @Summary      Simulate user login for imageStore tests
// @Description  Uses the /api/login endpoint to log in the fake user and extracts the authentication cookie.
// @Tags         testing, authentication, imageStore
func loginFakeUser(t *testing.T) *http.Cookie {
	form := url.Values{}
	form.Set("email", "fake@example.com")
	form.Set("password", "testpassword")
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// Call the login handler.
	login.LoginUser(rr, req)
	// Successful login redirects.
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect after login, got %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("no cookies set in login response")
	}
	return cookies[0]
}

// TestLoadImage_Unauthorized tests that LoadImage returns an unauthorized error when the user is not logged in.
//
// @Summary      Test LoadImage unauthorized access
// @Description  Verifies that calling LoadImage without a valid session cookie returns a 401 Unauthorized status.
// @Tags         testing, imageStore, images
func TestLoadImage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/data/image/somefile.png", nil)
	rr := httptest.NewRecorder()
	// Create mux vars to simulate a route variable.
	vars := map[string]string{"imagePath": "somefile.png"}
	req = mux.SetURLVars(req, vars)

	LoadImage(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "User is not logged in") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}
}

// TestLoadImage_InvalidFilename tests that LoadImage returns an error when the image record is not found.
//
// @Summary      Test LoadImage with invalid filename
// @Description  Verifies that a logged-in request for a non-existent image returns a 401 Unauthorized status with an appropriate error message.
// @Tags         testing, imageStore, images
func TestLoadImage_InvalidFilename(t *testing.T) {
	setupTestDB(t)
	createFakeUser(t)
	cookie := loginFakeUser(t)

	req := httptest.NewRequest("GET", "/api/data/image/nonexistent.png", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	vars := map[string]string{"imagePath": "nonexistent.png"}
	req = mux.SetURLVars(req, vars)

	LoadImage(rr, req)
	// In this implementation, if the record is not found, we return 401 with "Invalid filename".
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid filename") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}
}

// TestLoadImage_PermissionDenied tests that LoadImage returns a forbidden error when the image record exists
// but belongs to a different user.
//
// @Summary      Test LoadImage permission denied
// @Description  Inserts an image record with a different user ID than the logged-in user and verifies that LoadImage returns 403 Forbidden.
// @Tags         testing, imageStore, images
func TestLoadImage_PermissionDenied(t *testing.T) {
	db := setupTestDB(t)
	createFakeUser(t)
	cookie := loginFakeUser(t)

	// Insert an image record with a user ID different from the logged-in user (logged-in user has ID 1).
	imageRecord := ProductImage{
		Filename: "diffuser.png",
		ID:       999, // Different user ID.
	}
	if err := db.Create(&imageRecord).Error; err != nil {
		t.Fatalf("failed to create image record: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/data/image/diffuser.png", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	vars := map[string]string{"imagePath": "diffuser.png"}
	req = mux.SetURLVars(req, vars)

	LoadImage(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Permission denied") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}
	// Cleanup record.
	db.Delete(&ProductImage{}, "filename = ?", "diffuser.png")
}

// TestLoadImage_FileNotExist tests that LoadImage returns a 404 error when the image file does not exist,
// even if the image record exists and the user is authorized.
//
// @Summary      Test LoadImage with missing file
// @Description  Inserts an image record for which the file does not exist and verifies that LoadImage returns 404 Not Found.
// @Tags         testing, imageStore, images
func TestLoadImage_FileNotExist(t *testing.T) {
	db := setupTestDB(t)
	createFakeUser(t)
	cookie := loginFakeUser(t)

	// Insert an image record for a file that does not exist.
	imageRecord := ProductImage{
		Filename: "nonexistentfile.png",
		ID:       1, // Matching logged-in user.
	}
	if err := db.Create(&imageRecord).Error; err != nil {
		t.Fatalf("failed to create image record: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/data/image/nonexistentfile.png", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	vars := map[string]string{"imagePath": "nonexistentfile.png"}
	req = mux.SetURLVars(req, vars)

	LoadImage(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Requested image does not exist") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}
	// Cleanup record.
	db.Delete(&ProductImage{}, "filename = ?", "nonexistentfile.png")
}

// TestLoadImage_Success tests that LoadImage successfully serves the image file when all conditions are met.
//
// @Summary      Test successful LoadImage
// @Description  Inserts an image record with the matching user ID, creates a temporary file in the images directory,
//
//	and verifies that LoadImage serves the file with status 200.
//
// @Tags         testing, imageStore, images
func TestLoadImage_Success(t *testing.T) {
	db := setupTestDB(t)
	createFakeUser(t)
	cookie := loginFakeUser(t)

	// Create a temporary file in the "data/images" directory.
	imageDir := "data/images"
	tempFile, err := os.CreateTemp(imageDir, "testimage_*.png")
	if err != nil {
		t.Fatalf("failed to create temporary image file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.WriteString("fake image content")
	tempFile.Close()

	// Get the base filename.
	baseName := filepath.Base(tempFile.Name())
	// Insert an image record with matching user ID.
	imageRecord := ProductImage{
		Filename: baseName,
		ID:       1,
	}
	if err := db.Create(&imageRecord).Error; err != nil {
		t.Fatalf("failed to create image record: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/data/image/"+baseName, nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	vars := map[string]string{"imagePath": baseName}
	req = mux.SetURLVars(req, vars)

	LoadImage(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	// Optionally, check that the served file content matches.
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if !strings.Contains(string(body), "fake image content") {
		t.Errorf("expected response body to contain %q, got %q", "fake image content", string(body))
	}
	// Cleanup record.
	db.Delete(&ProductImage{}, "filename = ?", baseName)
}

// TestUploadImage_Unauthorized tests that UploadImage returns unauthorized when the user is not logged in.
//
// @Summary      Test UploadImage unauthorized access
// @Description  Verifies that a request to upload an image without a valid session returns 401 Unauthorized.
// @Tags         testing, imageStore, images
func TestUploadImage_Unauthorized(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("filename", "upload.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	// Write valid image header bytes.
	part.Write([]byte{0x89, 0x50, 0x4e, 0x47})
	writer.Close()

	req := httptest.NewRequest("POST", "/api/data/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	UploadImage(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "User is not logged in") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}
}

// TestUploadImage_InvalidFileType tests that UploadImage returns an error for non-image file uploads.
//
// @Summary      Test UploadImage with invalid file type
// @Description  Simulates an image upload with non-image content and verifies that the endpoint returns 415 Unsupported Media Type.
// @Tags         testing, imageStore, images
func TestUploadImage_InvalidFileType(t *testing.T) {
	setupTestDB(t)
	createFakeUser(t)
	cookie := loginFakeUser(t)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("filename", "notimage.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	// Write plain text data.
	part.Write([]byte("this is not an image"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/data/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()

	UploadImage(rr, req)
	if rr.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected status %d, got %d", http.StatusUnsupportedMediaType, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid file type") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}
}

// TestUploadImage_Success tests that UploadImage successfully uploads an image file.
// It verifies that the returned filename is non-empty and that a record is created in the database.
//
// @Summary      Test successful UploadImage
// @Description  Simulates a valid image upload with correct file content. Verifies that the endpoint returns a filename and a database record is created.
// @Tags         testing, imageStore, images
func TestUploadImage_Success(t *testing.T) {
	db := setupTestDB(t)
	createFakeUser(t)
	cookie := loginFakeUser(t)

	// Prepare a valid image file (simulate a PNG file with minimal header bytes).
	imageBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0D}
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("filename", "valid.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write(imageBytes)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/data/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()

	UploadImage(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	uploadedFilename := strings.TrimSpace(rr.Body.String())
	if uploadedFilename == "" {
		t.Fatalf("expected a filename in response, got empty string")
	}

	// Verify that the file was created in the "data/images" directory.
	uploadedFilePath := filepath.Join("data/images", uploadedFilename)
	if _, err := os.Stat(uploadedFilePath); os.IsNotExist(err) {
		t.Errorf("uploaded file does not exist at %s", uploadedFilePath)
	} else {
		// Clean up the uploaded file.
		os.Remove(uploadedFilePath)
	}

	// Verify that a record was created in the database.
	var record ProductImage
	result := db.First(&record, "filename = ?", uploadedFilename)
	if result.Error != nil {
		t.Errorf("failed to find uploaded image record: %v", result.Error)
	} else {
		// Cleanup the record.
		db.Delete(&ProductImage{}, "filename = ?", uploadedFilename)
	}
}
