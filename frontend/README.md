# donguri (仮) フロントエンド開発ガイド

## 使用技術

- Framework: Vue 3
- Language: TypeScript
- Package Manager: pnpm
- Build Tool: Vite
- Linter: ESLint, Stylelint
- Styling: Tailwind CSS v4

## 環境構築

- [**pnpm**](https://pnpm.io/ja/installation#posix-システムの場合) がインストールされている必要があります。 
- `frontend/` ディレクトリ内で、`pnpm install` コマンドを実行して依存関係をインストールしてください。

## 開発の流れ

<!-- TODO: モックサーバーの起動方法を記載する！ -->

- プロジェクトルートで `make frontend/dev` を実行することで、開発サーバーを起動できます。
  - ブラウザで <http://localhost:5173> にアクセスして、ゲームが動作していることを確認してください。
  - コードを編集すると、ブラウザが自動的に更新されて変更が反映されます。
- `package.json` を書き換えたり Git 操作によって `package.json` が変更された場合は、`make frontend/install` を実行して依存関係を更新する必要があります。
- その他のコマンドについては、`make help` を参照してください。

## 開発時の注意点

- Vue を書く際は、常に Composition API を使用し、Options API は使用しないでください。
- 基本的に CSS は直接書かず、Tailwind CSS のユーティリティクラスを使用してスタイリングしてください。
  - ただし、アニメーションの定義など、ユーティリティクラスだけでは実現が難しいスタイリングについては、Scoped CSS を使用してスタイリングしてください。
