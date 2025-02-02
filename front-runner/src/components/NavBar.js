import { Link } from "react-router-dom";
import "./NavBar.css";

const NavBar = () => {
    return (
        <nav className="nav-bar">
            <div className="nav-icons">
                <Link to="/">
                    <img src="../assets/Logo.svg" className="logo" alt="FR logo"/>
                </Link>
                <Link to="/products">
                    <img src="../assets/Product icon.svg" className="nav-icon" alt="products"/>
                </Link>
                <Link to="/storefronts">
                    <img src="../assets/Storefront icon.svg" className="nav-icon" alt="storefronts"/>
                </Link>
                <Link to="/orders">
                    <img src="../assets/Orders icon.svg" className="nav-icon" alt="orders"/>
                </Link>
            </div>
            <div className="bottom-nav-icons">
                <Link to="/settings">   
                    <img src="../assets/Settings icon.svg" className="bottom-nav-icon" alt="icon"/>
                </Link>
                <Link to="/logout">   
                    <img src="../assets/Logout icon.svg" className="nav-icon" alt="icon"/>
                </Link>
            </div>
        </nav>
    );
};

export default NavBar;
