import React, { useState, useEffect } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community';
import NavBar from './NavBar';
import './Home.css';
import ProductForm from './ProductForm';
import StorefrontLinkForm from './StorefrontLinkForm';
import { faAmazon, faEtsy, faPinterest } from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const storeIcons = {
    amazon: faAmazon,
    etsy: faEtsy,
    pinterest: faPinterest,
};

ModuleRegistry.registerModules([AllCommunityModule]);

const Home = () => {
    const [rowData] = useState([
        { order: "0001", product: "Product 1", quantity: 2, total: "$20.57", customer: "Customer 1" },
        { order: "0002", product: "Product 2", quantity: 1, total: "$10.99", customer: "Customer 2" },
        { order: "0003", product: "Product 3", quantity: 3, total: "$31.50", customer: "Customer 3" },
        { order: "0004", product: "Product 4", quantity: 5, total: "$55.00", customer: "Customer 4" },
        { order: "0005", product: "Product 5", quantity: 4, total: "$44.00", customer: "Customer 5" },
    ]);

    const [colDefs] = useState([
        { field: "order" },
        { field: "product" },
        { field: "quantity" },
        { field: "total" },
        { field: "customer" }
    ]);

    const [products, setProducts] = useState([]);
    const [storefronts, setStorefronts] = useState([]);

    const [isProductModalOpen, setIsProductModalOpen] = useState(false);
    const [selectedProduct, setSelectedProduct] = useState(null);

    const [isStorefrontModalOpen, setIsStorefrontModalOpen] = useState(false);
    const [selectedStorefront, setSelectedStorefront] = useState(null);

    // Fetch Products
    const fetchProducts = async () => {
        try {
            const response = await fetch('/api/get_products');
            const data = await response.json();
            setProducts(data.slice(0, 3)); // Show only 3 on home
        } catch (error) {
            console.error('Error fetching products:', error);
        }
    };

    useEffect(() => {
        fetchProducts();
    }, []);

    // Fetch Storefronts
    const fetchStorefronts = async () => {
        try {
            const response = await fetch('/api/get_storefronts');
            if (!response.ok) {
                if (response.status === 401) {
                    console.log("User not authenticated to fetch storefronts.");
                    setStorefronts([]);
                    return;
                }
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            setStorefronts(data ? data.slice(0, 3) : []);
        } catch (error) {
            console.error('Error fetching storefronts:', error);
            setStorefronts([]);
        }
    };

    useEffect(() => {
        fetchStorefronts();
    }, []);

    // Handlers
    const handleAddNewProductClick = () => {
        setSelectedProduct(null);
        setIsProductModalOpen(true);
    };

    const handleProductClick = (product) => {
        const formattedProduct = {
            id: product.prodID,
            name: product.prodName,
            description: product.prodDesc,
            price: product.prodPrice,
            count: product.prodCount,
            tags: product.prodTags,
            image: product.image
        };
        setSelectedProduct(formattedProduct);
        setIsProductModalOpen(true);
    };

    const handleAddNewStorefrontClick = () => {
        setSelectedStorefront(null);
        setIsStorefrontModalOpen(true);
    };

    const handleStorefrontSubmitSuccess = () => {
        setIsStorefrontModalOpen(false);
        fetchStorefronts();
    };

    return (
        <div className='home'>
            <NavBar />
            <div className='home-content'>
                <h1>Home</h1>
                <div className='home-tiles'>
                    {/* --- PRODUCTS --- */}
                    <div className='products-tile'>
                        <div className='tile-header'>
                            <h2>My Products</h2>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                <div className='view-all-products' onClick={() => window.location.href = '/products'}>
                                    <p>View all</p>
                                    <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                        <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white" />
                                    </svg>
                                </div>
                                <img
                                    src={'../assets/Add new.svg'}
                                    alt='add new product'
                                    className='add-new-icon'
                                    onClick={handleAddNewProductClick}
                                    style={{ cursor: 'pointer' }}
                                />
                            </div>
                        </div>
                        <div className='home-products'>
                            {products.length === 0 ? (
                                <p style={{
                                    width: '90%',
                                    position: 'absolute',
                                    textAlign: 'center',
                                    top: '50%',
                                    left: '50%',
                                    transform: 'translate(-50%, -50%)',
                                    fontStyle: 'italic',
                                    color: 'gray'
                                }}>
                                    Nothing to see here yet. <a href='/products?openModal=true' style={{
                                        textDecoration: 'underline',
                                        background: 'linear-gradient(to right, #FF4949, #FF8000)',
                                        backgroundClip: 'text',
                                        WebkitBackgroundClip: 'text',
                                        WebkitTextFillColor: 'transparent',
                                        borderBottom: '1px solid #FF4949'
                                    }}>Add a product to get started</a>
                                </p>
                            ) : (
                                products.map((product) => (
                                    <div
                                        key={product.prodID}
                                        className='product-tile'
                                        style={{ width: "100%", backgroundImage: `url(/api/get_product_image?image=${product.image})` }}
                                        onClick={() => handleProductClick(product)}
                                    >
                                        <div className='product-info'>
                                            <h2>{product.prodName}</h2>
                                            <p>{product.prodDesc}</p>
                                        </div>
                                        <img
                                            src={`/api/get_product_image?image=${product.image}`}
                                            alt='product-image'
                                            className='product-image-preview'
                                        />
                                    </div>
                                ))
                            )}
                        </div>
                    </div>

                    {/* --- SMALL TILES --- */}
                    <div className='small-home-tiles'>
                        {/* --- ORDERS --- */}
                        <div className='small-home-tile orders-small-home-tile'>
                            <div className='tile-header'>
                                <h2>My Orders</h2>
                                <div className='view-all-products' onClick={() => window.location.href = '/orders'}>
                                    <p>View all</p>
                                    <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                        <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white" />
                                    </svg>
                                </div>
                            </div>
                            <div className='ag-theme-alpine orders-table'>
                                <AgGridReact
                                    rowData={rowData}
                                    columnDefs={colDefs}
                                    pagination={true}
                                    paginationPageSize={3}
                                />
                            </div>
                        </div>

                        {/* --- STOREFRONTS --- */}
                        <div className='small-home-tile storefronts-small-home-tile'>
                            <div className='tile-header'>
                                <h2>My Storefronts</h2>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                    <div className='view-all-products' onClick={() => window.location.href = '/storefronts'}>
                                        <p>View all</p>
                                        <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                            <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white" />
                                        </svg>
                                    </div>
                                    <img
                                        src={'../assets/Add new.svg'}
                                        alt='add new storefront'
                                        className='add-new-icon'
                                        onClick={handleAddNewStorefrontClick}
                                        style={{ cursor: 'pointer' }}
                                    />
                                </div>
                            </div>
                            <div className='storefront-icons'>
                                {storefronts.map((storefront, index) => {
                                    const Icon = storeIcons[storefront.platform.toLowerCase()];
                                    return (
                                        <div key={index} className='storefront-icon'>
                                            {Icon ? (
                                                <FontAwesomeIcon icon={Icon} size='2x' />
                                            ) : (
                                                <span>{storefront.platform}</span>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    </div>
                </div>

                {/* MODALS */}
                {isProductModalOpen && (
                    <ProductForm
                        onClose={() => setIsProductModalOpen(false)}
                        product={selectedProduct}
                    />
                )}
                {isStorefrontModalOpen && (
                    <StorefrontLinkForm
                        storefront={selectedStorefront}
                        onClose={() => setIsStorefrontModalOpen(false)}
                        onSubmitSuccess={handleStorefrontSubmitSuccess}
                    />
                )}
            </div>
        </div>
    );
};

export default Home;
