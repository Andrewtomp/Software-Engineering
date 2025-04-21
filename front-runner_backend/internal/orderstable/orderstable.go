package orderstable

import (
	"encoding/json"
	"errors" // Import errors package
	"fmt"    // Import fmt for error formatting
	"front-runner/internal/coredbutils"
	"front-runner/internal/oauth" // Use oauth for authentication
	"front-runner/internal/prodtable"
	"log"
	"net/http"
	"strconv" // Import strconv for ID parsing
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	// db will hold the GORM DB instance
	db        *gorm.DB
	setupOnce sync.Once
)

// Setup initializes the database connection for the orderstable package.
func Setup() {
	setupOnce.Do(func() {
		// Get DB connection from coredbutils
		coredbutils.LoadEnv() // Ensure env vars are loaded if needed by GetDB
		var err error
		db, err = coredbutils.GetDB() // Get the singleton DB instance
		if err != nil {
			log.Fatalf("orderstable Setup: Failed to get database connection: %v", err)
		}
		if db == nil {
			log.Fatal("orderstable Setup: Database connection is nil after GetDB.")
		}
		// login.Setup() // Removed: Assume main.go handles setup order
		log.Println("orderstable package setup complete (DB connection obtained).")
	})
}

// Order represents the main order details.
type Order struct {
	ID             uint      `gorm:"primaryKey"`
	CustomerName   string    // Name provided by the buyer (might not be a registered user)
	CustomerEmail  string    // Email provided by the buyer
	OrderDate      time.Time `gorm:"autoCreateTime"`
	OrderStatus    string    // e.g., "Pending", "Processing", "Shipped", "Delivered", "Cancelled"
	TrackingNumber string
	TrackingImage  string      // URL or path to a tracking image/label if applicable
	OrderProds     []OrderProd `gorm:"foreignKey:OrderID"` // <--- Add this line
}

// OrderProd links an Order with a Product, specifying quantity and cost at the time of order.
type OrderProd struct {
	gorm.Model                   // Includes ID, CreatedAt, UpdatedAt, DeletedAt
	OrderID    uint              `gorm:"not null;index"`
	Order      Order             `gorm:"foreignKey:OrderID"`
	ProdID     uint              `gorm:"not null;index"`
	Prod       prodtable.Product `gorm:"foreignKey:ProdID"` // Link to the actual product
	Count      uint              `gorm:"not null"`
	Cost       float64           `gorm:"not null"` // Price per item at the time of order
}

// OrderOwner links an Order to the User who *owns* the products being sold in that order.
// This allows sellers (users) to see orders containing their products.
type OrderOwner struct {
	gorm.Model       // Includes ID, CreatedAt, UpdatedAt, DeletedAt
	UserID     uint  `gorm:"not null;index:idx_user_order,unique"` // Seller's User ID
	OrderID    uint  `gorm:"not null;index:idx_user_order,unique"` // Order ID
	Order      Order `gorm:"foreignKey:OrderID"`                   // Link to the order
}

// OrderProductPayload is used to decode the JSON body when creating an order.
type OrderProductPayload struct {
	ProdID uint `json:"productID"`
	Count  uint `json:"count"`
}

// OrderCreatePayload is used to decode the JSON body when creating an order.
type OrderCreatePayload struct {
	CustomerName    string                `json:"customerName"`    // Name of the customer placing the order
	CustomerEmail   string                `json:"customerEmail"`   // Email of the customer placing the order
	OrderedProducts []OrderProductPayload `json:"orderedProducts"` // List of ordered products
}

// OrderProductReturn is struct returned to the frontend containing information about an order's products.
type OrderProductReturn struct {
	ProdID   uint    `json:"productID"`
	ProdName string  `json:"productName"`
	Count    uint    `json:"count"`
	Price    float64 `json:"price"` // Price per item at the time of order
}

// OrderReturn is struct returned to the frontend containing relevant information about an order,
// filtered for the requesting user (seller).
type OrderReturn struct {
	OrderID         uint                 `json:"orderID"`       // ID of the order requested
	CustomerName    string               `json:"customerName"`  // Name of the customer that placed the order
	CustomerEmail   string               `json:"customerEmail"` // Email of the customer that placed the order
	OrderDate       string               `json:"orderDate"`     // Formatted date string
	OrderStatus     string               `json:"status"`
	TrackingNumber  string               `json:"trackingNumber"`
	Total           float64              `json:"total"`           // Total cost *for the items owned by the requesting user* in this order
	OrderedProducts []OrderProductReturn `json:"orderedProducts"` // List of ordered products *owned by the requesting user*
}

// MigrateOrdersDB runs the database migrations for the order-related tables.
func MigrateOrdersDB() {
	if db == nil {
		log.Fatal("Database connection is not initialized for orders migration")
	}
	log.Println("Running orders database migrations...")
	// AutoMigrate Order, OrderProd, OrderOwner
	err := db.AutoMigrate(&Order{}, &OrderProd{}, &OrderOwner{})
	if err != nil {
		log.Fatalf("Orders migration failed: %v", err)
	}
	log.Println("Orders database migration complete")
}

// CreateOrder creates a new order. This endpoint is typically public or requires buyer authentication.
// It processes the order, updates stock, and links the order to the sellers of the products.
//
// @Summary      Creates an order
// @Description  Creates a new order entry with customer details and products. Updates product stock and links sellers.
// @Tags         order
// @Accept       json
// @Produce      json // Return the created order ID or details
// @Param        orderInfo body OrderCreatePayload true "Order Details"
// @Success      201  {object} map[string]uint "Order created successfully, returns order ID" // Example success response
// @Failure      400  {string}  string "Invalid request body, missing fields, or invalid product data"
// @Failure      404  {string}  string "Product not found or insufficient stock"
// @Failure      500  {string}  string "Internal server error during order processing"
// @Router       /api/create_order [post]
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Authentication: This endpoint might be public or require buyer auth.
	// For now, we assume it's public, but seller info is derived from products.

	var payload OrderCreatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// --- Basic Validation ---
	if strings.TrimSpace(payload.CustomerName) == "" || strings.TrimSpace(payload.CustomerEmail) == "" {
		http.Error(w, "Customer name and email are required", http.StatusBadRequest)
		return
	}
	if len(payload.OrderedProducts) == 0 {
		http.Error(w, "Order must contain at least one product", http.StatusBadRequest)
		return
	}

	// --- Consolidate Products and Check Stock within a Transaction ---
	var createdOrderID uint
	err := db.Transaction(func(tx *gorm.DB) error {
		consolidatedCount := make(map[uint]uint)
		productDetails := make(map[uint]prodtable.Product) // Store fetched product details
		sellerIDs := make(map[uint]bool)

		for _, item := range payload.OrderedProducts {
			if item.Count <= 0 {
				return fmt.Errorf("invalid count for product ID %d", item.ProdID) // Return error to rollback
			}
			consolidatedCount[item.ProdID] += item.Count
		}

		// Fetch products, check stock, and gather seller IDs
		for prodID, requestedCount := range consolidatedCount {
			var product prodtable.Product
			// Use the transaction tx here
			if err := tx.First(&product, prodID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("product with ID %d not found", prodID) // Return error to rollback
				}
				log.Printf("Error fetching product %d: %v", prodID, err)
				return fmt.Errorf("database error fetching product %d", prodID) // Return error to rollback
			}

			if requestedCount > product.ProdCount {
				return fmt.Errorf("insufficient stock for product ID %d (requested: %d, available: %d)", prodID, requestedCount, product.ProdCount) // Return error to rollback
			}
			productDetails[prodID] = product
			sellerIDs[product.UserID] = true // Track the seller (user) of this product
		}

		// --- Create Order Record ---
		order := Order{
			CustomerName:  payload.CustomerName,
			CustomerEmail: payload.CustomerEmail,
			OrderStatus:   "Pending", // Initial status
			// TrackingNumber and TrackingImage are usually set later
		}
		// Use the transaction tx here
		if err := tx.Create(&order).Error; err != nil {
			log.Printf("Error creating order record: %v", err)
			return errors.New("failed to create order record") // Return error to rollback
		}
		createdOrderID = order.ID // Store the ID for response

		// --- Create OrderOwner Records (Link Sellers) ---
		for sellerID := range sellerIDs {
			orderOwner := OrderOwner{
				UserID:  sellerID,
				OrderID: order.ID,
				// Order:   order, // GORM can handle this via OrderID
			}
			// Use the transaction tx here
			if err := tx.Create(&orderOwner).Error; err != nil {
				// Check for unique constraint violation (shouldn't happen if logic is correct)
				log.Printf("Error creating order owner link (User: %d, Order: %d): %v", sellerID, order.ID, err)
				return fmt.Errorf("failed to link seller %d to order", sellerID) // Return error to rollback
			}
		}

		// --- Create OrderProd Records and Update Product Stock ---
		for prodID, count := range consolidatedCount {
			product := productDetails[prodID] // Get pre-fetched product details

			// Create OrderProd record
			orderProductRecord := OrderProd{
				OrderID: order.ID,
				ProdID:  prodID,
				Count:   count,
				Cost:    product.ProdPrice, // Record price at time of order
			}
			// Use the transaction tx here
			if err := tx.Create(&orderProductRecord).Error; err != nil {
				log.Printf("Error creating order product record (Order: %d, Prod: %d): %v", order.ID, prodID, err)
				return fmt.Errorf("failed to record product %d in order", prodID) // Return error to rollback
			}

			// Update product stock count
			newCount := product.ProdCount - count
			// Use the transaction tx here
			if err := tx.Model(&prodtable.Product{}).Where("id = ?", prodID).Update("prod_count", newCount).Error; err != nil {
				log.Printf("Error updating stock for product %d: %v", prodID, err)
				return fmt.Errorf("failed to update stock for product %d", prodID) // Return error to rollback
			}
		}

		return nil // nil error commits the transaction
	})

	// --- Handle Transaction Outcome ---
	if err != nil {
		// Determine appropriate HTTP status code based on the error
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "insufficient stock") || strings.Contains(err.Error(), "invalid count") {
			http.Error(w, err.Error(), http.StatusBadRequest) // Or StatusConflict (409)? Bad Request seems ok.
		} else {
			// Generic internal server error for other DB issues
			http.Error(w, "Internal server error during order processing", http.StatusInternalServerError)
		}
		return
	}

	// --- Success Response ---
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]uint{"orderID": createdOrderID})
}

// GetOrder retrieves the information about a specified order, filtered for the logged-in user (seller).
// It only shows products within the order that belong to the requesting user.
//
// @Summary      Retrieve an order (filtered for seller)
// @Description  Retrieves an existing order and its associated products *owned by the authenticated user (seller)*.
// @Tags         order
// @Produce      json
// @Param        id   query integer true "Order ID"
// @Success      200  {object}  OrderReturn "JSON representation of the order's information relevant to the user (empty object if user has no items in this order)"
// @Failure      400  {string}  string "Invalid Order ID format"
// @Failure      401  {string}  string "User not authenticated"
// @Failure      403  {string}  string "Permission denied (user is not a seller for any product in this order)"
// @Failure      404  {string}  string "Order not found"
// @Failure      500  {string}  string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/get_order [get]
func GetOrder(w http.ResponseWriter, r *http.Request) {
	// --- Authentication ---
	user, err := oauth.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetOrder: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID

	// --- Get and Validate Order ID ---
	orderIDStr := r.URL.Query().Get("id")
	orderID64, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Order ID format", http.StatusBadRequest)
		return
	}
	orderID := uint(orderID64)

	// --- Verify User Ownership (Check OrderOwner) ---
	var orderOwner OrderOwner
	if err := db.Where("order_id = ? AND user_id = ?", orderID, userID).First(&orderOwner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Check if the order exists at all before returning 403
			var orderExists Order
			if errExists := db.First(&orderExists, orderID).Error; errors.Is(errExists, gorm.ErrRecordNotFound) {
				http.Error(w, fmt.Sprintf("Order with ID %d not found", orderID), http.StatusNotFound)
			} else {
				http.Error(w, "Permission denied: You are not associated with this order", http.StatusForbidden)
			}
		} else {
			log.Printf("Error checking order ownership (User: %d, Order: %d): %v", userID, orderID, err)
			http.Error(w, "Database error checking order ownership", http.StatusInternalServerError)
		}
		return
	}

	// --- Fetch Order Details ---
	var order Order
	// Preload OrderProds and their associated Prod details
	if err := db.Preload("OrderProds.Prod").First(&order, orderID).Error; err != nil {
		// This shouldn't happen if OrderOwner check passed, but handle defensively
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Order with ID %d not found (inconsistency?)", orderID), http.StatusNotFound)
		} else {
			log.Printf("Error fetching order details for ID %d: %v", orderID, err)
			http.Error(w, "Database error fetching order details", http.StatusInternalServerError)
		}
		return
	}

	// --- Filter Products for the Requesting User ---
	var userProds []OrderProductReturn
	var totalCost float64 = 0.0
	for _, op := range order.OrderProds {
		// Check if the product within the OrderProd belongs to the current user
		if op.Prod.UserID == userID {
			userProd := OrderProductReturn{
				ProdID:   op.ProdID,
				ProdName: op.Prod.ProdName,
				Count:    op.Count,
				Price:    op.Cost, // Use the cost stored at the time of order
			}
			totalCost += (op.Cost * float64(op.Count))
			userProds = append(userProds, userProd)
		}
	}

	// --- Construct and Return Response ---
	// Even if userProds is empty, return the main order details
	orderRet := OrderReturn{
		OrderID:         order.ID,
		CustomerName:    order.CustomerName,
		CustomerEmail:   order.CustomerEmail,
		OrderDate:       order.OrderDate.Format(time.RFC3339), // Standard format
		OrderStatus:     order.OrderStatus,
		TrackingNumber:  order.TrackingNumber,
		Total:           totalCost, // Total for *user's items only*
		OrderedProducts: userProds, // Will be [] if user owns no items in this order
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orderRet); err != nil {
		log.Printf("Error encoding order response for ID %d: %v", orderID, err)
		// Can't send HTTP error here as headers/status likely sent
	}
}

// GetOrders retrieves all orders containing products sold by the logged-in user.
//
// @Summary      Retrieve user's sales orders
// @Description  Retrieves orders containing products sold by the authenticated user, along with the relevant product details for each order.
// @Tags         order
// @Produce      json
// @Success      200  {array}  OrderReturn "JSON array of orders relevant to the user (empty array if none)"
// @Failure      401  {string}  string "User not authenticated"
// @Failure      500  {string}  string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/get_orders [get]
func GetOrders(w http.ResponseWriter, r *http.Request) {
	// --- Authentication ---
	user, err := oauth.GetCurrentUser(r)

	if err != nil {
		log.Printf("GetOrders: Error getting current user: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := user.ID

	// --- Fetch Order IDs associated with the User (Seller) ---
	var userOrderOwners []OrderOwner
	// Preload the main Order details along with the OrderOwner link
	if err := db.Preload("Order").Where("user_id = ?", userID).Find(&userOrderOwners).Error; err != nil {
		log.Printf("Error fetching order ownerships for user %d: %v", userID, err)
		http.Error(w, "Database error fetching user orders", http.StatusInternalServerError)
		return
	}

	// --- Process Each Order ---
	var orderReturns []OrderReturn
	if len(userOrderOwners) > 0 {
		orderIDs := make([]uint, len(userOrderOwners))
		orderMap := make(map[uint]Order) // Map order ID to preloaded Order details
		for i, owner := range userOrderOwners {
			orderIDs[i] = owner.OrderID
			orderMap[owner.OrderID] = owner.Order // Store the preloaded Order
		}

		// Fetch all relevant OrderProd items in one go
		var allOrderProds []OrderProd
		if err := db.Preload("Prod").Where("order_id IN ?", orderIDs).Find(&allOrderProds).Error; err != nil {
			log.Printf("Error fetching order products for user %d orders: %v", userID, err)
			http.Error(w, "Database error fetching order products", http.StatusInternalServerError)
			return
		}

		// Group OrderProds by OrderID
		prodsByOrderID := make(map[uint][]OrderProd)
		for _, op := range allOrderProds {
			prodsByOrderID[op.OrderID] = append(prodsByOrderID[op.OrderID], op)
		}

		// Construct OrderReturn for each order the user is linked to
		for _, owner := range userOrderOwners {
			order := orderMap[owner.OrderID] // Get the preloaded Order details
			orderProds := prodsByOrderID[owner.OrderID]

			var userProdsInOrder []OrderProductReturn
			var totalCost float64 = 0.0

			for _, op := range orderProds {
				// Filter for products owned by the current user
				if op.Prod.UserID == userID {
					userProd := OrderProductReturn{
						ProdID:   op.ProdID,
						ProdName: op.Prod.ProdName,
						Count:    op.Count,
						Price:    op.Cost,
					}
					totalCost += (op.Cost * float64(op.Count))
					userProdsInOrder = append(userProdsInOrder, userProd)
				}
			}

			// Only include the order in the response if the user sold items in it
			if len(userProdsInOrder) > 0 {
				orderInfo := OrderReturn{
					OrderID:         order.ID,
					CustomerName:    order.CustomerName,
					CustomerEmail:   order.CustomerEmail,
					OrderDate:       order.OrderDate.Format(time.RFC3339),
					OrderStatus:     order.OrderStatus,
					TrackingNumber:  order.TrackingNumber,
					Total:           totalCost,
					OrderedProducts: userProdsInOrder,
				}
				orderReturns = append(orderReturns, orderInfo)
			}
		}
	}

	// --- Return Response ---
	// Return empty array `[]` if no relevant orders found
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orderReturns); err != nil {
		log.Printf("Error encoding orders response for user %d: %v", userID, err)
	}
}

// package orderstable

// import (
// 	"encoding/json"
// 	"front-runner/internal/coredbutils"
// 	"front-runner/internal/login"
// 	"front-runner/internal/prodtable"
// 	"log"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"gorm.io/gorm"
// )

// var (
// 	// db will hold the GORM DB instance
// 	db        *gorm.DB
// 	setupOnce sync.Once
// )

// func Setup() {
// 	setupOnce.Do(func() {
// 		coredbutils.LoadEnv()
// 		db = coredbutils.GetDB()
// 		login.Setup()
// 	})
// }

// type Order struct {
// 	ID             uint `gorm:"primaryKey"`
// 	CustomerName   string
// 	CustomerEmail  string
// 	OrderDate      time.Time `gorm:"autoCreateTime"`
// 	OrderStatus    string
// 	TrackingNumber string
// 	TrackingImage  string
// }

// type OrderProd struct {
// 	OrderID uint
// 	Order   Order `gorm:"foreignKey:OrderID"`
// 	ProdID  uint
// 	Prod    prodtable.Product `gorm:"foreignKey:ProdID"`
// 	Count   uint
// 	Cost    float64
// }

// type OrderOwner struct {
// 	UserID  uint  `gorm:"not null;index:idx_product,unique"`
// 	OrderID uint  `gorm:"not null;index:idx_product,unique"`
// 	Order   Order `gorm:"foreignKey:OrderID"`
// }

// // OrderProductPayload is used to decode the JSON body when creating an order.
// type OrderProductPayload struct {
// 	ProdID uint `json:"productID"`
// 	Count  uint `json:"count"`
// }

// // OrderCreatePayload is used to decode the JSON body when creating an order.
// type OrderCreatePayload struct {
// 	CustomerName    string                `json:"customerName"`  // Name of the customer that placed the order
// 	CustomerEmail   string                `json:"customerEmail"` // Email of the customer that placed the order
// 	OrderedProducts []OrderProductPayload // List of ordered products
// }

// // OrderProductReturn is struct returned to the frontend containing information about an order's products.
// type OrderProductReturn struct {
// 	ProdID   uint    `json:"productID"`
// 	ProdName string  `json:"productName"`
// 	Count    uint    `json:"count"`
// 	Price    float64 `json:"price"`
// }

// // OrderProductReturn is struct returned to the frontend containing releveant information about an order.
// type OrderReturn struct {
// 	OrderID         uint                 `json:"orderID"`       // ID of the order requested
// 	CustomerName    string               `json:"customerName"`  // Name of the customer that placed the order
// 	CustomerEmail   string               `json:"customerEmail"` // Name of the customer that placed the order
// 	OrderDate       string               `json:"orderDate"`
// 	OrderStatus     string               `json:"status"`
// 	TrackingNumber  string               `json:"trackingNumber"`
// 	Total           float64              `json:"total"`
// 	OrderedProducts []OrderProductReturn // List of ordered products
// }

// // MigrateProdDB runs the database migrations for the product and image tables.
// func MigrateProdDB() {
// 	if db == nil {
// 		log.Fatal("Database connection is not initialized")
// 	}
// 	log.Println("Running orders database migrations...")
// 	err := db.AutoMigrate(&Order{}, &OrderProd{}, &OrderOwner{})
// 	if err != nil {
// 		log.Fatalf("Migration failed: %v", err)
// 	}
// 	log.Println("Orders database migration complete")
// }

// // CreateOrder creates a new order.
// //
// // @Summary      Creates an order
// // @Description  Creates a new order entry with details including customer name, products, count, and order date.
// //
// // @Tags         order
// // @Accept       json
// // @Produce      plain
// // @Param        orderInfo body OrderCreatePayload true "Order Details"
// // @Success      201  {string}  string "Product added successfully"
// // @Failure      400  {string}  string "Error parsing form or uploading image"
// // @Failure      401  {string}  string "User not authenticated"
// // @Failure      500  {string}  string "Internal server error"
// // @Router       /api/create_order [post]
// func CreateOrder(w http.ResponseWriter, r *http.Request) {
// 	// Extract the logged in user's ID from the context.
// 	// if !login.IsLoggedIn(r) {
// 	// 	http.Error(w, "User not authenticated", http.StatusUnauthorized)
// 	// 	return
// 	// }

// 	var payload OrderCreatePayload
// 	// Decode JSON body
// 	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
// 		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close() // Good practice to close body

// 	// Parse json to consolidate any products. Accounts for the case of multiple instances of the same item.
// 	consolidatedCount := make(map[uint]uint)
// 	for _, c := range payload.OrderedProducts {
// 		consolidatedCount[c.ProdID] += c.Count
// 	}

// 	// Check valid stock levels and retrieve product seller ids for tagging.
// 	sellerIDs := make(map[uint]bool)
// 	for id, count := range consolidatedCount {
// 		var product prodtable.Product
// 		if err := db.Preload("Img").Where("id = ?", id).First(&product).Error; err != nil {
// 			http.Error(w, "No Product with specified ID", http.StatusNotFound)
// 			return
// 		}

// 		if count > product.ProdCount {
// 			http.Error(w, "Invalid stock of item", http.StatusNotFound)
// 			return
// 		}
// 		sellerIDs[product.UserID] = true
// 	}

// 	order := Order{
// 		CustomerName:  payload.CustomerName,
// 		CustomerEmail: payload.CustomerEmail,
// 		// TODO: Set other elements such as tracking and status.
// 	}

// 	result := db.Create(&order)

// 	if result.Error != nil {

// 	}

// 	for seller := range sellerIDs {
// 		sellerRecord := OrderOwner{
// 			UserID:  seller,
// 			OrderID: order.ID,
// 			Order:   order,
// 		}
// 		db.Create(&sellerRecord)
// 	}

// 	for id, count := range consolidatedCount {
// 		var product prodtable.Product
// 		if err := db.Preload("Img").Where("id = ?", id).First(&product).Error; err != nil {
// 			http.Error(w, "No Product with specified ID", http.StatusNotFound)
// 			return
// 		}

// 		// create order prod record
// 		orderProductRecord := OrderProd{
// 			OrderID: order.ID,
// 			ProdID:  id,
// 			Count:   count,
// 			Cost:    product.ProdPrice,
// 		}
// 		db.Create(&orderProductRecord)

// 		// update product record with new count
// 		updates := map[string]interface{}{}
// 		updates["ProdCount"] = uint(product.ProdCount - count)
// 		if err := db.Model(&product).Updates(updates).Error; err != nil {
// 			http.Error(w, "Error updating product: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	w.Write([]byte("Order created successfully"))
// }

// // GetOrder retrieves the information about a specified order if it belongs to the logged-in user.
// //
// // @Summary      Retrieve an order
// // @Description  Retreives an existing order and its associated metadata if the order belongs to the authenticated user.
// // @Tags         order
// // @Produce      json
// // @Param        id   query integer true "Order ID"
// // @Success      200  {object}  OrderReturn "JSON representation of an orders information (empty object if none)"
// // @Failure      401  {string}  string "User not authenticated or unauthorized"
// // @Failure      403  {string}  string "Permission denied"
// // @Failure      404  {string}  string "No Order with specified ID"
// // @Router       /api/get_order [get]
// func GetOrder(w http.ResponseWriter, r *http.Request) {
// 	if !login.IsLoggedIn(r) {
// 		http.Error(w, "User not authenticated", http.StatusUnauthorized)
// 		return
// 	}

// 	userID, err := login.GetUserID(r)
// 	if err != nil {
// 		http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	orderID := r.URL.Query().Get("id")

// 	var order Order
// 	if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
// 		http.Error(w, "No Order with specified ID", http.StatusNotFound)
// 		return
// 	}

// 	var orderProds []OrderProd
// 	db.Preload("Prod").Preload("Order").Where("order_id = ?", orderID).Find(&orderProds)

// 	if len(orderProds) == 0 { // no products associated with order
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("{}"))
// 		return
// 	}

// 	var userProds []OrderProductReturn
// 	var totalCost float64 = 0.0
// 	for _, product := range orderProds {
// 		if product.Prod.UserID == userID {
// 			userProd := OrderProductReturn{
// 				ProdID:   product.Prod.ID,
// 				ProdName: product.Prod.ProdName,
// 				Count:    product.Count,
// 				Price:    product.Cost,
// 			}
// 			totalCost += (product.Cost * float64(product.Count))
// 			userProds = append(userProds, userProd)
// 		}
// 	}

// 	if len(userProds) == 0 {
// 		// TODO: What should be sent in the case of no products owned by user present in order
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("{}"))
// 		return
// 	}

// 	orderRet := OrderReturn{
// 		OrderID:         order.ID,
// 		CustomerName:    order.CustomerName,
// 		CustomerEmail:   order.CustomerEmail,
// 		OrderDate:       order.OrderDate.String(),
// 		OrderStatus:     order.OrderStatus,
// 		TrackingNumber:  order.TrackingNumber,
// 		Total:           totalCost,
// 		OrderedProducts: userProds,
// 	}

// 	ret, _ := json.Marshal(orderRet)
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(ret))
// }

// // GetOrders retrieves the information about a all orders belonging to the logged-in user.
// //
// // @Summary      Retrieve an user's orders
// // @Description  Retreives orders and their associated metadata belonging to the authenticated user.
// // @Tags         order
// // @Produce      json
// // @Success      200  {array}  OrderReturn "JSON representation of an orders information (empty object if none)"
// // @Failure      401  {string}  string "User not authenticated or unauthorized"
// // @Failure      403  {string}  string "Permission denied"
// // @Failure      404  {string}  string "No Order with specified ID"
// // @Router       /api/get_orders [get]
// func GetOrders(w http.ResponseWriter, r *http.Request) {
// 	if !login.IsLoggedIn(r) {
// 		http.Error(w, "User not authenticated", http.StatusUnauthorized)
// 		return
// 	}

// 	userID, err := login.GetUserID(r)
// 	if err != nil {
// 		http.Error(w, "Error retrieving session: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var orders []OrderOwner
// 	db.Preload("Order").Where("user_id = ?", userID).Find(&orders)

// 	var orderRet []OrderReturn

// 	for _, order := range orders { // for each order referencing a user's id
// 		var orderProds []OrderProd // get products for order
// 		db.Preload("Prod").Preload("Order").Where("order_id = ?", order.Order.ID).Find(&orderProds)

// 		if len(orderProds) == 0 { // no products associated with order
// 			w.WriteHeader(http.StatusOK)
// 			w.Write([]byte("[]"))
// 			return
// 		}

// 		var userProds []OrderProductReturn
// 		var totalCost float64 = 0.0
// 		for _, product := range orderProds {
// 			if product.Prod.UserID == userID {
// 				userProd := OrderProductReturn{
// 					ProdID:   product.Prod.ID,
// 					ProdName: product.Prod.ProdName,
// 					Count:    product.Count,
// 					Price:    product.Cost,
// 				}
// 				totalCost += (product.Cost * float64(product.Count))
// 				userProds = append(userProds, userProd)
// 			}
// 		}

// 		if len(userProds) == 0 {
// 			// TODO: What should be sent in the case of no products owned by user present in order
// 			w.WriteHeader(http.StatusOK)
// 			w.Write([]byte("[]"))
// 			return
// 		}

// 		orderInfo := OrderReturn{
// 			OrderID:         order.Order.ID,
// 			CustomerName:    order.Order.CustomerName,
// 			CustomerEmail:   order.Order.CustomerEmail,
// 			OrderDate:       order.Order.OrderDate.String(),
// 			OrderStatus:     order.Order.OrderStatus,
// 			TrackingNumber:  order.Order.TrackingNumber,
// 			Total:           totalCost,
// 			OrderedProducts: userProds,
// 		}

// 		orderRet = append(orderRet, orderInfo)
// 	}

// 	ret, _ := json.Marshal(orderRet)
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(ret))
// }
