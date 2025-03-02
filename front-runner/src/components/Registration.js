import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';

// Define the JSON Schema for the registration form
const schema = {
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
    <div className="login-container">
      <div className='logo-header'>
        <img src="../assets/Logo.svg" className="logo" alt="FR logo"/>
        <h1>FrontRunner</h1>
      </div>
      <div className='login-card'>
        <h2 className="text-center mb-4">Create an account</h2>
        <Form
          schema={schema}
          uiSchema={uiSchema}
          validator={validator}
          onSubmit={onSubmit} // Handle form submission
        />
        <a href = '/login'>
            Already have an account? Login here.
        </a>
      </div>
    </div>
  );
};

export default RegistrationForm;
