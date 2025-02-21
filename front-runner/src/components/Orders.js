import React, {useState} from 'react';  
import './Orders.css';
import { AgGridReact } from 'ag-grid-react';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community'; 

const Orders = () => {
    const [rowData, setRowData] = useState([
        { orderNo: "0001", products: "Product 1", total: "$20.57", orderDate: "2024-02-21", orderStatus: "Shipped", customer: "Customer 1", shippingLabel: "Label 0001", action: "View" },
        { orderNo: "0002", products: "Product 2", total: "$10.99", orderDate: "2024-02-20", orderStatus: "Processing", customer: "Customer 2", shippingLabel: "Label 0002", action: "View" },
        { orderNo: "0003", products: "Product 3", total: "$31.50", orderDate: "2024-02-19", orderStatus: "Delivered", customer: "Customer 3", shippingLabel: "Label 0003", action: "View" },
        { orderNo: "0004", products: "Product 4", total: "$55.00", orderDate: "2024-02-18", orderStatus: "Canceled", customer: "Customer 4", shippingLabel: "Label 0004", action: "View" },
        { orderNo: "0005", products: "Product 5", total: "$44.00", orderDate: "2024-02-17", orderStatus: "Shipped", customer: "Customer 5", shippingLabel: "Label 0005", action: "View" },
        { orderNo: "0006", products: "Product 6", total: "$22.00", orderDate: "2024-02-16", orderStatus: "Processing", customer: "Customer 6", shippingLabel: "Label 0006", action: "View" },
        { orderNo: "0007", products: "Product 7", total: "$15.75", orderDate: "2024-02-15", orderStatus: "Pending", customer: "Customer 7", shippingLabel: "Label 0007", action: "View" },
        { orderNo: "0008", products: "Product 8", total: "$29.99", orderDate: "2024-02-14", orderStatus: "Delivered", customer: "Customer 8", shippingLabel: "Label 0008", action: "View" },
        { orderNo: "0009", products: "Product 9", total: "$39.99", orderDate: "2024-02-13", orderStatus: "Shipped", customer: "Customer 9", shippingLabel: "Label 0009", action: "View" },
        { orderNo: "0010", products: "Product 10", total: "$12.50", orderDate: "2024-02-12", orderStatus: "Canceled", customer: "Customer 10", shippingLabel: "Label 0010", action: "View" },
        { orderNo: "0011", products: "Product 11", total: "$28.75", orderDate: "2024-02-11", orderStatus: "Shipped", customer: "Customer 11", shippingLabel: "Label 0011", action: "View" },
        { orderNo: "0012", products: "Product 12", total: "$19.99", orderDate: "2024-02-10", orderStatus: "Processing", customer: "Customer 12", shippingLabel: "Label 0012", action: "View" },
        { orderNo: "0013", products: "Product 13", total: "$65.49", orderDate: "2024-02-09", orderStatus: "Pending", customer: "Customer 13", shippingLabel: "Label 0013", action: "View" },
        { orderNo: "0014", products: "Product 14", total: "$33.33", orderDate: "2024-02-08", orderStatus: "Delivered", customer: "Customer 14", shippingLabel: "Label 0014", action: "View" },
        { orderNo: "0015", products: "Product 15", total: "$77.77", orderDate: "2024-02-07", orderStatus: "Shipped", customer: "Customer 15", shippingLabel: "Label 0015", action: "View" },
        { orderNo: "0016", products: "Product 16", total: "$42.10", orderDate: "2024-02-06", orderStatus: "Processing", customer: "Customer 16", shippingLabel: "Label 0016", action: "View" },
        { orderNo: "0017", products: "Product 17", total: "$88.99", orderDate: "2024-02-05", orderStatus: "Canceled", customer: "Customer 17", shippingLabel: "Label 0017", action: "View" },
        { orderNo: "0018", products: "Product 18", total: "$18.49", orderDate: "2024-02-04", orderStatus: "Delivered", customer: "Customer 18", shippingLabel: "Label 0018", action: "View" },
        { orderNo: "0019", products: "Product 19", total: "$55.55", orderDate: "2024-02-03", orderStatus: "Pending", customer: "Customer 19", shippingLabel: "Label 0019", action: "View" },
        { orderNo: "0020", products: "Product 20", total: "$99.99", orderDate: "2024-02-02", orderStatus: "Shipped", customer: "Customer 20", shippingLabel: "Label 0020", action: "View" },
    ]);
    
    const [colDefs, setColDefs] = useState([
        { field: "orderNo", headerName: "Order No." },
        { field: "products", headerName: "Product(s)" },
        { field: "total", headerName: "Total" },
        { field: "orderDate", headerName: "Order Date" },
        { field: "orderStatus", headerName: "Order Status" },
        { field: "customer", headerName: "Customer" },
        { field: "shippingLabel", headerName: "Shipping Label" },
        { field: "action", headerName: "Action" }
    ]);
    
    return (
        <div className='my-orders'>
            <div className='my-orders-content'>
                <h1>My Orders</h1>
        
                <div className='order-tiles-container'>
                    <div className='order-info-tile'>
                        <h2>My Orders</h2>
                        <div className='view-all-products' onClick={() => window.location.href='/orders'}>
                            <p>Export</p>
                            <svg width="26" height="24" viewBox="0 0 26 24" fill="none" xmlns="http://www.w3.org/2000/svg" className='arrow-right'>
                                <path d="M25.0607 13.0607C25.6464 12.4749 25.6464 11.5251 25.0607 10.9393L15.5147 1.3934C14.9289 0.807612 13.9792 0.807612 13.3934 1.3934C12.8076 1.97918 12.8076 2.92893 13.3934 3.51472L21.8787 12L13.3934 20.4853C12.8076 21.0711 12.8076 22.0208 13.3934 22.6066C13.9792 23.1924 14.9289 23.1924 15.5147 22.6066L25.0607 13.0607ZM0 13.5H24V10.5H0L0 13.5Z" fill="white"/>
                            </svg>
                        </div>
                        <div className="orders-ag-theme">
                            <AgGridReact
                                rowData={rowData}
                                columnDefs={colDefs}
                                defaultColDef={{
                                    flex: 1,
                                    resizable: true,
                                }}
                                
                            />
                        </div>
                    </div>

                    <div className='small-order-tile'>

                    </div>

                    <div className='small-order-tile'>

                    </div>
                </div>
            </div>
        </div>
    );
};

export default Orders;