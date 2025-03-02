package prodtable

import (
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

type Image struct {
	ID  uint   `gorm:"primaryKey"`
	URL string `gorm:"not null"`
}

type Product struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"not null;index"`
	ProdName        string `gorm:"unique;not null"`
	ProdDescription string `gorm:"not null"`
	ImgID           uint
	Img             Image `gorm:"foreignKey:ImgID"`
	ProdPrice       float64
	ProdCount       uint
	ProdTags        string
}

var (
	// db will hold the GORM DB instance
	db *gorm.DB

	sessionStore *sessions.CookieStore
)

func init() {
	db = coredbutils.GetDB()

	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}
}

func MigrateProdDB() {
	if db == nil {
		log.Fatal("Database connection is not initialized")
	}
	log.Println("Running database migrations...")
	err := db.AutoMigrate(&Product{}, &Image{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migration complete")
}

func ClearProdTable(db *gorm.DB) error {
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Product{}).Error; err != nil {
		return fmt.Errorf("error clearing product table: %w", err)
	}
	return nil
}

// AddProduct creates a new product and associates it with the logged-in user.
func AddProduct(w http.ResponseWriter, r *http.Request) {
	// Extract the logged in user's ID from the context.
	if !login.IsLoggedIn(r) {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // Limit to 10MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	productName := r.FormValue("productName")
	productDescription := r.FormValue("description")
	productPrice, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	productCount, _ := strconv.Atoi(r.FormValue("count"))
	productTags := r.FormValue("tags")

	// Handle File Upload
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error uploading image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save image to disk (or upload to cloud storage)
	imagePath := fmt.Sprintf("uploads/%s%s", uuid.New().String(), filepath.Ext(handler.Filename))
	dst, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Error saving image", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Save Product & Image in Database, including the user ID.
	product := Product{
		UserID:          userID,
		ProdName:        productName,
		ProdDescription: productDescription,
		ProdPrice:       productPrice,
		ProdCount:       uint(productCount),
		ProdTags:        productTags,
	}

	// Save Image record
	image := Image{
		URL: imagePath, // Store path instead of image data
	}
	db.Create(&image)

	product.ImgID = image.ID

	// Save product first
	if err := db.Create(&product).Error; err != nil {
		http.Error(w, "Error saving product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Product added successfully"))
}

// DeleteProduct removes a product but only if it belongs to the logged-in user.
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	if !login.IsLoggedIn(r) {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	productID := r.URL.Query().Get("id")
	var product Product

	if err := db.Preload("Img").First(&product, productID).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Check if the product belongs to the logged in user.
	if product.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete image file
	if err := os.Remove(product.Img.URL); err != nil {
		fmt.Println("Error deleting image:", err)
	}

	// Delete product (Image gets deleted due to CASCADE)
	db.Delete(&product)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product deleted successfully"))
}

// UpdateProduct allows updating a product if it belongs to the logged-in user.
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	if !login.IsLoggedIn(r) {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	productID := r.URL.Query().Get("id")
	var product Product

	if err := db.First(&product, productID).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Ensure that the logged-in user owns this product.
	if product.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	productDescription := r.FormValue("product_description")
	productPrice, _ := strconv.ParseFloat(r.FormValue("item_price"), 64)
	productCount, _ := strconv.Atoi(r.FormValue("stock_amount"))

	// Update only non-empty values
	if productDescription != "" {
		product.ProdDescription = productDescription
	}
	if productPrice > 0 {
		product.ProdPrice = productPrice
	}
	if productCount >= 0 {
		product.ProdCount = uint(productCount)
	}

	db.Save(&product)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product updated successfully"))
}
