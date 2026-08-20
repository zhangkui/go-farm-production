# ---- Build stage ----
FROM node:20-bookworm-slim AS builder
WORKDIR /src
COPY package.json package-lock.json* ./
RUN npm install --no-audit --no-fund
COPY . .
RUN npm run build

# ---- Runtime: Nginx serves the SPA and proxies /api to the API container ----
FROM nginx:1.25-alpine
COPY --from=builder /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
