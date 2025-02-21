import React, { useState } from "react";
import "./Orders.css";
import { AgGridReact } from "ag-grid-react";
import { OrderInfoPopup } from "./OrderInfoPopup";

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
            <div className="my-orders-content">
                <h1>My Orders</h1>

                <div className="order-tiles-container">
                    <div className="order-info-tile">
                        <h2>My Orders</h2>
                        <div className="orders-ag-theme">
                            <AgGridReact rowData={rowData} columnDefs={colDefs} defaultColDef={{ flex: 1, resizable: true, cellStyle: { alignItems: "center", textAlign: "center" } }} />
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
