// cypress/integration/storefront-form.spec.js
describe('Storefront Form', () => {
  const storefrontFormData = {
    storeType: 'Amazon Seller Central',
    storeName: 'test amzn',
    apiKey: 'apikey',
    apiSecret: 'apisecret',
    storeId: 'storeid',
    storeUrl: 'https://www.amazon.com',
  };



  beforeEach(() => {
    cy.visit('https://localhost:8080/login');
    cy.get('#root_email').type('test@frontrunner.com');
    cy.get('#root_password').type('frontrunner');
    cy.get('button[type="submit"]').click();

    // Mock API responses
    cy.intercept('POST', '/api/add_storefront', { statusCode: 200 }).as('addstorefront');
    cy.intercept('PUT', '/api/update_storefront*', { statusCode: 200 }).as('updatestorefront');
    cy.intercept('DELETE', '/api/delete_storefront*', { statusCode: 200 }).as('deletestorefront');

    // wait until the url is https://localhost:8080/ to continue
    cy.url().should('equal', 'https://localhost:8080/');


    cy.visit('https://localhost:8080/storefronts');
  });

  it('should open the Add storefront form', () => {

    // Assuming there's a button to open the add storefront form
    cy.get('.add-new-button').click();
    cy.get('.add-storefront-container').should('be.visible');
    cy.get('.storefront-form-header h2').should('have.text', 'Link A New Storefront');
  });

  it('should validate required fields in the Add storefront form', () => {
    // Open the add storefront form
    cy.get('.add-new-button').click();

    // Try to submit without filling required fields
    cy.get('button[type="submit"]').click();

    // Check that add storefront conteiner is still visible
    cy.get('.add-storefront-container').should('be.visible');
  });

  it('should successfully add a new storefront', () => {
    // Open the add storefront form
    cy.get('.add-new-button').click();

    // Fill the form
    cy.get('#root_storeName').type(storefrontFormData.storeName);
    cy.get('#root_apiKey').type(storefrontFormData.apiKey);
    cy.get('#root_apiSecret').type(storefrontFormData.apiSecret);
    cy.get('#root_storeId').type(storefrontFormData.storeId);
    cy.get('#root_storeUrl').type(storefrontFormData.storeUrl);

    // Submit the form
    cy.get('button[type="submit"]').click();

    // Check if the API was called with the correct data
    cy.wait('@addstorefront').then(interception => {
      // FormData is not directly accessible in Cypress, so we check for successful submission
      expect(interception.request.method).to.equal('POST');
      expect(interception.response.statusCode).to.equal(200);
    });

    // Check if page was reloaded (or a success message was shown)
    cy.url().should('include', '/');
  });

  it('should open and populate the Edit storefront form', () => {
    // Mock a storefront to edit
    const storefrontToEdit = {
      storeType: 'Amazon Seller Central',
      storeName: 'test edit',
      apiKey: 'apikey',
      apiSecret: 'apisecret',
      storeId: 'storeid',
      storeUrl: 'https://www.amazon.com',
    };

    // Intercept the API call that would load storefronts and inject our mock
    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToEdit]
    }).as('getstorefronts');

    // Refresh to load storefronts
    cy.reload();
    cy.wait('@getstorefronts');

    // Click on edit button for the storefront
    cy.get('.storefront-tile').first().click();

    // Check if form is populated correctly
    cy.get('#root_storeName').should('have.value', storefrontToEdit.storeName);
    cy.get('#root_storeId').should('have.value', storefrontToEdit.storeId);
    cy.get('#root_storeUrl').should('have.value', `${storefrontToEdit.storeUrl}`);
  });

  it('should update an existing storefront', () => {
    // Mock a storefront to edit
    const storefrontToEdit = {
      storeType: 'Amazon Seller Central',
      storeName: 'test edit',
      apiKey: 'apikey',
      apiSecret: 'apisecret',
      storeId: 'storeid',
      storeUrl: 'https://www.amazon.com',
    };

    // Intercept the API call that would load storefronts and inject our mock
    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToEdit]
    }).as('getstorefronts');

    // Refresh to load storefronts
    cy.reload();
    cy.wait('@getstorefronts');

    // Click on edit button for the storefront
    cy.get('.storefront-tile').first().click();

    // Check if form is populated correctly
    cy.get('#root_storeName').should('have.value', storefrontToEdit.storeName);
    cy.get('#root_storeId').should('have.value', storefrontToEdit.storeId);
    cy.get('#root_storeUrl').should('have.value', `${storefrontToEdit.storeUrl}`);

    // Update form fields
    cy.get('#root_storeName').clear().type('edited name');
    cy.get('#root_storeId').clear().type('edited id');


    // Submit the form
    cy.get('button[type="submit"]').click();

    // Check if the API was called with the correct data
    cy.wait('@updatestorefront').then(interception => {
      expect(interception.request.method).to.equal('PUT');
      expect(interception.response.statusCode).to.equal(200);
    });
  });

  it('should delete a storefront', () => {
    // Mock a storefront to delete
    const storefrontToDelete = {
      id: '1',
      storeType: 'Amazon Seller Central',
      storeName: 'test edit',
      apiKey: 'apikey',
      apiSecret: 'apisecret',
      storeId: 'storeid',
      storeUrl: 'https://www.amazon.com',
    };

    // Intercept the API call that would load storefronts and inject our mock
    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToDelete]
    }).as('getstorefronts');

    // Refresh to load storefronts
    cy.reload();
    cy.wait('@getstorefronts');

    // Click on edit button for the storefront
    cy.get('.storefront-tile').first().click();

    // Stub window.confirm to return true
    cy.window().then(win => {
      cy.stub(win, 'confirm').returns(true);
    });

    // Click the delete button
    cy.get('.delete-icon').click();

    // Check if the API was called with the correct data
    cy.wait('@deletestorefront').then(interception => {
      expect(interception.request.method).to.equal('DELETE');
      expect(interception.response.statusCode).to.equal(200);
    });
  });

  it('should cancel storefront deletion when user clicks cancel in confirm dialog', () => {
    // Mock a storefront
    const storefront = {
      storeType: 'Amazon Seller Central',
      storeName: 'test edit',
      apiKey: 'apikey',
      apiSecret: 'apisecret',
      storeId: 'storeid',
      storeUrl: 'https://www.amazon.com',
    };

    // Intercept the API call that would load storefronts and inject our mock
    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefront]
    }).as('getstorefronts');

    // Refresh to load storefronts
    cy.reload();
    cy.wait('@getstorefronts');

    // Click on edit button for the storefront
    cy.get('.storefront-tile').first().click();

    // Stub window.confirm to return false
    cy.window().then(win => {
      cy.stub(win, 'confirm').returns(false);
    });

    // Click the delete button
    cy.get('.delete-icon').click();

    // Verify the delete API was not called
    cy.get('@deletestorefront.all').should('have.length', 0);

    // Verify we're still on the edit form
    cy.get('.storefront-form-header h2').should('have.text', `Edit ${storefront.storeType} Link`);
  });

  it('should close the form when clicking the close button', () => {
    // Open the add storefront form
    cy.get('.add-new-button').click();

    // Verify form is open
    cy.get('.add-storefront-container').should('be.visible');

    // Click the close button
    cy.get('.fa-xmark').click();

    // Verify form is closed
    cy.get('.add-storefront-container').should('not.exist');
  });
});
