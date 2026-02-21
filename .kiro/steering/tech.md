# Technology Stack

## Architecture

CLIエントリーポイント → HTTPサーバー → ルーター → 専用ハンドラー、という直線的なレイヤー構成。
依存は常に上位 → 下位の一方向。`internal/` パッケージ間に循環依存はない。

## Core Technologies

- **Language**: Go 1.24
- **Module**: `mdserve`
- **Runtime**: Go standard library + 最小限の外部ライブラリ

## Key Libraries

| Library | Purpose |
|---|---|
| `github.com/yuin/goldmark` | Markdownパーサー・レンダラー（拡張可能） |
| `goldmark-highlighting/v2` + `alecthomas/chroma/v2` | シンタックスハイライト（CSSクラス方式） |
| `go.abhg.dev/goldmark/frontmatter` | YAMLフロントマターの除去 |
| `github.com/fsnotify/fsnotify` | クロスプラットフォームのファイル監視 |

## Development Standards

### Code Quality

- `go vet` + `errcheck` + `staticcheck` によるlint
- エラーを意図的に無視する場合は `_ = expr` または `_, _ = expr` 構文を使用（`errcheck` は `//nolint` コメントを認識しないため）

### Testing

- `go test ./...` で全パッケージをテスト
- 統合テスト: `integration_test.go`（ルートに配置）
- 各 `internal/` パッケージは対応する `_test.go` を持つ

### Asset Embedding

- 静的アセット（CSS・JS・テンプレート）は `embed.go` でバイナリに埋め込む
- `//go:embed` ディレクティブを使用

## Development Environment

### Required Tools

- [mise](https://mise.jdx.dev/) - Goバージョン管理（`mise.toml` で自動適用）
- Go 1.24+（`mise install` で自動セットアップ）

### Common Commands

```bash
# Dev server (live):
go run ./cmd/mdserve/ [path]

# Build:
go build -o mdserve ./cmd/mdserve/

# Test:
go test ./...

# Cross-compile:
GOOS=linux GOARCH=amd64 go build -o mdserve-linux-amd64 ./cmd/mdserve/
```

## Key Technical Decisions

- **SSEによるライブリロード** - WebSocketではなくSSE（`/events` エンドポイント）を採用。実装がシンプルでHTTPに収まる
- **Chromaのクラス方式** - インラインスタイルではなくCSSクラスでシンタックスハイライトを適用。テーマ変更が容易
- **シンボリックリンク解決** - パストラバーサル防止のためルーターでシンボリックリンクを完全解決して二重チェック

---
_Document standards and patterns, not every dependency_
