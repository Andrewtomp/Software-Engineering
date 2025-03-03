describe('Sprint 2 Cypress Test', () => {
  it('opens the "Add Product" modal', () => {
    cy.visit('http://localhost:3000/products')
    
    // Ensure the modal is not visible initially
    cy.get('.add-product-card').should('not.exist');
  
    // Click the Add New button
    cy.get('.add-new-button').click();
    
    // Check if the modal appears
    cy.get('.add-product-card').should('be.visible');
    
  })
})