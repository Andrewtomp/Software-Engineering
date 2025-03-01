// LoginForm.tsx (or LoginForm.js if you're not using TypeScript)
import React from 'react';
import Form from '@rjsf/core';
import { RJSFSchema } from '@rjsf/utils';
import validator from '@rjsf/validator-ajv8';

// Define the JSON Schema for the login form
const schema: RJSFSchema = {
  title: 'Login Form',
  type: 'object',
  properties: {
    username: {
      type: 'string',
      title: 'Username',
    },
    password: {
      type: 'string',
      title: 'Password',
      minLength: 6, // Minimum length for password validation
    },
  },
  required: ['username', 'password'], // Make both fields required
};

// Define the UI Schema for the form fields (optional customization)
const uiSchema = {
  password: {
    'ui:widget': 'password', // Automatically uses a password input
    'ui:placeholder': 'Enter your password',
  },
  username: {
    'ui:placeholder': 'Enter your username or email',
  },
};

// Define the login form's onSubmit handler
const onSubmit = ({ formData }: any) => {
  // Here you can process the login form data
  console.log('Form data submitted:', formData);
  // For example, send the data to an API to authenticate the user
  alert('Login attempt: ' + JSON.stringify(formData));
};

// LoginForm Component
const Login = () => {
  return (
    <div style={{ width: '400px', margin: '0 auto', padding: '20px', backgroundColor: '#f9f9f9' }}>
      <h2>Login</h2>
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        onSubmit={onSubmit} // Handle form submission
      />
    </div>
  );
};

export default Login;
