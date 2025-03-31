describe('Sprint 2 Cypress Test', () => {
  it('opens the "Add Product" modal', () => {
    // Visit the orders page before each test
    cy.visit('https://localhost:8080/login');
    cy.get('#root_email').type('test@frontrunner.com');
    cy.get('#root_password').type('frontrunner');
    cy.get('button[type="submit"]').click();
    // wait until the url is https://localhost:8080/ to continue
    cy.url().should('equal', 'https://localhost:8080/');
    
    cy.visit('https://localhost:8080/products')
    
    // Ensure the modal is not visible initially
    cy.get('.add-product-card').should('not.exist');
  
    // Click the Add New button
    cy.get('.add-new-button').click();
    
    // Check if the modal appears
    cy.get('.add-product-card').should('be.visible');
    
  })
})