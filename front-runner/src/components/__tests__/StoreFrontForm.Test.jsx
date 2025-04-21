import { render, screen, fireEvent } from '@testing-library/react';
import StorefrontLinkFormRJSF from '../StorefrontLinkForm';

test('allows the user to input storefront details and submit the form', () => {
  // Mock fetch and success callback
  global.fetch = jest.fn(() => Promise.resolve({ ok: true }));
  const mockOnSubmitSuccess = jest.fn();
  const mockOnClose = jest.fn();

  render(<StorefrontLinkFormRJSF storefront={null} onSubmitSuccess={mockOnSubmitSuccess} onClose={mockOnClose} />);

  // Find input fields by their labels
  const linkNameInput = screen.getByLabelText(/Link Name/i);
  const apiKeyInput = screen.getByLabelText(/API Key/i);
  const apiSecretInput = screen.getByLabelText(/API Secret/i);
  const storeIdInput = screen.getByLabelText(/Store ID/i);
  const storeUrlInput = screen.getByLabelText(/Store URL/i);

  // Simulate user input
  fireEvent.change(linkNameInput, { target: { value: 'My Amazon Store' } });
  fireEvent.change(apiKeyInput, { target: { value: 'my-api-key' } });
  fireEvent.change(apiSecretInput, { target: { value: 'my-secret' } });
  fireEvent.change(storeIdInput, { target: { value: 'AMZ123' } });
  fireEvent.change(storeUrlInput, { target: { value: 'https://amazon.com/mystore' } });

  // Check the values
  expect(linkNameInput.value).toBe('My Amazon Store');
  expect(apiKeyInput.value).toBe('my-api-key');
  expect(apiSecretInput.value).toBe('my-secret');
  expect(storeIdInput.value).toBe('AMZ123');
  expect(storeUrlInput.value).toBe('https://amazon.com/mystore');

  // Submit the form
  fireEvent.submit(screen.getByRole('form'));
});
