# ==============================================================================
# NetFusion Frontend Dockerfile
# Multi-stage build: Vite production build -> Nginx Alpine reverse proxy
# ==============================================================================

FROM node:22-alpine AS builder

WORKDIR /app

# Cache package manifests
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# Copy source code and build production bundle
COPY frontend/ ./
RUN npm run build

# Nginx runtime stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY deploy/docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
