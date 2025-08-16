#!/bin/sh

# Inject runtime env config
cat <<EOF > /usr/share/nginx/html/env.js
window.RUNTIME_CONFIG = {
  VITE_API_URL: "${VITE_API_URL}",
  VITE_WEBSOCKET_URL: "${VITE_WEBSOCKET_URL}",
  VITE_JELLYSEERR_URL: "${VITE_JELLYSEERR_URL}"
};
EOF

# Start nginx
exec nginx -g "daemon off;"
