# syntax=docker/dockerfile:1

# ---- build ------------------------------------------------------------------
FROM node:24-alpine AS build
WORKDIR /app

# Install dependencies in their own layer so code changes don't reinstall them.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY . .
RUN npm run build

# ---- runtime ----------------------------------------------------------------
# nginx serves the static build and proxies /api to the backend, so the
# browser talks to a single origin and no CORS is needed.
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1/ || exit 1
