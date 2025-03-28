import "./NavBar.css";
import {
    logo,
    productIcon,
    storefrontIcon,
    ordersIcon,
    settingsIcon,
    logoutIcon
} from '../assets/icons';

const handleLogout = async () => {
    try {
      const response = await fetch("/api/logout", { method: "POST" });
      if (response.ok) {
        // Optionally log a message or check response text here
        window.location.href = "/login";
      } else {
        console.error("Logout failed");
        // Still redirect to login even if logout fails, or handle it differently
        window.location.href = "/login";
      }
    } catch (error) {
      console.error("Error during logout:", error);
      // Redirect to login in case of error
      window.location.href = "/login";
    }
  };

const NavBar = () => {
    return (
        <nav className="nav-bar">
            <div className="nav-icons">
                <div onClick={() => window.location.href = "/"} className="nav-option">
                    <img src={logo.default} className="logo" alt="FR logo"/>
                    <h2><i>FrontRunner </i></h2>
                </div>
                <div onClick={() => window.location.href = "/products"} className="nav-option">
                    <img src={productIcon.default} className="nav-icon" alt="products"/>
                    <h2>My Products</h2>
                </div>
                <div onClick={() => window.location.href = "/storefronts"} className="nav-option">
                    <img src={storefrontIcon.default} className="nav-icon" alt="storefronts"/>
                    <h2>My Storefronts</h2>
                </div>
                <div onClick={() => window.location.href = "/orders"} className="nav-option">
                    <img src={ordersIcon.default} className="nav-icon" alt="orders"/>
                    <h2>My Orders</h2>
                </div>
            </div>
            <div className="bottom-nav-icons">
                <div onClick={() => window.location.href = "/settings"} className="nav-option">   
                    <img src={settingsIcon.default} className="bottom-nav-icon" alt="icon"/>
                    <h2>Settings</h2>
                </div>
                <div onClick={handleLogout} className="nav-option">   
                    <img src={logoutIcon.default} className="nav-icon" alt="icon"/>
                    <h2>Logout</h2>
                </div>
            </div>
        </nav>
    );
};

export default NavBar;
