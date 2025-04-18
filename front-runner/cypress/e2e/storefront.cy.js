describe('Storefront Form', () => {
  const storefrontFormData = {
    storeType: 'amazon',
    storeName: 'Test Amazon Store',
    storeId: 'AMZ123456',
    storeUrl: 'https://www.amazon.com/teststore',
    apiKey: 'test-api-key',
    apiSecret: 'test-api-secret'
  };

  beforeEach(() => {
    cy.visit('https://localhost:8080/login');
    cy.get('#root_email').type('test@frontrunner.com');
    cy.get('#root_password').type('frontrunner');
    cy.get('button[type="submit"]').click();

    // API mocks
    cy.intercept('POST', '/api/add_storefront', { statusCode: 200 }).as('addStorefront');
    cy.intercept('PUT', '/api/update_storefront*', { statusCode: 200 }).as('updateStorefront');
    cy.intercept('DELETE', '/api/delete_storefront*', { statusCode: 200 }).as('deleteStorefront');

    cy.url().should('equal', 'https://localhost:8080/');
    cy.visit('https://localhost:8080/storefronts');
  });

  it('should open the Link Storefront form', () => {
    cy.get('.add-new-button').click();
    cy.get('.add-storefront-container').should('be.visible');
    cy.get('.storefront-form-header h2').should('contain', 'Link A New Storefront');
  });

  it('should validate required fields in the storefront form', () => {
    cy.get('.add-new-button').click();
    cy.get('button[type="submit"]').click();

    // Form should still be visible due to missing required fields
    cy.get('.add-storefront-container').should('be.visible');
  });

  it('should successfully link a new storefront', () => {
    cy.get('.add-new-button').click();

    cy.get('#root_storeType').select(storefrontFormData.storeType);
    cy.get('#root_storeName').type(storefrontFormData.storeName);
    cy.get('#root_storeId').type(storefrontFormData.storeId);
    cy.get('#root_storeUrl').type(storefrontFormData.storeUrl);
    cy.get('#root_apiKey').type(storefrontFormData.apiKey);
    cy.get('#root_apiSecret').type(storefrontFormData.apiSecret);

    cy.get('button[type="submit"]').click();

    cy.wait('@addStorefront').then((interception) => {
      expect(interception.request.method).to.equal('POST');
      expect(interception.response.statusCode).to.equal(200);
    });
  });

  it('should open and populate the Edit Storefront form', () => {
    const storefrontToEdit = {
      id: '123',
      storeType: 'etsy',
      storeName: 'My Etsy Store',
      storeId: 'ETSY123',
      storeUrl: 'https://www.etsy.com/shop/myshop'
    };

    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToEdit]
    }).as('getStorefronts');

    cy.reload();
    cy.wait('@getStorefronts');

    cy.get('.home-storefront').first().click();

    cy.get('#root_storeName').should('have.value', storefrontToEdit.storeName);
    cy.get('#root_storeId').should('have.value', storefrontToEdit.storeId);
    cy.get('#root_storeUrl').should('have.value', storefrontToEdit.storeUrl);
    cy.get('#root_storeType').should('have.value', storefrontToEdit.storeType);
  });

  it('should update an existing storefront', () => {
    const storefrontToEdit = {
      id: '123',
      storeType: 'etsy',
      storeName: 'Old Etsy Store',
      storeId: 'ETSY123',
      storeUrl: 'https://www.etsy.com/shop/oldstore'
    };

    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToEdit]
    }).as('getStorefronts');

    cy.reload();
    cy.wait('@getStorefronts');

    cy.get('.home-storefront').first().click();

    cy.get('#root_storeName').clear().type('Updated Etsy Store');
    cy.get('#root_storeUrl').clear().type('https://www.etsy.com/shop/updated');

    cy.get('button[type="submit"]').click();

    cy.wait('@updateStorefront').then((interception) => {
      expect(interception.request.method).to.equal('PUT');
      expect(interception.response.statusCode).to.equal(200);
    });
  });

  it('should delete an existing storefront', () => {
    const storefrontToDelete = {
      id: '123',
      storeType: 'pinterest',
      storeName: 'Pinterest Biz',
      storeId: 'P123',
      storeUrl: 'https://www.pinterest.com/pbiz'
    };

    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToDelete]
    }).as('getStorefronts');

    cy.reload();
    cy.wait('@getStorefronts');

    cy.get('.home-storefront').first().click();

    // Stub confirm dialog
    cy.window().then(win => cy.stub(win, 'confirm').returns(true));

    cy.get('[data-testid="delete-icon"]').click();

    cy.wait('@deleteStorefront').then((interception) => {
      expect(interception.request.method).to.equal('DELETE');
      expect(interception.response.statusCode).to.equal(200);
    });
  });

  it('should cancel storefront deletion when user rejects confirm', () => {
    const storefrontToKeep = {
      id: '999',
      storeType: 'amazon',
      storeName: 'Keep This Store',
      storeId: 'KEEP123',
      storeUrl: 'https://www.amazon.com/mykeepstore'
    };

    cy.intercept('GET', '/api/get_storefronts', {
      body: [storefrontToKeep]
    }).as('getStorefronts');

    cy.reload();
    cy.wait('@getStorefronts');

    cy.get('.home-storefront').first().click();

    cy.window().then(win => cy.stub(win, 'confirm').returns(false));
    cy.get('[data-testid="delete-icon"]').click();

    cy.get('@deleteStorefront.all').should('have.length', 0);
    cy.get('.storefront-form-header h2').should('contain', 'Edit');
  });

  it('should close the form when clicking the close button', () => {
    cy.get('.add-new-button').click();
    cy.get('.add-storefront-container').should('be.visible');
    cy.get('.fa-times').click();
    cy.get('.add-storefront-container').should('not.exist');
  });
});
