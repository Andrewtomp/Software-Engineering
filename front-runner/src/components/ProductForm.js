// AddProductForm.js
import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';
import './ProductForm.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTimes } from '@fortawesome/free-solid-svg-icons';

// Custom Image Widget
const ImageWidget = ({ value, onChange, options, schema }) => {
  const handleChange = (event) => {
    const file = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        onChange(e.target.result);
      };
      reader.readAsDataURL(file);
    }
  };

  return (
    <div className="image-widget">
      <input
        type="file"
        accept="image/*"
        onChange={handleChange}
        className="form-control"
      />
      {value && (
        <div className="image-preview">
          <img 
            src={value} 
            alt="Product preview" 
            style={{ maxWidth: '200px', maxHeight: '200px', marginTop: '10px' }}
          />
        </div>
      )}
    </div>
  );
};

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
      pattern: "^(\\$)?\\d+(\\.\\d{2})?$"
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
      pattern: "^(#\\w+(,\\s*#\\w+)*)?$"
    },
  },
  required: ['productName', 'description', 'price', 'image'],
};

// Define the UI Schema for the form fields
const uiSchema = {
  productName: {
    'ui:placeholder': 'Enter the product name',
  },
  description: {
    'ui:widget': 'textarea',
    'ui:placeholder': 'Enter a brief product description',
  },
  image: {
    "ui:widget": "image",
    "ui:options": {
      accept: "image/*",
    },
  },
  price: {
    "ui:widget": "text",
    "ui:placeholder": "$0.00"
  },
  count: {
    "ui:widget": "updown"
  },
  tags: {
    "ui:widget": "textarea",
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

// Define the form's onSubmit handler
const onSubmit = async ({ formData }, product) => {
  console.log('Product data submitted:', formData);
  try {
    // Send the product data to your API endpoint.
    var multipart = new FormData();
    for (var key in formData) {
      if (key === "image") {
        // Only process the image if it's a new upload (starts with data:)
        if (formData[key].startsWith('data:')) {
          var filename = formData[key].split(',')[0].split(';')[1].split('=')[1];
          var test = dataURItoBlob(formData[key]);
          multipart.append(key, test, filename);
        } else {
          // If it's an existing image URL, just send it as is
          multipart.append(key, formData[key]);
        }
      } else {
        multipart.append(key, formData[key]);
      }
    }

    const endpoint = product ? `/api/update_product/${product.id}` : "/api/add_product";
    const method = product ? 'PUT' : 'POST';

    const response = await fetch(endpoint, {
      method: method,
      body: multipart,
      redirect: 'manual'
    });

    if (response.ok) {
      alert(product ? 'Product updated successfully!' : 'Product added successfully!');
      window.location.reload();
    } else {
      const errorText = await response.text();
      console.error('Error with product:', errorText);
      alert(`Error ${product ? 'updating' : 'adding'} product: ${errorText}`);
    }
  } catch (error) {
    console.error('Error with product:', error);
    alert(`Error ${product ? 'updating' : 'adding'} product: ${error.message}`);
  }
  
};

// ProductForm Component
const ProductForm = ({ onClose, product }) => {
  const formData = product ? {
    productName: product.name,
    description: product.description,
    price: product.price,
    count: product.count,
    tags: product.tags,
    image: product.image
  } : undefined;

  const widgets = {
    image: ImageWidget
  };

  return (
    <div className="add-product-container" style={{backgroundColor: "rgba(0,0,0,0.8)"}}>
      <div className='add-product-card'>
        <h2>{product ? 'Edit Product' : 'Add Product'}</h2>
        <FontAwesomeIcon 
          icon={faTimes} 
          onClick={onClose} 
          style={{position: "absolute", top: "10", right: "10", width: "32px", height:"32px", cursor: "pointer"}}
        />
        <Form
          schema={schema}
          uiSchema={uiSchema}
          validator={validator}
          onSubmit={(e) => onSubmit(e, product)}
          formData={formData}
          widgets={widgets}
        />
      </div>
    </div>
  );
};

export default ProductForm;