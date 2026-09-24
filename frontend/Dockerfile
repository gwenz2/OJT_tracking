# syntax=docker/dockerfile:1

# Build stage — produces the static PWA bundle in /src/dist.
FROM node:22-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runtime — Caddy serves the static bundle, terminates TLS, and
# reverse-proxies same-origin /api + /health to the Go API.
FROM caddy:2-alpine
COPY --from=build /src/dist /srv
COPY deploy/Caddyfile /etc/caddy/Caddyfile
# Deploy-time env: SITE_ADDRESS (TLS hostname), API_UPSTREAM (api:8080).
EXPOSE 80 443
