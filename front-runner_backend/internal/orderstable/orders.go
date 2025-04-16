package orderstable

import (
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/prodtable"
	"log"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	// db will hold the GORM DB instance
	db        *gorm.DB
	setupOnce sync.Once
)

func Setup() {
	setupOnce.Do(func() {
		coredbutils.LoadEnv()
		db = coredbutils.GetDB()
		login.Setup()
	})
}

type Order struct {
	ID             uint `gorm:"primaryKey"`
	OwnerID        uint `gorm:"not null;index"`
	CustomerName   string
	Total          uint
	OrderDate      time.Time `gorm:"autoCreateTime"`
	OrderStatus    string
	TrackingNumber string
	TrackingImage  string
}

type OrderProd struct {
	OrderID uint
	Order   Order `gorm:"foreignKey:OrderID"`
	ProdID  uint
	Prod    prodtable.Product `gorm:"foreignKey:ProdID"`
	Count   uint
}

// OrderProductPayload is used to decode the JSON body when creating an order.
type OrderProductPayload struct {
	ProdID uint `json:"productID"`
	Count  uint `json:"count"`
}

// OrderCreatePayload is used to decode the JSON body when creating an order.
type OrderCreatePayload struct {
	CustomerName    string                `json:"customerName"` // Name of the customer that placed the order
	OrderedProducts []OrderProductPayload // List of ordered products
}

// MigrateProdDB runs the database migrations for the product and image tables.
func MigrateProdDB() {
	if db == nil {
		log.Fatal("Database connection is not initialized")
	}
	log.Println("Running orders database migrations...")
	err := db.AutoMigrate(&Order{}, &OrderProd{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Orders database migration complete")
}

// CreateOrder creates a new order.
//
// @Summary      Creates an order
// @Description  Creates a new order entry with details including customer name, products, count, and order date.
//
// @Tags         order
// @Accept       json
// @Produce      plain
// @Param        orderInfo body OrderCreatePayload true "Order Details"
// @Success      201  {string}  string "Product added successfully"
// @Failure      400  {string}  string "Error parsing form or uploading image"
// @Failure      401  {string}  string "User not authenticated"
// @Failure      500  {string}  string "Internal server error"
// @Router       /api/create_order [post]
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Extract the logged in user's ID from the context.
	if !login.IsLoggedIn(r) {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// userID, err := login.GetUserID(r)
	// if err != nil {
	// 	http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// err = r.ParseMultipartForm(10 << 20) // Limit to 10MB
	// if err != nil {
	// 	http.Error(w, "Error parsing form", http.StatusBadRequest)
	// 	return
	// }

	// productName := r.FormValue("productName")
	// productDescription := r.FormValue("description")
	// productPrice, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	// productCount, _ := strconv.Atoi(r.FormValue("count"))
	// productTags := r.FormValue("tags")

	// // Handle File Upload
	// file, handler, err := r.FormFile("image")
	// if err != nil {
	// 	http.Error(w, "Error uploading image", http.StatusBadRequest)
	// 	return
	// }
	// defer file.Close()

	// // Save image to disk (or upload to cloud storage)
	// imageFilename := uuid.New().String() + filepath.Ext(handler.Filename)
	// imagePath := filepath.Join("uploads", imageFilename)
	// dst, err := os.Create(imagePath)
	// if err != nil {
	// 	http.Error(w, "Error saving image", http.StatusInternalServerError)
	// 	return
	// }
	// defer dst.Close()
	// io.Copy(dst, file)

	// // Save Product & Image in Database, including the user ID.
	// product := Product{
	// 	UserID:          userID,
	// 	ProdName:        productName,
	// 	ProdDescription: productDescription,
	// 	ProdPrice:       productPrice,
	// 	ProdCount:       uint(productCount),
	// 	ProdTags:        productTags,
	// }

	// // Save Image record
	// image := Image{
	// 	URL:    imageFilename, // Store path instead of image data
	// 	UserID: userID,
	// }
	// db.Create(&image)

	// product.ImgID = image.ID

	// // Save product first
	// if err := db.Create(&product).Error; err != nil {
	// 	http.Error(w, "Error saving product", http.StatusInternalServerError)
	// 	return
	// }

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Product added successfully"))
}
