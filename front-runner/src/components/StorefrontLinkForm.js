import React, { useState, useEffect } from 'react';
import Form from '@rjsf/core';
import './StorefrontLinkForm.css';
import validator from '@rjsf/validator-ajv8';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTimes, faTrash } from '@fortawesome/free-solid-svg-icons';

const SUPPORTED_STOREFRONTS = [
    { value: 'amazon', label: 'Amazon Seller Central' },
    { value: 'pinterest', label: 'Pinterest Business' },
    { value: 'etsy', label: 'Etsy' },
];

const getSchema = (isEditing) => ({
    title: '',
    type: 'object',
    required: isEditing ? [] : ['storeType', 'apiKey', 'apiSecret'],
    properties: {
        ...(isEditing ? {} : {
            storeType: {
                type: 'string',
                title: 'Storefront Type',
                enum: SUPPORTED_STOREFRONTS.map((s) => s.value),
                enumNames: SUPPORTED_STOREFRONTS.map((s) => s.label),
            },
        }),
        storeName: {
            type: 'string',
            title: 'Link Name',
        },
        ...(isEditing ? {} : {
            apiKey: {
                type: 'string',
                title: 'API Key',
            },
            apiSecret: {
                type: 'string',
                title: 'API Secret / Token',
            },
        }),
        storeId: {
            type: 'string',
            title: 'Store ID / Seller ID',
        },
        storeUrl: {
            type: 'string',
            format: 'uri',
            title: 'Store URL',
        },
    },
});

const uiSchema = {
    storeType: {
        'ui:disabled': true,
    },
    apiKey: {
        'ui:widget': 'password',
        'ui:options': { inputType: 'password' },
    },
    apiSecret: {
        'ui:widget': 'password',
        'ui:options': { inputType: 'password' },
    },
    storeId: {
        'ui:placeholder': 'Platform-specific ID (e.g., Amazon Seller ID)',
    },
    storeUrl: {
        'ui:placeholder': 'e.g., https://www.amazon.com/yourstore',
    },
    storeName: {
        'ui:placeholder': 'e.g., My Primary Amazon Store',
    },
};

const StorefrontLinkFormRJSF = ({ storefront, onClose, onSubmitSuccess }) => {
    const isEditing = storefront !== null;
    const [formData, setFormData] = useState({});
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        if (storefront) {
            setFormData({
                storeType: storefront.storeType,
                storeName: storefront.storeName,
                storeId: storefront.storeId,
                storeUrl: storefront.storeUrl,
            });
        } else {
            setFormData({ storeType: SUPPORTED_STOREFRONTS[0]?.value });
        }
    }, [storefront]);

    const handleSubmit = async ({ formData }) => {
        setError('');
        setIsLoading(true);

        try {
            let payload = {};
            let url = '';
            let method = '';

            if (isEditing) {
                method = 'PUT';
                url = `/api/update_storefront?id=${storefront.id}`;
                payload = {
                    storeName: formData.storeName || `${storefront.storeType} Link`,
                    storeId: formData.storeId,
                    storeUrl: formData.storeUrl,
                };
            } else {
                method = 'POST';
                url = '/api/add_storefront';
                payload = {
                    ...formData,
                    storeName: formData.storeName || `${formData.storeType} Link`,
                };
            }

            const response = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText || 'Error occurred.');
            }

            onSubmitSuccess();
        } catch (err) {
            console.error(err);
            setError(err.message || 'Unexpected error.');
        } finally {
            setIsLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!storefront?.id) return;

        if (!window.confirm(`Are you sure you want to delete "${storefront.storeName || storefront.storeType}"?`)) return;

        setIsLoading(true);
        setError('');

        try {
            const res = await fetch(`/api/delete_storefront?id=${storefront.id}`, {
                method: 'DELETE',
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed to delete storefront.');
            }

            onSubmitSuccess();
        } catch (err) {
            console.error(err);
            setError(err.message || 'Error deleting storefront.');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="add-storefront-container" style={{ backgroundColor: "rgba(0,0,0,0.8)" }}>
            <div className='add-storefront-card'>
                <div className='storefront-form-header'>
                    <h2>
                        {isEditing ?
                            `Edit ${SUPPORTED_STOREFRONTS.find(s => s.value === formData.storeType)?.label || formData.storeType} Link`
                            :
                            'Link A New Storefront'
                        }
                    </h2>
                    {isEditing && (
                        <FontAwesomeIcon
                            icon={faTrash}
                            onClick={handleDelete}
                            className='delete-icon'
                            data-testid="delete-icon"
                        />
                    )}
                </div>
                <FontAwesomeIcon
                    icon={faTimes}
                    onClick={onClose}
                    style={{ position: "absolute", top: "10", right: "10", width: "32px", height: "32px", cursor: "pointer" }}
                />
                {error && <p className="form-error">{error}</p>}

                <Form
                    schema={getSchema(isEditing)}
                    uiSchema={{
                        ...uiSchema,
                        storeType: {
                            ...uiSchema.storeType,
                            'ui:disabled': isEditing,
                            'classNames': isEditing ? 'storetype-disabled' : '',
                        },
                    }}
                    formData={formData}
                    validator={validator}
                    onChange={(e) => setFormData(e.formData)}
                    onSubmit={handleSubmit}
                />

            </div>
        </div>
    );
};

export default StorefrontLinkFormRJSF;
