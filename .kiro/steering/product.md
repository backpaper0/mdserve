# Product Overview

`mdserve` は、MarkdownファイルをHTMLに変換してブラウザで閲覧できるローカルWebサーバーです。シングルバイナリとして配布され、インストール不要で即座にMarkdownディレクトリをWebサイト化できます。

## Core Capabilities

- **Markdownレンダリング** - 標準Markdown（テーブル・打ち消し線・タスクリスト）をシンタックスハイライト付きでHTML変換
- **Mermaid.js対応** - コードフェンス `mermaid` をブラウザ上でSVG図に変換
- **ディレクトリ閲覧** - `README.md`/`index.md` を優先表示、なければMarkdownファイル一覧を表示
- **ライブリロード** - ファイル変更をSSE（Server-Sent Events）でブラウザに通知して自動リロード
- **シングルバイナリ** - 静的アセット・テンプレートをすべて埋め込んで単一実行ファイルで配布

## Target Use Cases

- ローカルでMarkdownドキュメントをプレビューしながら執筆
- `docs/` や `notes/` ディレクトリをすばやくWebブラウズ
- Mermaid図を含む技術ドキュメントのレンダリング確認

## Value Proposition

依存関係ゼロのシングルバイナリ。`mdserve` を起動するだけで、任意のディレクトリのMarkdownファイルをライブリロード付きのWebサイトとして閲覧できる。

---
_Focus on patterns and purpose, not exhaustive feature lists_
