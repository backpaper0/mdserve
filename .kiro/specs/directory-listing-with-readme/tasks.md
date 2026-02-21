# Implementation Plan

- [x] 1. テンプレートデータ構造にナビゲーション URL フィールドを追加する
  - `PageData` 構造体に `DirListURL` フィールドを追加する（空文字でリンク非表示、非空でリンク表示）
  - `DirListData` 構造体に `IndexURL` フィールドを追加する（空文字でリンク非表示、非空でリンク表示）
  - 既存フィールドのゼロ値挙動と一貫した設計になっていることを確認する
  - _Requirements: 2.1, 3.1_

- [x] 2. ディレクトリハンドラーに `?list` 検出と URL 生成を実装する

- [x] 2.1 (P) `?list` クエリパラメータの検出と README 優先表示の分岐制御を実装する
  - `url.Values.Has("list")` でキーの存在をチェックし、`forceList` フラグとして利用する（値は問わない）
  - `forceList` が `true` のとき、`IndexFile` の有無に関わらず一覧表示ロジックへ進む
  - `forceList` が `false` のとき、`IndexFile` がある場合は従来の README 優先表示を維持する
  - セキュリティチェック（パストラバーサル防止・シンボリックリンク解決）はルーターで完結しているため変更不要
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 2.2 README レンダリング時とファイル一覧表示時のナビゲーション URL を組み立てる
  - README をレンダリングする際に `DirListURL` として `r.URL.Path + "?list"` をセットする
  - ファイル一覧表示かつ `IndexFile` が存在する場合、`IndexURL` として `r.URL.Path` をセットする
  - `IndexFile` が存在しない場合は `IndexURL` を空文字のままにする（リンク非表示）
  - `r.URL.Path` の末尾スラッシュはルーターが保証しているため、追加バリデーション不要
  - _Requirements: 2.1, 2.2, 3.1, 3.2, 3.3_

- [x] 3. HTML テンプレートにナビゲーションリンクを追加する

- [x] 3.1 (P) `page.html` のブレッドクラム直後にファイル一覧リンクを追加する
  - `DirListURL` が非空のときのみ表示される `{{if .DirListURL}}` 条件ブロックを追加する
  - 「ファイル一覧を表示」のテキストリンクをブレッドクラム直下の視覚的に識別しやすい位置に配置する
  - 既存の `nav.breadcrumb` スタイルと一貫した CSS クラスを適用する
  - `go html/template` の自動エスケープが URL に適用されることを確認する（2.1 の後で実施可能）
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 3.2 (P) `dirlist.html` のブレッドクラム直後に README リンクを追加する
  - `IndexURL` が非空のときのみ表示される `{{if .IndexURL}}` 条件ブロックを追加する
  - 「README を表示」のテキストリンクを `page.html` と一貫したスタイルで配置する
  - `IndexURL` が空のとき（README なし）はリンクが表示されないことを確認する（2.1 の後で実施可能）
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 4. テストを実装して全要件の動作を検証する

- [x] 4.1 (P) ディレクトリハンドラーのユニットテストを実装する
  - `?list` あり + `IndexFile` あり → 一覧表示、`DirListData.IndexURL` に値がセットされること
  - `?list` なし + `IndexFile` あり → README 表示、`PageData.DirListURL` に値がセットされること
  - `?list` あり + `IndexFile` なし → 一覧表示、`DirListData.IndexURL` が空文字であること
  - `?list=anything`（値あり）→ `?list` と同様に一覧表示されること（キー存在チェックの確認）
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 3.3_

- [x] 4.2 (P) 統合テストを実装して HTTP レスポンスの HTML 出力を検証する
  - `GET /dir/`（README あり）→ レスポンス HTML に `?list` リンクが含まれること
  - `GET /dir/?list`（README あり）→ レスポンス HTML にファイル一覧エントリと README リンクが含まれること
  - `GET /dir/?list`（README なし）→ レスポンス HTML に README リンクが含まれないこと
  - `GET /dir/?list` の一覧エントリに `README.md` を含む全 `.md` ファイルとサブディレクトリが含まれること
  - ブレッドクラムが一覧ページに正しく表示されること
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 4.4_
