package orderstable

import (
	"encoding/json"
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
	CustomerName   string
	CustomerEmail  string
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
	Cost    float64
}

type OrderOwner struct {
	UserID  uint  `gorm:"not null;index:idx_product,unique"`
	OrderID uint  `gorm:"not null;index:idx_product,unique"`
	Order   Order `gorm:"foreignKey:OrderID"`
}

// OrderProductPayload is used to decode the JSON body when creating an order.
type OrderProductPayload struct {
	ProdID uint `json:"productID"`
	Count  uint `json:"count"`
}

// OrderCreatePayload is used to decode the JSON body when creating an order.
type OrderCreatePayload struct {
	CustomerName    string                `json:"customerName"`  // Name of the customer that placed the order
	CustomerEmail   string                `json:"customerEmail"` // Email of the customer that placed the order
	OrderedProducts []OrderProductPayload // List of ordered products
}

// OrderProductReturn is struct returned to the frontend containing information about an order's products.
type OrderProductReturn struct {
	ProdID   uint    `json:"productID"`
	ProdName string  `json:"productName"`
	Count    uint    `json:"count"`
	Price    float64 `json:"price"`
}

// OrderProductReturn is struct returned to the frontend containing releveant information about an order.
type OrderReturn struct {
	OrderID         uint                 `json:"orderID"`       // ID of the order requested
	CustomerName    string               `json:"customerName"`  // Name of the customer that placed the order
	CustomerEmail   string               `json:"customerEmail"` // Name of the customer that placed the order
	OrderDate       string               `json:"orderDate"`
	OrderStatus     string               `json:"status"`
	TrackingNumber  string               `json:"trackingNumber"`
	Total           float64              `json:"total"`
	OrderedProducts []OrderProductReturn // List of ordered products
}

// MigrateProdDB runs the database migrations for the product and image tables.
func MigrateProdDB() {
	if db == nil {
		log.Fatal("Database connection is not initialized")
	}
	log.Println("Running orders database migrations...")
	err := db.AutoMigrate(&Order{}, &OrderProd{}, &OrderOwner{})
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
	// if !login.IsLoggedIn(r) {
	// 	http.Error(w, "User not authenticated", http.StatusUnauthorized)
	// 	return
	// }

	var payload OrderCreatePayload
	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Good practice to close body

	// Parse json to consolidate any products. Accounts for the case of multiple instances of the same item.
	consolidatedCount := make(map[uint]uint)
	for _, c := range payload.OrderedProducts {
		consolidatedCount[c.ProdID] += c.Count
	}

	// Check valid stock levels and retrieve product seller ids for tagging.
	sellerIDs := make(map[uint]bool)
	for id, count := range consolidatedCount {
		var product prodtable.Product
		if err := db.Preload("Img").Where("id = ?", id).First(&product).Error; err != nil {
			http.Error(w, "No Product with specified ID", http.StatusNotFound)
			return
		}

		if count > product.ProdCount {
			http.Error(w, "Invalid stock of item", http.StatusNotFound)
			return
		}
		sellerIDs[product.UserID] = true
	}

	order := Order{
		CustomerName:  payload.CustomerName,
		CustomerEmail: payload.CustomerEmail,
		// TODO: Set other elements such as tracking and status.
	}

	result := db.Create(&order)

	if result.Error != nil {

	}

	for seller := range sellerIDs {
		sellerRecord := OrderOwner{
			UserID:  seller,
			OrderID: order.ID,
			Order:   order,
		}
		db.Create(&sellerRecord)
	}

	for id, count := range consolidatedCount {
		var product prodtable.Product
		if err := db.Preload("Img").Where("id = ?", id).First(&product).Error; err != nil {
			http.Error(w, "No Product with specified ID", http.StatusNotFound)
			return
		}

		// create order prod record
		orderProductRecord := OrderProd{
			OrderID: order.ID,
			ProdID:  id,
			Count:   count,
			Cost:    product.ProdPrice,
		}
		db.Create(&orderProductRecord)

		// update product record with new count
		updates := map[string]interface{}{}
		updates["ProdCount"] = uint(product.ProdCount - count)
		if err := db.Model(&product).Updates(updates).Error; err != nil {
			http.Error(w, "Error updating product: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Order created successfully"))
}

// GetOrder retrieves the information about a specified order if it belongs to the logged-in user.
//
// @Summary      Retrieve an order
// @Description  Retreives an existing order and its associated metadata if the order belongs to the authenticated user.
// @Tags         order
// @Produce      json
// @Param        id   query integer true "Order ID"
// @Success      200  {object}  OrderReturn "JSON representation of an orders information (empty object if none)"
// @Failure      401  {string}  string "User not authenticated or unauthorized"
// @Failure      403  {string}  string "Permission denied"
// @Failure      404  {string}  string "No Order with specified ID"
// @Router       /api/get_order [get]
func GetOrder(w http.ResponseWriter, r *http.Request) {
	if !login.IsLoggedIn(r) {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	orderID := r.URL.Query().Get("id")

	var order Order
	if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
		http.Error(w, "No Order with specified ID", http.StatusNotFound)
		return
	}

	var orderProds []OrderProd
	db.Preload("Prod").Preload("Order").Where("order_id = ?", orderID).Find(&orderProds)

	if len(orderProds) == 0 { // no products associated with order
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	var userProds []OrderProductReturn
	var totalCost float64 = 0.0
	for _, product := range orderProds {
		if product.Prod.UserID == userID {
			userProd := OrderProductReturn{
				ProdID:   product.Prod.ID,
				ProdName: product.Prod.ProdName,
				Count:    product.Count,
				Price:    product.Cost,
			}
			totalCost += (product.Cost * float64(product.Count))
			userProds = append(userProds, userProd)
		}
	}

	if len(userProds) == 0 {
		// TODO: What should be sent in the case of no products owned by user present in order
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	orderRet := OrderReturn{
		OrderID:         order.ID,
		CustomerName:    order.CustomerName,
		CustomerEmail:   order.CustomerEmail,
		OrderDate:       order.OrderDate.String(),
		OrderStatus:     order.OrderStatus,
		TrackingNumber:  order.TrackingNumber,
		Total:           totalCost,
		OrderedProducts: userProds,
	}

	ret, _ := json.Marshal(orderRet)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ret))
}
