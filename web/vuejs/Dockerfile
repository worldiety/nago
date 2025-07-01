# Dockerfile to test production build of the nago-server
# Build with `docker build -t nago-server .`
# Run with `docker run -p 8100:80 nago-server` and go to localhost:8100
FROM node:22-alpine AS builder
WORKDIR /usr/share/nago
COPY package.json package-lock.json* ./
RUN npm install

COPY . .
RUN npm run build-only

FROM ghcr.io/cryptaliagy/httpget:0.1.23 AS httpget

FROM caddy:2-alpine AS server

COPY Caddyfile /etc/caddy/Caddyfile

COPY --from=builder /usr/share/nago/dist/ /usr/share/nago/
COPY --from=httpget /httpget /bin/httpget

HEALTHCHECK --interval=10s --timeout=10s --start-period=30s CMD ["httpget", "http://localhost/"]
