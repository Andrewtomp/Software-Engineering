// internal/orderstable/orderstable_test.go
package orderstable

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log" // May need for product creation helper
	"net/http"
	"net/http/httptest"
	"os" // May need for product creation helper
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid" // May need for product creation helper
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"front-runner/internal/oauth"
	"front-runner/internal/prodtable" // Need product table structs and functions
	"front-runner/internal/usertable"
)

const projectDirName = "front-runner_backend"

// Global test variables
var (
	testDB           *gorm.DB
	testSessionStore *sessions.CookieStore
	setupEnvOnce     sync.Once
)

// setupTestEnvironment loads environment variables, initializes DB and session store for tests.
// It also clears relevant tables before each test run.
func setupTestEnvironment(t *testing.T) {
	t.Helper()

	setupEnvOnce.Do(func() {
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

		// Initialize DB connection
		coredbutils.ResetDBStateForTests()
		err = coredbutils.LoadEnv()
		require.NoError(t, err, "Failed to load core DB environment")
		var dbErr error
		testDB, dbErr = coredbutils.GetDB()
		require.NoError(t, dbErr, "Failed to get DB connection for tests")

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
		prodtable.Setup()                     // Uses coredbutils.GetDB()
		oauth.Setup(testSessionStore)         // Uses session store
		login.Setup(testDB, testSessionStore) // Uses DB and session store
		Setup()                               // Setup orderstable package (uses coredbutils.GetDB())

		// Run migrations once after setup
		usertable.MigrateUserDB()
		prodtable.MigrateProdDB()
		MigrateOrdersDB() // Migrates Order, OrderProd, OrderOwner tables
	})

	// 1. Clear dependent tables first (OrderOwner, OrderProd)
	require.NoError(t, testDB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&OrderOwner{}).Error, "Failed to clear order_owners table")
	require.NoError(t, testDB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&OrderProd{}).Error, "Failed to clear order_prods table")

	// 2. Clear tables that depend on others (Order, Product)
	// Order uses Unscoped() just in case, though it doesn't have gorm.Model by default
	require.NoError(t, testDB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Order{}).Error, "Failed to clear orders table")
	// Product uses Unscoped() as it might have soft delete hooks or relations
	require.NoError(t, testDB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&prodtable.Product{}).Error, "Failed to clear product table")

	// 3. Clear tables that are depended upon (Image, User)
	// Image uses Unscoped()
	require.NoError(t, testDB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&prodtable.Image{}).Error, "Failed to clear image table")
	// Assuming ClearUserTable handles its own potential soft delete logic if needed, or use Unscoped here too if necessary.
	require.NoError(t, usertable.ClearUserTable(testDB), "Failed to clear user table")
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

	const sessionName = "front-runner-session" // Use constant consistent with oauth/login
	const userSessionKey = "userID"

	session, err := testSessionStore.New(req, sessionName)
	require.NoError(t, err, "Failed to create new session for auth request")
	session.Values[userSessionKey] = user.ID
	rrCookieSetter := httptest.NewRecorder()
	err = testSessionStore.Save(req, rrCookieSetter, session)
	require.NoError(t, err, "Failed to save session to get cookie header")
	cookieHeader := rrCookieSetter.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookieHeader, "Set-Cookie header should not be empty after saving session")

	req.Header.Set("Cookie", cookieHeader)
	return req
}

// Helper to create a test product directly in the DB
func createTestProduct(t *testing.T, owner *usertable.User, name string, price float64, count uint) *prodtable.Product {
	t.Helper()
	// Create a dummy image record first (required by product schema)
	// No need to create actual file for order tests unless testing image links later
	dummyImage := prodtable.Image{
		URL:    fmt.Sprintf("dummy_%s.jpg", uuid.NewString()),
		UserID: owner.ID,
	}
	err := testDB.Create(&dummyImage).Error
	require.NoError(t, err, "Failed to create dummy image for product %s", name)

	product := &prodtable.Product{
		UserID:          owner.ID,
		ProdName:        name,
		ProdDescription: fmt.Sprintf("Description for %s", name),
		ImgID:           dummyImage.ID,
		ProdPrice:       price,
		ProdCount:       count,
		ProdTags:        "test",
	}
	err = testDB.Create(product).Error
	require.NoError(t, err, "Failed to create test product %s", name)
	require.NotZero(t, product.ID, "Created product %s has zero ID", name)
	return product
}

// Helper to create a full test order structure in the DB
func createTestOrder(t *testing.T, customerName, customerEmail string, products map[*prodtable.Product]uint) *Order {
	t.Helper()

	order := &Order{
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		OrderStatus:   "Pending", // Default status for tests
	}
	err := testDB.Create(order).Error
	require.NoError(t, err, "Failed to create base order record")
	require.NotZero(t, order.ID, "Created order has zero ID")

	sellerIDs := make(map[uint]bool)
	for product, count := range products {
		orderProd := &OrderProd{
			OrderID: order.ID,
			ProdID:  product.ID,
			Count:   count,
			Cost:    product.ProdPrice, // Record price at time of order
		}
		err = testDB.Create(orderProd).Error
		require.NoError(t, err, "Failed to create order_prod record for product %d", product.ID)
		sellerIDs[product.UserID] = true
	}

	for sellerID := range sellerIDs {
		orderOwner := &OrderOwner{
			UserID:  sellerID,
			OrderID: order.ID,
		}
		err = testDB.Create(orderOwner).Error
		require.NoError(t, err, "Failed to create order_owner record for user %d, order %d", sellerID, order.ID)
	}

	// Fetch the order again with preloads to ensure relations are set if needed later
	var finalOrder Order
	err = testDB.Preload("OrderProds.Prod").First(&finalOrder, order.ID).Error
	require.NoError(t, err, "Failed to fetch final order with preloads")

	return &finalOrder
}

// --- Test Cases ---

func TestCreateOrder(t *testing.T) {
	setupTestEnvironment(t)
	seller := createTestUser(t, "seller@example.com", "password")
	product1 := createTestProduct(t, seller, "Gadget", 10.50, 5)
	product2 := createTestProduct(t, seller, "Widget", 5.25, 10)

	t.Run("Success", func(t *testing.T) {
		payload := OrderCreatePayload{
			CustomerName:  "Test Customer",
			CustomerEmail: "customer@test.com",
			OrderedProducts: []OrderProductPayload{
				{ProdID: product1.ID, Count: 2},
				{ProdID: product2.ID, Count: 3},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		// Note: CreateOrder doesn't require authentication in the current implementation
		req := httptest.NewRequest("POST", "/api/create_order", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		CreateOrder(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created, body: %s", rr.Body.String())

		var resp map[string]uint
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err, "Failed to unmarshal response")
		createdOrderID, ok := resp["orderID"]
		require.True(t, ok, "Response should contain orderID")
		require.NotZero(t, createdOrderID, "Created orderID should not be zero")

		// Verify DB state
		var order Order
		err = testDB.First(&order, createdOrderID).Error
		require.NoError(t, err, "Failed to find created order in DB")
		assert.Equal(t, payload.CustomerName, order.CustomerName)
		assert.Equal(t, payload.CustomerEmail, order.CustomerEmail)
		assert.Equal(t, "Pending", order.OrderStatus)

		var orderProds []OrderProd
		err = testDB.Where("order_id = ?", createdOrderID).Find(&orderProds).Error
		require.NoError(t, err)
		require.Len(t, orderProds, 2, "Expected 2 order_prod records")
		// Add more checks for orderProds content (count, cost)

		var orderOwner OrderOwner
		err = testDB.Where("order_id = ? AND user_id = ?", createdOrderID, seller.ID).First(&orderOwner).Error
		require.NoError(t, err, "Failed to find order_owner record")

		// Verify stock update
		var updatedProd1 prodtable.Product
		testDB.First(&updatedProd1, product1.ID)
		assert.Equal(t, uint(3), updatedProd1.ProdCount, "Product1 stock should be 3") // 5 - 2

		var updatedProd2 prodtable.Product
		testDB.First(&updatedProd2, product2.ID)
		assert.Equal(t, uint(7), updatedProd2.ProdCount, "Product2 stock should be 7") // 10 - 3
	})

	t.Run("InsufficientStock", func(t *testing.T) {
		payload := OrderCreatePayload{
			CustomerName:  "Stock Test Customer",
			CustomerEmail: "stock@test.com",
			OrderedProducts: []OrderProductPayload{
				{ProdID: product1.ID, Count: 10}, // Only 5 available initially (or 3 after previous test)
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/create_order", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Need to reset product stock before this test if running sequentially
		testDB.Model(&prodtable.Product{}).Where("id = ?", product1.ID).Update("prod_count", 5)

		CreateOrder(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 Bad Request for insufficient stock")
		assert.Contains(t, rr.Body.String(), "insufficient stock")

		// Verify transaction rollback - order should not exist
		var order Order
		err := testDB.Where("customer_email = ?", payload.CustomerEmail).First(&order).Error
		assert.Error(t, err, "Order should not have been created")
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Expected ErrRecordNotFound for rolled back order")

		// Verify stock was NOT updated
		var updatedProd1 prodtable.Product
		testDB.First(&updatedProd1, product1.ID)
		assert.Equal(t, uint(5), updatedProd1.ProdCount, "Product1 stock should remain 5")
	})

	t.Run("ProductNotFound", func(t *testing.T) {
		nonExistentProductID := uint(99999)
		payload := OrderCreatePayload{
			CustomerName:  "Not Found Customer",
			CustomerEmail: "notfound@test.com",
			OrderedProducts: []OrderProductPayload{
				{ProdID: nonExistentProductID, Count: 1},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/create_order", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		CreateOrder(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "Expected 404 Not Found for non-existent product")
		assert.Contains(t, rr.Body.String(), "not found")

		// Verify transaction rollback
		var order Order
		err := testDB.Where("customer_email = ?", payload.CustomerEmail).First(&order).Error
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	// Add more tests for:
	// - Invalid JSON
	// - Missing customer name/email
	// - Empty orderedProducts array
	// - Zero count for a product
}

func TestGetOrder(t *testing.T) {
	setupTestEnvironment(t)
	seller1 := createTestUser(t, "seller1@example.com", "password")
	seller2 := createTestUser(t, "seller2@example.com", "password")
	productS1 := createTestProduct(t, seller1, "S1 Prod", 20.00, 10)
	productS2 := createTestProduct(t, seller2, "S2 Prod", 30.00, 5)

	// Create a test order with items from both sellers
	testOrder := createTestOrder(t, "GetOrder Cust", "getorder@test.com", map[*prodtable.Product]uint{
		productS1: 2, // Seller 1 owns this
		productS2: 1, // Seller 2 owns this
	})

	t.Run("Success_Seller1_OwnsItems", func(t *testing.T) {
		url := fmt.Sprintf("/api/get_order?id=%d", testOrder.ID)
		req := createAuthenticatedRequest(t, seller1, "GET", url, nil) // Authenticated as Seller 1
		rr := httptest.NewRecorder()

		GetOrder(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK, body: %s", rr.Body.String())
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp OrderReturn
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err, "Failed to unmarshal response")

		assert.Equal(t, testOrder.ID, resp.OrderID)
		assert.Equal(t, testOrder.CustomerName, resp.CustomerName)
		assert.Equal(t, testOrder.CustomerEmail, resp.CustomerEmail)
		assert.Equal(t, testOrder.OrderStatus, resp.OrderStatus)
		require.Len(t, resp.OrderedProducts, 1, "Seller 1 should see 1 product they own")
		assert.Equal(t, productS1.ID, resp.OrderedProducts[0].ProdID)
		assert.Equal(t, productS1.ProdName, resp.OrderedProducts[0].ProdName)
		assert.Equal(t, uint(2), resp.OrderedProducts[0].Count)
		assert.Equal(t, productS1.ProdPrice, resp.OrderedProducts[0].Price) // Price at time of order
		assert.Equal(t, 20.00*2, resp.Total, "Total should be for Seller 1's items only")
	})

	t.Run("Success_Seller2_OwnsItems", func(t *testing.T) {
		url := fmt.Sprintf("/api/get_order?id=%d", testOrder.ID)
		req := createAuthenticatedRequest(t, seller2, "GET", url, nil) // Authenticated as Seller 2
		rr := httptest.NewRecorder()

		GetOrder(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp OrderReturn
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.Len(t, resp.OrderedProducts, 1, "Seller 2 should see 1 product they own")
		assert.Equal(t, productS2.ID, resp.OrderedProducts[0].ProdID)
		assert.Equal(t, uint(1), resp.OrderedProducts[0].Count)
		assert.Equal(t, 30.00*1, resp.Total, "Total should be for Seller 2's items only")
	})

	t.Run("Forbidden_UserNotLinked", func(t *testing.T) {
		unrelatedUser := createTestUser(t, "unrelated@example.com", "password")
		url := fmt.Sprintf("/api/get_order?id=%d", testOrder.ID)
		req := createAuthenticatedRequest(t, unrelatedUser, "GET", url, nil) // Authenticated as unrelated user
		rr := httptest.NewRecorder()

		GetOrder(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Permission denied")
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		nonExistentID := 99999
		url := fmt.Sprintf("/api/get_order?id=%d", nonExistentID)
		req := createAuthenticatedRequest(t, seller1, "GET", url, nil)
		rr := httptest.NewRecorder()

		GetOrder(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "not found")
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		url := fmt.Sprintf("/api/get_order?id=%d", testOrder.ID)
		req := httptest.NewRequest("GET", url, nil) // No auth cookie
		rr := httptest.NewRecorder()

		GetOrder(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not authenticated")
	})

	// Add tests for:
	// - Invalid order ID format (e.g., ?id=abc) -> 400
}

func TestGetOrders(t *testing.T) {
	setupTestEnvironment(t)
	seller1 := createTestUser(t, "getseller1@example.com", "password")
	seller2 := createTestUser(t, "getseller2@example.com", "password")
	prodS1A := createTestProduct(t, seller1, "S1 Prod A", 10.00, 10)
	prodS1B := createTestProduct(t, seller1, "S1 Prod B", 15.00, 5)
	prodS2 := createTestProduct(t, seller2, "S2 Prod Only", 25.00, 8)

	// Order 1: Contains items from Seller 1 only
	order1 := createTestOrder(t, "Cust 1", "cust1@test.com", map[*prodtable.Product]uint{
		prodS1A: 1,
		prodS1B: 2,
	})

	// Order 2: Contains items from Seller 1 and Seller 2
	order2 := createTestOrder(t, "Cust 2", "cust2@test.com", map[*prodtable.Product]uint{
		prodS1A: 3,
		prodS2:  1,
	})

	// Order 3: Contains items from Seller 2 only
	_ = createTestOrder(t, "Cust 3", "cust3@test.com", map[*prodtable.Product]uint{
		prodS2: 2,
	})

	t.Run("Success_Seller1", func(t *testing.T) {
		req := createAuthenticatedRequest(t, seller1, "GET", "/api/get_orders", nil) // Authenticated as Seller 1
		rr := httptest.NewRecorder()

		GetOrders(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK, body: %s", rr.Body.String())
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp []OrderReturn
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err, "Failed to unmarshal response")

		require.Len(t, resp, 2, "Seller 1 should see 2 orders (Order 1 and Order 2)")

		// Check content (order might not be guaranteed, find by ID)
		foundOrder1 := false
		foundOrder2 := false
		for _, orderResp := range resp {
			if orderResp.OrderID == order1.ID {
				foundOrder1 = true
				assert.Len(t, orderResp.OrderedProducts, 2, "Order 1 response should have 2 products for Seller 1")
				assert.Equal(t, (10.00*1)+(15.00*2), orderResp.Total) // 10 + 30 = 40
			} else if orderResp.OrderID == order2.ID {
				foundOrder2 = true
				assert.Len(t, orderResp.OrderedProducts, 1, "Order 2 response should have 1 product for Seller 1")
				assert.Equal(t, prodS1A.ID, orderResp.OrderedProducts[0].ProdID)
				assert.Equal(t, 10.00*3, orderResp.Total) // 30
			}
		}
		assert.True(t, foundOrder1, "Response for Order 1 not found")
		assert.True(t, foundOrder2, "Response for Order 2 not found")
	})

	t.Run("Success_Seller2", func(t *testing.T) {
		req := createAuthenticatedRequest(t, seller2, "GET", "/api/get_orders", nil) // Authenticated as Seller 2
		rr := httptest.NewRecorder()

		GetOrders(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []OrderReturn
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Seller 2 is linked to Order 2 and Order 3
		require.Len(t, resp, 2, "Seller 2 should see 2 orders (Order 2 and Order 3)")
		// Add checks for content similar to Seller 1 test
	})

	t.Run("Success_NoOrders", func(t *testing.T) {
		newUser := createTestUser(t, "noorders@example.com", "password")
		req := createAuthenticatedRequest(t, newUser, "GET", "/api/get_orders", nil)
		rr := httptest.NewRecorder()

		GetOrders(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Equal(t, "[]", strings.TrimSpace(rr.Body.String()), "Expected empty JSON array for user with no orders")
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/get_orders", nil) // No auth
		rr := httptest.NewRecorder()

		GetOrders(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not authenticated")
	})
}
