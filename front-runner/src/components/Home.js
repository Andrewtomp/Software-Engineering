import React from 'react';
import NavBar from './NavBar';
import './Home.css';

const Home = () => {
    return (
        <div className='home'>
            <NavBar />
            <div className='home-content'>
                <h1>Home</h1>
                <div className='home-tiles'>
                    <div className='products-tile'>
                        <h2>My Products</h2>
                        <div className='home-products'>
                            <div className='home-product'>
                                <h3>Product 1</h3>
                            </div>
                            <div className='home-product'>
                                <h3>Product 2</h3>
                            </div>
                            <div className='home-product'>
                                <h3>Product 3</h3>
                            </div>
                            {/* <div className='home-product'>
                                <h3>Product 4</h3>
                            </div> */}
                            {/* <div className='home-product'>
                                <h3>Product 5</h3>
                            </div> */}
                            <div className='view-all-products' onClick={() => window.location.href='/products'}>
                                <h3>View all</h3>
                                <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                    <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white"/>
                                </svg>
                            </div>
                        </div>
                    </div>
                    <div className='small-home-tiles'>
                        <div className='small-home-tile'>
                            <h2>My Storefronts</h2>
                        </div>
                        <div className='small-home-tile'>
                            <h2>My Orders</h2>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Home;