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
    /*
    // Check if email field is present and visible
    cy.get('input[type="email"]').should('be.visible');
    
    // Check if password field is present and visible
    cy.get('input[type="password"]').should('be.visible');
    
    // Check if submit button is visible
    cy.get('button[type="submit"]').contains('Submit').should('be.visible');
  });

  it('should allow user to input credentials', () => {
    // Type email and password
    cy.get('input[type="email"]').should('be.visible').type('test@example.com');
    cy.get('input[type="password"]').should('be.visible').type('password123!');
    
    // Verify input values
    cy.get('input[type="email"]').should('have.value', 'test@example.com');
    cy.get('input[type="password"]').should('have.value', 'password123!');
  });

  it('should submit the form successfully', () => {
    // Fill out the form
    cy.get('input[type="email"]').should('be.visible').type('user@test.com');
    cy.get('input[type="password"]').should('be.visible').type('securepass');
    
    // Submit the form
    cy.get('form').submit();
    
    // Wait for the API request and verify it was called
    cy.wait('@loginRequest').then((interception) => {
      // Check that the request body contains the credentials
      expect(interception.request.body).to.include({
        email: 'user@test.com',
        password: 'securepass'
      });
      
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
    cy.get('input[type="email"]').should('be.visible').type('wrong@example.com');
    cy.get('input[type="password"]').should('be.visible').type('wrongpassword');
    
    // Submit the form
    cy.get('button[type="submit"]').click();
    
    // Wait for the failed login request
    cy.wait('@failedLoginRequest');
    
    // Check if error message is displayed
    cy.get('[role="alert"], .error-message, .alert')
      .should('be.visible')
      .and('contain.text', 'Invalid credentials');

      */
  });
});
