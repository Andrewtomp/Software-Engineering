import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';

// Define the JSON Schema for the Add Product form
const schema = {
  title: 'Add Product',
  type: 'object',
  properties: {
    productName: {
      type: 'string',
      title: 'Product Name',
    },
    description: {
      type: 'string',
      title: 'Description',
    },
    price: {
      type: 'number',
      title: 'Price',
      minimum: 0, // Ensure price is non-negative
    },
    stock: {
      type: 'integer',
      title: 'Stock Quantity',
      minimum: 0, // Ensure stock is non-negative
    },
    tags: {
      type: 'array',
      title: 'Tags',
      items: {
        type: 'string',
      },
    },
  },
  required: ['productName', 'description', 'price', 'stock'], // Required fields
};

// Define the UI Schema for the form fields (optional customization)
const uiSchema = {
  productName: {
    'ui:placeholder': 'Enter the product name',
  },
  description: {
    'ui:widget': 'textarea', // Allows multiline input
    'ui:placeholder': 'Enter a brief product description',
  },
  price: {
    'ui:widget': 'updown', // Numeric input
    'ui:placeholder': 'Enter product price',
  },
  stock: {
    'ui:widget': 'updown', // Numeric input
    'ui:placeholder': 'Enter available stock',
  },
  tags: {
    'ui:widget': 'textarea', // Allows inputting multiple tags
    'ui:placeholder': 'Enter tags separated by commas',
  },
};

// Custom field template to ensure labels are properly associated with inputs
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

// AddProductForm Component
const AddProductForm = () => {
  // Define the add product form's onSubmit handler
  const onSubmit = ({ formData }) => {
    console.log('Product data submitted:', formData);
    // Use window.alert instead of just alert to ensure it's properly mocked in tests
    window.alert('Product Added: ' + JSON.stringify(formData));
  };
  
  return (
    <div style={{ width: '400px', margin: '0 auto', padding: '20px', backgroundColor: '#f9f9f9' }}>
      <h2>Add Product</h2>
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
    </div>
  );
};

// Setup for tests - mock alert if needed
if (typeof window !== 'undefined' && !window.alert) {
  window.alert = function() {};
}

export default AddProductForm;