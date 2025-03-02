import React, { useState } from "react";
import "./Orders.css";
import { AgGridReact } from "ag-grid-react";
import { OrderInfoPopup } from "./OrderInfoPopup";
import NavBar from "./NavBar";

const Orders = () => {
    // Generate dummy data for orders table
    const [rowData, setRowData] = useState([
        ...Array.from({ length: 20 }, (_, i) => ({
            orderNo: (i + 1).toString().padStart(4, "0"),
            products: `Product ${i + 1}`,
            total: `$${(Math.random() * 100).toFixed(2)}`,
            orderDate: `2024-02-${(21 - i).toString().padStart(2, "0")}`,
            orderStatus: ["Shipped", "Processing", "Delivered", "Pending", "Canceled"][i % 5],
            customer: `Customer ${i + 1}`,
            shippingLabel: `/path/to/shipping_label_${(i + 1).toString().padStart(4, "0")}.png`,
            action: "View"
        }))
    ]);

    const [selectedOrder, setSelectedOrder] = useState(null);

    const handleOpenModal = (order) => {
        setSelectedOrder(order);
    };

    const handleCloseModal = () => {
        setSelectedOrder(null);
    };

    const handleDownloadLabel = (labelUrl) => {
        const link = document.createElement("a");
        link.href = labelUrl;
        link.download = "shipping_label.png";
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    const colDefs = [
        { field: "orderNo", headerName: "Order No." },
        { field: "products", headerName: "Product(s)" },
        { field: "total", headerName: "Total" },
        { field: "orderDate", headerName: "Order Date" },
        { field: "orderStatus", headerName: "Order Status" },
        { field: "customer", headerName: "Customer" },
        {
            field: "shippingLabel",
            headerName: "Shipping Label",
            cellRenderer: (params) => (
                <button onClick={() => handleDownloadLabel(params.value)} className="table-button">
                    Download
                </button>
            )
        },
        {
            field: "action",
            headerName: "Action",
            cellRenderer: (params) => (
                <button onClick={() => handleOpenModal(params.data)} className="table-button">View</button>
            )
        }
    ];

    return (
        <div className="my-orders">
                        <NavBar />  
            <div className="my-orders-content">
                <h1>My Orders</h1>

                <div className="order-tiles-container">
                    <div className="order-info-tile">
                        <div className='export' onClick={() => window.location.href=''}>
                            <p>Export</p>
                            <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='export-arrow'>
                                <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white"/>
                            </svg>
                        </div>
                        <h2>Order Info</h2>
                        <div className="orders-ag-theme">
                            <AgGridReact rowData={rowData} columnDefs={colDefs} defaultColDef={{ flex: 1, resizable: true, cellStyle: { display: "flex", alignItems: "center", textAlign: "center", justifyContent: "center" } }} />
                        </div>
                    </div>
                    <div className='order-metrics-tile'>
                        <h2>Order Metrics</h2>
                    </div>
                </div>
            </div>

            {selectedOrder && <OrderInfoPopup order={selectedOrder} onClose={handleCloseModal} />}
        </div>
    );
};

export default Orders;
