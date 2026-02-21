# md2html - Markdown HTML Server

Mermaid.jsの図を含むMarkdownファイルをHTMLに変換してブラウザで閲覧できるローカルWebサーバーアプリケーションです。シングルバイナリとして配布され、インストール不要で即座にMarkdownディレクトリをWebサイト化できます。

## 特徴

- **Markdownレンダリング** - 見出し・リスト・テーブル・コードブロックなどの標準Markdownをシンタックスハイライト付きでレンダリング
- **Mermaid.js対応** - フローチャート・シーケンス図・クラス図などをブラウザ上でSVG表示
- **ディレクトリ閲覧** - ディレクトリ内のMarkdownファイル一覧をナビゲート可能
- **ライブリロード** - ファイル変更を自動検知してブラウザを自動リロード
- **シングルバイナリ** - 依存ライブラリ込みの単一実行ファイルで配布

## 使い方

```bash
# カレントディレクトリをサーブ
mdserve

# 指定ディレクトリをサーブ
mdserve /path/to/docs

# ポートを指定
mdserve --port 8080 /path/to/docs

# ファイル監視を無効化
mdserve --no-watch /path/to/docs

# ヘルプを表示
mdserve --help
```

デフォルトでポート `3333` でサーバーが起動します。

## 動作仕様

- `.md` ファイルへのアクセス → HTMLに変換してレスポンス
- ディレクトリへのアクセス → `README.md` または `index.md` を優先表示、なければファイル一覧を表示
- その他のファイル（画像・PDFなど）→ そのまま配信
- 存在しないパス → HTTP 404

## ビルド

### 必要な環境

- Go 1.22 以上

### ビルド手順

```bash
# 依存パッケージの取得
go mod download

# ビルド（カレントディレクトリに mdserve バイナリを生成）
go build -o mdserve ./cmd/mdserve/

# パスの通った場所にインストール
go install ./cmd/mdserve/
```

### クロスコンパイル

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o mdserve-linux-amd64 ./cmd/mdserve/

# macOS (arm64)
GOOS=darwin GOARCH=arm64 go build -o mdserve-darwin-arm64 ./cmd/mdserve/

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o mdserve-windows-amd64.exe ./cmd/mdserve/
```

### テスト

```bash
go test ./...
```

## ライセンス

[MIT License](LICENSE)

## 開発

本プロジェクトは [cc-sdd](https://github.com/gotalab/cc-sdd) による Spec-Driven Development で管理されています。

```
.kiro/specs/markdown-html-server/   # 仕様書
```

仕様の確認:

```bash
/kiro:spec-status markdown-html-server
```
