# リサーチ・設計判断ログ

---
**目的**: UIデザイン改善フィーチャーにおける調査結果・設計判断の根拠を記録する。

---

## サマリー

- **フィーチャー**: `ui-design-improvement`
- **ディスカバリースコープ**: Extension（既存テンプレート・CSSの拡張）
- **主な発見事項**:
  - `github-markdown.css` は CSS カスタムプロパティ（`--bgColor-default`, `--fgColor-default` 等）を使用しており、外部から上書き可能な設計になっている
  - `//go:embed assets` ディレクティブで `assets/` ディレクトリ全体が埋め込まれるため、新ファイルを追加するだけで `/assets/theme.css` として自動配信される
  - `page.html` と `dirlist.html` 双方にインラインスタイルが存在し、共通化の余地がある

## リサーチログ

### アセット配信の仕組み

- **調査のきっかけ**: 新しいCSSファイルをどのように配信するかの確認
- **参照元**: `embed.go`, `internal/server/server.go`
- **発見事項**:
  - `embed.go` が `//go:embed assets` で `assets/` ディレクトリ全体をバイナリに埋め込む
  - `server.go` が `/assets/` プレフィックスで `NewAssetHandler` を使ってファイルを配信
  - `assets/` に新ファイルを追加すれば、Go コード変更不要で自動的に利用可能になる
- **インパクト**: 新規 CSSファイルの追加のみで要件を満たせる。Go コード変更が不要なため影響範囲が最小限

### `github-markdown.css` のCSS変数設計

- **調査のきっかけ**: 既存スタイルを壊さずにカラーテーマを上書きする方法の確認
- **参照元**: `assets/github-markdown.css`
- **発見事項**:
  - `@media (prefers-color-scheme: light)` と `dark` のブロックで CSS カスタムプロパティを定義
  - ライトモード: `--bgColor-default: #ffffff`（白）、`--fgColor-accent: #0969da`（青）
  - ダークモード: `--bgColor-default: #0d1117`（濃紺）、`--fgColor-accent: #4493f8`（青）
  - `.markdown-body` のベースフォントサイズ: `font-size: 16px`（ハードコード）
- **インパクト**: CSS カスタムプロパティを後続の `<link>` で上書きすることで、`github-markdown.css` を変更せずにテーマを適用できる

### テンプレートの現状分析

- **調査のきっかけ**: `page.html` と `dirlist.html` の共通化・差異確認
- **参照元**: `internal/tmpl/templates/page.html`, `dirlist.html`
- **発見事項**:
  - 両テンプレートは同一のインラインスタイル（body, breadcrumb）を持つ — 重複あり
  - `dirlist.html` は `highlight.css` を読み込んでいない（コードハイライト不要なため正しい）
  - ナビゲーションリンク（`dir-list-link`）のスタイル未定義
- **インパクト**: `theme.css` に共通スタイルをまとめ、インラインスタイルを削除（または最小化）することでDRYを実現できる

## アーキテクチャパターン評価

| 選択肢 | 説明 | 強み | リスク・制限 | 備考 |
|--------|------|------|-------------|------|
| A: `github-markdown.css` 直接編集 | vendored CSSに直接変更を加える | 設定ファイル数増加なし | 上流との差分管理が困難、vendored file変更は避けるべき | 非推奨 |
| B: `assets/theme.css` 新規作成 | 上書き専用CSSファイルを追加 | 責務分離明確、Go変更不要、両テンプレートで共有可能 | `<link>` の順序に依存（必ず後続に配置） | **採用** |
| C: インラインスタイル拡張 | 各テンプレートの `<style>` ブロックを拡張 | 設定ファイル不要 | DRYに反する、両テンプレートの同期保守が必要 | 非推奨 |

## 設計判断

### 判断: `assets/theme.css` を新規作成してテーマを集約する

- **コンテキスト**: 既存の `github-markdown.css`（vendored）を変更せずに、フォントサイズ拡大とカラーテーマを実現する必要がある
- **検討した代替案**:
  1. `github-markdown.css` を直接編集 — vendored fileの変更はメンテナンス上問題
  2. インラインスタイルを両テンプレートに追加 — DRYに反し、同期が必要
- **採用アプローチ**: `assets/theme.css` を新規作成し、`github-markdown.css` の後に読み込むことでCSS変数とプロパティを上書き
- **根拠**: `//go:embed assets` の仕組みにより追加コストゼロ。CSS変数の上書きは設計的に安全
- **トレードオフ**: `<link>` 順序の厳守が必要（先に `github-markdown.css`、後に `theme.css`）
- **フォローアップ**: テンプレートの `<link>` 順序をテストで確認すること

### 判断: カラーパレットにパステルピンク系を採用

- **コンテキスト**: ユーザーが「かわいくしてほしい」と明示。参照URLはリコリス・リコイルの上映イベントページ（アニメ：ピンク・赤系の配色）
- **採用アプローチ**: ライトモードはパステルピンク系（`#fff5f7` 背景、`#d63384` リンク）、ダークモードは深みのある赤紫系（`#1e0d14` 背景、`#f48fb1` リンク）
- **根拠**: 参照ページのテイストに合わせつつ、可読性を損なわない明度・彩度に調整
- **トレードオフ**: 汎用性より個性を優先。特定テイストのため好みが分かれる可能性あり

## リスクと対策

- **CSSカスケードの順序ミス** — `theme.css` の `<link>` が `github-markdown.css` より前に来ると上書きが効かない。テンプレート修正時に順序を必ず確認する
- **highlight.css との色衝突** — `highlight.css`（Chroma）の背景色 `#f7f7f7` がライトモードのパステルピンク背景と微妙に異なる可能性。コードブロックの背景は独立しているため許容範囲とする
- **ダークモード可読性** — 深い暗色背景にパステルリンク色を配置した際のコントラスト比をWCAG AA（4.5:1以上）の観点で設計時に確認が必要

## 参考

- MDN CSS カスケード: https://developer.mozilla.org/ja/docs/Web/CSS/Cascade
- WCAG 2.1 カラーコントラスト基準: https://www.w3.org/TR/WCAG21/#contrast-minimum
- Go embed パッケージ: https://pkg.go.dev/embed
