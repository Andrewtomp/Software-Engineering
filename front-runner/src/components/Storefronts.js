import React, { useState, useEffect } from 'react';
import './Storefronts.css'; // We'll need to create this CSS file
import NavBar from './NavBar';
import StorefrontLinkForm from './StorefrontLinkForm';
import { faAmazon, faEtsy, faPinterest } from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const storeIcons = {
    amazon: faAmazon,
    etsy: faEtsy,
    pinterest: faPinterest,
};

const Storefronts = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [storefronts, setStorefronts] = useState([]);
    const [selectedStorefront, setSelectedStorefront] = useState(null);

    useEffect(() => {
            // Check URL parameters when component mounts
            const urlParams = new URLSearchParams(window.location.search);
            if (urlParams.get('openModal') === 'true') {
                setIsModalOpen(true);
                // Remove the parameter from the URL without refreshing the page
                window.history.replaceState({}, '', '/storefronts');
            }
        }, []);

    // --- Fetch Storefronts ---
    useEffect(() => {
        const fetchStorefronts = async () => {
            try {
                const response = await fetch('/api/get_storefronts');
                if (!response.ok) {
                    if (response.status === 401) {
                        console.error("User not authenticated.");
                        setStorefronts([]);
                        return;
                    }
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                setStorefronts(data || []);
            } catch (error) {
                console.error("Error fetching storefronts:", error);
                setStorefronts([]);
            }
        };
        fetchStorefronts();
    }, [isModalOpen]);

    // --- Handlers ---
    const handleAddNewClick = () => {
        setSelectedStorefront(null);
        setIsModalOpen(true);
    };

    const handleStorefrontClick = (storefront) => {
        console.log("Editing storefront:", storefront);
        setSelectedStorefront(storefront);
        setIsModalOpen(true);
    };

    const handleFormSubmitSuccess = () => {
        setIsModalOpen(false);
    };

    // --- Helper to get domain for favicon ---
    const getDomain = (url) => {
        if (!url) return null;
        try {
            // Prepend http:// if no protocol exists, otherwise URL parsing might fail
            const fullUrl = url.startsWith('http://') || url.startsWith('https://') ? url : `http://${url}`;
            return new URL(fullUrl).hostname;
        } catch (error) {
            console.error("Error parsing URL for domain:", url, error);
            return null; // Return null if URL is invalid
        }
    };

    // --- Render ---
    return (
        <div className='my-storefronts'>
            <NavBar />
            <div className='my-storefronts-content'>
                <div className='storefronts-header'>
                    <h1>My Storefronts</h1>
                    <div onClick={handleAddNewClick} className="add-new-button" style={{ cursor: 'pointer' }}>
                        <img src={'../assets/Add new.svg'} alt='add new storefront' className='add-new-icon'/>
                    </div>
                </div>

                <div className='storefronts-container'>
                    {storefronts.length === 0 ? (
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
                            Nothing to see here yet. <a href="#" onClick={(e) => { e.preventDefault(); handleAddNewClick(); }} style={{
                                textDecoration: 'underline',
                                background: 'linear-gradient(to right, #FF4949, #FF8000)',
                                backgroundClip: 'text',
                                WebkitBackgroundClip: 'text',
                                WebkitTextFillColor: 'transparent',
                                borderBottom: '1px solid #FF4949'
                            }}>Link a storefront to get started</a>
                        </p>
                    ) : (
                        storefronts.map((storefront) => {
                            // // Get domain for favicon service
                            // // Use storeUrl field which should contain the full URL
                            // const domain = getDomain(storefront.storeUrl);
                            // // Construct favicon URL using Google's service (adjust size with sz=)
                            // // Using domain_url is often more reliable than just domain=
                            // const faviconUrl = domain
                            //     ? `https://www.google.com/s2/favicons?sz=32&domain_url=${encodeURIComponent(storefront.storeUrl)}`
                            //     : null; // Fallback if URL is invalid or missing

                            return (
                                <div 
                                    key={storefront.id} 
                                    className='storefront-tile' 
                                    // style={{ backgroundImage: `url(${faviconUrl})` }}
                                    onClick={() => handleStorefrontClick(storefront)}
                                >
                                    <FontAwesomeIcon 
                                        icon={storeIcons[storefront.storeType]} 
                                        fontSize={'64px'}
                                        className="storefront-image"
                                        color="white"
                                    />
                                    <h2>{storefront.storeName || storefront.storeType}</h2>
                                    {/* <img 
                                        src={faviconUrl} 
                                        alt='storefront-image' 
                                        className='storefront-image-preview'
                                    /> */}
                                </div>
                            );
                        })
                    )}
                </div>

                {/* Modal Rendering */}
                {isModalOpen && (
                    <StorefrontLinkForm
                        storefront={selectedStorefront}
                        onClose={() => setIsModalOpen(false)}
                        onSubmitSuccess={handleFormSubmitSuccess}
                    />
                )}
            </div>
        </div>
    );
};

export default Storefronts;