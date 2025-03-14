// AddProductForm.js
import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';
import './ProductForm.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTimes } from '@fortawesome/free-solid-svg-icons';

// Define the JSON Schema for the Add Product form
const schema = {
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
    image: {
      type: 'string',
      title: 'Product Image',
      format: 'data-url'
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
  required: ['productName', 'description', 'price', 'image'], // Make both fields required
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
  image: {
    "ui:widget": "file", // Use file input for image upload
    "ui:options": {
      accept: "image/*", // Restrict to image files only
    },
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
  }
};

function dataURItoBlob(dataURI) {
  // convert base64/URLEncoded data component to raw binary data held in a string
  var byteString;
  if (dataURI.split(',')[0].indexOf('base64') >= 0)
      byteString = atob(dataURI.split(',')[1]);
  else
      byteString = unescape(dataURI.split(',')[1]);

  // separate out the mime component
  var mimeString = dataURI.split(',')[0].split(':')[1].split(';')[0];

  // write the bytes of the string to a typed array
  var ia = new Uint8Array(byteString.length);
  for (var i = 0; i < byteString.length; i++) {
      ia[i] = byteString.charCodeAt(i);
  }

  return new Blob([ia], {type:mimeString});
}



// Define the add product form's onSubmit handler
const onSubmit = async ({ formData }) => {
  console.log('Product data submitted:', formData);
  try {
    // Send the product data to your API endpoint.
    var multipart = new FormData();
    for ( var key in formData ) {
      if (key === "image"){
        var filename = formData[key].split(',')[0].split(';')[1].split('=')[1]
        var test = dataURItoBlob(formData[key]);
        multipart.append(key, test, filename);
      } else{
        multipart.append(key, formData[key]);
      }
    }

    const response = await fetch("/api/add_product", {
      method: 'POST',
      body: multipart,
      redirect: 'manual'
    });

    if (response.ok) {
      alert('Product added successfully!');
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
const AddProductForm = ({ onClose }) => {
  return (
    <div className="add-product-container" style={{backgroundColor: "rgba(0,0,0,0.8)"}}>
      <div className='add-product-card'>
        <h2>Add Product</h2>
        <FontAwesomeIcon icon={faTimes} onClick={onClose} style={{position: "absolute", top: "10", right: "10", width: "32px", height:"32px", cursor: "pointer"}}/>
        <Form
          schema={schema}
          uiSchema={uiSchema}
          validator={validator}
          onSubmit={onSubmit} // Handle form submission
        />
      </div>
    </div>
  );
};

export default AddProductForm;