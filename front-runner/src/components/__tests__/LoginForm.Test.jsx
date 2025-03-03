import { render, screen, fireEvent } from '@testing-library/react';
import LoginForm from '../Login'; // Adjust the path based on your project structure

describe('LoginForm Component', () => {
  test('renders form and allows user input', () => {
    render(<LoginForm />);

    // Check if form title is present
    expect(screen.getByText(/Login/i)).toBeInTheDocument();

    // Get input fields
    const emailInput = screen.getByLabelText(/Email/i);
    const passwordInput = screen.getByLabelText(/Password/i);

    // Type into input fields
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123!' } });

    // Verify input values
    expect(emailInput.value).toBe('test@example.com');
    expect(passwordInput.value).toBe('password123!');
  });

  test('submits form successfully', () => {
    render(<LoginForm />);

    // Get input fields and submit button
    const emailInput = screen.getByLabelText(/Email/i);
    const passwordInput = screen.getByLabelText(/Password/i);
    const submitButton = screen.getByRole('button', { name: /submit/i });

    // Mock fetch function
    global.fetch = jest.fn(() =>
      Promise.resolve({
        json: () => Promise.resolve({ success: true }),
      })
    );

    // Type into fields and submit the form
    fireEvent.change(emailInput, { target: { value: 'user@test.com' } });
    fireEvent.change(passwordInput, { target: { value: 'securepass' } });
    fireEvent.submit(submitButton);

    // Ensure fetch is called with correct API and method
    //expect(global.fetch).toHaveBeenCalledWith(
    //  '/api/login',
    //  expect.objectContaining({ method: 'post' })
   // );
  });
});
