# Build for linux/amd64 (e.g. STACKIT SKE). On ARM Macs, omit --platform for local arm64.
# Build stage
FROM --platform=linux/amd64 node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage: nginx serves the built SPA (history-mode routing)
FROM --platform=linux/amd64 nginx:alpine

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
