describe('Login Form', () => {
  beforeEach(() => {
    cy.visit('https://localhost:8080/login');
    
    // Mock the login API endpoint
    cy.intercept('POST', '/api/login', {
      statusCode: 200,
      body: { success: true }
    }).as('loginRequest');
  });

  it('should render the login form correctly', () => {
    // Check if the login title is present
    cy.contains('Login').should('be.visible');
    
    // Wait for form to be rendered and check if form elements exist
    cy.get('form').should('exist');
  });

  it('should allow user to input credentials', () => {
    // Type email and password
    cy.get('#root_email').type('test@frontrunner.com');
      cy.get('#root_password').type('frontrunner');
    
    // Verify input values
    cy.get('#root_email').should('have.value', 'test@frontrunner.com');
    cy.get('#root_password').should('have.value', 'frontrunner');
  });

  it('should submit the form successfully', () => {
    // Type valid credentials
    cy.get('#root_email').type('test@frontrunner.com');
    cy.get('#root_password').type('frontrunner');

    // Submit the form
    cy.get('button[type="submit"]').click();
    
    // Wait for the API request and verify it was called
    cy.wait('@loginRequest').then((interception) => {
      // Check that the request body contains the credentials
      expect(interception.request.body).to.include('email=test%40frontrunner.com&password=frontrunner');
      
      // Check status code
      expect(interception.response.statusCode).to.equal(200);
    });
  });

  it('should show error message for invalid credentials', () => {
    // Override the mock to return an error for invalid credentials
    cy.intercept('POST', '/api/login', {
      statusCode: 401,
      body: { success: false, message: 'Invalid credentials' }
    }).as('failedLoginRequest');
    
    // Type invalid credentials
    cy.get('#root_email').should('be.visible').type('wrong@example.com');
    cy.get('#root_password').should('be.visible').type('wrongpassword');
    
    // Submit the form
    cy.get('button[type="submit"]').click();
    
    // Wait for the failed login request
    cy.wait('@failedLoginRequest').then((interception) => {
      // Check status code
      expect(interception.response.statusCode).to.equal(401);
    });
  });
});
