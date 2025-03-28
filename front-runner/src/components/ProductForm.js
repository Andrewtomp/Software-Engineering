// AddProductForm.js
import React from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import 'bootstrap/dist/css/bootstrap.min.css';
import './ProductForm.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTimes, faTrash} from '@fortawesome/free-solid-svg-icons';


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

  const imageSource = value?.startsWith('/api/') ? value : value;

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
            src={imageSource} 
            alt="Product preview" 
            style={{ maxWidth: '200px', maxHeight: '200px'}}
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
    'ui:widget': 'text',
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
    "ui:widget": "text",
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
    var multipart = new FormData();
    
    // Map the form fields to what the backend expects
    if (product) {
      // For update endpoint
      multipart.append('productName', formData.productName);
      multipart.append('product_description', formData.description);
      multipart.append('item_price', formData.price.replace('$', ''));
      multipart.append('stock_amount', formData.count);
      multipart.append('tags', formData.tags || '');

      // Handle image updates for existing products
      if (formData.image) {
        if (formData.image.startsWith('data:')) {
          // New image uploaded
          const blob = dataURItoBlob(formData.image);
          multipart.append('image', blob, 'image.jpg');
        } else if (!formData.image.startsWith('/api/')) {
          // Existing image path
          multipart.append('image', formData.image);
        }
      }
    } else {
      // For add endpoint
      multipart.append('productName', formData.productName);
      multipart.append('description', formData.description);
      multipart.append('price', formData.price.replace('$', ''));
      multipart.append('count', formData.count);
      multipart.append('tags', formData.tags || '');
      
      if (formData.image && formData.image.startsWith('data:')) {
        const blob = dataURItoBlob(formData.image);
        multipart.append('image', blob, 'image.jpg');
      }
    }

    const endpoint = product ? `/api/update_product?id=${product.id || product.prodID}` : "/api/add_product";
    const method = product ? 'PUT' : 'POST';

    const response = await fetch(endpoint, {
      method: method,
      body: multipart,
      redirect: 'manual'
    });

    if (response.ok) {
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

const handleDeleteProduct = async (product) => {
  const isConfirmed = window.confirm(`Are you sure you want to delete ${product.prodName}?`);

  if (!isConfirmed) return;

  try {
    // Perform the delete operation here
    console.log(`Deleting product: ${product.prodName}`);
    const endpoint = `/api/delete_product?id=${product.id || product.prodID}`;
    const response = await fetch(endpoint, {
      method: 'DELETE',
      redirect: 'manual'
    });
    if (response.ok){
      console.log(`Deleted product: ${product.prodName}`);
      window.location.reload();
    }
    else {
      console.log(`Failed to delete product: ${product.prodName}`);
    }
  } catch (error) {
    console.error("Error deleting product:", error);
  }
}

// ProductForm Component
const ProductForm = ({ onClose, product }) => {
  // Create a dummy data URL for existing images to satisfy the format requirement
  const getInitialImageValue = (product) => {
    if (!product) return undefined;
    return `/api/get_product_image?image=${product.image}`;
  };

  const formData = product ? {
    productName: product.name || product.prodName,
    description: product.description || product.prodDesc,
    price: `$${product.price || product.prodPrice}`,
    count: product.count || product.prodCount,
    tags: product.tags || product.prodTags,
    image: getInitialImageValue(product)
  } : undefined;

  const widgets = {
    image: ImageWidget
  };

  const currentSchema = product ? {
    ...schema,
    properties: {
      ...schema.properties,
      image: {
        ...schema.properties.image,
        format: undefined // Remove the data-url format requirement for editing
      }
    }
  } : schema;

  return (
    <div className="add-product-container" style={{backgroundColor: "rgba(0,0,0,0.8)"}}>
      <div className='add-product-card'>
        <div className='product-form-header'>
          <h2>{product ? 'Edit Product' : 'Add Product'}</h2>
          {product && (
            <FontAwesomeIcon 
              icon={faTrash} 
              onClick={() => handleDeleteProduct(product)} 
              className='delete-icon'
            />
          )}
        </div>
        
        <FontAwesomeIcon 
          icon={faTimes} 
          onClick={onClose} 
          style={{position: "absolute", top: "10", right: "10", width: "32px", height:"32px", cursor: "pointer"}}
        />
        <Form
          schema={currentSchema}
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