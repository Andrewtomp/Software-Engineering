// AddProductForm.js
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
      type: "string",
      title: "Price",
      description: "Enter the price in dollars (e.g., $12.99)",
      pattern: "^(\\$)?\\d+(\\.\\d{2})?$" // Optional $ at start, then digits, then a decimal with two digits
    },
    count: {
      type: "integer",
      title: "Count",
      minimum: 0
    },
    tags: {
      type: "string",
      title: "Tags",
      description: "Enter tags starting with '#' separated by commas (e.g., #sale, #new)",
      pattern: "^(#\\w+(,\\s*#\\w+)*)?$" // Optional: validate comma-separated tags that start with #
    },
  },
  required: ['productName', 'description', 'price'], // Make both fields required
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
    "ui:widget": "text",
    "ui:placeholder": "$0.00"
  },
  count: {
    "ui:widget": "updown" // Provides a spinner widget, or you can use "number" for a standard input field.
  },
  tags: {
    "ui:widget": "textarea", // Alternatively, you could create a custom widget for tag entry.
    "ui:placeholder": "#tag1, #tag2, ..."
  },
};

// Define the add product form's onSubmit handler
const onSubmit = async ({ formData }) => {
  console.log('Product data submitted:', formData);
  try {
    // Send the product data to your API endpoint.
    const response = await fetch("/api/add_product", {
      method: 'POST',
      body: new URLSearchParams(formData),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded', // Change to application/json if your API expects JSON.
      },
      credentials: 'include',
      redirect: 'manual'
    });
    
    if (response.ok) {
      // Notify the user.
      alert('Product added successfully!');
      // Reload the page to clear the form.
      window.location.reload();
    } else {
      const errorText = await response.text();
      console.error('Error adding product:', errorText);
      alert('Error adding product: ' + errorText);
    }
  } catch (error) {
    console.error('Error adding product:', error);
    alert('Error adding product: ' + error.message);
  }
};

// AddProductForm Component
const AddProductForm = () => {
  return (
    <div style={{ width: '400px', margin: '0 auto', padding: '20px', backgroundColor: '#f9f9f9' }}>
      <h2>Add Product</h2>
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        onSubmit={onSubmit} // Handle form submission
      />
    </div>
  );
};

export default AddProductForm;
