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
    const isEditing = storefront !== null;
    const formTitle = isEditing ? 'Edit Storefront Link' : 'Link New Storefront';

    const [storeType, setStoreType] = useState(storefront?.storeType || SUPPORTED_STOREFRONTS[0]?.value || '');
    const [storeName, setStoreName] = useState(storefront?.storeName || '');
    const [apiKey, setApiKey] = useState('');
    const [apiSecret, setApiSecret] = useState('');
    const [storeId, setStoreId] = useState(storefront?.storeId || '');
    const [storeUrl, setStoreUrl] = useState(storefront?.storeUrl || '');

    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        if (isEditing) {
            setApiKey('');
            setApiSecret('');
        }
        setStoreType(storefront?.storeType || SUPPORTED_STOREFRONTS[0]?.value || '');
        setStoreName(storefront?.storeName || '');
        setStoreId(storefront?.storeId || '');
        setStoreUrl(storefront?.storeUrl || '');
        setError('');
    }, [storefront, isEditing]);

    const handleSubmit = async (event) => {
        event.preventDefault();
        setIsLoading(true);
        setError('');

        if (!storeType) {
            setError('Please select a storefront type.');
            setIsLoading(false);
            return;
        }
        if (!isEditing && (apiKey === '' || apiSecret === '')) {
            setError('API Key and Secret are required when linking a new storefront.');
            setIsLoading(false);
            return;
        }

        let payload = {};
        let apiUrl = '';
        let apiMethod = '';

        if (isEditing) {
            apiMethod = 'PUT';
            apiUrl = `/api/update_storefront?id=${storefront.id}`;
            payload = {
                storeName: storeName || `${storefront.storeType} Link`,
                storeId: storeId,
                storeUrl: storeUrl,
            };
        } else {
            apiMethod = 'POST';
            apiUrl = '/api/add_storefront';
            payload = {
                storeType,
                storeName: storeName || `${storeType} Link`,
                apiKey,
                apiSecret,
                storeId,
                storeUrl,
            };
        }

        try {
            const response = await fetch(apiUrl, {
                method: apiMethod,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `Failed to ${isEditing ? 'update' : 'add'} storefront. Status: ${response.status}`);
            }
            onSubmitSuccess(); // Close modal and trigger refresh via parent

        } catch (err) {
            console.error("Form submission error:", err);
            setError(err.message || 'An unexpected error occurred.');
        } finally {
            setIsLoading(false);
        }
    };

    // --- ADDED: Delete Handler ---
    const handleDelete = async () => {
        // Double-check we are in edit mode and have an ID
        if (!isEditing || !storefront || !storefront.id) {
            setError("Cannot delete - storefront data missing.");
            return;
        }

        // Confirmation dialog
        const confirmDelete = window.confirm(
            `Are you sure you want to unlink the storefront "${storefront.storeName || storefront.storeType}"? This action cannot be undone.`
        );

        if (!confirmDelete) {
            return; // User cancelled
        }

        setIsLoading(true);
        setError('');

        try {
            const apiUrl = `/api/delete_storefront?id=${storefront.id}`;
            const response = await fetch(apiUrl, {
                method: 'DELETE',
                headers: {
                    // Include auth headers if needed
                    // 'Authorization': `Bearer ${token}`
                },
            });

            if (!response.ok) {
                 const errorText = await response.text();
                 throw new Error(errorText || `Failed to delete storefront. Status: ${response.status}`);
            }

            // Success! Call the same success handler as add/update
            // This will close the modal and trigger the refresh in Storefronts.js
            onSubmitSuccess();

        } catch (err) {
             console.error("Delete error:", err);
             setError(err.message || 'An unexpected error occurred during deletion.');
        } finally {
            setIsLoading(false);
        }
    };
    // --- End Delete Handler ---

    return (
        <div className="modal-backdrop">
            <div className="modal-content storefront-form">
                <h2>{formTitle}</h2>
                <button className="modal-close-button" onClick={onClose} aria-label="Close">X</button>

                <form onSubmit={handleSubmit}>
                    {error && <p className="form-error">{error}</p>}

                    {/* --- Form Groups remain the same --- */}
                    {/* Store Type Dropdown - Disabled when editing */}
                    <div className="form-group">
                        <label htmlFor="storeType">Storefront Type *</label>
                        <select id="storeType" value={storeType} onChange={(e) => setStoreType(e.target.value)} required disabled={isEditing}>
                            <option value="" disabled={storeType !== ''}>Select a type...</option>
                            {SUPPORTED_STOREFRONTS.map(option => (<option key={option.value} value={option.value}>{option.label}</option>))}
                        </select>
                        {isEditing && <p className="form-note">Store type cannot be changed after linking.</p>}
                    </div>
                     {/* Store Name Input */}
                     <div className="form-group">
                        <label htmlFor="storeName">Link Name</label>
                        <input type="text" id="storeName" value={storeName} onChange={(e) => setStoreName(e.target.value)} placeholder="e.g., My Primary Amazon Store"/>
                        <p className="form-note">Give your link a nickname (optional).</p>
                    </div>
                    {/* Credentials Fields - ONLY visible when ADDING */}
                    {!isEditing && (
                         <>
                            <div className="form-group">
                                <label htmlFor="apiKey">API Key *</label>
                                <input type="password" id="apiKey" value={apiKey} onChange={(e) => setApiKey(e.target.value)} required placeholder="Enter API Key from store" autoComplete="new-password"/>
                            </div>
                             <div className="form-group">
                                <label htmlFor="apiSecret">API Secret / Token *</label>
                                <input type="password" id="apiSecret" value={apiSecret} onChange={(e) => setApiSecret(e.target.value)} required placeholder="Enter API Secret or Token" autoComplete="new-password"/>
                                <p className="form-note">Credentials are required for linking and are stored securely. They cannot be updated later; re-link if they change.</p>
                             </div>
                         </>
                    )}
                     {/* Store ID Input */}
                    <div className="form-group">
                        <label htmlFor="storeId">Store ID / Seller ID</label>
                        <input type="text" id="storeId" value={storeId} onChange={(e) => setStoreId(e.target.value)} placeholder="Platform-specific ID (e.g., Amazon Seller ID)"/>
                    </div>
                    {/* Store URL Input */}
                     <div className="form-group">
                        <label htmlFor="storeUrl">Store URL</label>
                        <input type="url" id="storeUrl" value={storeUrl} onChange={(e) => setStoreUrl(e.target.value)} placeholder="e.g., https://www.amazon.com/yourstore (optional)"/>
                     </div>
                    {/* --- End Form Groups --- */}

                    {/* --- MODIFIED: Form Actions --- */}
                    <div className="form-actions">
                         {/* ADDED: Delete button, only shown when editing */}
                         {isEditing && (
                            <button
                                type="button"
                                className="delete-button" // Add class for styling
                                onClick={handleDelete}
                                disabled={isLoading} // Disable while any action is loading
                            >
                                {isLoading ? 'Deleting...' : 'Delete Link'}
                            </button>
                        )}
                        {/* Existing Cancel and Submit/Update buttons */}
                        <button type="button" onClick={onClose} disabled={isLoading}>Cancel</button>
                        <button type="submit" disabled={isLoading}>
                            {isLoading ? 'Saving...' : (isEditing ? 'Update Link' : 'Link Storefront')}
                        </button>
                    </div>
                    {/* --- End Form Actions --- */}
                </form>
            </div>
        </div>
    );
};

export default StorefrontLinkForm;