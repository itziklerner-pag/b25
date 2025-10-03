<<<<<<< HEAD
# Multi-stage build for optimized production image
=======
# Multi-stage build for React Web UI

# Builder stage - builds the React application
>>>>>>> refs/remotes/origin/main
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build application
RUN npm run build

<<<<<<< HEAD
# Production stage
=======
# Development stage - runs dev server with hot reload
FROM node:20-alpine AS development

# Install development tools
RUN apk add --no-cache curl

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install all dependencies
RUN npm ci

# Expose port
EXPOSE 5173

# Health check for development (Vite dev server)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:5173/ || exit 1

# Note: Source code should be mounted as volume in docker-compose
# Development command with hot reload
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Production stage - serves static files with nginx
>>>>>>> refs/remotes/origin/main
FROM nginx:alpine

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

<<<<<<< HEAD
# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/health || exit 1
=======
# Create non-root user
RUN addgroup -g 1000 nginx-app && \
    adduser -D -u 1000 -G nginx-app nginx-app

# Expose port
EXPOSE 80

# Health check for production
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost/health || exit 1
>>>>>>> refs/remotes/origin/main

CMD ["nginx", "-g", "daemon off;"]
