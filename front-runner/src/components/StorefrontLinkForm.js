import React, { useState, useEffect } from 'react';
import './StorefrontLinkForm.css'; // Ensure CSS supports the form styles

// Define supported storefront types
const SUPPORTED_STOREFRONTS = [
    { value: 'amazon', label: 'Amazon Seller Central' },
    { value: 'pinterest', label: 'Pinterest Business' },
    { value: 'etsy', label: 'Etsy' },
    // Add more as you support them
];

const StorefrontLinkForm = ({ storefront, onClose, onSubmitSuccess }) => {
    // Determine if we are editing or adding new
    const isEditing = storefront !== null;
    const formTitle = isEditing ? 'Edit Storefront Link' : 'Link New Storefront';

    // --- State for Form Fields ---
    // Initialize state based on whether we are editing or adding
    // These initializations correctly handle pre-filling form for editing
    const [storeType, setStoreType] = useState(storefront?.storeType || SUPPORTED_STOREFRONTS[0]?.value || '');
    const [storeName, setStoreName] = useState(storefront?.storeName || '');
    const [apiKey, setApiKey] = useState(''); // Only used for adding
    const [apiSecret, setApiSecret] = useState(''); // Only used for adding
    const [storeId, setStoreId] = useState(storefront?.storeId || '');
    const [storeUrl, setStoreUrl] = useState(storefront?.storeUrl || '');

    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');

    // Effect to reset API key/secret fields if the mode changes (optional cleanup)
    useEffect(() => {
        if (isEditing) {
            setApiKey('');
            setApiSecret('');
        }
         // Reset fields if the passed storefront prop changes (e.g. closing and reopening modal for different item)
         setStoreType(storefront?.storeType || SUPPORTED_STOREFRONTS[0]?.value || '');
         setStoreName(storefront?.storeName || '');
         setStoreId(storefront?.storeId || '');
         setStoreUrl(storefront?.storeUrl || '');
         setError(''); // Clear errors when modal opens/changes mode

    }, [storefront, isEditing]); // Rerun when storefront prop changes


    // --- MODIFIED: Handle Form Submission ---
    const handleSubmit = async (event) => {
        event.preventDefault();
        setIsLoading(true);
        setError('');

        // Basic validation (remains the same)
        if (!storeType) {
             setError('Please select a storefront type.');
             setIsLoading(false);
             return;
        }
        // Add more validation if needed...
        // Example: Require API key/secret only when adding
        if (!isEditing && (apiKey === '' || apiSecret === '')) {
             setError('API Key and Secret are required when linking a new storefront.');
             setIsLoading(false);
             return;
        }


        // --- Prepare Payload based on mode ---
        let payload = {};
        let apiUrl = '';
        let apiMethod = '';

        if (isEditing) {
            // --- EDITING MODE ---
            apiMethod = 'PUT';
            apiUrl = `/api/update_storefront?id=${storefront.id}`;
            payload = {
                // Fields allowed in the StorefrontLinkUpdatePayload (Go backend)
                storeName: storeName || `${storefront.storeType} Link`, // Default name if cleared
                storeId: storeId,
                storeUrl: storeUrl,
                // DO NOT send storeType, apiKey, apiSecret for update
            };
        } else {
            // --- ADDING MODE ---
            apiMethod = 'POST';
            apiUrl = '/api/add_storefront';
            payload = {
                // Fields allowed in the StorefrontLinkAddPayload (Go backend)
                storeType,
                storeName: storeName || `${storeType} Link`, // Default name if empty
                apiKey, // Send credentials only when adding
                apiSecret, // Send credentials only when adding
                storeId,
                storeUrl,
                // Do not send 'id' when adding
            };
        }

        // --- Make API Call ---
        try {
            const response = await fetch(apiUrl, {
                method: apiMethod,
                headers: {
                    'Content-Type': 'application/json',
                    // Include auth headers if your API requires them
                    // 'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                // Try to get error message from backend response body
                 const errorText = await response.text();
                 // Use backend error text if available, otherwise provide a generic message
                 throw new Error(errorText || `Failed to ${isEditing ? 'update' : 'add'} storefront. Status: ${response.status}`);
            }

            // Success! Call the callback passed from parent
            onSubmitSuccess();

        } catch (err) {
            console.error("Form submission error:", err);
             // Display the error message (could be from backend or generic)
            setError(err.message || 'An unexpected error occurred.');
        } finally {
            setIsLoading(false);
        }
    };

    // --- Render Modal Form ---
    return (
        <div className="modal-backdrop">
            <div className="modal-content storefront-form">
                <h2>{formTitle}</h2>
                <button className="modal-close-button" onClick={onClose} aria-label="Close">X</button> {/* Added aria-label */}

                <form onSubmit={handleSubmit}>
                    {/* Display error message if present */}
                    {error && <p className="form-error">{error}</p>}

                    {/* Store Type Dropdown - Disabled when editing */}
                    <div className="form-group">
                        <label htmlFor="storeType">Storefront Type *</label>
                        <select
                            id="storeType"
                            value={storeType}
                            onChange={(e) => setStoreType(e.target.value)}
                            required
                            disabled={isEditing} // Can't change type after creation
                        >
                            <option value="" disabled={storeType !== ''}>Select a type...</option> {/* Ensure placeholder is selectable if initial value is empty */}
                            {SUPPORTED_STOREFRONTS.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                        {isEditing && <p className="form-note">Store type cannot be changed after linking.</p>}
                    </div>

                     {/* Store Name Input */}
                     <div className="form-group">
                        <label htmlFor="storeName">Link Name</label>
                        <input
                            type="text"
                            id="storeName"
                            value={storeName}
                            onChange={(e) => setStoreName(e.target.value)}
                            placeholder="e.g., My Primary Amazon Store"
                        />
                         <p className="form-note">Give your link a nickname (optional).</p>
                    </div>

                    {/* --- Credentials Fields - ONLY visible when ADDING --- */}
                    {!isEditing && (
                         <>
                            <div className="form-group">
                                <label htmlFor="apiKey">API Key *</label>
                                <input
                                    type="password" // Use password type
                                    id="apiKey"
                                    value={apiKey}
                                    onChange={(e) => setApiKey(e.target.value)}
                                    required
                                    placeholder="Enter API Key from store"
                                    autoComplete="new-password" // Prevent browser autofill issues sometimes
                                />
                            </div>
                             <div className="form-group">
                                <label htmlFor="apiSecret">API Secret / Token *</label>
                                <input
                                    type="password" // Use password type
                                    id="apiSecret"
                                    value={apiSecret}
                                    onChange={(e) => setApiSecret(e.target.value)}
                                    required
                                    placeholder="Enter API Secret or Token"
                                    autoComplete="new-password" // Prevent browser autofill issues sometimes
                                />
                                 <p className="form-note">Credentials are required for linking and are stored securely. They cannot be updated later; re-link if they change.</p>
                             </div>
                         </>
                    )}

                     {/* Store ID Input */}
                    <div className="form-group">
                        <label htmlFor="storeId">Store ID / Seller ID</label>
                        <input
                            type="text"
                            id="storeId"
                            value={storeId}
                            onChange={(e) => setStoreId(e.target.value)}
                            placeholder="Platform-specific ID (e.g., Amazon Seller ID)"
                        />
                    </div>

                    {/* Store URL Input */}
                     <div className="form-group">
                        <label htmlFor="storeUrl">Store URL</label>
                        <input
                            type="url"
                            id="storeUrl"
                            value={storeUrl}
                            onChange={(e) => setStoreUrl(e.target.value)}
                            placeholder="e.g., https://www.amazon.com/yourstore (optional)"
                        />
                     </div>


                    {/* Form Actions */}
                    <div className="form-actions">
                        <button type="button" onClick={onClose} disabled={isLoading}>Cancel</button>
                        <button type="submit" disabled={isLoading}>
                            {/* Dynamic button text */}
                            {isLoading ? 'Saving...' : (isEditing ? 'Update Link' : 'Link Storefront')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default StorefrontLinkForm;