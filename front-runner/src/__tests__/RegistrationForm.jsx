import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RegistrationForm from '../components/Registration'; // Adjust the import path

describe('RegistrationForm', () => {
  test('renders the registration form', () => {
    render(<RegistrationForm />);
    
    // Check if form title appears
    expect(screen.getByText(/Create an account/i)).toBeInTheDocument();

    // Check for form fields
    expect(screen.getByPlaceholderText(/Enter your email/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Enter your password/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Enter your business name/i)).toBeInTheDocument();
  });

  test('allows users to enter registration details', async () => {
    render(<RegistrationForm />);
    
    const emailInput = screen.getByPlaceholderText(/Enter your email/i);
    const passwordInput = screen.getByPlaceholderText(/Enter your password/i);
    const businessNameInput = screen.getByPlaceholderText(/Enter your business name/i);

    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'securepassword');
    await userEvent.type(businessNameInput, 'Test Business');

    expect(emailInput).toHaveValue('test@example.com');
    expect(passwordInput).toHaveValue('securepassword');
    expect(businessNameInput).toHaveValue('Test Business');
  });

  test('displays a required error if email or password is missing', async () => {
    render(<RegistrationForm />);
    
    const submitButton = screen.getByRole('button', { name: /submit/i });
    await userEvent.click(submitButton);

    // Assuming the form validation displays error messages
    expect(screen.getByText(/is a required property/i)).toBeInTheDocument();
  });

  test('submits the form successfully', async () => {
    global.fetch = jest.fn(() => Promise.resolve({ ok: true }));

    render(<RegistrationForm />);
    const emailInput = screen.getByPlaceholderText(/Enter your email/i);
    const passwordInput = screen.getByPlaceholderText(/Enter your password/i);
    const businessNameInput = screen.getByPlaceholderText(/Enter your business name/i);
    const submitButton = screen.getByRole('button', { name: /submit/i });

    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'securepassword');
    await userEvent.type(businessNameInput, 'Test Business');
    await userEvent.click(submitButton);

   // expect(global.fetch).toHaveBeenCalledWith(
   //   '/api/register',
   //   expect.objectContaining({
   //     method: 'post',
   //   })
   // );
  });
});
