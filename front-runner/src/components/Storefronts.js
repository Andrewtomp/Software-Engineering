// import NavBar from './NavBar';

// const Storefronts = () => {
//     return <div>
//         <NavBar />
//         <p style={{position: "absolute", top: "50%", right: "50%", color: "white"}}>not implemented yet</p>
//     </div>;
// }

// export default Storefronts;
import React, { useState, useEffect } from 'react';
import './Storefronts.css'; // We'll need to create this CSS file
import NavBar from './NavBar';
import StorefrontLinkForm from './StorefrontLinkForm'; // We'll create this form component next

const Storefronts = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [storefronts, setStorefronts] = useState([]);
    // 'selectedStorefront' can be used if you later want to edit existing links
    const [selectedStorefront, setSelectedStorefront] = useState(null);

    // --- Fetch Storefronts ---
    useEffect(() => {
        const fetchStorefronts = async () => {
            try {
                // TODO: Update API endpoint when backend is ready
                const response = await fetch('/api/get_storefronts');
                if (!response.ok) {
                    // Handle non-successful responses (e.g., 401 Unauthorized)
                    if (response.status === 401) {
                        console.error("User not authenticated.");
                        // Optionally redirect to login or show a message
                        setStorefronts([]); // Clear storefronts if unauthorized
                        return;
                    }
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                setStorefronts(data || []); // Ensure data is an array, even if response is empty/null
            } catch (error) {
                console.error("Error fetching storefronts:", error);
                setStorefronts([]); // Set to empty array on error
            }
        };
        fetchStorefronts();
    }, [isModalOpen]); // Re-fetch when the modal closes (in case a new one was added)

    // --- Handlers ---
    const handleAddNewClick = () => {
        setSelectedStorefront(null); // Ensure we are adding, not editing
        setIsModalOpen(true);
    };

    // Placeholder for future 'edit' functionality
    const handleStorefrontClick = (storefront) => {
        console.log("Editing storefront:", storefront);
        // Set the selected storefront data to pass to the modal
        setSelectedStorefront(storefront);
        // Open the modal
        setIsModalOpen(true);
    };

    // Function to be passed to the modal to refresh data after adding/editing
    const handleFormSubmitSuccess = () => {
        setIsModalOpen(false);
        // Optionally trigger a re-fetch immediately, though the useEffect dependency does this
        // fetchStorefronts(); // uncomment if needed, but useEffect covers it
    };


    // --- Render ---
    return (
        <div className='my-storefronts'> {/* Use specific class */}
            <NavBar />
            <div className='my-storefronts-content'>
                <div className='storefronts-header'> {/* Use specific class */}
                    <h1>My Storefronts</h1>
                    <div onClick={handleAddNewClick} className="add-new-button" style={{ cursor: 'pointer' }}>
                        {/* Assuming you have a similar 'Add new' icon */}
                        <img src={'../assets/Add new.svg'} alt='add new storefront' className='add-new-icon'/>
                    </div>
                </div>

                <div className='storefronts-container'> {/* Use specific class */}
                    {storefronts.length === 0 ? (
                         <p style={{ /* Same styles as Products page */
                            width: '90%',
                            position: 'absolute',
                            textAlign: 'center',
                            top: '50%',
                            left: '50%',
                            transform: 'translate(-50%, -50%)',
                            fontStyle: 'italic',
                            color: 'gray' // Adjust color for your theme if needed
                        }}>
                            No storefronts linked yet. <a href="#" onClick={(e) => { e.preventDefault(); handleAddNewClick(); }} style={{ /* Same styles as Products page */
                                textDecoration: 'underline',
                                background: 'linear-gradient(to right, #FF4949, #FF8000)', // Example gradient
                                backgroundClip: 'text',
                                WebkitBackgroundClip: 'text',
                                WebkitTextFillColor: 'transparent',
                                borderBottom: '1px solid #FF4949' // Example border
                            }}>Link a storefront to get started</a>
                        </p>
                    ) : (
                        storefronts.map((storefront) => (
                            // --- Storefront Tile ---
                            // Adjust structure based on what data you get from the backend
                            // For now, just showing the type and name
                            <div
                                key={storefront.id} // Assuming backend provides a unique 'id'
                                className='storefront-tile' // Use specific class
                                onClick={() => handleStorefrontClick(storefront)} // Add click handler
                                // Optional: Add background image/icon based on store type later
                            >
                                <div className='storefront-info'>
                                    {/* Display relevant info. Adjust field names based on your Go struct */}
                                    <h2>{storefront.storeName || storefront.storeType}</h2> {/* Show user-given name or fallback to type */}
                                    <p>{storefront.storeType}</p> {/* Show the type (e.g., "Amazon", "Pinterest") */}
                                    {/* Maybe add store ID or URL if available and desired */}
                                    {/* <p>ID: {storefront.storeId}</p> */}
                                </div>
                                {/* You might want an icon representing the store */}
                                {/* <img src={`path/to/icons/${storefront.storeType}.png`} alt={storefront.storeType} /> */}
                            </div>
                        ))
                    )}
                </div>

                {isModalOpen && (
                    <StorefrontLinkForm
                        // Pass selectedStorefront for potential editing in the future
                        storefront={selectedStorefront}
                        onClose={() => setIsModalOpen(false)}
                        onSubmitSuccess={handleFormSubmitSuccess} // Pass the success handler
                    />
                )}
            </div>
        </div>
    );
};

export default Storefronts;