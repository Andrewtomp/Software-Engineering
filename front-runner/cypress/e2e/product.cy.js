// cypress/integration/product-form.spec.js
describe('Product Form', () => {
    const productFormData = {
      productName: 'Test Product',
      description: 'This is a test product description',
      price: '$19.99',
      count: 10,
      tags: '#test, #cypress'
    };

    
  
    beforeEach(() => {
      cy.visit('https://localhost:8080/login');
      cy.get('#root_email').type('test@frontrunner.com');
      cy.get('#root_password').type('frontrunner');
      cy.get('button[type="submit"]').click();

      // Mock API responses
      cy.intercept('POST', '/api/add_product', { statusCode: 200 }).as('addProduct');
      cy.intercept('PUT', '/api/update_product*', { statusCode: 200 }).as('updateProduct');
      cy.intercept('DELETE', '/api/delete_product*', { statusCode: 200 }).as('deleteProduct');
      
      // Mock image endpoint
      cy.intercept('GET', '/api/get_product_image*', {
        fixture: 'product-image.png'
      }).as('getProductImage');

      // wait until the url is https://localhost:8080/ to continue
      cy.url().should('equal', 'https://localhost:8080/');
      

      cy.visit('https://localhost:8080/products');
    });
  
    it('should open the Add Product form', () => {

      // Assuming there's a button to open the add product form
      cy.get('.add-new-button').click();
      cy.get('.add-product-container').should('be.visible');
      cy.get('.product-form-header h2').should('have.text', 'Add Product');
    });
  
    it('should validate required fields in the Add Product form', () => {
      // Open the add product form
      cy.get('.add-new-button').click();
      
      // Try to submit without filling required fields
      cy.get('button[type="submit"]').click();
      
      // Check that add product conteiner is still visible
      cy.get('.add-product-container').should('be.visible');
    });
  
    it('should successfully add a new product', () => {
      // Open the add product form
      cy.get('.add-new-button').click();
      
      // Fill the form
      cy.get('#root_productName').type(productFormData.productName);
      cy.get('#root_description').type(productFormData.description);
      cy.get('#root_price').type(productFormData.price);
      cy.get('#root_count').clear().type(productFormData.count);
      cy.get('#root_tags').type(productFormData.tags);
      
      // Upload an image
      cy.fixture('product-image.png', 'base64').then(fileContent => {
        const blob = Cypress.Blob.base64StringToBlob(fileContent, 'image/png');
        const testFile = new File([blob], 'product-image.png', { type: 'image/png' });
        const dataTransfer = new DataTransfer();
        dataTransfer.items.add(testFile);
        
        cy.get('input[type="file"]').then(input => {
          const inputElement = input[0];
          inputElement.files = dataTransfer.files;
          cy.wrap(input).trigger('change', { force: true });
        });
      });
      
      // Check if image preview is shown
      cy.get('.image-preview img').should('be.visible');
      
      // Submit the form
      cy.get('button[type="submit"]').click();
      
      // Check if the API was called with the correct data
      cy.wait('@addProduct').then(interception => {
        // FormData is not directly accessible in Cypress, so we check for successful submission
        expect(interception.request.method).to.equal('POST');
        expect(interception.response.statusCode).to.equal(200);
      });
      
      // Check if page was reloaded (or a success message was shown)
      cy.url().should('include', '/');
    });
  
    it('should open and populate the Edit Product form', () => {
      // Mock a product to edit
      const productToEdit = {
        prodID: '123',
        prodName: 'Existing Product',
        prodDesc: 'This is an existing product',
        prodPrice: '29.99',
        prodCount: 5,
        prodTags: '#existing',
        image: 'product123.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/get_products', {
        body: [productToEdit]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('.product-tile').first().click();
      
      // Check if form is populated correctly
      cy.get('#root_productName').should('have.value', productToEdit.prodName);
      cy.get('#root_description').should('have.value', productToEdit.prodDesc);
      cy.get('#root_price').should('have.value', `$${productToEdit.prodPrice}`);
      cy.get('#root_count').should('have.value', productToEdit.prodCount.toString());
      cy.get('#root_tags').should('have.value', productToEdit.prodTags);
      cy.get('.image-preview img').should('be.visible');
    });
  
    it('should update an existing product', () => {
      // Mock a product to edit
      const productToEdit = {
        prodID: '123',
        prodName: 'Existing Product',
        prodDesc: 'This is an existing product',
        prodPrice: '29.99',
        prodCount: 5,
        prodTags: '#existing',
        image: 'product123.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/get_products', {
        body: [productToEdit]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('.product-tile').first().click();
      
      // Update form fields
      cy.get('#root_productName').clear().type('Updated Product');
      cy.get('#root_description').clear().type('This is an updated description');
      cy.get('#root_price').clear().type('$39.99');
      
      // Submit the form
      cy.get('button[type="submit"]').click();
      
      // Check if the API was called with the correct data
      cy.wait('@updateProduct').then(interception => {
        expect(interception.request.method).to.equal('PUT');
        expect(interception.response.statusCode).to.equal(200);
      });
    });
  
    it('should delete a product', () => {
      // Mock a product to delete
      const productToDelete = {
        prodID: '123',
        prodName: 'Existing Product',
        prodDesc: 'This is an existing product',
        prodPrice: '29.99',
        prodCount: 5,
        prodTags: '#existing',
        image: 'product123.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/get_products', {
        body: [productToDelete]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('.product-tile').first().click();
      
      // Stub window.confirm to return true
      cy.window().then(win => {
        cy.stub(win, 'confirm').returns(true);
      });
      
      // Click the delete button
      cy.get('.delete-icon').click();
      
      // Check if the API was called with the correct data
      cy.wait('@deleteProduct').then(interception => {
        expect(interception.request.method).to.equal('DELETE');
        expect(interception.response.statusCode).to.equal(200);
      });
    });
  
    it('should cancel product deletion when user clicks cancel in confirm dialog', () => {
      // Mock a product
      const product = {
        id: '123',
        name: 'Product to Keep',
        description: 'This product will not be deleted',
        price: '9.99',
        count: 3,
        tags: '#keep',
        image: 'product-keep.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/get_products', {
        body: [product]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('.product-tile').first().click();
      
      // Stub window.confirm to return false
      cy.window().then(win => {
        cy.stub(win, 'confirm').returns(false);
      });
      
      // Click the delete button
      cy.get('.delete-icon').click();
      
      // Verify the delete API was not called
      cy.get('@deleteProduct.all').should('have.length', 0);
      
      // Verify we're still on the edit form
      cy.get('.product-form-header h2').should('have.text', 'Edit Product');
    });
  
    it('should close the form when clicking the close button', () => {
      // Open the add product form
      cy.get('.add-new-button').click();
      
      // Verify form is open
      cy.get('.add-product-container').should('be.visible');
      
      // Click the close button
      cy.get('.fa-xmark').click();
      
      // Verify form is closed
      cy.get('.add-product-container').should('not.exist');
    });
  });
