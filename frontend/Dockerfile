# Build stage (Deno only — do not use npm)
FROM denoland/deno:alpine-2 AS builder

WORKDIR /app

COPY package.json deno.lock* ./
RUN deno install

COPY . .

# API base used by the frontend (baked in at build time)
ARG VITE_API_URL=/backend
ENV VITE_API_URL=${VITE_API_URL}

RUN deno task build

# Runtime stage
FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
