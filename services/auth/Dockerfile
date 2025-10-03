# Multi-stage build for Node.js Authentication Service

# Builder stage - compiles TypeScript
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install all dependencies (including dev dependencies for building)
RUN npm ci

# Copy source code
COPY src ./src

# Build TypeScript
RUN npm run build

# Development stage - includes dev tools and supports hot reload
FROM node:20-alpine AS development

# Set environment
ENV NODE_ENV=development

WORKDIR /app

# Install development tools
RUN apk add --no-cache curl

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install all dependencies (including dev dependencies)
RUN npm ci

# Expose port
EXPOSE 9097

# Health check for development
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD node -e "require('http').get('http://localhost:9097/health/live', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})"

# Note: Source code should be mounted as volume in docker-compose
# Development command with hot reload using ts-node-dev or nodemon
CMD ["npm", "run", "dev"]

# Production stage - minimal runtime image
FROM node:20-alpine

# Set environment
ENV NODE_ENV=production

# Create app user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

WORKDIR /app

# Copy package files and install production dependencies only
COPY package*.json ./
RUN npm ci --only=production && \
    npm cache clean --force

# Copy built files from builder
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/src/database/schema.sql ./dist/database/

# Change ownership
RUN chown -R nodejs:nodejs /app

# Switch to non-root user
USER nodejs

# Expose port
EXPOSE 9097

# Health check for production
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD node -e "require('http').get('http://localhost:9097/health/live', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})"

# Start the application
CMD ["node", "dist/server.js"]
