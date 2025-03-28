const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
  app.use(
    "/api", // Only proxy API requests
    createProxyMiddleware({
      target: "https://localhost:8080", // Change to your backend URL
      changeOrigin: true,
      secure: false, // Disable SSL verification
    })
  );
};
