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