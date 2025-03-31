import { render, screen, fireEvent } from '@testing-library/react';
import AddProductForm from '../ProductForm';

test('allows the user to input, edit, and delete a product', () => {
  // Mock alert and confirm before rendering
  global.alert = jest.fn();
  global.confirm = jest.fn(() => true);  // Mocking confirmation dialog

  // Step 1: Render the Add Product Form for a new product (simulating adding a product)
  render(<AddProductForm />);

  // Find input fields by their labels
  const productNameInput = screen.getByLabelText(/Product Name/i);
  const descriptionInput = screen.getByLabelText(/Description/i);
  const priceInput = screen.getByLabelText(/Price/i);
  const stockInput = screen.getByLabelText(/Count/i);
  const submitButton = screen.getByRole('button', { name: /Submit/i });

  // Simulate user input for a new product
  fireEvent.change(productNameInput, { target: { value: 'Test Product' } });
  fireEvent.change(descriptionInput, { target: { value: 'This is a test description.' } });
  fireEvent.change(priceInput, { target: { value: '29.99' } });
  fireEvent.change(stockInput, { target: { value: '10' } });

  // Ensure values are set correctly
  expect(productNameInput.value).toBe('Test Product');
  expect(descriptionInput.value).toBe('This is a test description.');
  expect(priceInput.value).toBe('29.99');
  expect(stockInput.value).toBe('10');

  // Simulate form submission (for adding the product)
  fireEvent.submit(submitButton);

  // Step 2: Re-render the form for editing by passing the existing product data
  render(<AddProductForm product={{
    name: 'Test Product',
    description: 'This is a test description.',
    price: '29.99',
    count: 10
  }} />);

  // Find input fields by their labels again for the edit form
  const editProductNameInput = screen.getByLabelText(/Product Name/i);
  const editDescriptionInput = screen.getByLabelText(/Description/i);
  const editPriceInput = screen.getByLabelText(/Price/i);
  const editStockInput = screen.getByLabelText(/Count/i);
  
  // Check for the correct label on the button (Update, Save, or whatever your button is named)
  const editSubmitButton = screen.getAllByRole('button', { name: /Submit/i })[0];  // Updated the name here

  // Simulate editing product details
  fireEvent.change(editProductNameInput, { target: { value: 'Edited Product' } });
  fireEvent.change(editDescriptionInput, { target: { value: 'Updated test description.' } });
  fireEvent.change(editPriceInput, { target: { value: '39.99' } });
  fireEvent.change(editStockInput, { target: { value: '20' } });

  // Ensure edited values are set correctly
  expect(editProductNameInput.value).toBe('Edited Product');
  expect(editDescriptionInput.value).toBe('Updated test description.');
  expect(editPriceInput.value).toBe('39.99');
  expect(editStockInput.value).toBe('20');

  // Simulate form submission (for editing the product)
  fireEvent.submit(editSubmitButton);

  render(<AddProductForm product={{
    name: 'Test Product',
    description: 'This is a test description.',
    price: '29.99',
    count: 10
  }} />);

  // Step 3: Simulate clicking the delete button for the product
  // get the delete icon thats a font-awesome icon
  const deleteButton = screen.getAllByTestId('delete-icon')[0];
  fireEvent.click(deleteButton);

  // Ensure that the delete confirmation dialog is triggered
  expect(global.confirm).toHaveBeenCalledWith('Are you sure you want to delete Test Product?');
});