# Research & Design Decisions

---
**Purpose**: 調査結果・アーキテクチャ評価・設計判断の根拠を記録する。

---

## Summary

- **Feature**: `graceful-shutdown-with-sse`
- **Discovery Scope**: Extension（既存 SSE ブローカー・サーバーへの機能追加）
- **Key Findings**:
  - `http.Server.Shutdown()` は SSE 接続を「アクティブなリクエスト」として扱い、クローズされるまで待機する
  - Go の HTTP サーバーは `Shutdown()` 呼び出し時に `r.Context()` を自動的にキャンセルしない（`BaseContext` が設定されていない場合）
  - SSE ハンドラーは `case _, ok := <-ch: if !ok { return }` でチャネルクローズを検出できる — この仕組みを利用してブローカー側からシャットダウンを伝播できる

## Research Log

### Go `http.Server.Shutdown()` の動作確認

- **Context**: SSEが接続中でCtrl+Cを押しても終了しない原因調査
- **Sources Consulted**: Go 1.24 `net/http/server.go` ソースコード分析
- **Findings**:
  - `Shutdown()` はリスナーをクローズし、アイドル接続をクローズし、アクティブ接続が idle 状態に戻るまでポーリングで待機する
  - SSE 接続はアクティブな HTTP 接続として扱われ、`closeIdleConns()` の対象にならない
  - `Shutdown()` 内部では `cancelCtx()` が呼ばれるが、個別リクエストの `r.Context()` は自動キャンセルされない（`BaseContext` が cancellable でないため）
  - 結果: タイムアウト（5秒）まで `Shutdown()` がブロックし、`context.DeadlineExceeded` を返す
- **Implications**: SSE チャネルをブローカー側から明示的にクローズすることで、ハンドラーを自律的に終了させる必要がある

### SSE ハンドラーの終了経路分析

- **Context**: ブローカーからシャットダウンを伝播する最適な方法の調査
- **Sources Consulted**: `internal/server/handlers.go` の `NewSSEHandler` 実装
- **Findings**:
  - SSE ハンドラーは `for { select { ... } }` ループを持ち、3つの終了経路がある
    1. `case <-r.Context().Done()`: リクエストコンテキストのキャンセル（クライアント切断時）
    2. `case _, ok := <-ch: if !ok { return }`: チャネルがクローズされた場合
    3. 書き込みエラー: `fmt.Fprintf` がエラーを返した場合
  - 経路2（チャネルクローズ）がブローカー主導のシャットダウンに最適
- **Implications**: `broker.Shutdown()` が全クライアントチャネルをクローズすれば、全 SSE ハンドラーが自律的に終了する

### 既存インターフェース・コード構造の確認

- **Context**: 変更範囲の最小化と既存パターンとの整合性確認
- **Findings**:
  - `sse.Broker` インターフェース: `Register()`, `Unregister()`, `Broadcast()` の3メソッド
  - `broker` struct: `sync.Mutex` + `map[chan struct{}]struct{}` で実装
  - `Broker` インターフェースの実装は `sse.New()` が返す `*broker` のみ（モック実装なし）
  - `Server.Shutdown()`: `watcher.Close()` → `http.Server.Shutdown(5s)` の順
  - テストで `sse.Broker` を直接使用している箇所は `sse_test.go` と `server/sse_test.go`（いずれも `sse.New()` 経由）
- **Implications**: `Broker` インターフェースへの `Shutdown()` 追加は破壊的変更だが、外部に公開されていないため影響範囲は `internal/` 内のみ

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| Broker.Shutdown() 追加 | `Broker` インターフェースに `Shutdown()` メソッドを追加し、全チャネルをクローズ | シンプル・明示的・既存パターンに沿う | インターフェース変更（内部のみ） | 採用 |
| context 経由の伝播 | `Register(ctx context.Context)` でコンテキストを受け取り、ctx キャンセルで終了 | 慣用的な Go スタイル | ハンドラー・インターフェース双方の変更が必要、変更範囲大 | 不採用 |
| BaseContext の設定 | `http.Server.BaseContext` に cancellable context を設定し、Shutdown時にキャンセル | http.Server の標準的な仕組みを活用 | SSE の長時間接続を正しく処理するには Shutdown() の明示的呼び出しが依然必要 | 不採用（Broker方式と組み合わせる必要があり複雑） |

## Design Decisions

### Decision: `Broker` インターフェースへの `Shutdown()` 追加

- **Context**: SSE 接続をサーバー側から一括クローズする方法が必要
- **Alternatives Considered**:
  1. `context.Context` を `Register()` に渡す — 変更範囲が大きく、既存テストへの影響も大きい
  2. `Broker` に `Shutdown()` を追加する — 変更は `sse` パッケージと `server` パッケージのみ
- **Selected Approach**: `Broker.Shutdown()` を追加。呼び出し時に全クライアントチャネルをクローズし、`shutdown` フラグを立てて新規登録を拒否する
- **Rationale**: 変更範囲が最小、既存の `!ok` チェックを活用、テスト容易性が高い
- **Trade-offs**: インターフェース変更が必要だが、`internal/sse` は外部公開されていないため影響は内部のみ
- **Follow-up**: `Unregister()` が `Shutdown()` 後に呼ばれてもパニックしないことをテストで確認

### Decision: `Register()` にシャットダウンガードを追加

- **Context**: `Shutdown()` 後に新規 SSE 接続が `Register()` を呼んだ場合の安全性確保
- **Selected Approach**: `shutdown == true` の場合、既に `close` 済みのチャネルを返す。SSE ハンドラーは `!ok` で即座に終了する
- **Rationale**: レースコンディションを防ぎ、SSE ハンドラーが正しく終了することを保証する

### Decision: `Unregister()` の安全性

- **Context**: `Shutdown()` が全チャネルをクローズしてマップをクリアした後に `Unregister()` が呼ばれる（`defer` によるもの）
- **Selected Approach**: `Shutdown()` がマップをクリアするため、その後の `Unregister()` はマップ内にチャネルを見つけられず no-op になる
- **Rationale**: 既存の `Unregister()` 実装が既に安全（マップを走査して見つからなければ何もしない）

## Risks & Mitigations

- SSE ハンドラーがチャネルクローズを検出してから HTTP 接続がクローズされるまでのラグ — `http.Server.Shutdown()` の 5 秒タイムアウトが安全網として機能する
- `Shutdown()` と `Register()` の同時呼び出しによるレースコンディション — `sync.Mutex` で保護済み
- テスト `TestServer_ShutdownGracefully` が `s.Shutdown()` で `context.DeadlineExceeded` を返さないことの確認 — 修正後は SSE 接続なしでテストするため問題なし、SSE 付きのシャットダウンテストを追加が必要

## References

- [net/http Server.Shutdown 公式ドキュメント](https://pkg.go.dev/net/http#Server.Shutdown)
- [Go HTTP サーバーのライフサイクル](https://pkg.go.dev/net/http#Server.Serve)
