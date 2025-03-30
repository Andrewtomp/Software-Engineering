// cypress/integration/login.spec.js
describe('Login Form', () => {
    beforeEach(() => {
      // Visit the page with the login form
      cy.visit('http://localhost:3000/login');
      
      // Mock the login API endpoint
      cy.intercept('POST', '/api/login', {
        statusCode: 200,
        body: { success: true }
      }).as('loginRequest');
    });
  
    it('should render the login form correctly', () => {
      // Check if the login title is present
      cy.contains(/login/i).should('be.visible');
      
      // Check if form fields are present
      cy.get('label').contains(/email/i).should('be.visible');
      cy.get('label').contains(/password/i).should('be.visible');
      
      // Check if submit button is present
      cy.get('button[type="submit"]').should('be.visible');
    });
  
    it('should allow user to input credentials', () => {
      // Type into email field
      cy.get('input[type="email"]').type('test@example.com');
      
      // Type into password field
      cy.get('input[type="password"]').type('password123!');
      
      // Verify input values
      cy.get('input[type="email"]').should('have.value', 'test@example.com');
      cy.get('input[type="password"]').should('have.value', 'password123!');
    });
  
    it('should submit the form successfully', () => {
      // Type into email field
      cy.get('input[type="email"]').type('user@test.com');
      
      // Type into password field
      cy.get('input[type="password"]').type('securepass');
      
      // Submit the form
      cy.get('button[type="submit"]').click();
      
      // Wait for the API request and verify it was called
      cy.wait('@loginRequest').then((interception) => {
        // Check request payload
        expect(interception.request.body).to.include({
          email: 'user@test.com',
          password: 'securepass'
        });
        
        // Check status code
        expect(interception.response.statusCode).to.equal(200);
      });
    });
  
    it('should show error message for invalid credentials', () => {
      // Override the mock to return an error
      cy.intercept('POST', '/api/login', {
        statusCode: 401,
        body: { success: false, message: 'Invalid credentials' }
      }).as('failedLoginRequest');
      
      // Type credentials
      cy.get('input[type="email"]').type('wrong@example.com');
      cy.get('input[type="password"]').type('wrongpassword');
      
      // Submit the form
      cy.get('button[type="submit"]').click();
      
      // Wait for the API request
      cy.wait('@failedLoginRequest');
      
      // Check if error message is displayed
      cy.contains('Invalid credentials').should('be.visible');
    });
  
    it('should validate required fields', () => {
      // Submit without entering any data
      cy.get('button[type="submit"]').click();
      
      // Check for validation messages
      cy.contains(/required/i).should('be.visible');
      
      // The login request should not be made
      cy.get('@loginRequest.all').should('have.length', 0);
    });
  
    it('should redirect after successful login', () => {
      // Mock the login response with a redirect
      cy.intercept('POST', '/api/login', {
        statusCode: 200,
        body: { success: true, redirectUrl: '/dashboard' }
      }).as('loginWithRedirect');
      
      // Stub the window.location.href to track redirects
      cy.window().then(win => {
        cy.stub(win, 'location').as('windowLocation');
      });
      
      // Fill and submit the form
      cy.get('input[type="email"]').type('user@test.com');
      cy.get('input[type="password"]').type('securepass');
      cy.get('button[type="submit"]').click();
      
      // Wait for the login request
      cy.wait('@loginWithRedirect');
      
      // Assert the page was redirected or navigation occurred
      cy.url().should('include', '/dashboard');
    });
  });