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
const onSubmit = async ({ formData }) => {
  try {
    console.log('Registration: Submitting registration request with data:', formData);
    // First, call the registration API
    const regResponse = await fetch("/api/register", {
      method: 'POST',
      body: new URLSearchParams(formData),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      },
    });
    
    if (!regResponse.ok) {
      const regErrorText = await regResponse.text();
      console.error("Registration failed:", regErrorText);
      return;
    }
    
    console.log("Registration succeeded, now auto-logging in");
    
    // Next, call the login API using the same credentials
    const loginResponse = await fetch("/api/login", {
      method: 'POST',
      body: new URLSearchParams({
        email: formData.email,
        password: formData.password
      }),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      },
      redirect: 'manual'
    });
    
    console.log("Login response status:", loginResponse.status, "type:", loginResponse.type);
    
    // Handle both opaqueredirect or explicit 303 from the server
    if (loginResponse.status === 303 || loginResponse.type === "opaqueredirect") {
      console.log("Auto-login redirect detected, navigating to home");
      window.location.href = "/";
    } else if (loginResponse.ok) {
      console.log("Auto-login succeeded, navigating to home");
      window.location.href = "/";
    } else {
      const loginErrorText = await loginResponse.text();
      console.error("Auto-login failed:", loginErrorText);
    }
    
  } catch (error) {
    console.error("Error during registration and auto-login:", error);
  }
};

// RegistrationForm Component
const RegistrationForm = () => {
  return (
    <div className="login-container" style={{ backgroundImage: `url("../assets/FrontRunner Login Background.png")`, backgroundSize: "cover", backgroundPosition: "center"}}>
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
