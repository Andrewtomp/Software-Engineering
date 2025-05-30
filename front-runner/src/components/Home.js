import React, { useState, useEffect } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community';
import NavBar from './NavBar';
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
    const [rowData, setRowData] = useState([
        { order: "0001", product: "Product 1", quantity: 2, total: "$20.57", customer: "Customer 1" },
        { order: "0002", product: "Product 2", quantity: 1, total: "$10.99", customer: "Customer 2" },
        { order: "0003", product: "Product 3", quantity: 3, total: "$31.50", customer: "Customer 3" },
        { order: "0004", product: "Product 4", quantity: 5, total: "$55.00", customer: "Customer 4" },
        { order: "0005", product: "Product 5", quantity: 4, total: "$44.00", customer: "Customer 5" },
        { order: "0006", product: "Product 6", quantity: 2, total: "$22.00", customer: "Customer 6" },
        { order: "0007", product: "Product 7", quantity: 1, total: "$15.75", customer: "Customer 7" },
        { order: "0007", product: "Product 7", quantity: 1, total: "$15.75", customer: "Customer 7" },
        { order: "0007", product: "Product 7", quantity: 1, total: "$15.75", customer: "Customer 7" },
        { order: "0007", product: "Product 7", quantity: 1, total: "$15.75", customer: "Customer 7" },
        { order: "0007", product: "Product 7", quantity: 1, total: "$15.75", customer: "Customer 7" },
    ]);

    const [colDefs, setColDefs] = useState([
        { field: "order" },
        { field: "product" },
        { field: "quantity" },
        { field: "total" },
        { field: "customer" }
    ]);

    const [products, setProducts] = useState([]);

    useEffect(() => {
        const fetchProducts = async () => {
            try {
                const response = await fetch('/api/get_products');
                const data = await response.json();
                // Take only the first 3 products
                setProducts(data.slice(0, 3));
            } catch (error) {
                console.error('Error fetching products:', error);
            }
        };
        fetchProducts();
    }, []);

    const [storefronts, setStorefronts] = useState([]);

    useEffect(() => {
        const fetchStorefronts = async () => {
            try {
                const response = await fetch('/api/get_storefronts'); // API endpoint
                if (!response.ok) {
                    if (response.status === 401) {
                        console.log("User not authenticated to fetch storefronts.");
                        setStorefronts([]);
                        return;
                    }
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                // Slice to first 3 to match existing CSS layout for now
                setStorefronts(data ? data.slice(0, 3) : []);
            } catch (error) {
                console.error('Error fetching storefronts:', error);
                setStorefronts([]);
            }
        };
        fetchStorefronts();
    }, []);

    return (
        <div className='home'>
            <NavBar />
            <div className='home-content'>
                <h1>Home</h1>
                <div className='home-tiles'>
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
                                        onClick={() => window.location.href = `/products?id=${product.prodID}`}
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
                    <div className='small-home-tiles'>
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
                            <div className="home-ag-theme">
                                <AgGridReact
                                    rowData={rowData}
                                    columnDefs={colDefs}
                                    defaultColDef={{
                                        flex: 1,
                                        resizable: true,
                                        cellStyle: { alignItems: "center", textAlign: "center" },
                                    }}

                                />
                            </div>

                        </div>
                        {/* --- Storefronts Tile (MODIFIED) --- */}
                        <div className='small-home-tile storefronts-small-home-tile'>
                            <div className='tile-header'>
                            <h2>My Storefronts</h2>
                                <div className='view-all-products' onClick={() => window.location.href = '/storefronts'}>
                                    <p>View all</p>
                                    {/* Arrow SVG */}
                                    <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                        <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white" />
                                    </svg>
                                </div>
                            </div>
                            {/* --- Dynamic Storefront Rendering --- */}
                            <div className='storefront-tiles'>
                                {storefronts.length === 0 ? (
                                    // Placeholder when no storefronts are linked
                                    <p style={{ /* Placeholder styles */
                                        width: '90%', position: 'absolute', textAlign: 'center',
                                        top: '50%', left: '50%', transform: 'translate(-50%, -50%)',
                                        fontStyle: 'italic', color: 'gray'
                                    }}>
                                        Nothing to see here yet. <a href='/storefronts?openModal=true' style={{ /* Link styles */
                                            textDecoration: 'underline', background: 'linear-gradient(to right, #FF4949, #FF8000)',
                                            backgroundClip: 'text', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
                                            borderBottom: '1px solid #FF4949'
                                        }}>Link a storefront to get started</a>.
                                    </p>
                                ) : (
                                    // Map over the fetched storefronts
                                    storefronts.map((storefront) => (
                                        <div className='home-storefront' key={storefront.id}
                                        onClick={() => window.location.href = `/storefronts?id=${storefront.id}`}
                                        >
                                            <FontAwesomeIcon
                                                icon={storeIcons[storefront.storeType]}
                                                fontSize={'64px'}
                                                className="storefront-image"
                                                color="white"
                                            />
                                            <h2>{storefront.storeName || storefront.storeType}</h2>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Home;