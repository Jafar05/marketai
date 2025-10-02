# Multi-stage Dockerfile for Vite React frontend served by Nginx

FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm ci --no-audit --no-fund

# Build-time API url (for Vite)
ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}

COPY . .
RUN npm run build

FROM nginx:1.27-alpine AS runner

WORKDIR /usr/share/nginx/html
COPY --from=builder /app/dist .

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]




