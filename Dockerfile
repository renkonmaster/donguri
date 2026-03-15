# syntax=docker/dockerfile:1

FROM node:24-alpine AS frontend-builder

WORKDIR /app

ARG KOYEB_PUBLIC_DOMAIN
ENV KOYEB_PUBLIC_DOMAIN=${KOYEB_PUBLIC_DOMAIN}
ENV VITE_PUBLIC_URL=https://${KOYEB_PUBLIC_DOMAIN}

RUN corepack enable

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./
RUN pnpm build

# ----------------------------------------
# Frontend runtime (Koyeb frontend service)
# ----------------------------------------
FROM nginx:1.29-alpine AS frontend

COPY --from=frontend-builder /app/dist /usr/share/nginx/html
EXPOSE 80

# ----------------------------------------
# Backend builder (Koyeb backend service)
# ----------------------------------------
FROM golang:1.26 AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

RUN \
  --mount=type=cache,target=${GOCACHE} \
  --mount=type=cache,target=${GOMODCACHE} \
  --mount=type=bind,source=backend/go.mod,target=go.mod \
  --mount=type=bind,source=backend/go.sum,target=go.sum \
  go mod download

RUN \
  --mount=type=cache,target=${GOCACHE} \
  --mount=type=cache,target=${GOMODCACHE} \
  --mount=type=bind,source=backend,target=.,readwrite \
  go build -o /usr/bin/server ./main.go

# use `debug-nonroot` for debug shell access
FROM gcr.io/distroless/static-debian11:nonroot

COPY --from=builder /usr/bin/server /usr/bin/server

CMD ["/usr/bin/server"]
