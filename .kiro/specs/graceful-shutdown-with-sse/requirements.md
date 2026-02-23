# Requirements Document

## Project Description (Input)
ホットリロードのためのSSEが接続されている状態でもmdserveを落としたい。

現在は1つでもSSEが接続されている場合、`control + c`を押しても終了できない。
mdserve側からSSE接続をクローズして、プロセスを終了できるようにしてほしい。

## Introduction

mdserveはライブリロード機能のためにSSE（Server-Sent Events）を使用してブラウザとの接続を維持する。現在の実装では、SSEクライアントが接続中の場合、Ctrl+Cによる終了シグナルを受け取っても正常に終了できない。本要件では、mdserveがOSシグナルを受信した際に、アクティブなSSE接続をサーバー側からクローズし、プロセスを正常終了できるグレースフルシャットダウン機能を定義する。

## Requirements

### Requirement 1: OSシグナルの検出とシャットダウン開始

**Objective:** ユーザーとして、Ctrl+Cを押したとき（またはSIGTERMを送ったとき）、SSEクライアントが接続中であっても mdserve が終了してほしい。そうすることで、開発中に確実にプロセスを止められる。

#### Acceptance Criteria
1. When SIGINT（Ctrl+C）またはSIGTERMシグナルを受信した, the mdserve shall グレースフルシャットダウンシーケンスを開始する
2. The mdserve shall プロセス起動時からOSシグナル（SIGINT、SIGTERM）をリッスンする

---

### Requirement 2: アクティブなSSE接続のクローズ

**Objective:** 開発者として、mdserve終了時にすべてのSSEクライアント接続がサーバー側から適切にクローズされてほしい。そうすることで、ブラウザ側のSSEセッションが放置されず、プロセスがブロックされない。

#### Acceptance Criteria
1. When グレースフルシャットダウンが開始された, the mdserve shall アクティブなすべてのSSEクライアント接続をクローズする
2. When グレースフルシャットダウンが開始された, the mdserve shall 新規のSSEクライアント接続を受け付けない
3. While SSEクライアントのクローズ処理が進行中, the mdserve shall 全接続のクローズが完了するまで次のシャットダウンステップへ進まない

---

### Requirement 3: HTTPサーバーの正常停止

**Objective:** 開発者として、HTTPサーバーが進行中のリクエストを適切に処理してから終了してほしい。そうすることで、ファイルレスポンスなどが途中で切断されることなく正常に完了できる。

#### Acceptance Criteria
1. When SSE接続のクローズが完了した, the mdserve shall HTTPサーバーに対してグレースフルシャットダウンを実行し、進行中のリクエストの完了を待つ
2. When HTTPサーバーのシャットダウンが完了した, the mdserve shall プロセスを終了する

---

### Requirement 4: シャットダウンタイムアウト

**Objective:** ユーザーとして、シャットダウン処理が何らかの理由でブロックされ続けた場合でも、一定時間内に mdserve が強制終了してほしい。そうすることで、プロセスが永遠に残り続けることがない。

#### Acceptance Criteria
1. When グレースフルシャットダウン開始から一定時間（デフォルト: 5秒）が経過してもシャットダウンが完了しない, the mdserve shall プロセスを強制終了する
2. When タイムアウトにより強制終了する, the mdserve shall その旨を標準エラー出力またはログに出力する
