# Project Structure

## Organization Philosophy

Goの標準レイアウトに従ったレイヤー構成。`cmd/` がエントリーポイント、`internal/` が責務ごとに分割されたパッケージ群。各パッケージは単一責務を持ち、外部からは隠蔽されている。

## Directory Patterns

### Entry Point
**Location**: `cmd/mdserve/`
**Purpose**: CLIフラグの解析とサーバー起動のみ。ビジネスロジックを持たない
**Example**: `main.go` がフラグを `server.Config` に変換して `server.New(cfg).Start()` を呼ぶ

### Internal Packages
**Location**: `internal/`
**Purpose**: 責務ごとに分割されたパッケージ群。各パッケージは単一責務を持つ
**Pattern**:
- `server/` - HTTPサーバー・ルーター・ハンドラー群（唯一の外部公開サーフェイス）
- `renderer/` - Markdown → HTML変換（goldmarkラッパー）
- `dirlist/` - ディレクトリ一覧の生成
- `tmpl/` - HTMLテンプレートのレンダリング
- `sse/` - Server-Sent Eventsブローカー
- `watcher/` - fsnotifyによるファイル監視

### Dev Tools
**Location**: `tools/`
**Purpose**: 開発時のコード生成など。バイナリには含まれない
**Example**: `gen-chroma-css/` - Chroma用CSSの生成

## Naming Conventions

- **Files**: `snake_case.go`（Go標準）
- **Test files**: `<name>_test.go`（パッケージと同じディレクトリ）
- **Interfaces**: 動詞 + `-er` サフィックス（`Renderer`, `Watcher`, `Broker`）
- **Constructor**: `New()` または `New<Type>()` パターン

## Import Organization

```go
import (
    // 標準ライブラリ
    "context"
    "net/http"

    // 内部パッケージ（モジュール名 `mdserve` を使った絶対パス）
    "mdserve/internal/renderer"
    "mdserve/internal/sse"

    // 外部ライブラリ
    "github.com/yuin/goldmark"
)
```

**パスエイリアス**: なし。すべて `mdserve/internal/<pkg>` 形式の絶対パスを使用。

## Code Organization Principles

- **インターフェース定義はコンシューマー側** - `renderer.Renderer` は `renderer` パッケージ内で定義
- **コンテキスト経由での値受け渡し** - ルーター → ハンドラー間のパス情報は `context.WithValue` で伝達
- **ハンドラーは `http.Handler` を実装** - 関数ではなく構造体でハンドラーを表現し、依存をコンストラクタで注入

---
_Document patterns, not file trees. New files following patterns shouldn't require updates_
