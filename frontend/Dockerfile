# Build stage (Deno only — do not use npm)
FROM docker.io/denoland/deno:alpine-2.9.8 AS builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /app

COPY package.json deno.lock ./
RUN --mount=type=cache,id=tournament-deno2.9.8-alpine-v1-${TARGETOS}-${TARGETARCH},target=/deno-dir,sharing=locked \
    deno install --frozen

COPY . .

# API base used by the frontend (baked in at build time)
ARG VITE_API_URL=/backend
ENV VITE_API_URL=${VITE_API_URL}

RUN --mount=type=cache,id=tournament-deno2.9.8-alpine-v1-${TARGETOS}-${TARGETARCH},target=/deno-dir,sharing=locked \
    deno task build

# Runtime stage
FROM docker.io/library/nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
