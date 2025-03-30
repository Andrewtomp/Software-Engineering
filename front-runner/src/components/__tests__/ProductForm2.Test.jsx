import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import ProductForm from '../ProductForm';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

// Mock the fetch API to avoid actual HTTP requests during tests
global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    json: () => Promise.resolve({}),
  })
);

// Mocking the FontAwesomeIcon component as we are not testing the icon behavior
jest.mock('@fortawesome/react-fontawesome', () => ({
  FontAwesomeIcon: () => <span>Icon</span>,
}));

describe('ProductForm', () => {
  // Test case: Check if the form renders correctly for adding a new product
  test('renders Add Product form', () => {
    render(<ProductForm onClose={() => {}} />);

    expect(screen.getByText('Add Product')).toBeInTheDocument();
    expect(screen.getByLabelText(/Product Name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Description/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Price/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Count/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Tags/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Product Image/i)).toBeInTheDocument();
  });

  // Test case: Check form submission with valid data
  test('submits form with valid data', async () => {
    render(<ProductForm onClose={() => {}} />);

    fireEvent.change(screen.getByLabelText(/Product Name/i), {
      target: { value: 'Product 1' },
    });
    fireEvent.change(screen.getByLabelText(/Description/i), {
      target: { value: 'This is a product' },
    });
    fireEvent.change(screen.getByLabelText(/Price/i), {
      target: { value: '$12.99' },
    });
    fireEvent.change(screen.getByLabelText(/Count/i), {
      target: { value: 10 },
    });
    fireEvent.change(screen.getByLabelText(/Tags/i), {
      target: { value: '#sale, #new' },
    });

    const submitButton = screen.getByText(/submit/i); // Assuming a submit button exists
    fireEvent.click(submitButton);

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1));
    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        method: 'POST',
        body: expect.any(FormData),
      })
    );
  });

  // Test case: Check file input and image preview functionality
  test('handles image upload and displays image preview', async () => {
    render(<ProductForm onClose={() => {}} />);

    const imageInput = screen.getByLabelText(/Product Image/i);
    const file = new File(['image'], 'image.jpg', { type: 'image/jpeg' });

    Object.defineProperty(imageInput, 'files', {
      value: [file],
    });

    fireEvent.change(imageInput);

    // Check if the image preview is shown
    expect(screen.getByAltText('Product preview')).toBeInTheDocument();
  });

  // Test case: Check if delete button works correctly
  test('calls delete function when delete icon is clicked', async () => {
    const mockProduct = {
      id: 1,
      name: 'Product 1',
      description: 'Product Description',
      price: '12.99',
      count: 10,
      tags: '#tag1',
      image: 'image.jpg',
    };

    render(<ProductForm onClose={() => {}} product={mockProduct} />);

    const deleteButton = screen.getByText(/Icon/); // The FontAwesomeIcon is mocked as 'Icon'
    fireEvent.click(deleteButton);

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1));
    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        method: 'DELETE',
      })
    );
  });

  // Test case: Check form submission with missing required fields
  test('shows validation errors when submitting incomplete form', async () => {
    render(<ProductForm onClose={() => {}} />);

    const submitButton = screen.getByText(/submit/i); // Assuming a submit button exists
    fireEvent.click(submitButton);

    // Check for validation errors
    expect(screen.getByText(/product name is required/i)).toBeInTheDocument();
    expect(screen.getByText(/description is required/i)).toBeInTheDocument();
    expect(screen.getByText(/price is required/i)).toBeInTheDocument();
    expect(screen.getByText(/image is required/i)).toBeInTheDocument();
  });

  // Test case: Edit an existing product
  test('edits an existing product', async () => {
    const mockProduct = {
      id: 1,
      name: 'Product 1',
      description: 'Product Description',
      price: '12.99',
      count: 10,
      tags: '#tag1',
      image: 'image.jpg',
    };

    render(<ProductForm onClose={() => {}} product={mockProduct} />);

    // Check if the form is populated with existing product data
    expect(screen.getByDisplayValue('Product 1')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Product Description')).toBeInTheDocument();
    expect(screen.getByDisplayValue('$12.99')).toBeInTheDocument();
    expect(screen.getByDisplayValue('10')).toBeInTheDocument();
    expect(screen.getByDisplayValue('#tag1')).toBeInTheDocument();

    // Simulate editing the product name
    fireEvent.change(screen.getByLabelText(/Product Name/i), {
      target: { value: 'Updated Product 1' },
    });

    // Simulate submitting the form
    const submitButton = screen.getByText(/submit/i);
    fireEvent.click(submitButton);

    // Check if the correct API call is made (PUT request for updating)
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1));
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/update_product'),
      expect.objectContaining({
        method: 'PUT',
        body: expect.any(FormData),
      })
    );
  });
});
