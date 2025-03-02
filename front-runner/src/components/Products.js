import React from 'react';  
import './Products.css';
import NavBar from './NavBar';
import ProductForm from './ProductForm';
import { useState } from 'react';

const Products = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);

    const handleAddNewClick = () => {
        setIsModalOpen(true);
    };
    return (
        <div className='my-products'>
                        <NavBar />  
            <div className='my-products-content'>
                <div className='products-header'>
                    <h1>My Products</h1>
                    <div onClick={handleAddNewClick} className="nav-option">
                        <img src='../assets/Add new.svg' alt='add new' className='add-new-icon'/>
                    </div>
                </div>
                
                <div className='products-container' >
                    <div className='product-tile' style={{ backgroundImage: `url("../images/image-1.png")` }}>
                        <div className='product-info'>
                            <h2>Product 1</h2>
                            <p>Product 1 is a great product that people should buy</p>
                        </div>
                        <img src='../images/image-1.png' alt='product-image' className='product-image-preview'/>
                    </div>
                    <div className='product-tile' style={{ backgroundImage: `url("../images/image-2.png")` }}>
                        <div className='product-info'>
                            <h2>Product 2</h2>
                            <p>Product 2 is a great product that people should buy</p>
                        </div>
                        <img src='../images/image-2.png' alt='product-image' className='product-image-preview'/>
                    </div>
                    <div className='product-tile' style={{ backgroundImage: `url("../images/image-3.png")` }}>
                        <div className='product-info'>
                            <h2>Product 3</h2>
                            <p>Product 3 is a great product that people should buy</p>
                        </div>
                        <img src='../images/image-3.png' alt='product-image' className='product-image-preview'/>
                    </div>
                    <div className='product-tile' style={{ backgroundImage: `url("../images/image.png")` }}>
                        <div className='product-info'>
                            <h2>Product 4</h2>
                            <p>Product 4 is a great product that people should buy</p>
                        </div>
                        <img src='../images/image.png' alt='product-image' className='product-image-preview'/>
                    </div>
                    <div className='product-tile' style={{ backgroundImage: `url("../images/image-1.png")` }}>
                        <div className='product-info'>
                            <h2>Product 5</h2>
                            <p>Product 5 is a great product that people should buy</p>
                        </div>
                        <img src='../images/image-1.png' alt='product-image' className='product-image-preview'/>
                    </div>
                </div>

                {isModalOpen && <ProductForm onClose={() => setIsModalOpen(false)} />}

            </div>
        </div>
    );
};

export default Products;