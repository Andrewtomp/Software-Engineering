import React, {useState} from 'react';
import { AgGridReact } from 'ag-grid-react';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community'; 
import NavBar from './NavBar';
import './Home.css';

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

    return (
        <div className='home'>
                        <NavBar />  
            <div className='home-content'>
                <h1>Home</h1>
                <div className='home-tiles'>
                    <div className='products-tile'>
                        <div className='tile-header'>
                            <h2>My Products</h2>
                            <div className='view-all-products' onClick={() => window.location.href='/products'}>
                                <p>View all</p>
                                <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                    <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white"/>
                                </svg>
                            </div>
                        </div>
                        <div className='home-products'>
                            <div className='home-product' style={{ backgroundImage: `url("../images/image-1.png")` }}>
                                <div className='product-info'>
                                    <h2>Product 1</h2>
                                    <p>Product 1 is a great product that people should buy</p>
                                </div>
                                <img src='../images/image-1.png' alt='product-image' className='product-image-preview'/>
                            </div>
                            <div className='home-product' style={{ backgroundImage: `url("../images/image-2.png")` }}>
                                <div className='product-info'>
                                    <h2>Product 2</h2>
                                    <p>Product 2 is a great product that people should buy</p>
                                </div>
                                <img src='../images/image-2.png' alt='product-image' className='product-image-preview'/>
                            </div>
                            <div className='home-product' style={{ backgroundImage: `url("../images/image-3.png")` }}>
                                <div className='product-info'>
                                    <h2>Product 3</h2>
                                    <p>Product 3 is a great product that people should buy</p>
                                </div>
                                <img src='../images/image-3.png' alt='product-image' className='product-image-preview'/>
                            </div>
                            
                        </div>
                    </div>
                    <div className='small-home-tiles'>
                        <div className='small-home-tile orders-small-home-tile'>
                            <div className='tile-header'>
                                <h2>My Orders</h2>
                                <div className='view-all-products' onClick={() => window.location.href='/orders'}>
                                    <p>View all</p>
                                    <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                        <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white"/>
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
                        <div className='small-home-tile storefronts-small-home-tile'>
                            <div className='tile-header'>
                                <h2>My Storefronts</h2>
                                <div className='view-all-products' onClick={() => window.location.href='/storefronts'}>
                                    <p>View all</p>
                                    <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                        <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white"/>
                                    </svg>
                                </div>
                            </div>
                            <div className='storefront-tiles'>
                                <div className='home-storefront'>
                                    <h3>Storefront 1</h3>
                                </div>
                                <div className='home-storefront'>
                                    <h3>Storefront 2</h3>
                                </div>
                                <div className='home-storefront'>
                                    <h3>Storefront 2</h3>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Home;