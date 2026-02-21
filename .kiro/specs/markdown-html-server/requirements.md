# 要件定義書

## プロジェクト説明（入力）
mermaid.jsの図を含むmarkdownファイルをHTML化して見せるサーバーアプリケーション。シングルバイナリで提供され、ローカルで簡単に動かせられて、markdownファイル群が格納されたディレクトリをすぐにウェブサイト化できる。

## はじめに

本仕様は、MarkdownファイルをHTMLに変換してブラウザで閲覧できるローカルWebサーバーアプリケーション「Markdown HTML Server」の要件を定義する。Mermaid.jsによる図のレンダリングをサポートし、シングルバイナリとして配布・起動することで、MarkdownドキュメントのWebサイト化を即座に実現する。

## 要件

### 要件 1: Markdownレンダリング

**目的:** ドキュメント閲覧者として、MarkdownファイルをブラウザでHTMLとして表示したい。そうすることで、Markdownのシンタックスを意識せずに読みやすい形式でコンテンツを確認できる。

#### 受け入れ基準
1. When ブラウザがMarkdownファイルのURLにアクセスした, the Markdown HTML Server shall MarkdownをHTMLに変換して返却する
2. The Markdown HTML Server shall 見出し・リスト・テーブル・コードブロック・リンク・画像・太字・斜体を含む標準Markdownシンタックスをレンダリングする
3. The Markdown HTML Server shall シンタックスハイライトを適用したコードブロックをレンダリングする
4. When MarkdownファイルにYAML Front Matterが含まれる, the Markdown HTML Server shall Front Matterを除外してHTMLをレンダリングする
5. The Markdown HTML Server shall 変換後のHTMLに適切なスタイルシートを適用して読みやすいレイアウトで表示する

---

### 要件 2: Mermaid.js図のレンダリング

**目的:** ドキュメント閲覧者として、Markdownファイル中のMermaid記法で書かれた図をブラウザ上でSVGとして表示したい。そうすることで、フローチャートやシーケンス図などを視覚的に確認できる。

#### 受け入れ基準
1. When ブラウザがMermaid記法のコードブロック（```mermaid）を含むMarkdownファイルを取得した, the Markdown HTML Server shall そのコードブロックをMermaid.jsで描画可能なHTML要素として出力する
2. The Markdown HTML Server shall フローチャート・シーケンス図・クラス図・ガントチャート・状態遷移図を含むMermaid.jsがサポートする図の種類をレンダリングする
3. If Mermaid記法に構文エラーがある, the Markdown HTML Server shall エラーメッセージを図の代わりに表示してページ全体のレンダリングを継続する
4. The Markdown HTML Server shall Mermaid.jsをサーバー側に埋め込み、ブラウザがCDNに依存せずにレンダリングできるようにする

---

### 要件 3: HTTPサーバー機能

**目的:** ユーザーとして、指定したディレクトリをすぐにHTTPサーバーとして公開したい。そうすることで、ブラウザからMarkdownファイルを閲覧できる環境をすぐに整えられる。

#### 受け入れ基準
1. The Markdown HTML Server shall 指定されたディレクトリをドキュメントルートとして、HTTPサーバーを起動する
2. When サーバー起動コマンドが実行された, the Markdown HTML Server shall デフォルトでポート3333でリッスンを開始し、起動アドレスをコンソールに表示する
3. Where コマンドラインオプションにポート番号が指定されている, the Markdown HTML Server shall 指定されたポートでリッスンする
4. The Markdown HTML Server shall `.md`拡張子のファイルリクエストに対してHTMLをレスポンスし、その他のファイル（画像・PDFなど）はそのままレスポンスする
5. If 存在しないファイルへのリクエストを受信した, the Markdown HTML Server shall HTTP 404レスポンスを返す
6. The Markdown HTML Server shall Ctrl+Cなどのシグナルを受信したとき、正常にシャットダウンする

---

### 要件 4: ディレクトリ閲覧とナビゲーション

**目的:** ドキュメント閲覧者として、ディレクトリ内のMarkdownファイル一覧をブラウザで確認してファイルを選択したい。そうすることで、複数のMarkdownファイルを簡単にナビゲートできる。

#### 受け入れ基準
1. When ブラウザがディレクトリのURLにアクセスした, the Markdown HTML Server shall そのディレクトリに含まれるMarkdownファイルとサブディレクトリの一覧をHTMLページとして表示する
2. The Markdown HTML Server shall ファイル一覧のHTMLページに各ファイルへのリンクを含め、クリックでMarkdownのHTMLプレビューに遷移できるようにする
3. When ディレクトリのURLにアクセスした, the Markdown HTML Server shall `README.md`・`index.md`の順で優先的にインデックスファイルを探し、最初に見つかったファイルをレンダリングしてインデックスページとして表示する
4. The Markdown HTML Server shall レンダリングされたHTMLページに、ドキュメントルートからの相対パスによるパンくずリストを表示する
5. The Markdown HTML Server shall `.md`拡張子を持たないファイルをディレクトリ一覧から除外する（画像などのアセットは除く）

---

### 要件 5: シングルバイナリ配布と起動設定

**目的:** ユーザーとして、インストール作業なしに単一の実行ファイルをダウンロードしてすぐに使いたい。そうすることで、環境構築の手間なく即座にMarkdownビューアーとして活用できる。

#### 受け入れ基準
1. The Markdown HTML Server shall 依存ライブラリを含む単一の実行バイナリとしてビルド・配布できる
2. The Markdown HTML Server shall `mdserve [ディレクトリパス]`の形式で、サーブするディレクトリを引数として受け取る
3. When ディレクトリ引数が省略された, the Markdown HTML Server shall カレントディレクトリをドキュメントルートとして使用する
4. The Markdown HTML Server shall `--port`オプションでポート番号を指定できる
5. The Markdown HTML Server shall `--help`オプションで使い方を表示する
6. If 指定されたディレクトリが存在しない, the Markdown HTML Server shall エラーメッセージを表示して起動を中止する

---

### 要件 6: ファイル変更の自動検知とブラウザリロード

**目的:** ドキュメント作成者として、Markdownファイルを編集した際にブラウザを手動でリロードせずに変更内容を確認したい。そうすることで、ライブプレビューによる快適な編集体験が得られる。

#### 受け入れ基準
1. While サーバーが起動中である, the Markdown HTML Server shall ドキュメントルート以下のMarkdownファイルの変更・追加・削除を監視する
2. When 監視対象のファイルが変更された, the Markdown HTML Server shall 開いているブラウザのページをWebSocketまたはServer-Sent Eventsを使って自動的にリロードさせる
3. The Markdown HTML Server shall ライブリロード用のスクリプトをレンダリングされたHTMLページに自動的に埋め込む
4. Where `--no-watch`オプションが指定されている, the Markdown HTML Server shall ファイル監視とライブリロードを無効にして動作する
