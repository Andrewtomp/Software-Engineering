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
      // Mock API responses
      cy.intercept('POST', '/api/add_product', { statusCode: 200 }).as('addProduct');
      cy.intercept('PUT', '/api/update_product*', { statusCode: 200 }).as('updateProduct');
      cy.intercept('DELETE', '/api/delete_product*', { statusCode: 200 }).as('deleteProduct');
      
      // Mock image endpoint
      cy.intercept('GET', '/api/get_product_image*', {
        fixture: 'product-image.jpg'
      }).as('getProductImage');
  
      // Visit the page with the product form
      cy.visit('/');
    });
  
    it('should open the Add Product form', () => {
      // Assuming there's a button to open the add product form
      cy.get('[data-cy=add-product-button]').click();
      cy.get('.add-product-container').should('be.visible');
      cy.get('.product-form-header h2').should('have.text', 'Add Product');
    });
  
    it('should validate required fields in the Add Product form', () => {
      // Open the add product form
      cy.get('[data-cy=add-product-button]').click();
      
      // Try to submit without filling required fields
      cy.get('button[type="submit"]').click();
      
      // Check for validation messages
      cy.get('.errors').should('be.visible');
    });
  
    it('should successfully add a new product', () => {
      // Open the add product form
      cy.get('[data-cy=add-product-button]').click();
      
      // Fill the form
      cy.get('#root_productName').type(productFormData.productName);
      cy.get('#root_description').type(productFormData.description);
      cy.get('#root_price').type(productFormData.price);
      cy.get('#root_count').clear().type(productFormData.count);
      cy.get('#root_tags').type(productFormData.tags);
      
      // Upload an image
      cy.fixture('test-image.jpg', 'base64').then(fileContent => {
        const blob = Cypress.Blob.base64StringToBlob(fileContent, 'image/jpeg');
        const testFile = new File([blob], 'test-image.jpg', { type: 'image/jpeg' });
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
        id: '123',
        name: 'Existing Product',
        description: 'This is an existing product',
        price: '29.99',
        count: 5,
        tags: '#existing',
        image: 'product123.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/products', {
        body: [productToEdit]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('[data-cy=edit-product-button]').first().click();
      
      // Check if form is populated correctly
      cy.get('#root_productName').should('have.value', productToEdit.name);
      cy.get('#root_description').should('have.value', productToEdit.description);
      cy.get('#root_price').should('have.value', `$${productToEdit.price}`);
      cy.get('#root_count').should('have.value', productToEdit.count.toString());
      cy.get('#root_tags').should('have.value', productToEdit.tags);
      cy.get('.image-preview img').should('be.visible');
    });
  
    it('should update an existing product', () => {
      // Mock a product to edit
      const productToEdit = {
        id: '123',
        name: 'Existing Product',
        description: 'This is an existing product',
        price: '29.99',
        count: 5,
        tags: '#existing',
        image: 'product123.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/products', {
        body: [productToEdit]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('[data-cy=edit-product-button]').first().click();
      
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
        id: '123',
        name: 'Product to Delete',
        description: 'This product will be deleted',
        price: '9.99',
        count: 3,
        tags: '#delete',
        image: 'product-delete.jpg'
      };
      
      // Intercept the API call that would load products and inject our mock
      cy.intercept('GET', '/api/products', {
        body: [productToDelete]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('[data-cy=edit-product-button]').first().click();
      
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
      cy.intercept('GET', '/api/products', {
        body: [product]
      }).as('getProducts');
      
      // Refresh to load products
      cy.reload();
      cy.wait('@getProducts');
      
      // Click on edit button for the product
      cy.get('[data-cy=edit-product-button]').first().click();
      
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
      cy.get('[data-cy=add-product-button]').click();
      
      // Verify form is open
      cy.get('.add-product-container').should('be.visible');
      
      // Click the close button
      cy.get('.fa-times').click();
      
      // Verify form is closed
      cy.get('.add-product-container').should('not.exist');
    });
  });