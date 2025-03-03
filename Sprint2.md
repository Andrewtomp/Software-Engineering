# Sprint 2
Video: [insert link here]

## Completed Work
### Front end
- Merged all Sprint 1 content to main branch
- Created skeleton for orders page
- Filled order info with large ag-grid table
	- Table contains detailed order information as well as placeholder buttons to download shipping labels and view more order info
- Significantly improved responsiveness of Home, Product, and Nav Bar components
- Implemented json based login forms using RJSF (Reach JSON Schema Form)
- Created README for dependencies and directions to start front end server
- Used RJSF to create a form to input information to create a new product
- Routing for Login and ProductForm components
- Completed styling for ProductForm component
- Moved ProductForm component to a modal rather than its own page
- Used RJSF to implement Registration (Create new account) functionality
- Completed styling for Registration component using existing Login styling
- Designed and created visually appealing Login/Register backgrounds for when users first land on the site
- Created Cypress Test for opening the ProductForm modal
- Created React unit tests
   
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

## Testing
### Front end
_Cypress Test_

Completed a simple Cypress Test testing the opening of a modal for product creation. This test starts by opening the products page, then checks to see if the modal exists in the html yet, clicks the button to open the modal, and verifies that it is now there.

![image](https://github.com/user-attachments/assets/7d3a070d-7cc6-4daa-8223-69a139b3b4d6)


_Unit Tests_


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

Will check swagger docs to ensure correctness, but then just update here (MAKE SURE TO OPEN NEW BRANCH)