package prodtable

import (
	"fmt"
	"front-runner/internal/coredbutils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Image struct {
	ID        uint   `gorm:"primaryKey"`
	URL       string `gorm:"not null"`
	ProductID uint   `gorm:"index"`
}

type Product struct {
	ID              uint   `gorm:"primaryKey"`
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
)

func init() {
	db = coredbutils.GetDB()
}

func ClearProdTable(db *gorm.DB) error {
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Product{}).Error; err != nil {
		return fmt.Errorf("error clearing users table: %w", err)
	}
	return nil
}

func AddProduct(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Limit to 10MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	productName := r.FormValue("product_name")
	productDescription := r.FormValue("product_description")
	productPrice, _ := strconv.ParseFloat(r.FormValue("item_price"), 64)
	productCount, _ := strconv.Atoi(r.FormValue("stock_amount"))
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

	// Save Product & Image in Database
	product := Product{
		ProdName:        productName,
		ProdDescription: productDescription,
		ProdPrice:       productPrice,
		ProdCount:       uint(productCount),
		ProdTags:        productTags,
	}

	// Save product first
	if err := db.Create(&product).Error; err != nil {
		http.Error(w, "Error saving product", http.StatusInternalServerError)
		return
	}

	// Save Image record
	image := Image{
		URL:       imagePath, // Store path instead of image data
		ProductID: product.ID,
	}
	db.Create(&image)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Product added successfully"))
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	var product Product

	if err := db.Preload("Img").First(&product, productID).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
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

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	var product Product

	if err := db.First(&product, productID).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
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
