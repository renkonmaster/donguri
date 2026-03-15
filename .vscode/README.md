# VS Code セットアップガイド

---

## 共通

### 拡張機能のインストール

推奨拡張機能は `.vscode/extensions.json` に定義されています。  
VS Code の通知、または以下の手順で一括インストールできます。

1. コマンドパレット (`Ctrl+Shift+P` / `Cmd+Shift+P`) を開く
2. `Extensions: Show Recommended Extensions` を実行
3. 表示された拡張機能をすべてインストールする

---

## フロントエンド

追加のセットアップは不要です。拡張機能をインストールすれば動作します。

### 使用ツール

| ツール    | 用途                                               |
| --------- | -------------------------------------------------- |
| Volar     | Vue 3 の Language Server                           |
| ESLint    | JavaScript / TypeScript / Vue のフォーマット・Lint |
| Stylelint | CSS / Tailwind CSS の Lint                         |
| Prettier  | その他ファイルのフォーマット                       |

---

## バックエンド

### 1. Go のインストール

[go.dev](https://go.dev/dl/) から Go をインストールしてください。

### 2. Go ツールチェーンと golangci-lint v2 のインストール

VS Code を開くと `golang.go` 拡張が gopls 等の必要ツールのインストールを促します。
通知に従ってインストールしてください。

次に、golangci-lint v2 を拡張経由でインストールします。  
`.vscode/settings.json` の Go 関連設定はすでに済んでいるため、以下の手順を実行するだけです。

1. コマンドパレット (`Ctrl+Shift+P` / `Cmd+Shift+P`) を開く
2. `Go: Install/Update Tools` を実行
3. `golangci-lint-v2` を選択してインストール

> **Note**
> v1 と v2 の共存が可能なため、既存の `golangci-lint` (v1) が入っていても問題ありません。  
> また、`.golangci.yml` はファイルの場所から自動検出されるため、VS Code 側での設定は不要です。

### 使用ツール

| ツール           | 用途                                    |
| ---------------- | --------------------------------------- |
| golang.go        | Go の Language Server (gopls)・デバッガ |
| golangci-lint v2 | Lint・フォーマット (gofmt 含む)         |

---

## おすすめ個人設定 (任意)

プロジェクト共有の設定ではなく、各自のユーザー設定に追加することを推奨します。  
`Ctrl+Shift+P` → `Preferences: Open User Settings (JSON)` で編集できます。

### 保存時の自動フォーマット

```jsonc
// すべてのファイルで保存時にフォーマット
"editor.formatOnSave": true,
```

### Go: 保存時に import を自動整理

未使用 import の削除と不足 import の追加を自動で行います。

```jsonc
"[go]": {
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  }
},
```

### Go: 保存時に Lint を実行

```jsonc
// "package" はカレントパッケージのみ、"workspace" はプロジェクト全体 (重い)
"go.lintOnSave": "package",
```
