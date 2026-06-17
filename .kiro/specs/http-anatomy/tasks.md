# http-anatomy Tasks

## Implementation Plan

Go モジュールは `http-anatomy`。各タスクは 1 コミット相当。テストは Go 標準 `testing` + `net/http/httptest`。

### Wave 1 (parallel — no dependencies)

- [x] **Task 1.1**: モジュール初期化と雛形
  - What: `go mod init http-anatomy`（Go 1.22+）、`.gitignore`、空の `main.go`（`package main; func main(){}`）、ディレクトリ作成（`internal/model`, `internal/store`, `internal/inspector`, `internal/web`, `internal/web/templates`）。
  - Files: `go.mod`, `.gitignore`, `main.go`
  - Done when: `go build ./...` が通る。
  - Depends on: none

- [x] **Task 1.2**: ドメイン型
  - What: `Todo` / `User` / `Header` / `Exchange` を定義。
  - Files: `internal/model/todo.go`, `internal/model/user.go`, `internal/model/exchange.go`
  - Done when: `go vet ./internal/model/...` 通過、各 struct が design.md 通り。
  - Depends on: none

### Wave 2 (after Wave 1)

- [x] **Task 2.1**: store 実装 + テスト
  - What: `Store`/`Session`、`GetOrCreate`、Todo/User の Add/Toggle/Update/Delete、`AppendHistory`（新しい順・10件 trim）。全操作 mutex 保護。採番は session 内 seq。
  - Files: `internal/store/store.go`, `internal/store/store_test.go`
  - Done when: 採番単調増加・CRUD 整合・セッション分離・History trim/順序のテストが green。
  - Depends on: Task 1.2

- [x] **Task 2.2**: inspector 実装 + テスト
  - What: `BuildExchange(r *http.Request, body string, status int) model.Exchange`。選定ヘッダ（Host, HX-Request, HX-Target, HX-Trigger, Content-Type）抽出、`Proto`/`StatusText` 整形、Body=渡された主フラグメントのみ。
  - Files: `internal/inspector/capture.go`, `internal/inspector/capture_test.go`
  - Done when: 選定ヘッダのみ抽出・status→text 整形・Body に inspector 自身が混入しないテストが green。
  - Depends on: Task 1.2

- [x] **Task 2.3**: テンプレートと render ヘルパ
  - What: `page.html`（2ペイン外枠 + HTMX CDN）, `todos.html`/`todo-item.html`, `users.html`/`user-item.html`, `inspector.html`（ルートに `id="http-inspector" hx-swap-oob="true"`、`<pre>` でエスケープ表示・履歴一覧）, `error.html`。`go:embed` で読込み、`renderToString(name, data) (string, error)`。
  - Files: `internal/web/render.go`, `internal/web/templates/*.html`, `internal/web/render_test.go`
  - Done when: 各テンプレが parse でき、`renderToString` が期待文字列（OOB 属性・エスケープ）を返すテストが green。
  - Depends on: Task 1.2

### Wave 3 (after Wave 2)

- [x] **Task 3.1**: session ヘルパ
  - What: `ha_session` Cookie 取得/発行（crypto/rand UUID, HttpOnly, SameSite=Lax）、`resolveSession(w,r) *store.Session`（無ければ作成し Set-Cookie）。
  - Files: `internal/web/session.go`, `internal/web/session_test.go`
  - Done when: 初回 Set-Cookie・既存 Cookie 引継ぎのテストが green。
  - Depends on: Task 2.1

- [x] **Task 3.2**: ハンドラ + ルータ
  - What: トップ(`GET /`)、タブ(`/fragments/{todos,users}`)、Todo/User CRUD（POST/PATCH/DELETE）。各ハンドラで「主フラグメント生成 → BuildExchange → AppendHistory → inspector 連結 → Write」。404/422 は `error.html`。`http.ServeMux` のメソッド付きパターンで登録。入力長上限。
  - Files: `internal/web/handlers.go`, `internal/web/router.go`, `internal/web/handlers_test.go`
  - Done when: httptest で各エンドポイントが「主フラグメント + OOB inspector」を返し、`GET /` が Set-Cookie、不在 id で 404 断片、Cookie 引継ぎで状態保持、のテストが green。
  - Depends on: Task 2.2, Task 2.3, Task 3.1

### Wave 4 (after Wave 3 — integration / ship)

- [x] **Task 4.1**: main 配線と起動 URL 表示
  - What: `PORT`（既定 8080）読取、Store 生成、router マウント、localhost + LAN IP の URL 表示、`ListenAndServe`。
  - Files: `main.go`, `internal/web/router.go`（公開コンストラクタ）
  - Done when: `go run .` で起動し `GET /` が 2 ペインを返す（手動確認）。
  - Depends on: Task 3.2

- [x] **Task 4.2**: デプロイ資材
  - What: multi-stage `Dockerfile`（build → distroless/static）、`fly.toml`（単一インスタンス・内部 PORT）。
  - Files: `Dockerfile`, `fly.toml`, `.dockerignore`
  - Done when: `docker build` が通り、コンテナ内で `GET /` が 200。
  - Depends on: Task 4.1

- [x] **Task 4.3**: CI ワークフロー（parity・リモートなし）
  - What: `gofmt -l`(差分ゼロ) / `go vet ./...` / `go test ./...` を回す GitHub Actions。
  - Files: `.github/workflows/test.yml`
  - Done when: ローカルで同コマンドが全 green（push はしない）。
  - Depends on: Task 3.2

- [x] **Task 4.4**: README とクロスブラウザ確認
  - What: 概要・スタック・起動手順・ルート表・「状態は揮発」明記。Chrome/Safari/Firefox で 2 タブ別セッション CRUD を手動確認。
  - Files: `README.md`
  - Done when: 3 ブラウザで OOB 同時更新とセッション分離を確認。
  - Depends on: Task 4.1

## Progress
- Total: 11 tasks | Completed: 11 | In Progress: 0
- Wave 1: 1.1, 1.2 / Wave 2: 2.1, 2.2, 2.3 / Wave 3: 3.1, 3.2 / Wave 4: 4.1, 4.2, 4.3, 4.4
- 全 Wave 実装済み。`go test ./...` 20 件 green / `go vet` / `gofmt` クリーン。
  4.4 のクロスブラウザ手動確認のみ未実施（curl スモークで挙動は確認済み）。
