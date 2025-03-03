package prodtable

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/usertable"
)

// setupTestDB initializes the database connection via coredbutils and runs migrations for the product, image,
// and user tables. It assumes that your test environment is configured to use a dedicated PostgreSQL test database.
//
// @Summary      Setup PostgreSQL test database
// @Description  Connects to the PostgreSQL database using coredbutils, and runs the necessary migrations.
// @Tags         testing, database, prodtable
func setupTestDB(t *testing.T) *gorm.DB {
	db := coredbutils.GetDB()
	// Run migrations for prodtable.
	db.Migrator().DropTable(&Product{})
	db.Migrator().DropTable(&Image{})
	// Create tables in proper order (Image first, then Product).
	if err := db.AutoMigrate(&Image{}, &Product{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
	// Also migrate the user table; assuming usertable.User is defined.
	if err := db.AutoMigrate(&usertable.User{}); err != nil {
		t.Fatalf("failed to migrate user table: %v", err)
	}
	return db
}

// createFakeUser inserts a fake user into the users table for testing purposes.
// The password is "testpassword".
// It clears the user table first to avoid duplicates.
//
// @Summary      Create fake user for tests
// @Description  Inserts a fake user into the users table with a known email and password.
// @Tags         testing, authentication, prodtable
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
//
// @Summary      Simulate user login for tests
// @Description  Uses the /api/login endpoint to log in the fake user and extract the authentication cookie.
// @Tags         testing, authentication, prodtable
func loginFakeUser(t *testing.T) *http.Cookie {
	form := url.Values{}
	form.Set("email", "fake@example.com")
	form.Set("password", "testpassword")
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// Call the login handler.
	login.LoginUser(rr, req)
	// Expect a redirect on successful login.
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect after login, got %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth" {
			authCookie = cookie
			break
		}
	}
	if authCookie == nil {
		t.Fatalf("auth cookie not found")
	}
	return authCookie
}

func TestMain(m *testing.M) {
	// Set the SESSION_SECRET before any package initialization occurs.
	os.Setenv("SESSION_SECRET", "testsecret")
	os.Exit(m.Run())
}

// TestAddProduct tests the AddProduct endpoint by simulating a multipart/form-data POST request
// that includes product details and an image file. It uses a valid session cookie from the fake user.
// It verifies that the product and associated image are stored in the database.
//
// @Summary      Test AddProduct endpoint with PostgreSQL
// @Description  Creates a fake user, logs in to obtain a valid session cookie, then sends a product creation request.
//
//	Checks that the product is inserted into the database.
//
// @Tags         testing, prodtable
func TestAddProduct(t *testing.T) {
	db := setupTestDB(t)
	// Create and log in fake user.
	_ = createFakeUser(t)
	cookie := loginFakeUser(t)

	// Prepare multipart form data.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("productName", "Test Product")
	writer.WriteField("description", "Test Description")
	writer.WriteField("price", "9.99")
	writer.WriteField("count", "5")
	writer.WriteField("tags", "test,product")
	part, err := writer.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	io.Copy(part, strings.NewReader("dummy image content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/add_product", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()

	AddProduct(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "Product added successfully" {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}

	// Verify that the product was inserted.
	var product Product
	result := db.First(&product, "prod_name = ?", "Test Product")
	if result.Error != nil {
		t.Errorf("failed to find product: %v", result.Error)
	}
	// Verify associated image.
	var image Image
	result = db.First(&image, product.ImgID)
	if result.Error != nil {
		t.Errorf("failed to find associated image: %v", result.Error)
	}
	// Cleanup the uploaded file.
	os.Remove(image.URL)
	// Clear products.
	db.Exec("DELETE FROM products")
}

// TestDeleteProduct tests the DeleteProduct endpoint by inserting a dummy product (with an associated image file)
// for the fake user, then simulating a deletion request with a valid session cookie.
// It verifies that the product is removed from the database and the image file is deleted.
//
// @Summary      Test DeleteProduct endpoint with PostgreSQL
// @Description  Creates a fake user and dummy product, then simulates a deletion request. Checks that the product and its image are removed.
// @Tags         testing, prodtable
func TestDeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	// Create and log in fake user.
	user := createFakeUser(t)
	cookie := loginFakeUser(t)

	// Create a temporary dummy image file.
	tmpFile, err := os.CreateTemp("", "test_image_*.png")
	if err != nil {
		t.Fatalf("failed to create temporary image file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	// Insert dummy image record.
	image := Image{URL: tmpFilePath}
	if err := db.Create(&image).Error; err != nil {
		t.Fatalf("failed to create image record: %v", err)
	}

	// Insert dummy product record associated with the fake user.
	product := Product{
		UserID:          user.ID, // Use the actual user ID
		ProdName:        "Delete Product",
		ProdDescription: "To be deleted",
		ProdPrice:       19.99,
		ProdCount:       3,
		ProdTags:        "delete,test",
		ImgID:           image.ID,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("failed to create product record: %v", err)
	}

	// Ensure the dummy image file exists.
	if _, err := os.Stat(tmpFilePath); os.IsNotExist(err) {
		t.Fatalf("dummy image file does not exist")
	}

	// Create a deletion request with query parameter "id" set to the product's ID.
	// Using GET method since the DeleteProduct handler appears to use GET based on its Swagger annotations
	req := httptest.NewRequest("GET", "/api/delete_product?id="+strconv.Itoa(int(product.ID)), nil)
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()

	DeleteProduct(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if strings.TrimSpace(rr.Body.String()) != "Product deleted successfully" {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}

	// Verify that the product was deleted.
	var count int64
	db.Model(&Product{}).Where("id = ?", product.ID).Count(&count)
	if count != 0 {
		t.Errorf("product was not deleted from the database")
	}

	// Verify that the image file was deleted.
	// Note: If the file doesn't exist anymore, that's actually what we want
	if _, err := os.Stat(tmpFilePath); !os.IsNotExist(err) {
		t.Errorf("expected image file to be removed, but it exists")
		os.Remove(tmpFilePath)
	}

	// Clear products.
	db.Exec("DELETE FROM products")
}

// TestUpdateProduct tests the UpdateProduct endpoint by creating a dummy product for the fake user,
// then simulating an update request with new description, price, and stock count.
// It verifies that the product is updated in the database.
//
// @Summary      Test UpdateProduct endpoint with PostgreSQL
// @Description  Creates a fake user and dummy product, then sends an update request to change product details.
//
//	Checks that the database record is updated accordingly.
//
// @Tags         testing, prodtable
func TestUpdateProduct(t *testing.T) {
	db := setupTestDB(t)
	// Create and log in fake user.
	user := createFakeUser(t)
	cookie := loginFakeUser(t)

	// Create a temporary dummy image file.
	tmpFile, err := os.CreateTemp("", "test_image_*.png")
	if err != nil {
		t.Fatalf("failed to create temporary image file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	// Insert dummy image record first
	image := Image{URL: tmpFilePath}
	if err := db.Create(&image).Error; err != nil {
		t.Fatalf("failed to create image record: %v", err)
	}

	// Insert a dummy product with the image ID.
	product := Product{
		UserID:          user.ID,
		ProdName:        "Update Product",
		ProdDescription: "Old description",
		ProdPrice:       10.00,
		ProdCount:       2,
		ProdTags:        "update,test",
		ImgID:           image.ID, // Set the ImgID to reference the created image
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Prepare form data for updating the product.
	form := url.Values{}
	form.Set("product_description", "New description")
	form.Set("item_price", "15.50")
	form.Set("stock_amount", "5")

	// Using PUT method as indicated in the UpdateProduct Swagger annotation
	req := httptest.NewRequest("PUT", "/api/update_product?id="+strconv.Itoa(int(product.ID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()

	UpdateProduct(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if strings.TrimSpace(rr.Body.String()) != "Product updated successfully" {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}

	// Verify that the product was updated.
	var updated Product
	if err := db.First(&updated, product.ID).Error; err != nil {
		t.Fatalf("failed to retrieve updated product: %v", err)
	}
	if updated.ProdDescription != "New description" {
		t.Errorf("expected description %q, got %q", "New description", updated.ProdDescription)
	}
	if updated.ProdPrice != 15.50 {
		t.Errorf("expected price %f, got %f", 15.50, updated.ProdPrice)
	}
	if updated.ProdCount != 5 {
		t.Errorf("expected count %d, got %d", 5, updated.ProdCount)
	}

	// Clean up the image file
	os.Remove(tmpFilePath)

	// Clear products and users.
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM images")
}
