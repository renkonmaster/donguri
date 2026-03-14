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

## Docker でのビルド手順

> [!Warning]
> このビルド手順では開発に必要なセットアップは行わず、あくまでデプロイ環境と同様に Docker のみでただ動かしたい時のための手順であることに注意してください。開発環境のセットアップは上記のガイドを参照してください。

1.  ターミナルでプロジェクトのルートディレクトリに移動します。
2.  以下のコマンドを実行して Docker イメージをビルドします。
    ```bash
    docker build -t donguri:latest .
    ```
3.  以下のコマンドを実行して Docker コンテナを起動します。
    ```bash
    docker run -p 8080:8080 donguri:latest
    ```
4.  ブラウザで <http://localhost:8080> にアクセスし、ゲームが動作していることを確認してください。
