import { BrowserRouter, Routes, Route } from "react-router-dom";
import Home from "./components/Home";
import Products from "./components/Products";
import Storefronts from "./components/Storefronts";
import Orders from "./components/Orders";
import Settings from "./components/Settings";

function App() {
    return (
        <BrowserRouter>            
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/products" element={<Products />} />
                <Route path="/storefronts" element={<Storefronts />} />
                <Route path="/orders" element={<Orders />} />
                <Route path="/settings" element={<Settings />} />
            </Routes>
        </BrowserRouter>
    );
}

export default App;
