# 実装タスク

## Markdown HTML Server

---

- [x] 1. プロジェクト初期構成とGoモジュールセットアップ
  - `go mod init` でGoモジュールを初期化し、goldmark・goldmark-highlighting・goldmark-frontmatter・fsnotify・Chromaなど全依存パッケージを `go.mod` に追加する
  - 設計書のアーキテクチャに従い `cmd/mdserve/`、`internal/renderer/`、`internal/server/`、`internal/watcher/`、`internal/sse/`、`internal/dirlist/`、`internal/tmpl/`、`assets/` のディレクトリ構造を作成する
  - mermaid.min.js（v11.12.3）を公式リリースからダウンロードし `assets/` ディレクトリに配置する
  - github-markdown-css（v5.9.0）を公式ソースから取得し `assets/` ディレクトリに配置する
  - `go build ./cmd/mdserve` が通ることを確認し、シングルバイナリとして出力できる状態を確立する
  - _Requirements: 5.1_

- [x] 2. Markdownレンダリングパイプラインの構築

- [x] 2.1 標準Markdown構文のHTMLレンダリング実装
  - goldmarkを初期化し、見出し・段落・リスト・テーブル・コードブロック・リンク・画像・太字・斜体を含む標準Markdown構文をHTMLに変換する機能を実装する
  - goldmark標準拡張（Table・Strikethrough・TaskList）を有効化する
  - ファイルパスを受け取りHTMLバイト列を返す `Renderer` インターフェースを定義し、goldmarkを使った実装を提供する
  - _Requirements: 1.1, 1.2_

- [x] 2.2 YAML Front Matter除去の実装
  - `go.abhg.dev/goldmark/frontmatter` 拡張をgoldmarkパイプラインに組み込む
  - `.md` ファイル先頭のYAML Front MatterがHTML出力に含まれないことを確認する
  - _Requirements: 1.4_

- [x] 2.3 コードブロックのシンタックスハイライト実装
  - `goldmark-highlighting/v2` とChroma v2をgoldmarkパイプラインに統合し、コードブロックにGitHubスタイルのシンタックスハイライトを適用する
  - `html.WithClasses(true)` でCSSクラスベースのハイライトを使用し、Chroma CSSをファイルとして生成して `assets/` ディレクトリに保存する
  - _Requirements: 1.3_

- [x] 2.4 Mermaid.jsダイアグラムのHTMLレンダリング対応
  - ````mermaid` コードフェンスを検出して `<div class="mermaid">` タグに変換するgoldmarkカスタム拡張またはgoldmark-mermaidライブラリを組み込む
  - Mermaid構文エラーはクライアントサイド（mermaid.js）が表示を担うため、エラーが発生してもページ全体のレンダリングを継続する
  - フローチャート・シーケンス図・クラス図・ガントチャート・状態遷移図はmermaid.js v11が対応するため、サーバー側では追加実装不要
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 3. (P) 静的アセット管理とHTMLテンプレートエンジンの実装
  *(Task 2と並列実行可能。アセット配置と埋め込み設定はTask 1完了後に開始できる)*

- [x] 3.1 静的アセットのバイナリ埋め込みとアセット配信ハンドラー
  - `//go:embed assets/` ディレクティブでmermaid.min.js・github-markdown.css・chroma highlight.cssをGoバイナリに埋め込む
  - `embed.FS` を `http.FileServer(http.FS(...))` でサーブし、`/assets/*` パスで各アセットを配信するハンドラーを実装する
  - _Requirements: 2.4, 1.5_

- [x] 3.2 HTMLページテンプレートエンジンの実装
  - `html/template` を使い、Markdownページ全体（`<!DOCTYPE html>` 〜 `</html>`）を生成するテンプレートを `internal/tmpl/` に作成する
  - テンプレートに `/assets/github-markdown.css`・`/assets/highlight.css` の参照、Markdownコンテンツ（`<article class="markdown-body">` でラップ）、パンくずリスト表示を含める
  - Mermaid初期化スクリプト（`/assets/mermaid.min.js` 読み込み + `mermaid.initialize({ startOnLoad: true })`）をテンプレートに埋め込む
  - `liveReload` フラグが有効な場合、SSEクライアントスクリプト（`new EventSource('/events')` + `location.reload()` 呼び出し）をHTML末尾に自動挿入する
  - ディレクトリ一覧ページ用テンプレートも作成し、`.md` ファイルとサブディレクトリの一覧をリンク付きで表示する
  - _Requirements: 1.5, 4.2, 4.4, 6.3_

- [x] 4. (P) CLIとHTTPサーバー基盤の実装
  *(Task 2・3と並列実行可能。Config構造体とServer骨格はTask 1完了後に開始できる)*

- [x] 4.1 コマンドライン引数解析とアプリケーション設定
  - `flag` パッケージで `--port`（デフォルト3333）・`--no-watch`・`--help` フラグと、サーブするディレクトリの位置引数を解析する
  - ディレクトリ引数が省略された場合は `os.Getwd()` でカレントディレクトリを使用する
  - 指定ディレクトリが存在しない場合は、エラーメッセージを標準エラー出力に表示して終了する（`os.Exit(1)`）
  - 解析結果を `Config{DocRoot, Port, NoWatch}` 構造体にまとめる
  - _Requirements: 3.2, 3.3, 5.2, 5.3, 5.4, 5.5, 5.6, 6.4_

- [x] 4.2 HTTPサーバー起動とグレースフルシャットダウンの実装
  - `http.ListenAndServe` で指定ポートにバインドし、起動アドレス（例: `Serving /path on http://localhost:3333`）をコンソールに出力する
  - `os/signal` で `SIGINT`・`SIGTERM` を捕捉し、`http.Server.Shutdown(context.WithTimeout(..., 5*time.Second))` でアクティブな接続を安全に終了する
  - _Requirements: 3.1, 3.2, 3.6_

- [x] 5. リクエストルーティングとファイル配信ハンドラーの実装

- [x] 5.1 URLパス解決とリクエストディスパッチの実装
  - URLパスをドキュメントルート以下の実ファイルパスに変換し、`.md` ファイル・ディレクトリ・その他ファイルをそれぞれ対応するハンドラーに委譲するルーターを実装する
  - `filepath.Clean` とドキュメントルートへの前方一致チェック、`filepath.EvalSymlinks` でパストラバーサル（`../` を含むパスなど）を防止する
  - ファイルが存在しない場合はHTTP 404レスポンスを返す
  - _Requirements: 1.1, 3.4, 3.5_

- [x] 5.2 (P) Markdownファイルの変換・レスポンスハンドラー
  - `.md` ファイルへのリクエストを受けてRendererでHTMLフラグメントに変換し、TemplateEngineでページ全体にラップして `text/html; charset=utf-8` でレスポンスする
  - _Requirements: 1.1_

- [x] 5.3 (P) ディレクトリ閲覧とナビゲーションハンドラーの実装
  - ディレクトリへのリクエストを受けて、`README.md` → `index.md` の優先順でインデックスファイルを探し、見つかった場合はそのMarkdownをレンダリングして表示する
  - インデックスファイルが存在しない場合はディレクトリ内の `.md` ファイルとサブディレクトリの一覧をHTMLページとして表示し、非 `.md` ファイルは一覧から除外する
  - ドキュメントルートからの相対パスに基づきパンくずリストデータを生成し、テンプレートに渡す
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 5.4 (P) 静的ファイルの配信ハンドラー
  - `.md` 以外のファイル（画像・PDF・動画など）を `http.FileServer` でドキュメントルートから直接配信する
  - _Requirements: 3.4_

- [x] 6. (P) SSEライブリロードとファイル監視の実装
  *(Task 5と並列実行可能。Task 4完了後に開始できる)*

- [x] 6.1 fsnotifyを用いた再帰的なファイル変更監視
  - `github.com/fsnotify/fsnotify` を使って `filepath.WalkDir` でドキュメントルート以下の全サブディレクトリを起動時に監視対象へ登録する
  - `Create` イベントで新規ディレクトリが作成された場合は動的に監視対象へ追加し、再帰監視を維持する
  - `Write`・`Create`・`Remove`・`Rename` イベントを検知してSSEブローカーの `Broadcast()` を呼び出す（`Chmod` イベントは無視する）
  - `Config.NoWatch == true` の場合はファイル監視全体をスキップしてSSEブローカーへの通知を行わない
  - _Requirements: 6.1, 6.4_

- [x] 6.2 SSEブローカーとSSEエンドポイントの実装
  - 複数のSSEクライアント接続を管理するブローカー（`Register`・`Unregister`・`Broadcast` メソッド）を実装し、`sync.Mutex` でクライアントマップへの並行アクセスを保護する
  - `Broadcast` はノンブロッキング送信（`select` + デフォルトケース）で実装し、クライアント切断による遅延を防ぐ
  - `/events` エンドポイントで `Content-Type: text/event-stream` を返し、ファイル変更通知を `data: reload\n\n` イベントとしてブラウザに送信する
  - SSE接続切断時（`r.Context().Done()`）にクライアントを登録解除してチャンネルをクリーンアップする
  - 15秒ごとにキープアライブコメント（`: keepalive\n\n`）を送信してプロキシ経由での接続を維持する
  - _Requirements: 6.2_

- [x] 7. 統合テストと動作検証

- [x] 7.1 Markdownレンダリングパイプラインのユニットテスト
  - 見出し・テーブル・コードブロック・リンク・画像を含むMarkdownが期待通りのHTMLに変換されることを検証する
  - Front Matter付きMarkdownでFront MatterがHTML出力に含まれないことを検証する
  - シンタックスハイライトが適用されたコードブロックのHTML出力にChromaのCSSクラスが含まれることを検証する
  - Mermaidコードブロックが `<div class="mermaid">` タグに変換されることを検証する
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1_

- [x] 7.2 HTTPサーバーE2Eテスト
  - テスト用一時ディレクトリを作成し実際にHTTPサーバーを起動して、`.md` ファイルへのGETリクエストが200 HTMLを返すことをHTTPクライアントで検証する
  - ディレクトリへのリクエストでインデックスファイル優先表示（`README.md`・`index.md`）が機能することを検証する
  - 存在しないパスへのリクエストがHTTP 404を返すことを検証する
  - パストラバーサル攻撃（`../` を含むパス）が適切に拒否されることを検証する
  - `/assets/mermaid.min.js` と `/assets/github-markdown.css` のリクエストが200で返ることを確認する
  - _Requirements: 1.1, 3.1, 3.4, 3.5, 4.1, 4.3, 2.4_

- [x] 7.3 SSEライブリロードの統合テスト
  - テスト用.mdファイルをサーブした状態でSSE接続を確立し、ファイルを更新した後に `data: reload` イベントが受信されることを検証する
  - `--no-watch` モードで起動した場合にファイル変更後もSSEイベントが発行されないことを検証する
  - _Requirements: 6.1, 6.2, 6.4_
