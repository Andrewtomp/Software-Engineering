import React, { useState, useEffect } from 'react';
import './Storefronts.css'; // We'll need to create this CSS file
import NavBar from './NavBar';
import StorefrontLinkForm from './StorefrontLinkForm';

const Storefronts = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [storefronts, setStorefronts] = useState([]);
    const [selectedStorefront, setSelectedStorefront] = useState(null);

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
                         <p style={{ /* Placeholder styles */ }}>
                            No storefronts linked yet. <a href="#" onClick={(e) => { e.preventDefault(); handleAddNewClick(); }} style={{ /* Link styles */ }}>Link a storefront to get started</a>
                        </p>
                    ) : (
                        storefronts.map((storefront) => {
                            // Get domain for favicon service
                            // Use storeUrl field which should contain the full URL
                            const domain = getDomain(storefront.storeUrl);
                            // Construct favicon URL using Google's service (adjust size with sz=)
                            // Using domain_url is often more reliable than just domain=
                            const faviconUrl = domain
                                ? `https://www.google.com/s2/favicons?sz=32&domain_url=${encodeURIComponent(storefront.storeUrl)}`
                                : null; // Fallback if URL is invalid or missing

                            return (
                                <div
                                    key={storefront.id}
                                    className='storefront-tile' // Apply styles for layout
                                    onClick={() => handleStorefrontClick(storefront)}
                                    style={{ cursor: 'pointer' }}
                                >
                                    {/* --- ADDED: Favicon Image --- */}
                                    {faviconUrl && (
                                        <img
                                            src={faviconUrl}
                                            alt={`${storefront.storeName || storefront.storeType} favicon`}
                                            className="storefront-favicon" // Class for styling
                                            // Add error handling for broken images
                                            onError={(e) => {
                                                e.target.style.display = 'none'; // Hide if favicon fails to load
                                                // Optionally show a default icon/placeholder
                                            }}
                                        />
                                    )}
                                    {/* --- End Favicon Image --- */}

                                    <div className='storefront-info'>
                                        <h2>{storefront.storeName || storefront.storeType}</h2>
                                        <p>{storefront.storeType}</p>
                                        {/* Optionally display URL if desired */}
                                        {/* {storefront.storeUrl && <a href={storefront.storeUrl.startsWith('http') ? storefront.storeUrl : `http://${storefront.storeUrl}`} target="_blank" rel="noopener noreferrer" className="storefront-link" onClick={(e) => e.stopPropagation()}>Visit Store</a>} */}
                                    </div>
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