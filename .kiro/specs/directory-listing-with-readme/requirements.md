# Requirements Document

## Project Description (Input)
README.mdがある場合でもディレクトリ内のファイルを一覧表示できる導線を整備する。

## はじめに

現在の mdserve は、ディレクトリに `README.md` または `index.md` が存在する場合、自動的にその内容を優先表示する。このため、ユーザーはディレクトリ内の他のファイルを一覧から確認する手段がない。本機能は、`README.md` が存在する場合でもファイル一覧へアクセスできる UI 導線と URL アクセス手段を整備する。

## Requirements

### Requirement 1: クエリパラメータによるファイル一覧アクセス

**Objective:** Markdown ドキュメントの閲覧者として、`?list` クエリパラメータを URL に付与することでディレクトリのファイル一覧を表示したい。そうすることで、`README.md` が存在するディレクトリでも全ファイルを確認できる。

#### Acceptance Criteria

1. When ユーザーがディレクトリ URL に `?list` クエリパラメータを付与してアクセスしたとき, the mdserve shall `README.md` や `index.md` の有無に関わらず、ディレクトリのファイル一覧ページを表示する
2. When ユーザーが `?list` なしのディレクトリ URL にアクセスしたとき, the mdserve shall 現行通り `README.md` / `index.md` を優先表示し、存在しない場合はファイル一覧を表示する
3. The mdserve shall `?list` クエリパラメータが付与されたディレクトリ URL に対してセキュリティチェック（パストラバーサル防止・シンボリックリンク解決）を変わらず適用する

---

### Requirement 2: READMEページからファイル一覧への導線

**Objective:** Markdown ドキュメントの閲覧者として、`README.md` が表示されているページから一覧ページへのリンクをクリックして遷移したい。そうすることで、URL を手動で編集することなくファイル一覧を確認できる。

#### Acceptance Criteria

1. While `README.md` または `index.md` がインデックスファイルとしてレンダリングされているとき, the mdserve shall ディレクトリ一覧ページへ遷移するリンク（「ファイル一覧」など）をページ内に表示する
2. When ユーザーがそのリンクをクリックしたとき, the mdserve shall 同一ディレクトリの `?list` URL（例: `/path/to/dir/?list`）に遷移する
3. The mdserve shall そのリンクをページコンテンツの一部として視覚的に識別しやすい位置（ヘッダー、ページ先頭など）に表示する

---

### Requirement 3: ファイル一覧ページからREADMEへの導線

**Objective:** Markdown ドキュメントの閲覧者として、`?list` で表示されているファイル一覧ページから `README.md` へ簡単に戻れるリンクが欲しい。そうすることで、ファイル一覧と README コンテンツをスムーズに行き来できる。

#### Acceptance Criteria

1. While `?list` クエリパラメータでファイル一覧ページが表示されており、かつディレクトリに `README.md` または `index.md` が存在するとき, the mdserve shall ページ内にインデックスファイルへ戻るリンク（「README を表示」など）を表示する
2. When ユーザーがそのリンクをクリックしたとき, the mdserve shall `?list` なしの同一ディレクトリ URL（例: `/path/to/dir/`）に遷移する
3. If ディレクトリに `README.md` も `index.md` も存在しないとき, the mdserve shall インデックスファイルへ戻るリンクを表示しない

---

### Requirement 4: ファイル一覧ページの内容

**Objective:** Markdown ドキュメントの閲覧者として、`?list` で表示されるファイル一覧に全 Markdown ファイルとサブディレクトリが含まれている必要がある。そうすることで、インデックスファイル（`README.md`）も含めてディレクトリ全体の構造を把握できる。

#### Acceptance Criteria

1. The mdserve shall `?list` 表示のファイル一覧に、`README.md` / `index.md` を含むすべての `.md` ファイルとサブディレクトリのエントリを表示する
2. The mdserve shall 既存のファイル一覧と同様に、エントリをアルファベット順またはファイルシステムの並び順で表示する
3. The mdserve shall `?list` 表示のファイル一覧に対しても、現行のブレッドクラムナビゲーションを適切に表示する
4. When ユーザーが一覧内のファイルリンクをクリックしたとき, the mdserve shall 該当ファイルの Markdown レンダリングページに遷移する
