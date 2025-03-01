import { BrowserRouter, Routes, Route } from "react-router-dom";
import Home from "./components/Home";
import Products from "./components/Products";
import Storefronts from "./components/Storefronts";
import Orders from "./components/Orders";
import Settings from "./components/Settings";
import NavBar from "./components/NavBar";
import Login from "./components/Login";
import ProductForm from "./components/ProductForm";
import RegistrationForm from "./components/Registration";

function App() {
    return (
        <BrowserRouter>  
            <NavBar />          
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/products" element={<Products />} />
                <Route path="/storefronts" element={<Storefronts />} />
                <Route path="/orders" element={<Orders />} />
                <Route path="/settings" element={<Settings />} />
                <Route path="/login" element={<Login />} />
                <Route path="/add-product" element={<ProductForm />} />
                <Route path= "/register" element={<RegistrationForm />} />

            </Routes>
        </BrowserRouter>
    );
}

export default App;
