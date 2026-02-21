# Research & Design Decisions

---
**Purpose**: `directory-listing-with-readme` フィーチャーの設計判断と調査記録

---

## Summary

- **Feature**: `directory-listing-with-readme`
- **Discovery Scope**: Extension（既存システムの拡張）
- **Key Findings**:
  - `directoryHandler.ServeHTTP`（`internal/server/handlers.go`）が index ファイルとリスト表示の分岐を担っており、最小変更でクエリパラメータ分岐を挿入できる
  - `PageData` / `DirListData` 構造体にフィールドを追加するだけでテンプレート側へ URL を渡せる。テンプレートエンジンの `TemplateEngine` インターフェース変更は不要
  - 新規外部依存はなし。標準ライブラリの `url.Values.Has()`（Go 1.17+）を利用するが、プロジェクトは Go 1.24 なので互換性問題なし

---

## Research Log

### `?list` クエリパラメータ検出箇所の特定

- **Context**: `?list` によるリスト表示強制をどのレイヤーで処理すべきか
- **Sources Consulted**: `internal/server/router.go`、`internal/server/handlers.go`、Go 標準ライブラリドキュメント
- **Findings**:
  - `requestRouter` はファイルシステムタイプ（ディレクトリ / .md / その他）でルーティングするだけで、クエリパラメータを見ていない
  - クエリパラメータは `directoryHandler.ServeHTTP` の `r.URL.Query()` で取得可能
  - `url.Values.Has("list")` でキーの存在確認（値は問わない）ができる
- **Implications**: `requestRouter` の変更は不要。`directoryHandler` 内の分岐条件に `!forceList` を加えるだけで対応可能

### テンプレートへの URL 伝達方法

- **Context**: ナビゲーションリンク URL をテンプレートにどう渡すか
- **Sources Consulted**: `internal/tmpl/tmpl.go`、`internal/tmpl/templates/page.html`、`internal/tmpl/templates/dirlist.html`
- **Findings**:
  - `PageData` と `DirListData` は単純な Go 構造体でフィールド追加が容易
  - テンプレートは `{{if .DirListURL}}` のような条件分岐で任意フィールドを表示制御できる
  - `TemplateEngine` インターフェースのシグネチャ変更は不要（既存の引数型を拡張するだけ）
- **Implications**: `PageData.DirListURL string` と `DirListData.IndexURL string` を追加するだけでテンプレート側の条件レンダリングが実現できる

### ナビゲーション URL の構築

- **Context**: `serveIndexFile` および `serveDirList` からどのように URL を組み立てるか
- **Sources Consulted**: `internal/server/handlers.go`、`net/url` ドキュメント
- **Findings**:
  - `r.URL.Path` にはディレクトリの URL パス（末尾 `/` あり）が格納されている
  - `DirListURL = r.URL.Path + "?list"` で一覧ページの URL を生成できる
  - `IndexURL = r.URL.Path`（クエリパラメータなし）で README ページへ戻れる
- **Implications**: 外部ライブラリ不要。シンプルな文字列連結で完結

---

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| クエリパラメータ `?list` | URL に `?list` を付与してリスト強制表示 | 実装最小、URL の意味が明確、ブックマーク可能 | `?list=anything` など値ありでも動作するが意味的に混乱しない | **採用** |
| 専用サブパス `/_list/` | `/_list/path/to/dir/` で一覧 | URL が明示的 | ルーターへの追加ハンドリングが必要、相対リンクへの影響あり | 不採用 |
| UI トグルのみ（URL 変更なし） | JavaScript で表示切替 | サーバー変更最小 | ブックマーク不可、SSR の恩恵なし | 不採用 |

---

## Design Decisions

### Decision: クエリパラメータのキー存在チェックで一覧モードを判定

- **Context**: `?list` を値付き（`?list=true`）か値なし（`?list`）かで処理する方法の選択
- **Alternatives Considered**:
  1. 値チェック（`r.URL.Query().Get("list") == "true"`) — 厳密だが UX に余計な制約
  2. キー存在チェック（`r.URL.Query().Has("list")`) — 値を問わない直感的な挙動
- **Selected Approach**: `url.Values.Has("list")` によるキー存在チェック
- **Rationale**: `?list` のみ（値なし）でも動作させることでユーザーが手動入力しやすい。値の意味を定義・ドキュメント化する必要がない
- **Trade-offs**: `?list=false` でも一覧表示になるが、ユーザーが `false` を指定するシナリオは想定されない
- **Follow-up**: テストで `?list`、`?list=anything`、`?list=` の各ケースを確認

### Decision: PageData / DirListData へのフィールド追加

- **Context**: テンプレートへの URL 渡し方法
- **Alternatives Considered**:
  1. `context.WithValue` でハンドラーからテンプレートへ — 複雑で型安全性が下がる
  2. 構造体フィールド追加 — 型安全、テンプレートとの接続がシンプル
- **Selected Approach**: 構造体フィールド追加（`PageData.DirListURL`、`DirListData.IndexURL`）
- **Rationale**: 既存パターン（`PageData` 拡張）と一致し、ゼロ値（空文字）でリンクを非表示にできる
- **Trade-offs**: 将来的にフィールドが増えすぎると構造体が肥大化する可能性があるが、現時点では許容範囲

---

## Risks & Mitigations

- `DirListURL` が誤って空でない値で設定された場合、README ページに不要なリンクが表示される — ハンドラーで `listing.IndexFile != ""` の条件を厳密にチェック
- クエリパラメータ付き URL がライブリロード時に正しく再ロードされるか — SSE リロードは `location.reload()` を呼ぶだけなので URL はそのまま保持される。問題なし

---

## References

- `internal/server/handlers.go` — `directoryHandler.serveIndexFile` / `serveDirList` 実装
- `internal/tmpl/tmpl.go` — `PageData` / `DirListData` 構造体定義
- Go 標準ライブラリ `net/url` — `url.Values.Has()` (Go 1.17+)
