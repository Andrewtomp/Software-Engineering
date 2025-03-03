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
        Submit
      </button>
    </div>
  );
};

// LoginForm Component
const LoginForm = () => {
  const onSubmit = ({ formData }) => {
    const headers = new Headers();
    headers.set("Content-Type", "application/x-www-form-urlencoded");
    fetch("/api/login", {
      method: 'post',
      body: new URLSearchParams(formData)
    });
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
        <a href='/register'>
          New here? Create an account.
        </a>
      </div>
    </div>
  );
};

export default LoginForm;