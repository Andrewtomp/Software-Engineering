import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';

// Define the JSON Schema for the registration form
const schema = {
  title: 'User Registration',
  type: 'object',
  properties: {
    email: {
      type: 'string',
      title: 'Email',
      format: 'email',
    },
    password: {
      type: 'string',
      title: 'Password',
      minLength: 6, // Minimum length for password validation
    },
    businessName: {
      type: 'string',
      title: 'Business Name',
    },
  },
  required: ['email', 'password'], // Make email and password required
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
  businessName: {
    'ui:placeholder': 'Enter your business name (optional)',
  },
};

// Define the form's onSubmit handler
const onSubmit = ({ formData }) => {
        const headers = new Headers();
        headers.set("Content-Type", "application/x-www-form-urlencoded")
        fetch("/api/register", {
            method: 'post',
            body: new URLSearchParams(formData)
        });
};

// RegistrationForm Component
const RegistrationForm = () => {
  return (
    <div style={{ width: '400px', margin: '0 auto', padding: '20px', backgroundColor: '#f9f9f9' }}>
      <h2>User Registration</h2>
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        onSubmit={onSubmit} // Handle form submission
      />
      <div style={{ marginTop: '10px', textAlign: 'center' }}>
      <button onClick={() => window.location.href = '/login'} style={{ padding: '10px', backgroundColor: '#28a745', color: '#fff', border: 'none', cursor: 'pointer' }}>
          Login
        </button>
      </div>
    </div>
  );
};

export default RegistrationForm;
