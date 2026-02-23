# 実装計画

## タスクリスト

- [ ] 1. テーマCSSファイルを作成する
- [x] 1.1 タイポグラフィ（フォントサイズ・行間）を定義する
  - `assets/theme.css` を新規作成し、`.markdown-body` のベースフォントサイズを 18px に設定する
  - 行間（line-height）を 1.7 に設定して読みやすさを向上する
  - コードブロック（`code`, `pre code`）のフォントサイズを本文比 0.875em に設定し、比率を維持する
  - 見出し（h1〜h6）は `github-markdown.css` のサイズ比を引き継ぐため、ベースサイズの変更のみで自動的に拡大されることを確認する
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 1.2 ライトモードのカラーテーマを定義する
  - `@media (prefers-color-scheme: light)` ブロックで CSS カスタムプロパティを上書きする
  - 背景色（`--bgColor-default`）を `#fff5f7`（パステルピンク）に設定する
  - リンク色（`--fgColor-accent`）を `#d63384`（ピンク）に設定し、`a:hover` に下線とわずかな色変化を追加する
  - 見出し（h1〜h3）のボーダー下線色を `#f3b8cc` に設定してアクセントを与える
  - パンくずナビゲーション（`nav.breadcrumb`）とリンクの色をテーマに合わせて設定する
  - ディレクトリ・READMEナビゲーションリンク（`.dir-list-link`）のスタイルをテーマに沿って設定する
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 3.2_

- [ ] 1.3 ダークモードのカラーテーマを定義する
  - `@media (prefers-color-scheme: dark)` ブロックで CSS カスタムプロパティを上書きする
  - 背景色（`--bgColor-default`）を `#1e0d14`（深みのある赤紫）に設定する
  - リンク色（`--fgColor-accent`）を `#f48fb1`（淡いピンク）に設定し、可読性を確保する
  - 見出しのボーダー下線・パンくずリンク色をダークモード用に調整する
  - テキスト色（`--fgColor-default`）と境界線色（`--borderColor-default`）もテーマに合わせて設定する
  - _Requirements: 2.1, 2.2, 2.4, 2.6, 3.4_

- [ ] 2. HTMLテンプレートを更新してテーマCSSを読み込む
- [ ] 2.1 (P) Markdownページテンプレートにテーマ参照を追加する
  - `page.html` の `highlight.css` の直後に `<link rel="stylesheet" href="/assets/theme.css">` を追加する
  - `<link>` の順序が `github-markdown.css` → `highlight.css` → `theme.css` であることを確認する
  - テンプレート内のインライン `<style>` からパンくず（`nav.breadcrumb`）スタイルを削除し、`theme.css` 側でカバーする
  - `body` の padding・max-width は `theme.css` に移管するか、インラインのまま維持するか判断する（`github-markdown.css` を非変更にするという制約のもとで）
  - _Requirements: 1.1, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 3.1, 3.3_

- [ ] 2.2 (P) ディレクトリ一覧テンプレートにテーマ参照を追加する
  - `dirlist.html` の `github-markdown.css` の直後に `<link rel="stylesheet" href="/assets/theme.css">` を追加する
  - `<link>` の順序が `github-markdown.css` → `theme.css` であることを確認する（`highlight.css` は不要）
  - テンプレート内のインライン `<style>` からパンくずスタイルを削除し、`theme.css` 側でカバーする
  - _Requirements: 1.3, 2.5, 3.1, 3.2, 3.3_

- [ ] 3. テストを追加・更新する
- [ ] 3.1 (P) テンプレートの出力に `theme.css` の参照が含まれることをテストする
  - `tmpl_test.go` に `page.html` の出力に `theme.css` の `<link>` タグが含まれることを検証するテストを追加する
  - `dirlist.html` の出力にも同様のテストを追加する
  - `<link>` の順序（`github-markdown.css` より後）が正しいことを確認する
  - _Requirements: 3.1, 3.3_

- [ ] 3.2 (P) `/assets/theme.css` のアセット配信を統合テストで検証する
  - `integration_test.go` に `/assets/theme.css` への GET リクエストが HTTP 200 を返すことを確認するテストを追加する
  - レスポンスのContent-Typeが `text/css` であることを確認する
  - _Requirements: 3.3_
