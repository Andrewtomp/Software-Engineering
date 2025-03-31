describe('Orders Page', () => {
    beforeEach(() => {
      // Visit the orders page before each test
      cy.visit('https://localhost:8080/login');
      cy.get('#root_email').type('test@frontrunner.com');
      cy.get('#root_password').type('frontrunner');
      cy.get('button[type="submit"]').click();
      // wait until the url is https://localhost:8080/ to continue
      cy.url().should('equal', 'https://localhost:8080/');
      
      cy.visit('https://localhost:8080/orders'); // Replace with your actual URL
    });
    it('should trigger a download and verify a PNG file is downloaded', () => {
      // Step 1: Visit the orders page
      cy.visit('https://localhost:8080/orders'); // Replace with your actual URL
  
      // Step 2: Intercept the request for downloading the shipping label (if needed)
      cy.intercept('GET', '/path/to/shipping_label.png',{
        statusCode: 200}).as('downloadRequest');
  
      // Step 3: Click the "Download" button (first button in the "Shipping Label" column)
      cy.get('button.table-button', { timeout: 10000 }).first().click();
        

    });
  });
  