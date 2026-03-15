# donguri (仮) バックエンド開発ガイド

## 使用技術

- Language: Go
- API Schema: ogen (OpenAPI)
- DB: PostgreSQL (PostGIS)
- ORM: GORM
- Linter: golangci-lint v2

## 環境構築

- [**Go**](https://go.dev/dl/) が `go.mod` に記載のバージョン以上でインストールされている必要があります。
- [**Docker**](https://www.docker.com/) と Docker Compose v2.22 以上が必要です。
- Linter を手元で実行する場合は [**golangci-lint v2**](https://golangci-lint.run/welcome/install/) のインストールが必要です。

## 開発の流れ

- プロジェクトルートで `make backend/dev` を実行することで、開発サーバーを起動できます。
  - ソースコードの変更を検知して自動でリビルド・再起動されます。
  - 起動後、以下の URL にアクセスできます。
    - <http://localhost:8080/> (API)
    - <http://localhost:8081/> (DB 管理画面)
- その他のコマンドについては、`make help` を参照してください。

## 開発時の注意点

- ルーティングは `internal/handler/` に、DB アクセスは `internal/repository/` にそれぞれ実装してください。
- `internal/` パッケージは外部モジュールから参照できないため、結合テスト (`integration_tests/`) に公開する必要があるコードは `infrastructure/` に配置してください。
- OpenAPI スキーマを変更した場合は、`go generate ./...` でコードを再生成してください。
