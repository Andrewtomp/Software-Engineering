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
	// var orderReturns []OrderReturn
	orderReturns := make([]OrderReturn, 0)
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
