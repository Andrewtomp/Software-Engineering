describe('Orders Page', () => {
    it('should trigger a download and verify a PNG file is downloaded', () => {
      // Step 1: Visit the orders page
      cy.visit('http://localhost:3000/orders'); // Replace with your actual URL
  
      // Step 2: Intercept the request for downloading the shipping label (if needed)
      cy.intercept('GET', '/path/to/shipping_label_0001.png').as('downloadRequest');
  
      // Step 3: Click the "Download" button (first button in the "Shipping Label" column)
      cy.get('button.table-button', { timeout: 10000 }).first().click();
  
      // Step 4: Wait for a few seconds to ensure the file has been downloaded
      cy.wait(2000); // Waiting for the download to complete, feel free to adjust time
  
    });
  });
  