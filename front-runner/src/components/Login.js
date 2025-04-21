import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';
import './Login.css'; 

// Define the JSON Schema for the login form
const schema = {
  type: 'object',
  properties: {
    email: {
      type: 'string',
      title: 'Email',
    },
    password: {
      type: 'string',
      title: 'Password',
      minLength: 6, // Minimum length for password validation
    },
  },
  required: ['email', 'password'], // Make both fields required
};

// Define the UI Schema for the form fields (optional customization)
const uiSchema = {
  password: {
    'ui:widget': 'password', // Automatically uses a password input
    'ui:placeholder': 'Enter your password',
  },
  email: {
    'ui:placeholder': 'Enter your email',
  },
};

// Custom templates for Form to ensure proper accessibility and test compatibility
const CustomFieldTemplate = (props) => {
  const { id, label, children, rawErrors, required } = props;
  
  return (
    <div className="form-group mb-3">
      <label htmlFor={id}>{label}{required ? "*" : ""}</label>
      {children}
      {rawErrors && rawErrors.length > 0 && (
        <div className="text-danger">{rawErrors.join(', ')}</div>
      )}
    </div>
  );
};

// Custom submit button template to add proper role
const CustomButtonTemplate = (props) => {
  return (
    <div className="d-flex justify-content-center mb-3">
      <button 
        type="submit" 
        className="btn btn-primary"
        role="button"
      >
        Login with Email
      </button>
    </div>
  );
};

// LoginForm Component
const LoginForm = () => {
  const onSubmit = async ({ formData }) => {
    try {
      console.log('onSubmit: Submitting login request with data:', formData);
      const response = await fetch("/api/login", {
        method: 'POST',
        body: new URLSearchParams(formData),
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        },
        redirect: 'manual' // Prevent automatic following of redirects
      });
      console.log('onSubmit: Response status:', response.status);
  
      // if (response.status === 303) {
      if (response.type === "opaqueredirect") {
        // If server returns a redirect, update the browser location manually
        // const redirectUrl = response.headers.get('Location');
        // console.log('onSubmit: Redirect URL from server:', redirectUrl);
        // window.location.href = redirectUrl ? redirectUrl : '/';
        console.log('onSubmit: Redirect detected or implied, navigating to /');
        window.location.href = '/'; // Navigate to dashboard after successful login/redirect
      } else if (response.ok) {
        console.log('onSubmit: Login succeeded without explicit redirect, navigating to home');
        // If login is successful but no explicit redirect, navigate to home
        window.location.href = '/';
      } else {
        const errorText = await response.text();
        console.error('Login failed:', errorText);
        // Optionally display an error message to the user
        alert(`Login failed: ${errorText}`);
      }
    } catch (error) {
      console.error('Error during login:', error);
      alert('An error occurred during login. Please try again.');
    }
  };
  // const onSubmit = ({ formData }) => {
  //   const headers = new Headers();
  //   headers.set("Content-Type", "application/x-www-form-urlencoded");
  //   fetch("/api/login", {
  //     method: 'post',
  //     body: new URLSearchParams(formData)
  //   });
  // };

  const handleGoogleLogin = () => {
    window.location.href = '/auth/google';
  };

  return (
    <div className="login-container" style={{ backgroundImage: `url("../assets/FrontRunner Login Background.png")`, backgroundSize: "cover", backgroundPosition: "center"}}>
      <div className='login-card'>
        <h2 className="text-center mb-4">Login</h2>
        <Form
          schema={schema}
          uiSchema={uiSchema}
          validator={validator}
          onSubmit={onSubmit}
          templates={{
            FieldTemplate: CustomFieldTemplate,
            ButtonTemplates: { SubmitButton: CustomButtonTemplate }
          }}
        />
        {/* Divider */}
        <div className="or-divider">OR</div>

        {/* Google Login Button */}
        <div className="d-flex justify-content-center mb-3">
          <button
            type="button" // Important: type="button" prevents form submission
            className="btn btn-primary" // Using Bootstrap's red color for Google
            onClick={handleGoogleLogin}
            role="button"
          >
            <i className="bi bi-google me-2"></i> {/* Optional: Add Google icon if using Bootstrap Icons */}
            Login with Google
          </button>
        </div>

        <div className="text-center">
          <a href='/register'>
            New here? Create an account.
          </a>
        </div>
      </div>
    </div>
  );
};

export default LoginForm;