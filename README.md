# donguri (仮)

三大学合同ハッカソンＡ班「チームどんぐり」による Web ゲームのリポジトリです。

---

## 環境構築

- Windows の場合は WSL2 をセットアップし、そこにこのリポジトリをクローンしてください。このプロジェクトでの操作は全て WSL2 上で行うことを想定しています。
  - Git の初期設定が済んでいる必要があります。
- [**Docker Desktop**](https://www.docker.com/ja-jp/get-started/) をインストールしてください。
- 開発環境は [**VS Code**](https://code.visualstudio.com/) (または VS Code 互換エディタ) を使用することを推奨します。セットアップ方法は [**VS Code セットアップガイド**](.vscode/README.md) を参照してください。
- フロントエンド開発時の環境構築は [**donguri (仮) フロントエンド開発ガイド**](frontend/README.md) を参照してください。
- バックエンド開発時の環境構築は [**donguri (仮) バックエンド開発ガイド**](backend/README.md) を参照してください。

---

## コマンド一覧

プロジェクトルートで `make help` を実行するとコマンド一覧が確認できます。

```sh
make help
```

よく使うコマンドは以下の通りです。

| コマンド | 説明 |
|---|---|
| `make frontend/dev` | フロントエンド開発サーバーを起動する |
| `make backend/dev` | バックエンド開発サーバーを起動する (Docker Compose Watch) |
| `make lint` | 全 Linter を実行する |
| `make backend/test` | バックエンドのユニットテストを実行する |

---

## Docker でのビルド手順

> [!Warning]
>
> このビルド手順では開発に必要なセットアップは行わず、あくまでデプロイ環境と同様に Docker のみでただ動かしたい時のための手順であることに注意してください。開発環境のセットアップは上記のガイドを参照してください。

1.  ターミナルでプロジェクトのルートディレクトリに移動します。
2.  以下のコマンドを実行して、バックエンドアプリケーションと PostgreSQL を含む必要なコンテナ群を起動します。
    ```bash
    docker compose -f backend/compose.yml up
    ```
3.  コンテナの起動が完了したら、ブラウザで <http://localhost:8080> にアクセスし、ゲームが動作していることを確認してください。

---

## AI Coding Agent を使用する場合

`pnpm dlx skills experimental_install` コマンドを実行することで、AI 用の推奨スキルを一括で `.agents/skills/` 以下にインストールすることができます。使用する Agent によってはディレクトリが異なる場合があるため、必要に応じてシンボリックリンクを貼ってください。
