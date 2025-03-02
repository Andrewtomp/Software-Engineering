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

// Define the login form's onSubmit handler
const onSubmit = ({ formData }) => {
  const headers = new Headers();
  headers.set("Content-Type", "application/x-www-form-urlencoded")
  fetch("/api/login", {
      method: 'post',
      body: new URLSearchParams(formData)
  });
};

// LoginForm Component
const LoginForm = () => {
  return (
    <div className="login-container">
      <div className='logo-header'>
        <img src="../assets/Logo.svg" className="logo" alt="FR logo"/>
        <h1>FrontRunner</h1>
      </div>
      <div className='login-card'>
        <h2 className="text-center mb-4">Login</h2>
        <Form
          schema={schema}
          uiSchema={uiSchema}
          validator={validator}
          onSubmit={onSubmit} // Handle form submission
        />
          <a href= '/register'>
          New here? Create an account.
          </a>
      </div>
    </div>
  );
};

export default LoginForm;