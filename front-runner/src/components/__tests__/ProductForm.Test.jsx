import { render, screen, fireEvent } from '@testing-library/react';
import AddProductForm from '../ProductForm';

test('allows the user to input product details and submit the form', () => {
  // Mock alert before rendering
  global.alert = jest.fn();
  
  render(<AddProductForm />);

  // Find input fields by their labels
  const productNameInput = screen.getByLabelText(/Product Name/i);
  const descriptionInput = screen.getByLabelText(/Description/i);
  const priceInput = screen.getByLabelText(/Price/i);
  const stockInput = screen.getByLabelText(/Count/i);
  const submitButton = screen.getByRole('button', { name: /Submit/i });

  // Simulate user input
  fireEvent.change(productNameInput, { target: { value: 'Test Product' } });
  fireEvent.change(descriptionInput, { target: { value: 'This is a test description.' } });
  fireEvent.change(priceInput, { target: { value: '29.99' } });
  fireEvent.change(stockInput, { target: { value: '10' } });

  // Ensure values are set correctly using standard Jest expectations
  expect(productNameInput.value).toBe('Test Product');
  expect(descriptionInput.value).toBe('This is a test description.');
  expect(parseFloat(priceInput.value)).toBe(29.99);
  expect(parseInt(stockInput.value)).toBe(10);

  // Simulate form submission
  fireEvent.submit(submitButton);
});