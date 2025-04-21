package prodtable

import (
	"encoding/json"
	"errors"
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/oauth" // Import oauth

	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Image struct definition
type Image struct {
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	URL    string `gorm:"unique;not null"`
	UserID uint   `gorm:"not null;index"`
}

// Product struct definition
type Product struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"not null;index:idx_product,unique"`
	ProdName        string `gorm:"not null;index:idx_product,unique"`
	ProdDescription string `gorm:"not null"`
	ImgID           uint
	Img             Image `gorm:"foreignKey:ImgID"`
	ProdPrice       float64
	ProdCount       uint
	ProdTags        string
}

// AfterDelete hook to clean up associated image file and record.
func (p *Product) AfterDelete(tx *gorm.DB) (err error) {
	// Find the image associated with the product
	var img Image
	if err := tx.First(&img, "id = ?", p.ImgID).Error; err == nil {
		// Delete the image file from disk
		imagePath := filepath.Join("uploads", img.URL)
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			// Log error if deletion fails for reasons other than file not existing
			log.Printf("Error deleting image file %s during product deletion: %v", imagePath, err)
			// Decide if this should halt the transaction or just be logged
			// return fmt.Errorf("failed to delete image file: %w", err) // Uncomment to make it critical
		}
		// Delete the image record from the database
		if err := tx.Delete(&Image{}, "id = ?", p.ImgID).Error; err != nil {
			log.Printf("Error deleting image record %d during product deletion: %v", p.ImgID, err)
			return err // Return DB error
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Log error if fetching image fails for reasons other than not found
		log.Printf("Error finding image record %d during product deletion: %v", p.ImgID, err)
		return err // Return DB error
	}
	// If image not found or successfully deleted, proceed
	return nil
}

var (
	db        *gorm.DB
	setupOnce sync.Once
)

// Setup initializes database connection and ensures uploads directory exists.
func Setup() {
	setupOnce.Do(func() {
		coredbutils.LoadEnv() // Ensure env vars are loaded if needed by GetDB
		var err error
		db, err = coredbutils.GetDB() // Get the singleton DB instance
		if err != nil {
			// If GetDB can fail, handle it here. Assuming it fatals or test setup checks it.
			log.Fatalf("prodtable Setup: Failed to get database connection: %v", err)
		}
		if db == nil {
			log.Fatal("prodtable Setup: Database connection is nil after GetDB.")
		}
		log.Println("prodtable package setup complete (DB connection obtained).")
	})

	// Ensure uploads directory exists
	uploadsDir := "uploads"
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		log.Printf("Creating uploads directory: %s", uploadsDir)
		if err := os.Mkdir(uploadsDir, 0755); err != nil { // Use 0755 permissions
			log.Fatalf("Failed to create uploads directory '%s': %v", uploadsDir, err)
		}
	} else if err != nil {
		log.Fatalf("Error checking uploads directory '%s': %v", uploadsDir, err)
	}
}

// doesFileExist checks if a file exists at the given path.
func doesFileExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return !errors.Is(err, os.ErrNotExist)
}

// MigrateProdDB runs GORM auto-migration for Product and Image models.
func MigrateProdDB() {
	if db == nil {
		log.Fatal("Database connection is not initialized")
	}
	log.Println("Running product and image database migrations...")
	err := db.AutoMigrate(&Product{}, &Image{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Product and Image database migration complete")
}

// ClearProdTable removes all records from the Product and Image tables. USE WITH CAUTION.
func ClearProdTable(db *gorm.DB) error {
	// It's safer to delete images first if there's no strict foreign key constraint ensuring cascade delete
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Image{}).Error; err != nil {
		return fmt.Errorf("error clearing images table: %w", err)
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Product{}).Error; err != nil {
		return fmt.Errorf("error clearing product table: %w", err)
	}
	// Optionally clear the uploads directory
	// files, err := filepath.Glob(filepath.Join("uploads", "*"))
	// if err == nil {
	// 	for _, f := range files {
	// 		os.Remove(f)
	// 	}
	// }
	return nil
}

// AddProduct creates a new product and associates it with the logged-in user.
// It expects product details and an image file via multipart/form-data.
//
// @Summary      Add a new product
// @Description  Creates a new product listing associated with the authenticated user. Requires product details and an image upload.
// @Tags         Products
// @Accept       multipart/form-data
// @Produce      text/plain
// @Param        productName  formData  string  true  "Name of the product"
// @Param        description  formData  string  true  "Description of the product"
// @Param        price        formData  number  true  "Price of the product (e.g., 19.99)" Format(float)
// @Param        count        formData  integer true  "Available stock count" Format(int32)
// @Param        tags         formData  string  false "Comma-separated tags for the product"
// @Param        image        formData  file    true  "Product image file"
// @Success      201  {string}  string "Product added successfully"
// @Failure      400  {string}  string "Bad Request: Missing required fields, invalid data format, or image error"
// @Failure      401  {string}  string "Unauthorized: User not authenticated"
// @Failure      500  {string}  string "Internal Server Error: Database or file system error"
// @Security     ApiKeyAuth
// @Router       /api/products [post]
func AddProduct(w http.ResponseWriter, r *http.Request) {
	// --- Updated Auth Check ---
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("AddProduct: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID // Use the ID from the authenticated user
	// --- End Updated Auth Check ---

	err = r.ParseMultipartForm(10 << 20) // Limit to 10MB
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	productName := r.FormValue("productName")
	productDescription := r.FormValue("description")
	productPriceStr := r.FormValue("price")
	productCountStr := r.FormValue("count")
	productTags := r.FormValue("tags")

	// Basic validation for required fields
	if productName == "" || productDescription == "" || productPriceStr == "" || productCountStr == "" {
		http.Error(w, "Missing required fields (productName, description, price, count)", http.StatusBadRequest)
		return
	}

	productPrice, err := strconv.ParseFloat(productPriceStr, 64)
	if err != nil || productPrice < 0 {
		http.Error(w, "Invalid product price", http.StatusBadRequest)
		return
	}
	productCount, err := strconv.Atoi(productCountStr)
	if err != nil || productCount < 0 {
		http.Error(w, "Invalid product count", http.StatusBadRequest)
		return
	}

	// Handle File Upload
	file, handler, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			http.Error(w, "Product image is required", http.StatusBadRequest)
		} else {
			http.Error(w, "Error retrieving the image file: "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	defer file.Close()

	// Generate unique filename and save image
	imageFilename := uuid.New().String() + filepath.Ext(handler.Filename)
	imagePath := filepath.Join("uploads", imageFilename)
	dst, err := os.Create(imagePath)
	if err != nil {
		log.Printf("Error creating image file %s: %v", imagePath, err)
		http.Error(w, "Error saving image", http.StatusInternalServerError)
		return
	}
	defer dst.Close() // Ensure destination file is closed

	_, err = io.Copy(dst, file) // Copy file content
	if err != nil {
		log.Printf("Error copying image data to %s: %v", imagePath, err)
		http.Error(w, "Error saving image data", http.StatusInternalServerError)
		// Attempt to clean up partially created file
		os.Remove(imagePath)
		return
	}
	// Explicitly close dst here before proceeding, to ensure data is flushed
	if err := dst.Close(); err != nil {
		log.Printf("Error closing image file %s after write: %v", imagePath, err)
		// Continue, but log the error
	}

	// Use transaction for database operations
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		os.Remove(imagePath) // Clean up saved image file
		return
	}

	// Save Image record
	image := Image{
		URL:    imageFilename,
		UserID: userID,
	}
	if err := tx.Create(&image).Error; err != nil {
		tx.Rollback()
		log.Printf("Error saving image record for user %d: %v", userID, err)
		http.Error(w, "Error saving image metadata", http.StatusInternalServerError)
		os.Remove(imagePath) // Clean up saved image file
		return
	}

	// Create Product record
	product := Product{
		UserID:          userID,
		ProdName:        productName,
		ProdDescription: productDescription,
		ProdPrice:       productPrice,
		ProdCount:       uint(productCount),
		ProdTags:        productTags,
		ImgID:           image.ID, // Link the image ID
	}
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		log.Printf("Error saving product record for user %d: %v", userID, err)
		http.Error(w, "Error saving product", http.StatusInternalServerError)
		os.Remove(imagePath) // Clean up saved image file
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback() // Attempt rollback on commit failure (might be redundant)
		log.Printf("Failed to commit transaction for user %d: %v", userID, err)
		http.Error(w, "Database error during commit", http.StatusInternalServerError)
		os.Remove(imagePath) // Clean up saved image file
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Product added successfully") // Use fmt.Fprint for consistency
}

// DeleteProduct removes a product if it belongs to the logged-in user.
// It uses the product's ID provided as a query parameter.
//
// @Summary      Delete a product
// @Description  Deletes a specific product owned by the authenticated user, identified by its ID. Also deletes the associated image file and record.
// @Tags         Products
// @Produce      text/plain
// @Param        id   query     int  true  "ID of the product to delete" Format(uint64)
// @Success      200  {string}  string "Product deleted successfully"
// @Failure      400  {string}  string "Bad Request: Invalid Product ID"
// @Failure      401  {string}  string "Unauthorized: User not authenticated"
// @Failure      403  {string}  string "Forbidden: User does not own this product"
// @Failure      404  {string}  string "Not Found: Product not found"
// @Failure      500  {string}  string "Internal Server Error: Database or file system error during deletion"
// @Security     ApiKeyAuth
// @Router       /api/products [delete]
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// --- Updated Auth Check ---
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("DeleteProduct: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID
	// --- End Updated Auth Check ---

	productIDStr := r.URL.Query().Get("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	var product Product
	// Use transaction for find and delete
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction for delete: %v", tx.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Find the product, preloading the image for deletion hook/file cleanup
	if err := tx.Preload("Img").First(&product, "id = ?", uint(productID)).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			log.Printf("Error finding product %d for delete: %v", productID, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Check ownership
	if product.UserID != userID {
		tx.Rollback()
		http.Error(w, "Unauthorized: You do not own this product", http.StatusForbidden) // Use 403 Forbidden
		return
	}

	// Delete the product (AfterDelete hook handles image record and file)
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback()
		log.Printf("Error deleting product %d: %v", productID, err)
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback() // Attempt rollback
		log.Printf("Failed to commit transaction for delete product %d: %v", productID, err)
		http.Error(w, "Database error during commit", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Product deleted successfully")
}

// UpdateProduct updates an existing product's details if it belongs to the logged-in user.
// It accepts optional fields via multipart/form-data and an optional new image.
//
// @Summary      Update a product
// @Description  Updates details (name, description, price, count, tags) and/or the image for a specific product owned by the authenticated user. Fields not provided are left unchanged.
// @Tags         Products
// @Accept       multipart/form-data
// @Produce      text/plain
// @Param        id           query     int     true  "ID of the product to update" Format(uint64)
// @Param        productName  formData  string  false "New name for the product"
// @Param        description  formData  string  false "New description for the product"
// @Param        price        formData  number  false "New price for the product (e.g., 29.99)" Format(float)
// @Param        count        formData  integer false "New available stock count" Format(int32)
// @Param        tags         formData  string  false "New comma-separated tags (replaces old tags)"
// @Param        image        formData  file    false "New product image file (replaces old image)"
// @Success      200  {string}  string "Product updated successfully"
// @Failure      400  {string}  string "Bad Request: Invalid Product ID or data format"
// @Failure      401  {string}  string "Unauthorized: User not authenticated"
// @Failure      403  {string}  string "Forbidden: User does not own this product"
// @Failure      404  {string}  string "Not Found: Product not found"
// @Failure      409  {string}  string "Conflict: Product name already exists for this user" // If name is updated
// @Failure      500  {string}  string "Internal Server Error: Database or file system error during update"
// @Security     ApiKeyAuth
// @Router       /api/products [put] // Or PATCH if partial updates are the primary intent
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// --- Updated Auth Check ---
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("UpdateProduct: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID
	// --- End Updated Auth Check ---

	productIDStr := r.URL.Query().Get("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	// Use transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction for update: %v", tx.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var product Product
	// Find the product, preloading image info
	if err := tx.Preload("Img").First(&product, "id = ?", uint(productID)).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			log.Printf("Error finding product %d for update: %v", productID, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Check ownership
	if product.UserID != userID {
		tx.Rollback()
		http.Error(w, "Unauthorized: You do not own this product", http.StatusForbidden)
		return
	}

	// Parse the multipart form for updates
	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Prepare map for product updates
	productUpdates := map[string]interface{}{}
	if productName := r.FormValue("productName"); productName != "" {
		productUpdates["ProdName"] = productName
	}
	// Note: Swagger doc uses 'product_description', form likely uses 'description' based on AddProduct
	if productDescription := r.FormValue("description"); productDescription != "" {
		productUpdates["ProdDescription"] = productDescription
	}
	// Note: Swagger doc uses 'item_price', form likely uses 'price'
	if productPriceStr := r.FormValue("price"); productPriceStr != "" {
		if productPrice, err := strconv.ParseFloat(productPriceStr, 64); err == nil && productPrice >= 0 {
			productUpdates["ProdPrice"] = productPrice
		} else {
			tx.Rollback()
			http.Error(w, "Invalid product price format", http.StatusBadRequest)
			return
		}
	}
	// Note: Swagger doc uses 'stock_amount', form likely uses 'count'
	if productCountStr := r.FormValue("count"); productCountStr != "" {
		if productCount, err := strconv.Atoi(productCountStr); err == nil && productCount >= 0 {
			productUpdates["ProdCount"] = uint(productCount)
		} else {
			tx.Rollback()
			http.Error(w, "Invalid product count format", http.StatusBadRequest)
			return
		}
	}
	if productTags := r.FormValue("tags"); productTags != "" { // Allow empty tags to clear? Decide policy.
		productUpdates["ProdTags"] = productTags
	}

	// Handle image update if provided
	file, handler, err := r.FormFile("image")
	newImageFilename := ""
	newImagePath := ""
	if err == nil { // New image provided
		defer file.Close()

		// Save new image
		newImageFilename = uuid.New().String() + filepath.Ext(handler.Filename)
		newImagePath = filepath.Join("uploads", newImageFilename)
		dst, err := os.Create(newImagePath)
		if err != nil {
			tx.Rollback()
			log.Printf("Error creating new image file %s for update: %v", newImagePath, err)
			http.Error(w, "Error saving new image", http.StatusInternalServerError)
			return
		}
		defer dst.Close() // Ensure destination file is closed

		_, err = io.Copy(dst, file)
		if err != nil {
			tx.Rollback()
			log.Printf("Error copying new image data to %s: %v", newImagePath, err)
			http.Error(w, "Error saving new image data", http.StatusInternalServerError)
			os.Remove(newImagePath) // Clean up partially created file
			return
		}
		// Explicitly close dst here before proceeding
		if err := dst.Close(); err != nil {
			log.Printf("Error closing new image file %s after write: %v", newImagePath, err)
			// Continue, but log the error
		}

		// Update image record in DB
		imageUpdates := map[string]interface{}{"URL": newImageFilename}
		if err := tx.Model(&Image{}).Where("id = ?", product.ImgID).Updates(imageUpdates).Error; err != nil {
			tx.Rollback()
			log.Printf("Error updating image record %d in DB: %v", product.ImgID, err)
			http.Error(w, "Error updating image metadata", http.StatusInternalServerError)
			os.Remove(newImagePath) // Clean up new image file
			return
		}

		// Delete old image file *after* DB update is successful
		if product.Img.URL != "" && product.Img.URL != newImageFilename {
			oldImagePath := filepath.Join("uploads", product.Img.URL)
			if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
				// Log error but don't fail the request, DB is updated
				log.Printf("Warning: Failed to delete old image file %s during update: %v", oldImagePath, err)
			}
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		// Error occurred other than missing file
		tx.Rollback()
		http.Error(w, "Error processing image upload: "+err.Error(), http.StatusBadRequest)
		return
	}
	// If err is http.ErrMissingFile, proceed without image update

	// Apply product updates if any were provided
	if len(productUpdates) > 0 {
		if err := tx.Model(&product).Updates(productUpdates).Error; err != nil {
			tx.Rollback()
			log.Printf("Error updating product %d: %v", productID, err)
			http.Error(w, "Error updating product details", http.StatusInternalServerError)
			if newImagePath != "" {
				os.Remove(newImagePath)
			} // Clean up new image if product update failed
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback() // Attempt rollback
		log.Printf("Failed to commit transaction for update product %d: %v", productID, err)
		http.Error(w, "Database error during commit", http.StatusInternalServerError)
		if newImagePath != "" {
			os.Remove(newImagePath)
		} // Clean up new image if commit failed
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Product updated successfully")
}

// ProductReturn struct definition for returning product data via the API.
type ProductReturn struct {
	ProdID          uint    `json:"prodID"`
	ProdName        string  `json:"prodName"`
	ProdDescription string  `json:"prodDesc"`
	ImgPath         string  `json:"image"` // Consider renaming to imageURL or similar
	ProdPrice       float64 `json:"prodPrice"`
	ProdCount       uint    `json:"prodCount"`
	ProdTags        string  `json:"prodTags"`
}

// setProductReturn converts a Product DB model to a ProductReturn API model.
func setProductReturn(product Product) ProductReturn {
	var ret ProductReturn
	ret.ProdID = product.ID
	ret.ProdName = product.ProdName
	ret.ProdDescription = product.ProdDescription
	ret.ImgPath = product.Img.URL // This is just the filename, client needs to construct full URL or use GetProductImage
	ret.ProdPrice = product.ProdPrice
	ret.ProdCount = product.ProdCount
	ret.ProdTags = product.ProdTags
	return ret
}

// GetProduct retrieves the information about a specified product if it belongs to the logged-in user.
//
// @Summary      Get a specific product
// @Description  Retrieves details for a specific product owned by the authenticated user, identified by its ID.
// @Tags         Products
// @Produce      application/json
// @Param        id   query     int  true  "ID of the product to retrieve" Format(uint64)
// @Success      200  {object}  ProductReturn "Successfully retrieved product details"
// @Failure      400  {string}  string "Bad Request: Invalid Product ID"
// @Failure      401  {string}  string "Unauthorized: User not authenticated"
// @Failure      403  {string}  string "Forbidden: User does not own this product"
// @Failure      404  {string}  string "Not Found: Product not found"
// @Failure      500  {string}  string "Internal Server Error: Database error"
// @Security     ApiKeyAuth
// @Router       /api/products/details [get] // Changed path slightly to avoid conflict with GetProducts
func GetProduct(w http.ResponseWriter, r *http.Request) {
	// --- Updated Auth Check ---
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetProduct: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID
	// --- End Updated Auth Check ---

	productIDStr := r.URL.Query().Get("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	var product Product
	// Preload image data when fetching the product
	if err := db.Preload("Img").First(&product, "id = ?", uint(productID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			log.Printf("Error finding product %d: %v", productID, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Check ownership
	if product.UserID != userID {
		http.Error(w, "Permission denied: You do not own this product", http.StatusForbidden)
		return
	}

	retrieve := setProductReturn(product)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(retrieve); err != nil {
		log.Printf("Error encoding product %d to JSON: %v", productID, err)
		// Hard to send error response here as headers/status might be sent
	}
}

// GetProducts retrieves information about all products belonging to the logged-in user.
//
// @Summary      Get all products for the user
// @Description  Retrieves a list of all products owned by the authenticated user.
// @Tags         Products
// @Produce      application/json
// @Success      200  {array}   ProductReturn "Successfully retrieved list of products"
// @Failure      401  {string}  string "Unauthorized: User not authenticated"
// @Failure      500  {string}  string "Internal Server Error: Database error"
// @Security     ApiKeyAuth
// @Router       /api/products [get]
func GetProducts(w http.ResponseWriter, r *http.Request) {
	// --- Updated Auth Check ---
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetProducts: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID
	// --- End Updated Auth Check ---

	var products []Product
	// Find all products for the user, preloading image data
	if err := db.Preload("Img").Where("user_id = ?", userID).Find(&products).Error; err != nil {
		log.Printf("Error fetching products for user %d: %v", userID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return empty JSON array if no products found, instead of just "[]" string
	if len(products) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "[]")
		return
	}

	var productsRet []ProductReturn
	for _, product := range products {
		retrieve := setProductReturn(product)
		productsRet = append(productsRet, retrieve)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(productsRet); err != nil {
		log.Printf("Error encoding products for user %d to JSON: %v", userID, err)
	}
}

// GetProductImage serves the image file associated with a product.
// It requires the image filename (e.g., the UUID.ext stored in the Image record) as a query parameter.
// Authentication checks if the user owns the image record.
//
// @Summary      Get a product image
// @Description  Retrieves and serves the image file associated with a product, identified by its filename. Requires the user to be authenticated and own the product/image.
// @Tags         Products
// @Produce      image/*
// @Param        image  query     string true  "Filename of the image to retrieve (e.g., 'uuid.jpg')"
// @Success      200    {file}    binary "Product image file"
// @Failure      400    {string}  string "Bad Request: Missing or invalid image filename"
// @Failure      401    {string}  string "Unauthorized: User not authenticated"
// @Failure      403    {string}  string "Forbidden: User does not own this image"
// @Failure      404    {string}  string "Not Found: Image metadata or file not found"
// @Failure      500    {string}  string "Internal Server Error: Database or file system error"
// @Security     ApiKeyAuth
// @Router       /api/products/image [get]
func GetProductImage(w http.ResponseWriter, r *http.Request) {
	// --- Updated Auth Check ---
	// Note: Authentication might not be strictly necessary if image URLs are non-guessable UUIDs
	// and considered public once known. However, checking ownership adds a layer of security.
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetProductImage: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID
	// --- End Updated Auth Check ---

	imageFilename := r.URL.Query().Get("image")
	if imageFilename == "" {
		http.Error(w, "Missing image filename", http.StatusBadRequest)
		return
	}

	// Clean the filename to prevent path traversal
	imageFilename = filepath.Base(imageFilename)
	if imageFilename == "." || imageFilename == "/" {
		http.Error(w, "Invalid image filename", http.StatusBadRequest)
		return
	}

	var image Image
	// Find the image record by URL (filename)
	if err := db.Where("url = ?", imageFilename).First(&image).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Image metadata not found", http.StatusNotFound)
		} else {
			log.Printf("Error finding image record %s: %v", imageFilename, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Check ownership (important if URLs aren't inherently secret)
	if image.UserID != userID {
		http.Error(w, "Permission denied: You do not own this image", http.StatusForbidden)
		return
	}

	imagePath := filepath.Join("uploads", imageFilename)

	// Check if file exists on disk *after* checking DB and ownership
	if !doesFileExist(imagePath) {
		log.Printf("Image file not found on disk: %s (DB record exists, ID: %d)", imagePath, image.ID)
		http.Error(w, "Image file not found", http.StatusNotFound)
		return
	}

	// Serve the file - http.ServeFile handles Content-Type, caching headers etc.
	http.ServeFile(w, r, imagePath)
}
