import React, { useState } from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';

// Define the JSON Schema for the registration form
const schema = {
  type: 'object',
  properties: {
    name: { // <-- ADDED name property
      type: 'string',
      title: 'Full Name',
    },
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
      title: 'Business Name (Optional)',
    },
  },
  required: ['name', 'email', 'password'], // Make email and password required
};

// Define the UI Schema for the form fields (optional customization)
const uiSchema = {
  name: {
    'ui:placeholder': 'Enter your full name',
  },
  email: {
    'ui:placeholder': 'Enter your email',
  },
  password: {
    'ui:widget': 'password', // Automatically uses a password input
    'ui:placeholder': 'Enter your password (min 6 characters)',
  },
  businessName: {
    'ui:placeholder': 'Enter your business name (Optional)',
  },
};

// Custom templates for Form to ensure proper accessibility and test compatibility
// This template already handles marking required fields
const CustomFieldTemplate = (props) => {
  const { id, label, children, rawErrors, required, description, help } = props; // Added description and help

  return (
    <div className="form-group mb-3">
      {/* Display label with asterisk if required */}
      <label htmlFor={id}>{label}{required ? "*" : ""}</label>
      {/* Display description/help text if provided */}
      {description}
      {help}
      {children}
      {/* Display errors */}
      {rawErrors && rawErrors.length > 0 && (
        <div className="text-danger mt-1" style={{ fontSize: '0.8rem' }}>{rawErrors.join(', ')}</div>
      )}
    </div>
  );
};

// Custom field template to add proper role to submit button
// const CustomButtonTemplate = (props) => {
//   const { id, label, children, rawErrors, required, description, help} = props;
  
//   return (
//     <div className="d-flex justify-content-center">
//       <button 
//         type="submit" 
//         className="btn btn-primary"
//         role="button"
//       >
//         Submit
//       </button>
//     </div>
//   );
// };

// Custom submit button template to add proper role
const CustomButtonTemplate = (props) => {
  // const { uiSchema, onSubmit } = props; // onSubmit is not needed here
  return (
    <div className="d-flex justify-content-center">
      <button
        type="submit"
        className="btn btn-primary" // Use the custom button style from Login.css
        role="button"
      >
        Create Account
      </button>
    </div>
  );
};

// RegistrationForm Component
const RegistrationForm = () => {
  // const [validationError, setValidationError] = useState(null);

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
        alert(`Registration failed: ${regErrorText}`);
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
        alert("Registration successful, but auto-login failed. Please try logging in manually.");
        window.location.href = "/login"; // Redirect to login page
      }
      
    } catch (error) {
      console.error("Error during registration and auto-login:", error);
      alert("An error occurred during the registration process. Please try again.");
    }
  };

  // Handle form validation errors
  const onError = (errors) => {
    // setValidationError("is a required property");
    console.log("Form validation errors:", errors);
  };

  return (
    <div className="login-container" style={{ backgroundImage: `url("../assets/FrontRunner Login Background.png")`, backgroundSize: "cover", backgroundPosition: "center"}}>
      <div className='login-card'>
        <h2 className="text-center mb-4">Create an account</h2>
        <Form
          schema={schema}
          uiSchema={uiSchema}
          validator={validator}
          onSubmit={onSubmit}
          onError={onError}
          templates={{
            FieldTemplate: CustomFieldTemplate,
            ButtonTemplates: { SubmitButton: CustomButtonTemplate } 
          }}
          showErrorList={false} // Hide the default top-level error list if desired
        />
        {/* validationError && <div className="error-message">{validationError}</div> */}
        <div className="text-center mt-3">
          <a href='/login'>
            Already have an account? Login here.
          </a>
        </div>
      </div>
    </div>
  );
};

export default RegistrationForm;