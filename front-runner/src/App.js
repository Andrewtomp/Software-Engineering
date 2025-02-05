import { BrowserRouter, Routes, Route } from "react-router-dom";
import Home from "./components/Home";
import Products from "./components/Products";
import Storefronts from "./components/Storefronts";
import Orders from "./components/Orders";
import Settings from "./components/Settings";
import Login from "./components/Login";
import Registration from "./components/Registration";

function App() {
    return (
        <BrowserRouter>            
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/products" element={<Products />} />
                <Route path="/storefronts" element={<Storefronts />} />
                <Route path="/orders" element={<Orders />} />
                <Route path="/settings" element={<Settings />} />
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Registration />} />
            </Routes>
        </BrowserRouter>
    );
}

export default App;
