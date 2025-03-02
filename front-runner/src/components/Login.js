import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';

// Define the JSON Schema for the login form
const schema = {
  title: 'Login',
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

// Define the login form's onSubmit handler
// const onSubmit = ({ formData }) => {
//   const headers = new Headers();
//   headers.set("Content-Type", "application/x-www-form-urlencoded")
//   fetch("/api/login", {
//       method: 'post',
//       body: new URLSearchParams(formData)
//   });
// };
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
      const redirectUrl = response.headers.get('Location');
      console.log('onSubmit: Redirect URL from server:', redirectUrl);
      window.location.href = redirectUrl ? redirectUrl : '/';
    } else if (response.ok) {
      console.log('onSubmit: Login succeeded without explicit redirect, navigating to home');
      // If login is successful but no explicit redirect, navigate to home
      window.location.href = '/';
    } else {
      const errorText = await response.text();
      console.error('Login failed:', errorText);
      // Optionally display an error message to the user
    }
  } catch (error) {
    console.error('Error during login:', error);
  }
};

// LoginForm Component
const LoginForm = () => {
  return (
    <div style={{ width: '400px', margin: '0 auto', padding: '20px', backgroundColor: '#f9f9f9' }}>
      <h2>Login</h2>
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        onSubmit={onSubmit} // Handle form submission
      />
      <div style={{ marginTop: '10px', textAlign: 'center' }}>
        <button onClick={() => window.location.href = '/register'} style={{ padding: '10px', backgroundColor: '#28a745', color: '#fff', border: 'none', cursor: 'pointer' }}>
          Register
        </button>
      </div>
    </div>
  );
};

export default LoginForm;
