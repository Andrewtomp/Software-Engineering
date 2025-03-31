describe('StorefrontLinkForm', () => {
    beforeEach(() => {
      // Visit the orders page before each test
      cy.visit('https://localhost:8080/login');
      cy.get('#root_email').type('test@frontrunner.com');
      cy.get('#root_password').type('frontrunner');
      cy.get('button[type="submit"]').click();
      // wait until the url is https://localhost:8080/ to continue
      cy.url().should('equal', 'https://localhost:8080/');
      
      
      // Mock the storefront data for testing (if required) or visit the page where the component is rendered.
      cy.visit('https://localhost:8080/storefronts'); // Replace with the appropriate route to load the StorefrontLinkForm
    });
  
    it('should render the form and show the correct title', () => {
      // Check if the form exists
      cy.get('form').should('exist');
  
      // Check if the form title is visible
      cy.get('h2').should('be.visible').and('contain.text', 'Link New Storefront');
  
      // Optionally, you can also check if the close button is visible
      cy.get('.modal-close-button').should('be.visible');
    });
  
    it('should display the error message when there is an error', () => {
      // Simulate an error state (you may need to trigger this state in your app first)
      cy.get('form').within(() => {
        // Manually trigger an error state or make the API return an error for testing
        cy.get('.form-error').should('not.exist'); // Ensure error message does not exist initially
  
        // Example error
        cy.get('.form-error').should('be.visible').and('contain.text', 'Please select a storefront type.');
      });
    });
  });
  