const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
    app.use(createProxyMiddleware("/api", { target: "http://localhost:3333" }));
    app.use(createProxyMiddleware("/setup", { target: "http://localhost:3333" }));
    app.use(createProxyMiddleware("/ws", { target: "http://localhost:3333", ws: true }));
};
