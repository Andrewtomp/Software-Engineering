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
    
    // More flexible ways to find the form elements
    cy.get('form').should('exist');
    cy.get('button').contains(/login|sign in|submit/i).should('be.visible');
  });

  it('should allow user to input credentials', () => {
    // Multiple strategies to find email field
    cy.getEmailField().type('test@example.com');
    
    // Multiple strategies to find password field
    cy.getPasswordField().type('password123!');
    
    // Verify input values
    cy.getEmailField().should('have.value', 'test@example.com');
    cy.getPasswordField().should('have.value', 'password123!');
  });

  it('should submit the form successfully', () => {
    // Fill out the form
    cy.getEmailField().type('user@test.com');
    cy.getPasswordField().type('securepass');
    
    // Submit the form - try multiple strategies
    cy.get('form').submit();
    
    // Wait for the API request and verify it was called
    cy.wait('@loginRequest').then((interception) => {
      // Check that the request contains our credentials (the exact format may vary)
      expect(interception.request.body).to.include.any.keys('email', 'username');
      
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
    cy.getEmailField().type('wrong@example.com');
    cy.getPasswordField().type('wrongpassword');
    
    // Submit the form
    cy.get('button[type="submit"], button:contains("Login"), button:contains("Sign In")').click();
    
    // Wait for the API request
    cy.wait('@failedLoginRequest');
    
    // Check if error message is displayed - trying multiple approaches
    cy.get('[role="alert"], .error-message, .alert')
      .should('be.visible')
      .or('contain.text', 'Invalid')
      .or('contain.text', 'failed')
      .or('contain.text', 'incorrect');
  });
});

// Custom commands to find form fields with multiple strategies
Cypress.Commands.add('getEmailField', () => {
  // Try different ways to find the email field
  return cy.get('body').then($body => {
    // Try by input type
    if ($body.find('input[type="email"]').length > 0) {
      return cy.get('input[type="email"]');
    }
    
    // Try by placeholder
    if ($body.find('input[placeholder*="email" i]').length > 0) {
      return cy.get('input[placeholder*="email" i]');
    }
    
    // Try by ID
    if ($body.find('#email, #emailInput, #username').length > 0) {
      return cy.get('#email, #emailInput, #username');
    }
    
    // Try by name attribute
    if ($body.find('input[name="email"], input[name="username"]').length > 0) {
      return cy.get('input[name="email"], input[name="username"]');
    }
    
    // Try by aria-label
    if ($body.find('input[aria-label*="email" i], input[aria-label*="username" i]').length > 0) {
      return cy.get('input[aria-label*="email" i], input[aria-label*="username" i]');
    }
    
    // Try by looking at labels
    if ($body.find('label:contains("Email"), label:contains("Username")').length > 0) {
      const label = $body.find('label:contains("Email"), label:contains("Username")').first();
      const forAttr = label.attr('for');
      if (forAttr) {
        return cy.get(`#${forAttr}`);
      }
    }
    
    // Fallback - get the first input in the form
    return cy.get('form input').first();
  });
});

Cypress.Commands.add('getPasswordField', () => {
  // Try different ways to find the password field
  return cy.get('body').then($body => {
    // Try by input type
    if ($body.find('input[type="password"]').length > 0) {
      return cy.get('input[type="password"]');
    }
    
    // Try by placeholder
    if ($body.find('input[placeholder*="password" i]').length > 0) {
      return cy.get('input[placeholder*="password" i]');
    }
    
    // Try by ID
    if ($body.find('#password, #passwordInput').length > 0) {
      return cy.get('#password, #passwordInput');
    }
    
    // Try by name attribute
    if ($body.find('input[name="password"]').length > 0) {
      return cy.get('input[name="password"]');
    }
    
    // Try by aria-label
    if ($body.find('input[aria-label*="password" i]').length > 0) {
      return cy.get('input[aria-label*="password" i]');
    }
    
    // Try by looking at labels
    if ($body.find('label:contains("Password")').length > 0) {
      const label = $body.find('label:contains("Password")').first();
      const forAttr = label.attr('for');
      if (forAttr) {
        return cy.get(`#${forAttr}`);
      }
    }
    
    // Fallback - get the second input in the form assuming it's the password
    return cy.get('form input').eq(1);
  });
});