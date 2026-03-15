.DEFAULT_GOAL := help
.PHONY: help lint \
	frontend/install frontend/dev frontend/build frontend/typecheck frontend/lint frontend/lint/fix \
	backend/dev backend/build backend/lint backend/lint/fix backend/test

help: ## コマンド一覧を表示する
	@grep -E '^[a-zA-Z][a-zA-Z/]*:.*## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*## "}; {printf "  %-28s%s\n", $$1, $$2}'

lint: frontend/lint backend/lint ## 全 Linter を実行する

# ----------------------------------------
# Frontend
# ----------------------------------------

frontend/install: ## 依存関係をインストールする
	pnpm -C frontend install

frontend/dev: ## 開発サーバーを起動する
	pnpm -C frontend dev

frontend/build: ## 本番ビルドを実行する
	pnpm -C frontend build

frontend/typecheck: ## 型チェックを実行する
	pnpm -C frontend typecheck

frontend/lint: ## Linter を実行する
	pnpm -C frontend lint

frontend/lint/fix: ## Linter を実行して自動修正する
	pnpm -C frontend lint:fix

# ----------------------------------------
# Backend
# ----------------------------------------

backend/dev: ## 開発サーバーを起動する (Docker Compose Watch)
	docker compose -f backend/compose.yml watch

backend/build: ## バイナリをビルドする
	go -C backend mod download
	go -C backend build -o ./bin/server ./main.go

backend/lint: ## Linter を実行する
	cd backend && golangci-lint run --timeout=5m ./...

backend/lint/fix: ## Linter を実行して自動修正する
	cd backend && golangci-lint run --timeout=5m --fix ./...

backend/test: ## ユニットテストを実行する
	go -C backend test -v -cover -race -shuffle=on ./internal/...
