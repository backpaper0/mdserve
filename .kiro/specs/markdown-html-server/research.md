# リサーチ & 設計決定ログ

---
**目的**: ディスカバリーフェーズの調査結果、アーキテクチャ検討、設計根拠を記録する。

**利用方法**:
- ディスカバリーフェーズ中のリサーチ活動と結果を記録する。
- `design.md` に記載するには詳細すぎる設計決定のトレードオフを文書化する。
- 今後の監査や再利用のための参考資料・根拠を提供する。
---

## Summary

- **Feature**: `markdown-html-server`
- **Discovery Scope**: 新規機能（グリーンフィールド）— 既存コードベースなし
- **Key Findings**:
  - Goはシングルバイナリ配布の主要言語として最適。`//go:embed`ディレクティブで静的アセット（CSS・JS）をバイナリに内包できる
  - goldmark は CommonMark 準拠・高拡張性の Go Markdown パーサーであり、blackfriday より積極的にメンテナンスされている。豊富な拡張エコシステム（front matter、シンタックスハイライト、Mermaid）が存在する
  - Mermaid.jsはクライアントサイドレンダリング（mermaid.min.js v11.12.3 をバイナリに埋め込みローカル配信）が最適解。サーバーサイドレンダリングはChromiumへの依存を生むため採用不可
  - ライブリロードは SSE（Server-Sent Events）で実装可能で、Go stdlibのみで完結できる。WebSocket より実装コストが低く、一方向通信に特化して適切

---

## Research Log

### Markdown パーサー選定

- **Context**: シングルバイナリで配布するため、外部プロセスなしに Go 内でMarkdown→HTML変換が必要
- **Sources Consulted**:
  - goldmark GitHub: https://github.com/yuin/goldmark
  - blackfriday GitHub: https://github.com/russross/blackfriday
  - goldmark-highlighting: https://github.com/yuin/goldmark-highlighting
  - goldmark-frontmatter: https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter
- **Findings**:
  - goldmark: CommonMark 準拠、インターフェースベースのAST設計で外部拡張が容易、Hugoでも採用、2025年も活発にメンテナンス中
  - blackfriday v2: CommonMark非準拠、AST設計が外部拡張を許可しない構造、実質メンテナンスモード
  - goldmark 拡張エコシステム: Table・Strikethrough・TaskList（標準拡張）、Front Matter（`go.abhg.dev/goldmark/frontmatter`）、シンタックスハイライト（`goldmark-highlighting/v2` + Chroma）、Mermaid（`go.abhg.dev/goldmark/mermaid`）
- **Implications**: goldmark を採用。拡張パッケージで全要件を満たせる

### Mermaid.js 組み込み方式

- **Context**: 要件 2.4「Mermaid.jsをサーバー側に埋め込み、ブラウザがCDNに依存せずにレンダリングできるようにする」
- **Sources Consulted**:
  - mermaid.js releases: https://github.com/mermaid-js/mermaid/releases
  - goldmark-mermaid: https://pkg.go.dev/go.abhg.dev/goldmark/mermaid
  - mermaid usage docs: https://mermaid.js.org/config/usage.html
- **Findings**:
  - mermaid.min.js 最新バージョン: v11.12.3（2026年2月時点）
  - サーバーサイドレンダリング（go-rod / chromedp）はChromium依存が発生し、シングルバイナリ要件に違反する
  - クライアントサイドレンダリング（CDN経由）はオフライン動作不可でありCDN依存が生じる
  - **最適解**: `//go:embed`で mermaid.min.js をバイナリに埋め込み、`/assets/mermaid.min.js`として配信。HTMLテンプレートにローカルパスのscriptタグを埋め込む。goldmark-mermaidのカスタムURL設定またはカスタムコードブロックレンダラーで `<div class="mermaid">` を出力する
- **Implications**: mermaid.min.js（~3-5 MB）がバイナリサイズに加算される。機能面でのトレードオフ（オフライン動作・CDN非依存）と許容できる

### ファイル監視ライブラリ

- **Context**: 要件 6.1「ドキュメントルート以下のMarkdownファイルの変更・追加・削除を監視する」
- **Sources Consulted**:
  - fsnotify: https://pkg.go.dev/github.com/fsnotify/fsnotify
  - fsnotify GitHub: https://github.com/fsnotify/fsnotify
- **Findings**:
  - fsnotify v1.9.0（2025年4月リリース）がデファクトスタンダード
  - **重要な制限**: デフォルトでは再帰監視に対応していない（issue #18で追跡中）
  - 対応策: 起動時にディレクトリツリーをwalkして全サブディレクトリを個別に `watcher.Add()` し、`Create` イベントで新規ディレクトリを追加登録する
  - クロスプラットフォーム対応: Linux（inotify）、macOS（kqueue/FSEvents）、Windows（ReadDirectoryChangesW）
- **Implications**: FileWatcher コンポーネントは再帰監視ロジックを内部実装する必要がある

### ライブリロード方式（SSE vs WebSocket）

- **Context**: 要件 6.2「WebSocketまたはServer-Sent Eventsを使って自動的にリロードさせる」
- **Sources Consulted**:
  - SSE vs WebSocket: https://ably.com/blog/websockets-vs-sse
  - Go SSE 実装: https://www.freecodecamp.org/news/how-to-implement-server-sent-events-in-go/
- **Findings**:
  - ライブリロードは サーバー→ブラウザの一方向通信のみ必要
  - SSEは `net/http` と `http.Flusher` インターフェースのみで実装可能（外部ライブラリ不要）
  - ブラウザの `EventSource` APIは自動再接続機能を持つ
  - WebSocketはプロトコルアップグレードが必要で実装コストが高く、この用途では過剰
- **Implications**: SSEを採用。外部依存なしでシンプルに実装できる

### シンタックスハイライト

- **Context**: 要件 1.3「シンタックスハイライトを適用したコードブロックをレンダリングする」
- **Sources Consulted**:
  - Chroma: https://pkg.go.dev/github.com/alecthomas/chroma/v2
  - goldmark-highlighting: https://github.com/yuin/goldmark-highlighting
- **Findings**:
  - Chroma v2.23.1（2026年1月リリース）—250以上の言語レキサーをサポート
  - goldmark-highlighting/v2 が goldmark との統合を提供
  - `WithClasses(true)` モードでインラインスタイルの代わりにCSSクラスを使用し、CSSファイルを別途生成できる
  - GitHubスタイル（`github`）などのビルトインスタイルが利用可能
- **Implications**: Chroma CSSをバイナリに埋め込むか、`<style>`タグでHTMLに埋め込む

### 静的アセット埋め込み（`//go:embed`）

- **Sources Consulted**:
  - embed package: https://pkg.go.dev/embed
  - Go embed blog: https://oneuptime.com/blog/post/2026-01-25-bundle-static-assets-go-embed/view
- **Findings**:
  - Go 1.16以降で利用可能。外部ライブラリ不要
  - `//go:embed`ディレクティブで `embed.FS` または `[]byte` としてアセットを埋め込める
  - 隠しファイル（`.`始まり）はデフォルト除外。`all:`プレフィックスで含めることができる
  - パッケージレベルの `var` 宣言にのみ使用可能（関数内不可）
  - 埋め込みファイルはランタイムで変更不可（イミュータブル）
- **Implications**: `assets/` ディレクトリを `embed.FS` で埋め込み、`http.FileServer(http.FS(assetsFS))` でサーブする

### Markdown CSS スタイル

- **Sources Consulted**:
  - github-markdown-css: https://github.com/sindresorhus/github-markdown-css
- **Findings**:
  - github-markdown-css v5.9.0（2026年2月リリース、MIT ライセンス）
  - GitHub の Markdown レンダリングを再現する最小 CSS
  - ライト・ダーク・オート切り替えの3バリアントが存在
  - ファイルサイズ: 約10-15KB（バイナリサイズへの影響は軽微）
- **Implications**: `github-markdown.css`（自動ライト/ダーク）を採用。`<article class="markdown-body">` でラップして適用

---

## Architecture Pattern Evaluation

| オプション | 説明 | 強み | リスク・制限 | 備考 |
|----------|------|------|------------|------|
| レイヤードアーキテクチャ（採用） | CLI → Server → Business Logic（Renderer/Watcher）の明確な層分離 | シンプル、保守しやすい、テスト容易 | 大規模化時に制約が出る可能性 | 本アプリケーションの規模・複雑度に最適 |
| ヘキサゴナル（Ports & Adapters） | ドメインコアを抽象化し、アダプター経由で外部と接続 | テスト容易、依存性逆転 | アダプター実装のオーバーヘッド | 本アプリには過剰 |
| モノリシック（単一ファイル） | 全処理を `main.go` に集約 | 実装速度が高い | 保守性が低い、テスト困難 | 小規模ツールでは有効だが拡張性に欠ける |

---

## Design Decisions

### Decision: `プログラミング言語の選定`

- **Context**: シングルバイナリ配布（要件 5.1）とクロスプラットフォーム動作が必要
- **Alternatives Considered**:
  1. Go — `go build`で依存なしシングルバイナリ生成、標準ライブラリにHTTPサーバーが内包
  2. Rust — 高性能だが生態系がGo比で成熟度が低く、開発コストが高い
  3. Node.js — エコシステムが豊富だが、ランタイムが必要でシングルバイナリ化が複雑
- **Selected Approach**: Go（1.22+）
- **Rationale**: シングルバイナリ要件に最適、`net/http` でHTTPサーバーが標準提供、`//go:embed` でアセット埋め込みが容易、クロスプラットフォームコンパイルが標準サポート
- **Trade-offs**: GoのMarkdownエコシステムはNode.jsより小さいが、goldmarkで十分に機能要件を満たせる
- **Follow-up**: 本番バイナリのサイズ検証（目安: 20-30 MB with mermaid.min.js）

### Decision: `Mermaid.js のローカル埋め込み方式`

- **Context**: 要件 2.4 「CDNに依存せずにレンダリングできるようにする」
- **Alternatives Considered**:
  1. `//go:embed` で mermaid.min.js を内包し `/assets/mermaid.min.js` として配信（採用）
  2. サーバーサイドレンダリング（headless Chrome）— Chromium依存でシングルバイナリ要件に違反
  3. CDN経由— オフライン不可、要件違反
- **Selected Approach**: `//go:embed assets/mermaid.min.js`、HTMLテンプレートにローカルパス参照のscriptタグを埋め込み、goldmarkカスタムレンダラーで mermaid コードブロックを `<div class="mermaid">` に変換
- **Rationale**: シングルバイナリ要件を満たしつつ、クライアントサイドレンダリングの実装シンプルさを維持
- **Trade-offs**: バイナリサイズが3-5 MB増加（mermaid.min.js分）。Mermaidバージョン更新にはリコンパイルが必要
- **Follow-up**: goldmark-mermaid の `ScriptURL` 設定オプション確認（内部実装によりカスタム renderer が必要な場合がある）

### Decision: `ライブリロード方式`

- **Context**: 要件 6.2「WebSocketまたはServer-Sent Eventsを使って自動的にリロードさせる」
- **Alternatives Considered**:
  1. SSE（Server-Sent Events）— stdlib のみで実装可能、ブラウザ自動再接続対応（採用）
  2. WebSocket — 双方向通信が可能だが、外部ライブラリが必要で実装コストが高い
- **Selected Approach**: SSE（`net/http` + `http.Flusher`）
- **Rationale**: ライブリロードは一方向通信（サーバー→ブラウザ）のみ必要。SSEはその用途に特化し、外部依存なし
- **Trade-offs**: WebSocketほど汎用性はないが、本用途では十分
- **Follow-up**: HTTP/2環境でのSSEの動作確認

### Decision: `ディレクトリ監視の再帰実装`

- **Context**: 要件 6.1「ドキュメントルート以下のMarkdownファイルの変更・追加・削除を監視する」。fsnotify が再帰監視をネイティブサポートしない
- **Selected Approach**: 起動時に `filepath.WalkDir` で全サブディレクトリを列挙して個別に `watcher.Add()` し、`Create` イベントで新規ディレクトリが追加された場合に動的に登録する
- **Rationale**: fsnotify v1.9.0の現行APIの制約に対する標準的な回避策
- **Trade-offs**: 深いディレクトリ構造では起動時の初期化コストが増加するが、一般的なドキュメントリポジトリでは問題ない規模
- **Follow-up**: 大量サブディレクトリ環境でのパフォーマンス確認

---

## Risks & Mitigations

- **Mermaid.js バイナリサイズ肥大化** — mermaid.min.js（~3-5 MB）がバイナリサイズを増加させる。UPXなどのバイナリ圧縮や、`--no-mermaid` フラグによる軽量版ビルドをリリースで提供することで軽減可能
- **fsnotify 再帰監視の制限** — 起動後に手動で `filepath.WalkDir` を実行し全サブディレクトリを登録。新規ディレクトリの `Create` イベントをキャッチして動的追加することで対処
- **SSE 接続管理のメモリリーク** — SSEBroker にクライアント登録・解除の仕組みを設け、コンテキストキャンセル（`r.Context().Done()`）で接続切断を検知してクリーンアップする
- **パストラバーサル攻撃** — `filepath.Clean` と `docRoot` との前方一致チェックで、ドキュメントルート外へのアクセスを防止する
- **大きなMarkdownファイルのパフォーマンス** — goldmark はストリーミングではなくインメモリ変換を行う。通常のドキュメントサイズ（数MB以内）であれば問題ない。必要に応じてレスポンスキャッシュを追加

---

## References

- [goldmark](https://github.com/yuin/goldmark) — CommonMark準拠Markdownパーサー
- [goldmark-highlighting/v2](https://github.com/yuin/goldmark-highlighting) — Chroma統合シンタックスハイライト拡張
- [goldmark-frontmatter](https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter) — YAML/TOML Front Matter除去拡張
- [goldmark-mermaid](https://pkg.go.dev/go.abhg.dev/goldmark/mermaid) — Mermaid.js統合拡張
- [Chroma v2](https://pkg.go.dev/github.com/alecthomas/chroma/v2) — Goシンタックスハイライトライブラリ v2.23.1
- [mermaid.js v11.12.3](https://github.com/mermaid-js/mermaid/releases) — ダイアグラムレンダリングライブラリ
- [fsnotify v1.9.0](https://github.com/fsnotify/fsnotify) — Goファイルシステム監視ライブラリ
- [github-markdown-css v5.9.0](https://github.com/sindresorhus/github-markdown-css) — GitHub風Markdownスタイルシート
- [Go embed package](https://pkg.go.dev/embed) — 静的アセット埋め込みディレクティブ（Go 1.16+）
- [SSE vs WebSocket](https://ably.com/blog/websockets-vs-sse) — ライブリロード方式の比較
