import React, { useState, useEffect } from 'react';
import './StorefrontLinkForm.css'; // We'll need to create this CSS file

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
    const [storeType, setStoreType] = useState(storefront?.storeType || SUPPORTED_STOREFRONTS[0]?.value || ''); // Default to first option or empty
    const [storeName, setStoreName] = useState(storefront?.storeName || ''); // User-defined name for the link
    const [apiKey, setApiKey] = useState(''); // Sensitive - will be sent to backend
    const [apiSecret, setApiSecret] = useState(''); // Sensitive - will be sent to backend
    const [storeId, setStoreId] = useState(storefront?.storeId || ''); // Platform specific ID (e.g., Amazon Seller ID)
    const [storeUrl, setStoreUrl] = useState(storefront?.storeUrl || ''); // Optional URL

    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');

    // --- TODO: Dynamic Fields (Optional Advanced Feature) ---
    // In a real-world scenario, different store types require different fields.
    // You might dynamically render fields based on the selected 'storeType'.
    // For now, we'll include a few common ones. Adjust as needed per platform.

    // --- Handle Form Submission ---
    const handleSubmit = async (event) => {
        event.preventDefault();
        setIsLoading(true);
        setError('');

        // Basic validation
        if (!storeType) {
             setError('Please select a storefront type.');
             setIsLoading(false);
             return;
        }
        // Add more validation as needed (e.g., check required fields based on storeType)

        // Prepare data payload for the backend
        // IMPORTANT: Only send necessary fields. Sensitive data like keys/secrets
        // should be handled securely by the backend (e.g., encryption).
        const payload = {
            storeType,
            storeName: storeName || `${storeType} Link`, // Default name if empty
            // --- Include credentials ONLY for the ADD request ---
            // For UPDATE, you might handle credential changes differently or not allow updates via form
            apiKey: isEditing ? undefined : apiKey, // Only send on add, or if specifically changed
            apiSecret: isEditing ? undefined : apiSecret, // Only send on add, or if specifically changed
            storeId,
            storeUrl,
            // If editing, include the ID of the storefront being edited
            id: isEditing ? storefront.id : undefined,
        };

        // Determine API endpoint and method
        const apiUrl = isEditing ? `/api/update_storefront?id=${storefront.id}` : '/api/add_storefront';
        const apiMethod = isEditing ? 'PUT' : 'POST';

        try {
            const response = await fetch(apiUrl, {
                method: apiMethod,
                headers: {
                    'Content-Type': 'application/json',
                    // Add authorization headers if needed (e.g., JWT token)
                    // 'Authorization': `Bearer ${your_token}`
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                 const errorData = await response.text(); // Get error details from backend
                 throw new Error(errorData || `Failed to ${isEditing ? 'update' : 'add'} storefront`);
            }

            // Success!
            onSubmitSuccess(); // Call the success handler passed from parent

        } catch (err) {
            console.error("Form submission error:", err);
            setError(err.message || `An error occurred.`);
        } finally {
            setIsLoading(false);
        }
    };

    // --- Render Modal Form ---
    return (
        <div className="modal-backdrop"> {/* Style this to cover the background */}
            <div className="modal-content storefront-form"> {/* Style the modal container */}
                <h2>{formTitle}</h2>
                <button className="modal-close-button" onClick={onClose}>X</button>

                <form onSubmit={handleSubmit}>
                    {error && <p className="form-error">{error}</p>}

                    {/* Store Type Dropdown */}
                    <div className="form-group">
                        <label htmlFor="storeType">Storefront Type *</label>
                        <select
                            id="storeType"
                            value={storeType}
                            onChange={(e) => setStoreType(e.target.value)}
                            required
                            disabled={isEditing} // Usually can't change type after creation
                        >
                            <option value="" disabled>Select a type...</option>
                            {SUPPORTED_STOREFRONTS.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                    </div>

                     {/* Store Name (Optional Nickname) */}
                     <div className="form-group">
                        <label htmlFor="storeName">Link Name (Optional)</label>
                        <input
                            type="text"
                            id="storeName"
                            value={storeName}
                            onChange={(e) => setStoreName(e.target.value)}
                            placeholder="e.g., My Primary Amazon Store"
                        />
                    </div>

                    {/* --- Fields specific to ADDING (Credentials) --- */}
                    {!isEditing && (
                         <>
                            <div className="form-group">
                                <label htmlFor="apiKey">API Key *</label>
                                <input
                                    type="password" // Use password type for sensitive fields
                                    id="apiKey"
                                    value={apiKey}
                                    onChange={(e) => setApiKey(e.target.value)}
                                    required
                                    placeholder="Enter API Key from store"
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
                                />
                             </div>
                         </>
                    )}
                     {/* --- Fields common to ADD/EDIT (or specific types) --- */}
                    {/* Example: Store ID (Adjust label based on selected storeType if needed) */}
                    <div className="form-group">
                        <label htmlFor="storeId">Store ID / Seller ID</label>
                        <input
                            type="text"
                            id="storeId"
                            value={storeId}
                            onChange={(e) => setStoreId(e.target.value)}
                            placeholder="e.g., A1B2C3D4E5F6G7 (Amazon)"
                        />
                    </div>

                    {/* Example: Store URL (Optional) */}
                     <div className="form-group">
                        <label htmlFor="storeUrl">Store URL (Optional)</label>
                        <input
                            type="url"
                            id="storeUrl"
                            value={storeUrl}
                            onChange={(e) => setStoreUrl(e.target.value)}
                            placeholder="e.g., https://www.amazon.com/yourstore"
                        />
                     </div>


                    {/* Form Actions */}
                    <div className="form-actions">
                        <button type="button" onClick={onClose} disabled={isLoading}>Cancel</button>
                        <button type="submit" disabled={isLoading}>
                            {isLoading ? 'Saving...' : (isEditing ? 'Update Link' : 'Link Storefront')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default StorefrontLinkForm;