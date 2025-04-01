# Sprint 3
Video: [Sprint 3 VIDEO](https://youtu.be/DJhXZlULoso)

## Completed Work
### Front end
- Merged all Sprint 2 content to main branch
- Implemented image widget within Product form so that when an image is selected, there is a preview
- Updated README for dependencies and directions to start front end server
- Added support for a scrollable products page
  - Included fade off for clean view rather than products getting cut off
- Implemented API calls for fetching, updating, and deleting products
- Implemented API call for fetching image data
- Implemented front end logic for product viewing on home page and products page
- Implemented front end logic for updating an existing product
- Implemented front end logic for deleting products
- Added delete option to product form
- Made home page products more responsive
- *Incomplete Sprint2 Issue:* Fixed bug where product tags aren't being received correctly when submitting a new product
- *Incomplete Sprint2 Issue:* Connected Product page and preview on dashboard with back end to show users' products in real time
- Created Cypress Tests for this sprint's functionality

### Back end
- Merged all Sprint 2 content to main branch
- Created `storefronttable` package
  - `AddStorefront()`
  - `GetStorefronts()`
  - `UpdateStorefront()`
  - `DeleteStorefront()`
- Added unit tests for `storefronttable`
- Added separate `encryption` module for `storfronttable`
- Created the `generateCert.sh` script that generates the tls certificate required to access the site and generates the `.storefrontkey` for encrypted API info storage
- Updated the routing to add api calls to `storefronttable`
- Added very basic front end for Store Fronts
- Added pop-up form for adding a Store Front
- Made it so that clicking on a Store Front allows a user to edit it
- Made it so that only the logged in user's store fronts appear
- Updated `routes` to include calls to `storefronttable`
- Updated `main` to correct initalize the `storefronttable` db
- Updated `Product` struct so that `UserID` and `ProdName` have to be a unique combination
- Updated how products are updated in the `prodtable` to account for new restriction
- Implemented back end product filter so only products added by a specific user appear when that user is authorized (logged in)
- Connected the ability to edit values for each item to the front end
- Connected the ability to delete an item to the front end
- Swagger docs have been updated
- Added makefile for automatic deployment

## Incomplete Work
### Front end
- Need to complete Storefronts Figma pages so development can start
- Enable product and storefront editing on home page

### Back end
- Need to implement OAuth 2.0 for truely connecting to external stores
  - The site needs to be hosted somewhere
  - We need to create developer accounts for all supported storefronts
  - We need to register our site in our developer accounts
  - We need to set up the token management for the OAuth (many steps here)
- Need to implement order tracking
- Need to implement ag-grid updates for front end
- Need to refine `routes` unit tests to account for added routes
- improve makefile


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

Tests passing:

![image](https://github.com/user-attachments/assets/fdfc2e8a-f08a-415b-9ed6-96dd560ea00a)
![image](https://github.com/user-attachments/assets/f13d9d21-5ee9-43b4-9e1b-b4f01d011d35)
![image](https://github.com/user-attachments/assets/e8cba87d-5fda-4dcf-8f9f-7f20f6178d51)
![image](https://github.com/user-attachments/assets/0ed22aae-86a3-420f-b765-ae444db0142c)
![image](https://github.com/user-attachments/assets/b3b4d064-66a0-48df-9caf-0afef86c35d9)




_Unit Tests_

| Unit Test | Test Description |
| --- | --- |
| `LoginForm.Test` | Tests the rendering of login form, allowing user input, and submitting the form |
| `RegistrationForm.Test` | Tests the rendering of login form, allowing user input, validating the input, and submitting the form |
| `NavBar.Test` | Tests the routing from the nav bar to the home, products, storefronts, and orders pages |
| `ProductForm.Test` | Tests that the Product Form popup allows user input, validates the input, and submits the form |

Tests passing:

![image](https://github.com/user-attachments/assets/9a3be6b0-f1d4-4c70-915a-b72c9c46a4f1)


### Back end
Each internal package has an associated unit test that can be run by entering the following command from the `front-runner_backend` directory:

```bash
go test ./internal/login # replace login with the desired internal package
```

Alternatively, the tests can be automatically run with an extension in vscode.

![backend_tests_sprint3](https://github.com/user-attachments/assets/3db2cb0b-94cb-4f82-9e74-562a3d23dd85)

_Unit Tests List_

| Unit Test | Test Description |
| --- | --- |
| `TestGetDB` | Tests the GetDB function for initializing a database connection and verifies that the connection can be pinged successfully. |
| `TestLoginUser` | Checks that logging in with valid credentials works. |
| `TestLoginUSerInvalid` | Checks that an invalid login attempt returns an error. |
| `TestLogoutUser` | Verifies that logging out clears the session. |
| `TestAddProduct` | Tests the AddProduct endpoint by simulating a multipart/form-data POST request that includes product details and an image file. It uses a valid session cookie from the fake user. It verifies that the product and associated image are stored in the database. |
| `TestDeleteProduct` | Tests the DeleteProduct endpoint by inserting a dummy product (with an associated image file) for the fake user, then simulating a deletion request with a valid session cookie. It verifies that the product is removed from the database and the image file is deleted. |
| `TestUpdateProduct` | Tests the UpdateProduct endpoint by creating a dummy product for the fake user, then simulating an update request with new description, price, and stock count. It verifies that the product is updated in the database. |
| `TestGetProduct` | Verifies the `/api/get_product` endpoint successfully retrieves the correct product details when queried by its ID after creating a test product. |
| `TestGetProductImage` | Checks if the `/api/get_product_image` endpoint returns a success status when attempting to retrieve a product's image file using its filename. |
| `TestGetProducts` | Tests if the `/api/get_products` endpoint correctly retrieves a list containing the user's product(s) after creating a test product. |
| `TestRegisterRoutes_AllRoutes` | Verifies that the router correctly matches the expected routes and HTTP methods. |
| `TestRegisterRoutes_WithDummyStaticFile` | Verifies that the static file server returns the dummy index file. |
| `TestDirectUserEntry` | Tests direct insertion of user records into the database. |
| `TestRegisterUser` | Tests the RegisterUser HTTP handler for successful user registration. |
| `TestRegisterUserEmptyFields` | Verifies that the registration endpoint returns an error when required fields are missing. |
| `TestValidEmail` | Verifies that a properly formatted email address is considered valid. |
| `TestInvalidEmail`| Verifies that improperly formatted email addresses are considered invalid. |
| `TestAddStorefront_RealLogin` | Verifies the `/api/add_storefront` endpoint correctly creates links for authenticated users (via real login) and denies unauthorized requests. |
| `TestGetUpdateDeleteFlow_RealLogin` | Simulates an authenticated user successfully adding, viewing, updating, and deleting a storefront link through the relevant API endpoints. |
| `TestSpecificErrors_RealLogin` | Checks that the update and delete endpoints correctly return errors like 'Forbidden' (when accessing another user's link) and 'Not Found' (when targeting a non-existent link). |

## Front Runner API Documentation

API documentation for the Front Runner application.

## Version: 1.0

**Contact information:**  
API Support  
jonathan.bravo@ufl.edu  

**License:** [MIT](http://www.apache.org/licenses/LICENSE-2.0.html)

---
### /api/add_product

#### POST
##### Summary

Add a new product

##### Description

Creates a new product with details including name, description, price, count, tags, and an associated image.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| productName | formData | Product name | Yes | string |
| description | formData | Product description | Yes | string |
| price | formData | Product price | Yes | number |
| count | formData | Product stock count | Yes | integer |
| tags | formData | Product tags | No | string |
| image | formData | Product image file | Yes | file |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Product added successfully | string |
| 400 | Error parsing form or uploading image | string |
| 401 | User not authenticated | string |
| 500 | Internal server error | string |

### /api/delete_product

#### DELETE
##### Summary

Delete a product

##### Description

Deletes an existing product and its associated image if the product belongs to the authenticated user.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | query | Product ID | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Product deleted successfully | string |
| 401 | User not authenticated or unauthorized | string |
| 404 | Product not found | string |

### /api/get_product

#### GET
##### Summary

Retrieve a product

##### Description

Retreives an existing product and its associated metadata if the product belongs to the authenticated user.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | query | Product ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | JSON representation of a product's information | string |
| 401 | User not authenticated or unauthorized | string |
| 403 | Permission denied | string |
| 404 | No Product with specified ID | string |

### /api/get_product_image

#### GET
##### Summary

Retrieve a product image

##### Description

Retreives an existing product image if it exists and belongs to the authenticated user.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| image | query | Filepath of image | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Image's data | string |
| 401 | User not authenticated or unauthorized | string |
| 403 | Permission denied | string |
| 404 | Requested image does not exist | string |

### /api/get_products

#### GET
##### Summary

Retrieves all product information for authenticated user.

##### Description

Retreives existing products and their associated metadata for the authenticated user.

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | JSON representation of a user's product information | string |
| 401 | User not authenticated or unauthorized | string |

### /api/update_product

#### PUT
##### Summary

Update a product

##### Description

Updates the details of an existing product (description, price, stock count) that belongs to the authenticated user.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | query | Product ID | Yes | string |
| product_description | formData | New product description | No | string |
| item_price | formData | New product price | No | number |
| stock_amount | formData | New product stock count | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Product updated successfully | string |
| 401 | User not authenticated or unauthorized | string |
| 404 | Product not found | string |

---
### /api/add_storefront

#### POST
##### Summary

Link a new storefront

##### Description

Links a new external storefront (e.g., Amazon, Pinterest) to the user's account, storing credentials securely. Requires authentication.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| storefrontLink | body | Storefront Link Details (including credentials like apiKey, apiSecret) | Yes | [storefronttable.StorefrontLinkAddPayload](#storefronttablestorefrontlinkaddpayload) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Successfully linked storefront (credentials omitted) | [storefronttable.StorefrontLinkReturn](#storefronttablestorefrontlinkreturn) |
| 400 | Bad Request - Invalid input, missing fields, or JSON parsing error | string |
| 401 | Unauthorized - User session invalid or expired | string |
| 409 | Conflict - A link with this name/type already exists for the user | string |
| 500 | Internal Server Error - E.g., failed to encrypt, database error | string |

##### Security

| Security Schema | Scopes |
| --------------- | ------ |
| ApiKeyAuth |  |

### /api/delete_storefront

#### DELETE
##### Summary

Unlink a storefront

##### Description

Removes the link to an external storefront specified by its unique ID. User must own the link. Requires authentication.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | query | ID of the Storefront Link to delete | Yes | integer (uint) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Storefront unlinked successfully | string |
| 204 | Storefront unlinked successfully (No Content) | string |
| 400 | Bad Request - Invalid or missing 'id' query parameter | string |
| 401 | Unauthorized - User session invalid or expired | string |
| 403 | Forbidden - User does not own this storefront link | string |
| 404 | Not Found - Storefront link with the specified ID not found | string |
| 500 | Internal Server Error - Database deletion failed | string |

##### Security

| Security Schema | Scopes |
| --------------- | ------ |
| ApiKeyAuth |  |

### /api/get_storefronts

#### GET
##### Summary

Get linked storefronts

##### Description

Retrieves a list of all external storefronts linked by the currently authenticated user. Credentials are *never* included. Requires authentication.

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | List of linked storefronts (empty array if none) | [ [storefronttable.StorefrontLinkReturn](#storefronttablestorefrontlinkreturn) ] |
| 401 | Unauthorized - User session invalid or expired | string |
| 500 | Internal Server Error - Database query failed | string |

##### Security

| Security Schema | Scopes |
| --------------- | ------ |
| ApiKeyAuth |  |

### /api/update_storefront

#### PUT
##### Summary

Update a storefront link

##### Description

Updates the name, store ID, or store URL of an existing storefront link belonging to the authenticated user. Store type and credentials cannot be updated via this endpoint.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | query | ID of the Storefront Link to update | Yes | integer (uint) |
| storefrontUpdate | body | Fields to update (storeName, storeId, storeUrl) | Yes | [storefronttable.StorefrontLinkUpdatePayload](#storefronttablestorefrontlinkupdatepayload) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Successfully updated storefront link details | [storefronttable.StorefrontLinkReturn](#storefronttablestorefrontlinkreturn) |
| 400 | Bad Request - Invalid input, missing ID, or JSON parsing error | string |
| 401 | Unauthorized - User session invalid or expired | string |
| 403 | Forbidden - User does not own this storefront link | string |
| 404 | Not Found - Storefront link with the specified ID not found | string |
| 409 | Conflict - Update would violate a unique constraint (e.g., duplicate name) | string |
| 500 | Internal Server Error - Database update failed | string |

##### Security

| Security Schema | Scopes |
| --------------- | ------ |
| ApiKeyAuth |  |

---
### /api/get_product_image

#### GET
##### Summary

Retrieve a product image

##### Description

Retreives an existing product image if it exists and belongs to the authenticated user.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| image | query | Filepath of image | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Image's data | string |
| 401 | User not authenticated or unauthorized | string |
| 403 | Permission denied | string |
| 404 | Requested image does not exist | string |

---
### /api/login

#### POST
##### Summary

User login

##### Description

Authenticates a user and creates a session.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| email | formData | User email | Yes | string |
| password | formData | User password | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Logged in successfully. | string |
| 400 | Email and password are required | string |
| 401 | Invalid credentials | string |

### /api/logout

#### POST
##### Summary

User logout

##### Description

Logs out the current user by clearing the session.

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Logged out successfully | string |

### /api/register

#### POST
##### Summary

Register a new user

##### Description

Registers a new user using email, password, and an optional business name.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| email | formData | User email | Yes | string |
| password | formData | User password | Yes | string |
| business_name | formData | Business name | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | User registered successfully | string |
| 400 | Email and password are required or invalid email format | string |
| 409 | Email already in use or database error | string |

---
### Models

#### storefronttable.StorefrontLinkAddPayload

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| apiKey | string | Example credential field | No |
| apiSecret | string | Example credential field | No |
| storeId | string | Platform-specific ID | No |
| storeName | string | User-defined nickname | No |
| storeType | string |  | No |
| storeUrl | string | Storefront URL | No |

#### storefronttable.StorefrontLinkReturn

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | integer |  | No |
| storeId | string | Match frontend JSON keys | No |
| storeName | string |  | No |
| storeType | string |  | No |
| storeUrl | string |  | No |

#### storefronttable.StorefrontLinkUpdatePayload

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| storeId | string | Platform-specific ID | No |
| storeName | string | User-defined nickname | No |
| storeUrl | string | Storefront URL | No |
