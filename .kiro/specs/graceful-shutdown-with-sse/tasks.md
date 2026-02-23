# Implementation Plan

## Task Format

- [ ] 1. SSEブローカーにシャットダウン機能を追加する

- [ ] 1.1 `Broker` インターフェースに `Shutdown()` を追加し、`broker` 構造体に実装する
  - `sse.Broker` インターフェースに `Shutdown()` メソッドを追加する
  - `broker` 構造体に `shutdown bool` フィールドを追加する
  - `broker.Shutdown()` を実装する: `shutdown = true` をセットし、全クライアントチャネルをクローズしてからマップをクリアする
  - 複数回呼び出しても安全であること（クローズ済みチャネルへの二重クローズを防ぐ）
  - _Requirements: 2.1, 2.3_

- [ ] 1.2 `broker.Register()` にシャットダウン済みガードを追加する
  - `Register()` の先頭で `shutdown` フラグを確認する
  - `shutdown == true` の場合、新規チャネルを生成してすぐにクローズし、そのチャネルを返す
  - SSE ハンドラーが `!ok` を検出して即座に終了できるようにする
  - _Requirements: 2.2_

- [ ] 2. サーバーのシャットダウン順序を修正する
  - `Server.Shutdown()` 内で `http.Server.Shutdown()` の直前に `broker.Shutdown()` を呼び出す順序に変更する
  - 順序: `watcher.Close()` → `broker.Shutdown()` → `http.Server.Shutdown(5s timeout)`
  - `broker` が `nil` の場合は安全にスキップすること
  - 既存の5秒タイムアウトはSSEクローズ失敗時の安全網として維持する
  - _Requirements: 1.1, 2.1, 3.1, 3.2, 4.1_

- [ ] 3. テストを追加する

- [ ] 3.1 (P) `sse.Broker.Shutdown()` のユニットテストを追加する
  - `Shutdown()` が全クライアントチャネルをクローズすること（`!ok` を確認）
  - `Shutdown()` 後の `Register()` がクローズ済みチャネルを返すこと
  - `Shutdown()` 後の `Unregister()` がパニックしないこと（no-op）
  - クライアントなしで `Shutdown()` を呼んでもパニックしないこと
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 3.2 (P) SSE 接続が確立した状態でのサーバーシャットダウン統合テストを追加する
  - `/events` エンドポイントへの SSE 接続を確立した状態でサーバーを起動する
  - `Shutdown()` を呼び出し、タイムアウトなし（`context.DeadlineExceeded` が返らないこと）かつ3秒以内に完了することを確認する
  - `Shutdown()` がエラーなし（`nil`）を返すことを確認する
  - _Requirements: 1.1, 2.1, 2.3, 3.1, 3.2, 4.1_
