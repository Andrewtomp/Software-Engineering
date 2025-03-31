# Sprint 3
Video: [Sprint 3 VIDEO]()

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
- Merged all Sprint 1 content to main branch
- Updated the routing so that logging in actually redirects to main page
  - Linked login page to correct route and created function to submit login api request
- Updated so that registering an account from the registration form actually registers a user AND logs in automatically
  - Linked registration page to correct route and created function that submits registration api call AND login api call
- Updated nav bar logout button so user is actually logged out and redirected to the login screen
  - Created function so that logout button to sends a logout api request and redirects
- Updated so that clicking on the add item button from the product page actually redirects to the product form
- Updated the product form to have the required fields
  - Created function to gather image and send add_product api request
- Updated swagger docs for API
- Updated unit tests for routes
- Added unit tests for `prodtable`
- Added unit tests for `imageStore`
- Added migration functions for `user`, `product` and, `image` database tables so they are automatically made/ migrated

## Incomplete Work
### Front end
- Need to complete Storefronts Figma pages so development can start

### Back end
- Need to implement back end product filter so only products added by a specific user appear when that user is authorized (logged in)
- Need to connect the ability to edit values for each item to the front end
- Need to connect the ability to delete an item to the front end

## Testing
### Front end
_Cypress Tests_

For this sprint, we moved to a full Cypress test suite rather than React Unit Tests.

| Unit Test | Test Description |
| --- | --- |
| `Sprint 2 Test` | Tests that the add new button opens the Add Product modal |
| `Render Login` | Tests the redering of login form |
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

These are the tests previously made that still work.

| Unit Test | Test Description |
| --- | --- |
| `LoginForm.Test` | Tests the redering of login form, allowing user input, and submitting the form |
| `RegistrationForm.Test` | Tests the redering of login form, allowing user input, validating the input, and submitting the form |
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

![image](backend_tests.png)

_Unit Tests List_

| Unit Test | Test Description |
| --- | --- |
| `TestGetDB` | Tests the GetDB function for initializing a database connection and verifies that the connection can be pinged successfully. |
| `TestLoadImage_Unauthorized` | Tests that LoadImage returns an unauthorized error when the user is not logged in. |
| `TestLoadImage_InvalidFilename` | Tests that LoadImage returns an error when the image record is not found. |
| `TestLoadImage_PermissionDenied`| Tests that LoadImage returns a forbidden error when the image record exists but belongs to a different user. |
| `TestLoadImage_FileNotExist` | tests that LoadImage returns a 404 error when the image file does not exist, even if the image record exists and the user is authorized. |
| `TestLoadImage_Success` | Tests that LoadImage successfully serves the image file when all conditions are met. |
| `TestUploadImage_Unauthorized` | Tests that UploadImage returns unauthorized when the user is not logged in. |
| `TestUploadImage_InvalidFileType` | Tests that UploadImage returns an error for non-image file uploads. |
| `TestUploadImage_Success` | Tests that UploadImage successfully uploads an image file. |
| `TestLoginUser` | Checks that logging in with valid credentials works. |
| `TestLoginUSerInvalid` | Checks that an invalid login attempt returns an error. |
| `TestLogoutUser` | Verifies that logging out clears the session. |
| `TestAddProduct` | Tests the AddProduct endpoint by simulating a multipart/form-data POST request that includes product details and an image file. It uses a valid session cookie from the fake user. It verifies that the product and associated image are stored in the database. |
| `TestDeleteProduct` | Tests the DeleteProduct endpoint by inserting a dummy product (with an associated image file) for the fake user, then simulating a deletion request with a valid session cookie. It verifies that the product is removed from the database and the image file is deleted. |
| `TestUpdateProduct` | Tests the UpdateProduct endpoint by creating a dummy product for the fake user, then simulating an update request with new description, price, and stock count. It verifies that the product is updated in the database. |
| `TestRegisterRoutes_AllRoutes` | Verifies that the router correctly matches the expected routes and HTTP methods. |
| `TestRegisterRoutes_WithDummyStaticFile` | Verifies that the static file server returns the dummy index file. |
| `TestDirectUserEntry` | Tests direct insertion of user records into the database. |
| `TestRegisterUser` | Tests the RegisterUser HTTP handler for successful user registration. |
| `TestRegisterUserEmptyFields` | Verifies that the registration endpoint returns an error when required fields are missing. |
| `TestValidEmail` | Verifies that a properly formatted email address is considered valid. |
| `TestInvalidEmail`| Verifies that improperly formatted email addresses are considered invalid. |

## API Documentation
