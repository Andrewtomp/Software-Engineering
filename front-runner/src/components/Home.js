import React, { useState, useEffect } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community';
import NavBar from './NavBar';
import ProductForm from './ProductForm';
import StorefrontLinkForm from './StorefrontLinkForm';
import './Home.css';
import { faAmazon, faEtsy, faPinterest } from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const storeIcons = {
    amazon: faAmazon,
    etsy: faEtsy,
    pinterest: faPinterest,
};

ModuleRegistry.registerModules([AllCommunityModule]);

const Home = () => {
    // --- Product State ---
    const [products, setProducts] = useState([]);
    const [isProductModalOpen, setIsProductModalOpen] = useState(false);
    const [selectedProduct, setSelectedProduct] = useState(null);

    // --- Storefront State ---
    const [storefronts, setStorefronts] = useState([]);
    const [isStorefrontModalOpen, setIsStorefrontModalOpen] = useState(false);
    const [selectedStorefront, setSelectedStorefront] = useState(null);

    // --- Orders ---
    const [rowData, setRowData] = useState([
        { order: "0001", product: "Product 1", quantity: 2, total: "$20.57", customer: "Customer 1" },
        { order: "0002", product: "Product 2", quantity: 1, total: "$10.99", customer: "Customer 2" },
        { order: "0003", product: "Product 3", quantity: 3, total: "$31.50", customer: "Customer 3" },
    ]);
    const [colDefs] = useState([
        { field: "order" },
        { field: "product" },
        { field: "quantity" },
        { field: "total" },
        { field: "customer" }
    ]);

    // --- Fetch Products ---
    useEffect(() => {
        const fetchProducts = async () => {
            try {
                const response = await fetch('/api/get_products');
                const data = await response.json();
                setProducts(data.slice(0, 3));
            } catch (error) {
                console.error('Error fetching products:', error);
            }
        };
        fetchProducts();
    }, []);

    // --- Fetch Storefronts ---
    useEffect(() => {
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
                setStorefronts(data.slice(0, 3));
            } catch (error) {
                console.error('Error fetching storefronts:', error);
                setStorefronts([]);
            }
        };
        fetchStorefronts();
    }, []);

    // --- Handlers ---
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
                            <div className='view-all-products' onClick={() => window.location.href = '/products'}>
                                <p>View all</p>
                                <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                    <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white" />
                                </svg>
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
                                    Nothing to see here yet. <span
                                        onClick={handleAddNewProductClick}
                                        style={{
                                            textDecoration: 'underline',
                                            background: 'linear-gradient(to right, #FF4949, #FF8000)',
                                            backgroundClip: 'text',
                                            WebkitBackgroundClip: 'text',
                                            WebkitTextFillColor: 'transparent',
                                            borderBottom: '1px solid #FF4949',
                                            cursor: 'pointer'
                                        }}
                                    >
                                        Add a product to get started
                                    </span>
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
                                            alt='product-preview'
                                            className='product-image-preview'
                                        />
                                    </div>
                                ))
                            )}
                        </div>
                    </div>

                    {/* --- ORDERS + STOREFRONTS --- */}
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
                            <div className='orders-grid ag-theme-alpine'>
                                <AgGridReact
                                    rowData={rowData}
                                    columnDefs={colDefs}
                                    domLayout='autoHeight'
                                />
                            </div>
                        </div>

                        {/* --- STOREFRONTS --- */}
                        <div className='small-home-tile storefronts-small-home-tile'>
                            <div className='tile-header'>
                                <h2>Storefronts</h2>
                                <div className="add-new-button" onClick={handleAddNewStorefrontClick}>
                                    <img src={'../assets/Add new.svg'} alt='add new' className='add-new-icon' />
                                </div>
                            </div>
                            <div className='storefronts-container'>
                                {storefronts.length === 0 ? (
                                    <p style={{ fontStyle: 'italic', color: 'gray', textAlign: 'center' }}>
                                        No storefronts connected yet.
                                    </p>
                                ) : (
                                    storefronts.map((storefront, index) => (
                                        <div key={index} className='storefront-tile'>
                                            <div className='storefront-icon'>
                                                <FontAwesomeIcon icon={storeIcons[storefront.platform]} size="2x" />
                                            </div>
                                            <div className='storefront-info'>
                                                <h3>{storefront.platform}</h3>
                                                <p>{storefront.name}</p>
                                            </div>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* --- Modals --- */}
            {isProductModalOpen && (
                <ProductForm
                    onClose={() => setIsProductModalOpen(false)}
                    product={selectedProduct}
                />
            )}
            {isStorefrontModalOpen && (
                <StorefrontLinkForm
                    onClose={() => setIsStorefrontModalOpen(false)}
                    storefront={selectedStorefront}
                    onSubmitSuccess={handleStorefrontSubmitSuccess}
                />
            )}
        </div>
    );
};

export default Home;
