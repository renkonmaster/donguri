# syntax=docker/dockerfile:1

# ----------------------------------------
# 1. Frontend Builder (Node.js)
# ----------------------------------------
FROM node:24-alpine AS frontend-builder
WORKDIR /app

ARG KOYEB_PUBLIC_DOMAIN
# KoyebのダッシュボードでVITE_PUBLIC_URLを直接設定する場合は以下のドメイン結合は不要です
ENV VITE_PUBLIC_URL=https://${KOYEB_PUBLIC_DOMAIN}

RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build
# → /app/dist に静的ファイルが生成される

# ----------------------------------------
# 2. Backend Builder (Go)
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

# ----------------------------------------
# 3. Final Runtime (統合された1つのアプリ)
# ----------------------------------------
FROM gcr.io/distroless/static-debian11:nonroot

WORKDIR /app

# Goの実行ファイルをコピー
COPY --from=builder /usr/bin/server /app/server

# Nginxを使わず、Viteのビルド結果(dist)をそのまま最終コンテナにコピー
COPY --from=frontend-builder /app/dist /app/dist

# アプリケーションの実行
CMD ["/app/server"]
