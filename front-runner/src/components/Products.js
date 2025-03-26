import React from 'react';  
import './Products.css';
import NavBar from './NavBar';
import ProductForm from './ProductForm';
import { useState, useEffect } from 'react';

const Products = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [products, setProducts] = useState([]);
    const [selectedProduct, setSelectedProduct] = useState(null);
    // const testProduct = {
    //     id: 'test-1',
    //     name: 'Test Product',
    //     description: 'This is a test product description',
    //     price: '$19.99',
    //     image: './images/image-1.png',
    //     count: 10,
    //     tags: '#test,#sample'   
    // };

    // useEffect(() => {
    //     setProducts(prevProducts => [...prevProducts, testProduct]);
    // }, []);

    useEffect(() => {
        // Check URL parameters when component mounts
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('openModal') === 'true') {
            setIsModalOpen(true);
            // Remove the parameter from the URL without refreshing the page
            window.history.replaceState({}, '', '/products');
        }
    }, []);

    const handleAddNewClick = () => {
        setSelectedProduct(null);
        setIsModalOpen(true);
    };

    const handleProductClick = (product) => {
        setSelectedProduct(product);
        setIsModalOpen(true);
    };

    useEffect(() => {
        const fetchProducts = async () => {
            const response = await fetch('/api/products');
            const data = await response.json();
            setProducts(data);
        };
        fetchProducts();
    }, []);

    return (
        <div className='my-products'>
            <NavBar />  
            <div className='my-products-content'>
                <div className='products-header'>
                    <h1>My Products</h1>
                    <div onClick={handleAddNewClick} className="add-new-button" style={{ cursor: 'pointer' }}>
                        <img src='./assets/Add new.svg' alt='add new' className='add-new-icon'/>
                    </div>
                </div>
                
                <div className='products-container'>
                    {products.map((product) => (
                        <div 
                            key={product.id} 
                            className='product-tile' 
                            style={{ backgroundImage: `url(${product.image})` }}
                            onClick={() => handleProductClick(product)}
                        >
                            <div className='product-info'>
                                <h2>{product.name}</h2>
                                <p>{product.description}</p>
                            </div>
                            <img src={product.image} alt='product-image' className='product-image-preview'/>
                        </div>
                    ))}
                    
                    {/* <div className='product-tile' style={{ backgroundImage: `url("../images/image-1.png")` }}>
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
                    </div> */}
                </div>

                {isModalOpen && (
                    <ProductForm 
                        onClose={() => setIsModalOpen(false)} 
                        product={selectedProduct}
                    />
                )}
            </div>
        </div>
    );
};

export default Products;