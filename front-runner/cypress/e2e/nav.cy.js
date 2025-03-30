// cypress/integration/navbar.spec.js
describe('Navbar Navigation', () => {
    beforeEach(() => {
      // Visit the application and login if necessary
      cy.visit('http://localhost:3000/');
      
      // Mock API responses for smoother testing
      cy.intercept('POST', '/api/logout', {
        statusCode: 200,
        body: { success: true }
      }).as('logoutRequest');
      
      // If there's a login required, handle it here
      // cy.login(); // Uncomment and implement if needed
    });

    it('should navigate to Products page', () => {
      // Click on Products nav option
      cy.contains('.nav-option', 'My Products').click();
      
      // Verify URL
      cy.url().should('include', '/products');
      
      // Additional assertions to confirm we're on the products page
      cy.get('h1, h2').contains(/products|inventory/i).should('exist');
    });
  
    it('should navigate to Storefronts page', () => {
      // Click on Storefronts nav option
      cy.contains('.nav-option', 'My Storefronts').click();
      
      // Verify URL
      cy.url().should('include', '/storefronts');
      
      // Additional assertions
      cy.get('h1, h2').contains(/storefronts|stores/i).should('exist');
    });
  
    it('should navigate to Orders page', () => {
      // Click on Orders nav option
      cy.contains('.nav-option', 'My Orders').click();
      
      // Verify URL
      cy.url().should('include', '/orders');
      
      // Additional assertions
      cy.get('h1, h2').contains(/orders/i).should('exist');
    });
  
    it('should navigate to Settings page', () => {
      // Click on Settings nav option
      cy.contains('.nav-option', 'Settings').click();
      
      // Verify URL
      cy.url().should('include', '/settings');
      
      // Additional assertions
      cy.get('h1, h2').contains(/settings/i).should('exist');
    });
  
    it('should logout successfully and redirect to login page', () => {
      // Click on Logout nav option
      cy.contains('.nav-option', 'Logout').click();
      
      // Wait for the logout API call
      cy.wait('@logoutRequest');
      
      // Verify URL is now login page
      cy.url().should('include', '/login');
      
      // Verify login form is visible
      cy.get('form').should('exist');
      cy.contains(/login|sign in/i).should('exist');
    });
  
  
    it('should verify all navbar elements are visible', () => {
      // Check that all nav options are visible
      cy.get('.nav-option').should('have.length', 6); // Total number of options
      
      // Check that all icons are loaded
      cy.get('.nav-icon, .bottom-nav-icon, .logo')
        .should('be.visible')
        .and($imgs => {
          // Check that images are actually loaded
          $imgs.each((i, img) => {
            expect(img.naturalWidth).to.be.greaterThan(0);
          });
        });
      
    });
  
    it('should verify navbar is responsive', () => {
      // Test on mobile viewport
      cy.viewport('iphone-8');
      
      // Check that navbar adapts to smaller screen
      cy.get('.nav-bar').should('be.visible');
      
      // Test on tablet viewport
      cy.viewport('ipad-2');
      cy.get('.nav-bar').should('be.visible');
      
      // Return to desktop
      cy.viewport(1280, 720);
      cy.get('.nav-bar').should('be.visible');
    });
  });
  
  // Custom command for login if needed
  Cypress.Commands.add('login', () => {
    cy.visit('/login');
    cy.get('input[name="email"]').type('test@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.get('button[type="submit"]').click();
    cy.url().should('eq', Cypress.config().baseUrl + '/');
  });