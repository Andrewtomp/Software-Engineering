# Sprint 4
Video: [Sprint 4 VIDEO]()

## Completed Work
### Front end
- Merged all Sprint 3 content to main branch
- Updated README for dependencies and directions to start front end server
- Implemented API calls for fetching, updating, and deleting storefronts
- Implemented front end logic for storefront viewing on home page and storefronts page
- Implemented front end logic for updating an existing storefront
- Implemented front end logic for deleting storefronts
- Added delete option to storefront form
- Created navigation from home products and storefronts to their respective pages (e.g. ProductA is clicked on homepage, app navigates to products page, edit form for ProductA is opened)
- *Incomplete Sprint3 Issue:* Completed user interface for storefronts page
- Created Cypress and Unit Tests for this sprint's functionality

### Back end
- Merged all Sprint 3 content to main branch
- Updated the `generateCerts.sh` script to work on linux and mac
- Added option to host the page through ngrok
- Added oauth through google (only works through ngrok)
- Added additional option in the make file for running with ngrok
- Updated the `.env_template` to include the required variable for Ngrok and Google Auth
- Added backend functionality to retrieve order information
- Added backend functionality to create orders
- Updated backend to work locally and through oauth
- Updated user id retrieval to use oauth
- Updated all unit tests
- Updated Documentation to be more consistent

### Ngrok Traffic

![Ngrok_Traffic](https://github.com/user-attachments/assets/719168f6-7a31-4d8d-882c-7a3bd4e1ac50)


## Incomplete Work (For Both)
- Hook in order retrieval to front end
- Add order updater to front end
- Hook in with store front APIs to either retieve orders or automatically get sent orders

NOTE: Could not connect to other businesses because we are not a business entity

## Testing
### Front end
_Cypress Tests_

| Unit Test | Test Description |
| --- | --- |
| `Sprint 2 Test` | Tests that the add new button opens the Add Product modal |
| `Render Login` | Tests the rendering of login form |
| `Login Input` | Tests that the login form allows input |
| `Login Submission` | Tests login submission with valid information |
| `Login Error` | Tests that login submission fails with invalid information |
| `Navbar Navigation` | Tests that the Navbar navigates to the Products, Storefronts, Orders, and Settings pages |
| `Navbar Logout` | Tests that the Navbar logout button logs the user out and redirects them to the login page |
| `Navbar Visibility` | Tests that all navbar elements are visible |
| `Navbar Responsiveness` | Tests that the Navbar is responsive with a different screen size |
| `Order Label Download` | Tests that the download button in the orders table downloads a shipping label for that item |
| `Open Add Product` | Tests that the add new button opens the Add Product modal |
| `Product Form Validation` | Tests the validation of required fields in the Add Product form |
| `Add New Product` | Tests that a new product is added after the necessary, correct information is inputted |
| `Select Existing Product` | Tests that an existing product can be clicked and its information is populated in the form |
| `Update Product` | Tests that existing products can be edited and the changes are upheld |
| `Delete Product` | Tests that a product can be deleted |
| `Cancel Product Deletion` | Tests that deleting a product can be aborted after clicking the delete button |
| `Close Product Form` | Tests that clicking the close button on the product form closes the form |
| `Open Add Storefront` | Tests that the add new button opens the Add Storefront modal |
| `Storefront Form Validation` | Tests the validation of required fields in the Add Storefront form |
| `Add New Storefront` | Tests that a new storefront is added after the necessary, correct information is inputted |
| `Select Existing Storefront` | Tests that an existing storefront can be clicked and its information is populated in the form |
| `Update Storefront` | Tests that existing storefronts can be edited and the changes are upheld |
| `Delete Storefront` | Tests that a storefront can be deleted |
| `Cancel Storefront Deletion` | Tests that deleting a storefront can be aborted after clicking the delete button |
| `Close Storefront Form` | Tests that clicking the close button on the storefront form closes the form |

Tests passing:

![image](https://github.com/user-attachments/assets/fdfc2e8a-f08a-415b-9ed6-96dd560ea00a)
![image](https://github.com/user-attachments/assets/f13d9d21-5ee9-43b4-9e1b-b4f01d011d35)
![image](https://github.com/user-attachments/assets/e8cba87d-5fda-4dcf-8f9f-7f20f6178d51)
![image](https://github.com/user-attachments/assets/0ed22aae-86a3-420f-b765-ae444db0142c)
![image](https://github.com/user-attachments/assets/b3b4d064-66a0-48df-9caf-0afef86c35d9)
![image](https://github.com/user-attachments/assets/6465f370-5a5e-48b2-9b24-0dd2d71539ce)




_Unit Tests_

| Unit Test | Test Description |
| --- | --- |
| `LoginForm.Test` | Tests the rendering of login form, allowing user input, and submitting the form |
| `RegistrationForm.Test` | Tests the rendering of login form, allowing user input, validating the input, and submitting the form |
| `NavBar.Test` | Tests the routing from the nav bar to the home, products, storefronts, and orders pages |
| `ProductForm.Test` | Tests that the Product Form popup allows user input, validates the input, and submits the form |
| `StoreFrontForm.Test` | Tests that the Storefront Form popup allows user input, validates the input, and submits the form |

Tests passing:

![image](https://github.com/user-attachments/assets/602e8a84-fb83-49fc-b3c7-9e7bd85bc1c2)


### Back end
Each internal package has an associated unit test that can be run by entering the following command from the `front-runner_backend` directory:

```bash
go test ./internal/login # replace login with the desired internal package
```

Alternatively, the tests can be automatically run with an extension in vscode.

![backend_tests_sprint4](https://github.com/user-attachments/assets/a30916f3-e07d-4831-9524-7bb5b804034b)


_Unit Tests List_

| Unit Test | Test Description |
| :--- | :--- |
| `TestIsLocalHost` | Tests the `isLocalHost` helper function in `coredbutils`. |
| `TestLoadEnv` | Tests the `LoadEnv` function in `coredbutils` under various environment variable conditions (success, failure, runs once). |
| `TestGetDB_Integration` | Integration test for `GetDB` in `coredbutils`, checking connection to a live database based on `.env`. |
| `TestGetDB_Singleton` | Verifies `GetDB` in `coredbutils` returns the same instance on multiple calls. |
| `TestGetDB_Errors` | Tests error conditions for `GetDB` in `coredbutils` (LoadEnv failure, connection failure). |
| `TestLoginUser` | Tests successful login via the `LoginUser` handler in the `login` package. |
| `TestLoginUserInvalid` | Tests the `LoginUser` handler with an incorrect password. |
| `TestLoginUserNotFound` | Tests the `LoginUser` handler with a non-existent email. |
| `TestLoginUserMissingFields` | Tests the `LoginUser` handler with missing email or password form fields. |
| `TestLoginUserAlreadyLoggedIn` | Tests attempting to log in via `LoginUser` when already logged in. |
| `TestLogoutUser` | Tests successful logout via the `LogoutUser` handler when logged in. |
| `TestLogoutUserNotLoggedIn` | Tests the `LogoutUser` handler when not logged in (should still redirect). |
| `TestHandleGoogleLogin` | Tests the initiation of the Google OAuth flow via `HandleGoogleLogin` in the `oauth` package. |
| `TestHandleGoogleCallback` | Tests the `HandleGoogleCallback` handler for new user registration, existing user login, and error scenarios. |
| `TestHandleLogout` | Tests the OAuth logout handler `HandleLogout` in the `oauth` package. |
| `TestGetCurrentUser` | Tests the `GetCurrentUser` helper function in `oauth` for retrieving user data from a session under various conditions. |
| `TestCreateOrder` | Tests the `CreateOrder` handler in `orderstable` for successful order creation, stock updates, and error handling (insufficient stock, etc.). |
| `TestGetOrder` | Tests the `GetOrder` handler in `orderstable` for retrieving a specific order, ensuring correct filtering based on the authenticated seller. |
| `TestGetOrders` | Tests the `GetOrders` handler in `orderstable` for retrieving all orders relevant to the authenticated seller. |
| `TestAddProduct` | Tests the `AddProduct` handler in `prodtable` for creating a new product with an image upload. |
| `TestDeleteProduct` | Tests the `DeleteProduct` handler in `prodtable`, including deletion of the associated image file and record. |
| `TestUpdateProduct` | Tests the `UpdateProduct` handler in `prodtable` for updating product details and optionally replacing the image. |
| `TestGetProduct` | Tests the `GetProduct` handler in `prodtable` for retrieving details of a specific product owned by the user. |
| `TestGetProducts` | Tests the `GetProducts` handler in `prodtable` for retrieving all products owned by the user. |
| `TestGetProductImage` | Tests the `GetProductImage` handler in `prodtable` for serving a product's image file. |
| `TestGetProductImage_NotFound` | Tests `GetProductImage` scenarios where the image record or file is not found. |
| `TestProduct_Auth` | Tests authentication and authorization failures across various product endpoints (unauthenticated, wrong user). |
| `TestRouteExistenceAndBasicHandling` | Checks if routes defined in the `routes` package are registered and return expected basic status codes (including redirects for auth). |
| `TestSPAHandler` | Tests the `spaHandler` logic directly for serving static files and the index fallback. |
| `TestAuthMiddleware` | Tests the `authMiddleware` logic directly for allowing/blocking requests based on session state. |
| `TestInvalidAPI` | Tests the `InvalidAPI` handler for undefined API routes. |
| `TestAddStorefront` | Tests the `AddStorefront` handler in `storefronttable` for linking storefronts (success, unauthorized, missing fields, duplicates). |
| `TestGetUpdateDeleteFlow` | Tests the full lifecycle (add, get, update, delete) for storefront links via their respective handlers in `storefronttable`. |
| `TestSpecificErrors` | Tests error scenarios (forbidden, not found, missing ID) for storefront update and delete handlers in `storefronttable`. |
| `TestCreateAndGetUser` | Tests the `CreateUser` and `GetUserByEmail` functions directly in the `usertable` package. |
| `TestRegisterUserHandler` | Tests the `RegisterUser` HTTP handler, including success, missing fields, invalid email, and duplicate email scenarios. |
| `TestCreateUserFunction` | Tests the `CreateUser` function directly with various valid and invalid inputs. |
| `TestGetUserByID` | Tests the `GetUserByID` function for retrieving users by their primary key. |
| `TestGetUserByProviderID` | Tests the `GetUserByProviderID` function for retrieving users by OAuth provider details. |
| `TestUpdateUser` | Tests the `UpdateUser` function for updating existing user records and handling errors (no ID, non-existent user). |
| `TestValidEmail` | Tests the `Valid` function in the `validemail` package with a correctly formatted email address. |
| `TestInvalidEmail` | Tests the `Valid` function in the `validemail` package with various improperly formatted email addresses. |

## Front Runner API Documentation

API documentation for the Front Runner application.

### Version
1.0

### License

[MIT](http://www.apache.org/licenses/LICENSE-2.0.html)

### Contact

API Support jonathan.bravo@ufl.edu 

### URI Schemes
  * http

### Consumes
  * application/json
  * multipart/form-data
  * application/x-www-form-urlencoded

### Produces
  * image/*
  * application/json
  * text/plain

## All endpoints

###  authentication

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/login | [post API login](#post-api-login) | User Login (Email/Password) |
| POST | /api/register | [post API register](#post-api-register) | Register a new local user |

###  authentication_o_auth

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /auth/google | [get auth google](#get-auth-google) | Initiate Google Login |
| GET | /auth/google/callback | [get auth google callback](#get-auth-google-callback) | Google Login Callback |
| GET | /logout | [get logout](#get-logout) | User Logout |

###  order

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/get_order | [get API get order](#get-api-get-order) | Retrieve an order (filtered for seller) |
| GET | /api/get_orders | [get API get orders](#get-api-get-orders) | Retrieve user's sales orders |
| POST | /api/create_order | [post API create order](#post-api-create-order) | Creates an order |

###  products

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/products | [delete API products](#delete-api-products) | Delete a product |
| GET | /api/products | [get API products](#get-api-products) | Get all products for the user |
| GET | /api/products/details | [get API products details](#get-api-products-details) | Get a specific product |
| GET | /api/products/image | [get API products image](#get-api-products-image) | Get a product image |
| POST | /api/products | [post API products](#post-api-products) | Add a new product |
| PUT | /api/products | [put API products](#put-api-products) | Update a product |

###  storefronts

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/delete_storefront | [delete API delete storefront](#delete-api-delete-storefront) | Unlink a storefront |
| GET | /api/get_storefronts | [get API get storefronts](#get-api-get-storefronts) | Get linked storefronts |
| POST | /api/add_storefront | [post API add storefront](#post-api-add-storefront) | Link a new storefront |
| PUT | /api/update_storefront | [put API update storefront](#put-api-update-storefront) | Update a storefront link |

## Paths

### <span id="delete-api-delete-storefront"></span> Unlink a storefront (*DeleteAPIDeleteStorefront*)

```
DELETE /api/delete_storefront
```

Removes the link to an external storefront specified by its unique ID. User must own the link. Requires authentication.

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `query` | uint (formatted integer) | `uint64` |  | ✓ |  | ID of the Storefront Link to delete |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-api-delete-storefront-200) | OK | Storefront unlinked successfully |  | [schema](#delete-api-delete-storefront-200-schema) |
| [204](#delete-api-delete-storefront-204) | No Content | Storefront unlinked successfully (No Content)" // Added 204 as an alternative success |  | [schema](#delete-api-delete-storefront-204-schema) |
| [400](#delete-api-delete-storefront-400) | Bad Request | Bad Request - Invalid or missing 'id' query parameter |  | [schema](#delete-api-delete-storefront-400-schema) |
| [401](#delete-api-delete-storefront-401) | Unauthorized | Unauthorized - User session invalid or expired |  | [schema](#delete-api-delete-storefront-401-schema) |
| [403](#delete-api-delete-storefront-403) | Forbidden | Forbidden - User does not own this storefront link |  | [schema](#delete-api-delete-storefront-403-schema) |
| [404](#delete-api-delete-storefront-404) | Not Found | Not Found - Storefront link with the specified ID not found |  | [schema](#delete-api-delete-storefront-404-schema) |
| [500](#delete-api-delete-storefront-500) | Internal Server Error | Internal Server Error - Database deletion failed |  | [schema](#delete-api-delete-storefront-500-schema) |

#### Responses

##### <span id="delete-api-delete-storefront-200"></span> 200 - Storefront unlinked successfully
Status: OK

###### <span id="delete-api-delete-storefront-200-schema"></span> Schema
   
##### <span id="delete-api-delete-storefront-204"></span> 204 - Storefront unlinked successfully (No Content)" // Added 204 as an alternative success
Status: No Content

###### <span id="delete-api-delete-storefront-204-schema"></span> Schema
   
##### <span id="delete-api-delete-storefront-400"></span> 400 - Bad Request - Invalid or missing 'id' query parameter
Status: Bad Request

###### <span id="delete-api-delete-storefront-400-schema"></span> Schema
   

##### <span id="delete-api-delete-storefront-401"></span> 401 - Unauthorized - User session invalid or expired
Status: Unauthorized

###### <span id="delete-api-delete-storefront-401-schema"></span> Schema

##### <span id="delete-api-delete-storefront-403"></span> 403 - Forbidden - User does not own this storefront link
Status: Forbidden

###### <span id="delete-api-delete-storefront-403-schema"></span> Schema
   
##### <span id="delete-api-delete-storefront-404"></span> 404 - Not Found - Storefront link with the specified ID not found
Status: Not Found

###### <span id="delete-api-delete-storefront-404-schema"></span> Schema
   
##### <span id="delete-api-delete-storefront-500"></span> 500 - Internal Server Error - Database deletion failed
Status: Internal Server Error

###### <span id="delete-api-delete-storefront-500-schema"></span> Schema
   
### <span id="delete-api-products"></span> Delete a product (*DeleteAPIProducts*)

```
DELETE /api/products
```

Deletes a specific product owned by the authenticated user, identified by its ID. Also deletes the associated image file and record.

#### Produces
  * text/plain

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `query` | uint64 (formatted integer) | `uint64` |  | ✓ |  | ID of the product to delete |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-api-products-200) | OK | Product deleted successfully |  | [schema](#delete-api-products-200-schema) |
| [400](#delete-api-products-400) | Bad Request | Bad Request: Invalid Product ID |  | [schema](#delete-api-products-400-schema) |
| [401](#delete-api-products-401) | Unauthorized | Unauthorized: User not authenticated |  | [schema](#delete-api-products-401-schema) |
| [403](#delete-api-products-403) | Forbidden | Forbidden: User does not own this product |  | [schema](#delete-api-products-403-schema) |
| [404](#delete-api-products-404) | Not Found | Not Found: Product not found |  | [schema](#delete-api-products-404-schema) |
| [500](#delete-api-products-500) | Internal Server Error | Internal Server Error: Database or file system error during deletion |  | [schema](#delete-api-products-500-schema) |

#### Responses

##### <span id="delete-api-products-200"></span> 200 - Product deleted successfully
Status: OK

###### <span id="delete-api-products-200-schema"></span> Schema
   
##### <span id="delete-api-products-400"></span> 400 - Bad Request: Invalid Product ID
Status: Bad Request

###### <span id="delete-api-products-400-schema"></span> Schema
   
##### <span id="delete-api-products-401"></span> 401 - Unauthorized: User not authenticated
Status: Unauthorized

###### <span id="delete-api-products-401-schema"></span> Schema
   
##### <span id="delete-api-products-403"></span> 403 - Forbidden: User does not own this product
Status: Forbidden

###### <span id="delete-api-products-403-schema"></span> Schema
   
##### <span id="delete-api-products-404"></span> 404 - Not Found: Product not found
Status: Not Found

###### <span id="delete-api-products-404-schema"></span> Schema
   
##### <span id="delete-api-products-500"></span> 500 - Internal Server Error: Database or file system error during deletion
Status: Internal Server Error

###### <span id="delete-api-products-500-schema"></span> Schema
   
### <span id="get-api-get-order"></span> Retrieve an order (filtered for seller) (*GetAPIGetOrder*)

```
GET /api/get_order
```

Retrieves an existing order and its associated products *owned by the authenticated user (seller)*.

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `query` | integer | `int64` |  | ✓ |  | Order ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-api-get-order-200) | OK | JSON representation of the order's information relevant to the user (empty object if user has no items in this order) |  | [schema](#get-api-get-order-200-schema) |
| [400](#get-api-get-order-400) | Bad Request | Invalid Order ID format |  | [schema](#get-api-get-order-400-schema) |
| [401](#get-api-get-order-401) | Unauthorized | User not authenticated |  | [schema](#get-api-get-order-401-schema) |
| [403](#get-api-get-order-403) | Forbidden | Permission denied (user is not a seller for any product in this order) |  | [schema](#get-api-get-order-403-schema) |
| [404](#get-api-get-order-404) | Not Found | Order not found |  | [schema](#get-api-get-order-404-schema) |
| [500](#get-api-get-order-500) | Internal Server Error | Internal server error |  | [schema](#get-api-get-order-500-schema) |

#### Responses


##### <span id="get-api-get-order-200"></span> 200 - JSON representation of the order's information relevant to the user (empty object if user has no items in this order)
Status: OK

###### <span id="get-api-get-order-200-schema"></span> Schema
   
[OrderstableOrderReturn](#orderstable-order-return)

##### <span id="get-api-get-order-400"></span> 400 - Invalid Order ID format
Status: Bad Request

###### <span id="get-api-get-order-400-schema"></span> Schema
   
##### <span id="get-api-get-order-401"></span> 401 - User not authenticated
Status: Unauthorized

###### <span id="get-api-get-order-401-schema"></span> Schema
   
##### <span id="get-api-get-order-403"></span> 403 - Permission denied (user is not a seller for any product in this order)
Status: Forbidden

###### <span id="get-api-get-order-403-schema"></span> Schema
   
##### <span id="get-api-get-order-404"></span> 404 - Order not found
Status: Not Found

###### <span id="get-api-get-order-404-schema"></span> Schema
   
##### <span id="get-api-get-order-500"></span> 500 - Internal server error
Status: Internal Server Error

###### <span id="get-api-get-order-500-schema"></span> Schema
   
### <span id="get-api-get-orders"></span> Retrieve user's sales orders (*GetAPIGetOrders*)

```
GET /api/get_orders
```

Retrieves orders containing products sold by the authenticated user, along with the relevant product details for each order.

#### Security Requirements
  * ApiKeyAuth

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-api-get-orders-200) | OK | JSON array of orders relevant to the user (empty array if none) |  | [schema](#get-api-get-orders-200-schema) |
| [401](#get-api-get-orders-401) | Unauthorized | User not authenticated |  | [schema](#get-api-get-orders-401-schema) |
| [500](#get-api-get-orders-500) | Internal Server Error | Internal server error |  | [schema](#get-api-get-orders-500-schema) |

#### Responses

##### <span id="get-api-get-orders-200"></span> 200 - JSON array of orders relevant to the user (empty array if none)
Status: OK

###### <span id="get-api-get-orders-200-schema"></span> Schema
   
[][OrderstableOrderReturn](#orderstable-order-return)

##### <span id="get-api-get-orders-401"></span> 401 - User not authenticated
Status: Unauthorized

###### <span id="get-api-get-orders-401-schema"></span> Schema
   
##### <span id="get-api-get-orders-500"></span> 500 - Internal server error
Status: Internal Server Error

###### <span id="get-api-get-orders-500-schema"></span> Schema
   
### <span id="get-api-get-storefronts"></span> Get linked storefronts (*GetAPIGetStorefronts*)

```
GET /api/get_storefronts
```

Retrieves a list of all external storefronts linked by the currently authenticated user. Credentials are *never* included. Requires authentication.

#### Security Requirements
  * ApiKeyAuth

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-api-get-storefronts-200) | OK | List of linked storefronts (empty array if none) |  | [schema](#get-api-get-storefronts-200-schema) |
| [401](#get-api-get-storefronts-401) | Unauthorized | Unauthorized - User session invalid or expired |  | [schema](#get-api-get-storefronts-401-schema) |
| [500](#get-api-get-storefronts-500) | Internal Server Error | Internal Server Error - Database query failed |  | [schema](#get-api-get-storefronts-500-schema) |

#### Responses

##### <span id="get-api-get-storefronts-200"></span> 200 - List of linked storefronts (empty array if none)
Status: OK

###### <span id="get-api-get-storefronts-200-schema"></span> Schema
   
[][StorefronttableStorefrontLinkReturn](#storefronttable-storefront-link-return)

##### <span id="get-api-get-storefronts-401"></span> 401 - Unauthorized - User session invalid or expired
Status: Unauthorized

###### <span id="get-api-get-storefronts-401-schema"></span> Schema
   
##### <span id="get-api-get-storefronts-500"></span> 500 - Internal Server Error - Database query failed
Status: Internal Server Error

###### <span id="get-api-get-storefronts-500-schema"></span> Schema

### <span id="get-api-products"></span> Get all products for the user (*GetAPIProducts*)

```
GET /api/products
```

Retrieves a list of all products owned by the authenticated user.

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-api-products-200) | OK | Successfully retrieved list of products |  | [schema](#get-api-products-200-schema) |
| [401](#get-api-products-401) | Unauthorized | Unauthorized: User not authenticated |  | [schema](#get-api-products-401-schema) |
| [500](#get-api-products-500) | Internal Server Error | Internal Server Error: Database error |  | [schema](#get-api-products-500-schema) |

#### Responses


##### <span id="get-api-products-200"></span> 200 - Successfully retrieved list of products
Status: OK

###### <span id="get-api-products-200-schema"></span> Schema
  
[][ProdtableProductReturn](#prodtable-product-return)

##### <span id="get-api-products-401"></span> 401 - Unauthorized: User not authenticated
Status: Unauthorized

###### <span id="get-api-products-401-schema"></span> Schema
   
##### <span id="get-api-products-500"></span> 500 - Internal Server Error: Database error
Status: Internal Server Error

###### <span id="get-api-products-500-schema"></span> Schema
   
### <span id="get-api-products-details"></span> Get a specific product (*GetAPIProductsDetails*)

```
GET /api/products/details
```

Retrieves details for a specific product owned by the authenticated user, identified by its ID.

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `query` | uint64 (formatted integer) | `uint64` |  | ✓ |  | ID of the product to retrieve |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-api-products-details-200) | OK | Successfully retrieved product details |  | [schema](#get-api-products-details-200-schema) |
| [400](#get-api-products-details-400) | Bad Request | Bad Request: Invalid Product ID |  | [schema](#get-api-products-details-400-schema) |
| [401](#get-api-products-details-401) | Unauthorized | Unauthorized: User not authenticated |  | [schema](#get-api-products-details-401-schema) |
| [403](#get-api-products-details-403) | Forbidden | Forbidden: User does not own this product |  | [schema](#get-api-products-details-403-schema) |
| [404](#get-api-products-details-404) | Not Found | Not Found: Product not found |  | [schema](#get-api-products-details-404-schema) |
| [500](#get-api-products-details-500) | Internal Server Error | Internal Server Error: Database error |  | [schema](#get-api-products-details-500-schema) |

#### Responses


##### <span id="get-api-products-details-200"></span> 200 - Successfully retrieved product details
Status: OK

###### <span id="get-api-products-details-200-schema"></span> Schema
   
[ProdtableProductReturn](#prodtable-product-return)

##### <span id="get-api-products-details-400"></span> 400 - Bad Request: Invalid Product ID
Status: Bad Request

###### <span id="get-api-products-details-400-schema"></span> Schema
   
##### <span id="get-api-products-details-401"></span> 401 - Unauthorized: User not authenticated
Status: Unauthorized

###### <span id="get-api-products-details-401-schema"></span> Schema
   
##### <span id="get-api-products-details-403"></span> 403 - Forbidden: User does not own this product
Status: Forbidden

###### <span id="get-api-products-details-403-schema"></span> Schema

##### <span id="get-api-products-details-404"></span> 404 - Not Found: Product not found
Status: Not Found

###### <span id="get-api-products-details-404-schema"></span> Schema
   
##### <span id="get-api-products-details-500"></span> 500 - Internal Server Error: Database error
Status: Internal Server Error

###### <span id="get-api-products-details-500-schema"></span> Schema
   
### <span id="get-api-products-image"></span> Get a product image (*GetAPIProductsImage*)

```
GET /api/products/image
```

Retrieves and serves the image file associated with a product, identified by its filename. Requires the user to be authenticated and own the product/image.

#### Produces
  * image/*

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| image | `query` | string | `string` |  | ✓ |  | Filename of the image to retrieve (e.g., 'uuid.jpg') |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-api-products-image-200) | OK | Product image file |  | [schema](#get-api-products-image-200-schema) |
| [400](#get-api-products-image-400) | Bad Request | Bad Request: Missing or invalid image filename |  | [schema](#get-api-products-image-400-schema) |
| [401](#get-api-products-image-401) | Unauthorized | Unauthorized: User not authenticated |  | [schema](#get-api-products-image-401-schema) |
| [403](#get-api-products-image-403) | Forbidden | Forbidden: User does not own this image |  | [schema](#get-api-products-image-403-schema) |
| [404](#get-api-products-image-404) | Not Found | Not Found: Image metadata or file not found |  | [schema](#get-api-products-image-404-schema) |
| [500](#get-api-products-image-500) | Internal Server Error | Internal Server Error: Database or file system error |  | [schema](#get-api-products-image-500-schema) |

#### Responses

##### <span id="get-api-products-image-200"></span> 200 - Product image file
Status: OK

###### <span id="get-api-products-image-200-schema"></span> Schema
   
##### <span id="get-api-products-image-400"></span> 400 - Bad Request: Missing or invalid image filename
Status: Bad Request

###### <span id="get-api-products-image-400-schema"></span> Schema
   
##### <span id="get-api-products-image-401"></span> 401 - Unauthorized: User not authenticated
Status: Unauthorized

###### <span id="get-api-products-image-401-schema"></span> Schema
   
##### <span id="get-api-products-image-403"></span> 403 - Forbidden: User does not own this image
Status: Forbidden

###### <span id="get-api-products-image-403-schema"></span> Schema

##### <span id="get-api-products-image-404"></span> 404 - Not Found: Image metadata or file not found
Status: Not Found

###### <span id="get-api-products-image-404-schema"></span> Schema
   
##### <span id="get-api-products-image-500"></span> 500 - Internal Server Error: Database or file system error
Status: Internal Server Error

###### <span id="get-api-products-image-500-schema"></span> Schema
   
### <span id="get-auth-google"></span> Initiate Google Login (*GetAuthGoogle*)

```
GET /auth/google
```

Redirects the user to Google for authentication as part of the OAuth2 flow.

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [307](#get-auth-google-307) | Temporary Redirect | Redirects to Google's authentication endpoint |  | [schema](#get-auth-google-307-schema) |
| [500](#get-auth-google-500) | Internal Server Error | Internal Server Error (if Goth setup fails) |  | [schema](#get-auth-google-500-schema) |

#### Responses

##### <span id="get-auth-google-307"></span> 307 - Redirects to Google's authentication endpoint
Status: Temporary Redirect

###### <span id="get-auth-google-307-schema"></span> Schema
   
##### <span id="get-auth-google-500"></span> 500 - Internal Server Error (if Goth setup fails)
Status: Internal Server Error

###### <span id="get-auth-google-500-schema"></span> Schema
   
### <span id="get-auth-google-callback"></span> Google Login Callback (*GetAuthGoogleCallback*)

```
GET /auth/google/callback
```

Handles the callback from Google after authentication. Creates a user session upon successful authentication and redirects to the homepage.

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [307](#get-auth-google-callback-307) | Temporary Redirect | Redirects to / on successful login |  | [schema](#get-auth-google-callback-307-schema) |
| [400](#get-auth-google-callback-400) | Bad Request | Bad Request (e.g., state mismatch)" // Goth might handle this |  | [schema](#get-auth-google-callback-400-schema) |
| [500](#get-auth-google-callback-500) | Internal Server Error | Internal Server Error (session, database, or Goth issue) |  | [schema](#get-auth-google-callback-500-schema) |

#### Responses


##### <span id="get-auth-google-callback-307"></span> 307 - Redirects to / on successful login
Status: Temporary Redirect

###### <span id="get-auth-google-callback-307-schema"></span> Schema

##### <span id="get-auth-google-callback-400"></span> 400 - Bad Request (e.g., state mismatch)" // Goth might handle this
Status: Bad Request

###### <span id="get-auth-google-callback-400-schema"></span> Schema

##### <span id="get-auth-google-callback-500"></span> 500 - Internal Server Error (session, database, or Goth issue)
Status: Internal Server Error

###### <span id="get-auth-google-callback-500-schema"></span> Schema

### <span id="get-logout"></span> User Logout (*GetLogout*)

```
GET /logout
```

Logs out the current user by clearing the session cookie and redirects to the homepage.

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [307](#get-logout-307) | Temporary Redirect | Redirects to / after logout |  | [schema](#get-logout-307-schema) |
| [500](#get-logout-500) | Internal Server Error | Internal Server Error (if saving cleared session fails) |  | [schema](#get-logout-500-schema) |

#### Responses

##### <span id="get-logout-307"></span> 307 - Redirects to / after logout
Status: Temporary Redirect

###### <span id="get-logout-307-schema"></span> Schema

##### <span id="get-logout-500"></span> 500 - Internal Server Error (if saving cleared session fails)
Status: Internal Server Error

###### <span id="get-logout-500-schema"></span> Schema

### <span id="post-api-add-storefront"></span> Link a new storefront (*PostAPIAddStorefront*)

```
POST /api/add_storefront
```

Links a new external storefront (e.g., Amazon, Pinterest) to the user's account, storing credentials securely. Requires authentication.

#### Consumes
  * application/json

#### Security Requirements
  * ApiKeyAuth // Assuming ApiKeyAuth is defined for session/token auth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| storefrontLink | `body` | [StorefronttableStorefrontLinkAddPayload](#storefronttable-storefront-link-add-payload) | `models.StorefronttableStorefrontLinkAddPayload` | | ✓ | | Storefront Link Details (including credentials like apiKey, apiSecret) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#post-api-add-storefront-201) | Created | Successfully linked storefront (credentials omitted) |  | [schema](#post-api-add-storefront-201-schema) |
| [400](#post-api-add-storefront-400) | Bad Request | Bad Request - Invalid input, missing fields, or JSON parsing error |  | [schema](#post-api-add-storefront-400-schema) |
| [401](#post-api-add-storefront-401) | Unauthorized | Unauthorized - User session invalid or expired |  | [schema](#post-api-add-storefront-401-schema) |
| [409](#post-api-add-storefront-409) | Conflict | Conflict - A link with this name/type already exists for the user |  | [schema](#post-api-add-storefront-409-schema) |
| [500](#post-api-add-storefront-500) | Internal Server Error | Internal Server Error - E.g., failed to encrypt, database error |  | [schema](#post-api-add-storefront-500-schema) |

#### Responses

##### <span id="post-api-add-storefront-201"></span> 201 - Successfully linked storefront (credentials omitted)
Status: Created

###### <span id="post-api-add-storefront-201-schema"></span> Schema

[StorefronttableStorefrontLinkReturn](#storefronttable-storefront-link-return)

##### <span id="post-api-add-storefront-400"></span> 400 - Bad Request - Invalid input, missing fields, or JSON parsing error
Status: Bad Request

###### <span id="post-api-add-storefront-400-schema"></span> Schema
   
##### <span id="post-api-add-storefront-401"></span> 401 - Unauthorized - User session invalid or expired
Status: Unauthorized

###### <span id="post-api-add-storefront-401-schema"></span> Schema

##### <span id="post-api-add-storefront-409"></span> 409 - Conflict - A link with this name/type already exists for the user
Status: Conflict

###### <span id="post-api-add-storefront-409-schema"></span> Schema

##### <span id="post-api-add-storefront-500"></span> 500 - Internal Server Error - E.g., failed to encrypt, database error
Status: Internal Server Error

###### <span id="post-api-add-storefront-500-schema"></span> Schema

### <span id="post-api-create-order"></span> Creates an order (*PostAPICreateOrder*)

```
POST /api/create_order
```

Creates a new order entry with customer details and products. Updates product stock and links sellers.

#### Consumes
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| orderInfo | `body` | [OrderstableOrderCreatePayload](#orderstable-order-create-payload) | `models.OrderstableOrderCreatePayload` | | ✓ | | Order Details |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#post-api-create-order-201) | Created | Order created successfully, returns order ID" // Example success response |  | [schema](#post-api-create-order-201-schema) |
| [400](#post-api-create-order-400) | Bad Request | Invalid request body, missing fields, or invalid product data |  | [schema](#post-api-create-order-400-schema) |
| [404](#post-api-create-order-404) | Not Found | Product not found or insufficient stock |  | [schema](#post-api-create-order-404-schema) |
| [500](#post-api-create-order-500) | Internal Server Error | Internal server error during order processing |  | [schema](#post-api-create-order-500-schema) |

#### Responses

##### <span id="post-api-create-order-201"></span> 201 - Order created successfully, returns order ID" // Example success response
Status: Created

###### <span id="post-api-create-order-201-schema"></span> Schema
   
map of integer

##### <span id="post-api-create-order-400"></span> 400 - Invalid request body, missing fields, or invalid product data
Status: Bad Request

###### <span id="post-api-create-order-400-schema"></span> Schema

##### <span id="post-api-create-order-404"></span> 404 - Product not found or insufficient stock
Status: Not Found

###### <span id="post-api-create-order-404-schema"></span> Schema

##### <span id="post-api-create-order-500"></span> 500 - Internal server error during order processing
Status: Internal Server Error

###### <span id="post-api-create-order-500-schema"></span> Schema

### <span id="post-api-login"></span> User Login (Email/Password) (*PostAPILogin*)

```
POST /api/login
```

Authenticates a user using email and password. Creates a session cookie upon successful authentication and redirects to the homepage.

#### Consumes
  * application/x-www-form-urlencoded

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| email | `formData` | string | `string` |  | ✓ |  | User's Email Address |
| password | `formData` | string | `string` |  | ✓ |  | User's Password |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [303](#post-api-login-303) | See Other | Redirects to / on successful login |  | [schema](#post-api-login-303-schema) |
| [400](#post-api-login-400) | Bad Request | Bad Request: Email and password are required |  | [schema](#post-api-login-400-schema) |
| [401](#post-api-login-401) | Unauthorized | Unauthorized: Invalid credentials |  | [schema](#post-api-login-401-schema) |
| [500](#post-api-login-500) | Internal Server Error | Internal Server Error |  | [schema](#post-api-login-500-schema) |

#### Responses

##### <span id="post-api-login-303"></span> 303 - Redirects to / on successful login
Status: See Other

###### <span id="post-api-login-303-schema"></span> Schema

##### <span id="post-api-login-400"></span> 400 - Bad Request: Email and password are required
Status: Bad Request

###### <span id="post-api-login-400-schema"></span> Schema

##### <span id="post-api-login-401"></span> 401 - Unauthorized: Invalid credentials
Status: Unauthorized

###### <span id="post-api-login-401-schema"></span> Schema

##### <span id="post-api-login-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="post-api-login-500-schema"></span> Schema

### <span id="post-api-products"></span> Add a new product (*PostAPIProducts*)

```
POST /api/products
```

Creates a new product listing associated with the authenticated user. Requires product details and an image upload.

#### Consumes
  * multipart/form-data

#### Produces
  * text/plain

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| count | `formData` | int32 (formatted integer) | `int32` |  | ✓ |  | Available stock count |
| description | `formData` | string | `string` |  | ✓ |  | Description of the product |
| image | `formData` | file | `io.ReadCloser` |  | ✓ |  | Product image file |
| price | `formData` | float (formatted number) | `float32` |  | ✓ |  | Price of the product (e.g., 19.99) |
| productName | `formData` | string | `string` |  | ✓ |  | Name of the product |
| tags | `formData` | string | `string` |  |  |  | Comma-separated tags for the product |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#post-api-products-201) | Created | Product added successfully |  | [schema](#post-api-products-201-schema) |
| [400](#post-api-products-400) | Bad Request | Bad Request: Missing required fields, invalid data format, or image error |  | [schema](#post-api-products-400-schema) |
| [401](#post-api-products-401) | Unauthorized | Unauthorized: User not authenticated |  | [schema](#post-api-products-401-schema) |
| [500](#post-api-products-500) | Internal Server Error | Internal Server Error: Database or file system error |  | [schema](#post-api-products-500-schema) |

#### Responses

##### <span id="post-api-products-201"></span> 201 - Product added successfully
Status: Created

###### <span id="post-api-products-201-schema"></span> Schema

##### <span id="post-api-products-400"></span> 400 - Bad Request: Missing required fields, invalid data format, or image error
Status: Bad Request

###### <span id="post-api-products-400-schema"></span> Schema

##### <span id="post-api-products-401"></span> 401 - Unauthorized: User not authenticated
Status: Unauthorized

###### <span id="post-api-products-401-schema"></span> Schema

##### <span id="post-api-products-500"></span> 500 - Internal Server Error: Database or file system error
Status: Internal Server Error

###### <span id="post-api-products-500-schema"></span> Schema

### <span id="post-api-register"></span> Register a new local user (*PostAPIRegister*)

```
POST /api/register
```

Registers a new user account using email and password for local authentication.

#### Consumes
  * application/x-www-form-urlencoded

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| businessName | `formData` | string | `string` |  |  |  | User's Business Name (Optional) |
| email | `formData` | string | `string` |  | ✓ |  | User's Email Address |
| name | `formData` | string | `string` |  | ✓ |  | User's Full Name |
| password | `formData` | string | `string` |  | ✓ |  | User's Password (min length recommended) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-api-register-200) | OK | User registered successfully |  | [schema](#post-api-register-200-schema) |
| [400](#post-api-register-400) | Bad Request | Bad Request: Missing required fields (email, password, name), or invalid email format |  | [schema](#post-api-register-400-schema) |
| [409](#post-api-register-409) | Conflict | Conflict: Email address is already registered |  | [schema](#post-api-register-409-schema) |
| [500](#post-api-register-500) | Internal Server Error | Internal Server Error: Failed to hash password or save user to database |  | [schema](#post-api-register-500-schema) |

#### Responses

##### <span id="post-api-register-200"></span> 200 - User registered successfully
Status: OK

###### <span id="post-api-register-200-schema"></span> Schema

##### <span id="post-api-register-400"></span> 400 - Bad Request: Missing required fields (email, password, name), or invalid email format
Status: Bad Request

###### <span id="post-api-register-400-schema"></span> Schema

##### <span id="post-api-register-409"></span> 409 - Conflict: Email address is already registered
Status: Conflict

###### <span id="post-api-register-409-schema"></span> Schema

##### <span id="post-api-register-500"></span> 500 - Internal Server Error: Failed to hash password or save user to database
Status: Internal Server Error

###### <span id="post-api-register-500-schema"></span> Schema

### <span id="put-api-products"></span> Update a product (*PutAPIProducts*)

```
PUT /api/products
```

Updates details (name, description, price, count, tags) and/or the image for a specific product owned by the authenticated user. Fields not provided are left unchanged.

#### Consumes
  * multipart/form-data

#### Produces
  * text/plain

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `query` | uint64 (formatted integer) | `uint64` |  | ✓ |  | ID of the product to update |
| count | `formData` | int32 (formatted integer) | `int32` |  |  |  | New available stock count |
| description | `formData` | string | `string` |  |  |  | New description for the product |
| image | `formData` | file | `io.ReadCloser` |  |  |  | New product image file (replaces old image) |
| price | `formData` | float (formatted number) | `float32` |  |  |  | New price for the product (e.g., 29.99) |
| productName | `formData` | string | `string` |  |  |  | New name for the product |
| tags | `formData` | string | `string` |  |  |  | New comma-separated tags (replaces old tags) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#put-api-products-200) | OK | Product updated successfully |  | [schema](#put-api-products-200-schema) |
| [400](#put-api-products-400) | Bad Request | Bad Request: Invalid Product ID or data format |  | [schema](#put-api-products-400-schema) |
| [401](#put-api-products-401) | Unauthorized | Unauthorized: User not authenticated |  | [schema](#put-api-products-401-schema) |
| [403](#put-api-products-403) | Forbidden | Forbidden: User does not own this product |  | [schema](#put-api-products-403-schema) |
| [404](#put-api-products-404) | Not Found | Not Found: Product not found |  | [schema](#put-api-products-404-schema) |
| [409](#put-api-products-409) | Conflict | Conflict: Product name already exists for this user" // If name is updated |  | [schema](#put-api-products-409-schema) |
| [500](#put-api-products-500) | Internal Server Error | Internal Server Error: Database or file system error during update |  | [schema](#put-api-products-500-schema) |

#### Responses

##### <span id="put-api-products-200"></span> 200 - Product updated successfully
Status: OK

###### <span id="put-api-products-200-schema"></span> Schema

##### <span id="put-api-products-400"></span> 400 - Bad Request: Invalid Product ID or data format
Status: Bad Request

###### <span id="put-api-products-400-schema"></span> Schema

##### <span id="put-api-products-401"></span> 401 - Unauthorized: User not authenticated
Status: Unauthorized

###### <span id="put-api-products-401-schema"></span> Schema

##### <span id="put-api-products-403"></span> 403 - Forbidden: User does not own this product
Status: Forbidden

###### <span id="put-api-products-403-schema"></span> Schema

##### <span id="put-api-products-404"></span> 404 - Not Found: Product not found
Status: Not Found

###### <span id="put-api-products-404-schema"></span> Schema

##### <span id="put-api-products-409"></span> 409 - Conflict: Product name already exists for this user" // If name is updated
Status: Conflict

###### <span id="put-api-products-409-schema"></span> Schema

##### <span id="put-api-products-500"></span> 500 - Internal Server Error: Database or file system error during update
Status: Internal Server Error

###### <span id="put-api-products-500-schema"></span> Schema

### <span id="put-api-update-storefront"></span> Update a storefront link (*PutAPIUpdateStorefront*)

```
PUT /api/update_storefront
```

Updates the name, store ID, or store URL of an existing storefront link belonging to the authenticated user. Store type and credentials cannot be updated via this endpoint.

#### Consumes
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `query` | uint (formatted integer) | `uint64` |  | ✓ |  | ID of the Storefront Link to update |
| storefrontUpdate | `body` | [StorefronttableStorefrontLinkUpdatePayload](#storefronttable-storefront-link-update-payload) | `models.StorefronttableStorefrontLinkUpdatePayload` | | ✓ | | Fields to update (storeName, storeId, storeUrl) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#put-api-update-storefront-200) | OK | Successfully updated storefront link details |  | [schema](#put-api-update-storefront-200-schema) |
| [400](#put-api-update-storefront-400) | Bad Request | Bad Request - Invalid input, missing ID, or JSON parsing error |  | [schema](#put-api-update-storefront-400-schema) |
| [401](#put-api-update-storefront-401) | Unauthorized | Unauthorized - User session invalid or expired |  | [schema](#put-api-update-storefront-401-schema) |
| [403](#put-api-update-storefront-403) | Forbidden | Forbidden - User does not own this storefront link |  | [schema](#put-api-update-storefront-403-schema) |
| [404](#put-api-update-storefront-404) | Not Found | Not Found - Storefront link with the specified ID not found |  | [schema](#put-api-update-storefront-404-schema) |
| [409](#put-api-update-storefront-409) | Conflict | Conflict - Update would violate a unique constraint (e.g., duplicate name) |  | [schema](#put-api-update-storefront-409-schema) |
| [500](#put-api-update-storefront-500) | Internal Server Error | Internal Server Error - Database update failed |  | [schema](#put-api-update-storefront-500-schema) |

#### Responses

##### <span id="put-api-update-storefront-200"></span> 200 - Successfully updated storefront link details
Status: OK

###### <span id="put-api-update-storefront-200-schema"></span> Schema

[StorefronttableStorefrontLinkReturn](#storefronttable-storefront-link-return)

##### <span id="put-api-update-storefront-400"></span> 400 - Bad Request - Invalid input, missing ID, or JSON parsing error
Status: Bad Request

###### <span id="put-api-update-storefront-400-schema"></span> Schema

##### <span id="put-api-update-storefront-401"></span> 401 - Unauthorized - User session invalid or expired
Status: Unauthorized

###### <span id="put-api-update-storefront-401-schema"></span> Schema

##### <span id="put-api-update-storefront-403"></span> 403 - Forbidden - User does not own this storefront link
Status: Forbidden

###### <span id="put-api-update-storefront-403-schema"></span> Schema

##### <span id="put-api-update-storefront-404"></span> 404 - Not Found - Storefront link with the specified ID not found
Status: Not Found

###### <span id="put-api-update-storefront-404-schema"></span> Schema

##### <span id="put-api-update-storefront-409"></span> 409 - Conflict - Update would violate a unique constraint (e.g., duplicate name)
Status: Conflict

###### <span id="put-api-update-storefront-409-schema"></span> Schema

##### <span id="put-api-update-storefront-500"></span> 500 - Internal Server Error - Database update failed
Status: Internal Server Error

###### <span id="put-api-update-storefront-500-schema"></span> Schema

## Models

### <span id="orderstable-order-create-payload"></span> orderstable.OrderCreatePayload

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| customerEmail | string| `string` |  | | Email of the customer placing the order |  |
| customerName | string| `string` |  | | Name of the customer placing the order |  |
| orderedProducts | [][OrderstableOrderProductPayload](#orderstable-order-product-payload)| `[]*OrderstableOrderProductPayload` |  | | List of ordered products |  |

### <span id="orderstable-order-product-payload"></span> orderstable.OrderProductPayload

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| count | integer| `int64` |  | |  |  |
| productID | integer| `int64` |  | |  |  |

### <span id="orderstable-order-product-return"></span> orderstable.OrderProductReturn

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| count | integer| `int64` |  | |  |  |
| price | number| `float64` |  | | Price per item at the time of order |  |
| productID | integer| `int64` |  | |  |  |
| productName | string| `string` |  | |  |  |

### <span id="orderstable-order-return"></span> orderstable.OrderReturn

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| customerEmail | string| `string` |  | | Email of the customer that placed the order |  |
| customerName | string| `string` |  | | Name of the customer that placed the order |  |
| orderDate | string| `string` |  | | Formatted date string |  |
| orderID | integer| `int64` |  | | ID of the order requested |  |
| orderedProducts | [][OrderstableOrderProductReturn](#orderstable-order-product-return)| `[]*OrderstableOrderProductReturn` |  | | List of ordered products *owned by the requesting user* |  |
| status | string| `string` |  | |  |  |
| total | number| `float64` |  | | Total cost *for the items owned by the requesting user* in this order |  |
| trackingNumber | string| `string` |  | |  |  |

### <span id="prodtable-product-return"></span> prodtable.ProductReturn

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| image | string| `string` |  | | Consider renaming to imageURL or similar |  |
| prodCount | integer| `int64` |  | |  |  |
| prodDesc | string| `string` |  | |  |  |
| prodID | integer| `int64` |  | |  |  |
| prodName | string| `string` |  | |  |  |
| prodPrice | number| `float64` |  | |  |  |
| prodTags | string| `string` |  | |  |  |

### <span id="storefronttable-storefront-link-add-payload"></span> storefronttable.StorefrontLinkAddPayload

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| apiKey | string| `string` |  | | Example credential field |  |
| apiSecret | string| `string` |  | | Example credential field |  |
| storeId | string| `string` |  | | Platform-specific ID |  |
| storeName | string| `string` |  | | User-defined nickname |  |
| storeType | string| `string` |  | |  |  |
| storeUrl | string| `string` |  | | Storefront URL |  |

### <span id="storefronttable-storefront-link-return"></span> storefronttable.StorefrontLinkReturn

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | integer| `int64` |  | |  |  |
| storeId | string| `string` |  | | Match frontend JSON keys |  |
| storeName | string| `string` |  | |  |  |
| storeType | string| `string` |  | |  |  |
| storeUrl | string| `string` |  | |  |  |

### <span id="storefronttable-storefront-link-update-payload"></span> storefronttable.StorefrontLinkUpdatePayload

**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| storeId | string| `string` |  | | Platform-specific ID |  |
| storeName | string| `string` |  | | User-defined nickname |  |
| storeUrl | string| `string` |  | | Storefront URL |  |
