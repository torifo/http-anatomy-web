# http-anatomy Requirements

## Overview
HTMX の本質（任意の HTML 要素から GET 以外の HTTP メソッドを直接発行し、単一レスポンスで画面内の複数領域を同時に書き換える HATEOAS）を視覚的に「解剖」して学ぶインタラクティブ教材。左ペインで Todo / Users を CRUD 操作すると、右ペインの「HTTP インスペクター」に実際に飛んだリクエスト/レスポンスの生情報が同じ 1 レスポンス（Out-of-Band swap）でリアルタイム表示される。JavaScript による状態管理は持たず、サーバが返す HTML 断片だけで画面を駆動する。

## User Stories

### US-001: Todo の CRUD を HTMX で操作する
**As a** 学習者 **I want to** Todo を追加・完了トグル・編集・削除する **So that** GET 以外の HTTP メソッド（POST/PATCH/DELETE）が HTML 要素から直接発行される様子を体感できる

**Acceptance Criteria:**
- WHEN 学習者が新規追加フォームを送信する THE SYSTEM SHALL `POST /api/todos` を受け、新しい Todo 行の HTML 断片を返す
- WHEN 学習者が Todo の完了ボタンを押す THE SYSTEM SHALL `PATCH /api/todos/{id}` を受け、完了状態を反転した行の HTML 断片を返す
- WHEN 学習者が Todo の編集を確定する THE SYSTEM SHALL `PATCH /api/todos/{id}` を受け、更新後の行の HTML 断片を返す
- WHEN 学習者が Todo の削除ボタンを押す THE SYSTEM SHALL `DELETE /api/todos/{id}` を受け、その行を取り除く（空または削除済み表現の）HTML 断片を返す
- IF 指定 id の Todo がセッション内に存在しない THEN THE SYSTEM SHALL 404 を HTML 断片で返す

### US-002: Users の CRUD を HTMX で操作する
**As a** 学習者 **I want to** User を追加・編集・削除する **So that** 同じ HTMX パターンが別リソースでも一様に成立することを確認できる

**Acceptance Criteria:**
- WHEN 学習者が User 追加フォームを送信する THE SYSTEM SHALL `POST /api/users` を受け、新しい User 行の HTML 断片を返す
- WHEN 学習者が User の編集を確定する THE SYSTEM SHALL `PATCH /api/users/{id}` を受け、更新後の行の HTML 断片を返す
- WHEN 学習者が User の削除ボタンを押す THE SYSTEM SHALL `DELETE /api/users/{id}` を受け、その行を取り除く HTML 断片を返す
- IF 指定 id の User がセッション内に存在しない THEN THE SYSTEM SHALL 404 を HTML 断片で返す

### US-003: HTTP インスペクターで交信を可視化する
**As a** 学習者 **I want to** 直前の操作で飛んだリクエスト/レスポンスの生情報を右ペインで見る **So that** 「裏側で何が起きたか」を解剖して理解できる

**Acceptance Criteria:**
- WHEN いずれかの CRUD 操作が行われる THE SYSTEM SHALL 同一レスポンス内に `hx-swap-oob="true"` を付けたインスペクター断片を含め、右ペインを丸ごと差し替える
- WHEN インスペクターが更新される THE SYSTEM SHALL リクエストの method / path / HTTP version / `Host` / `HX-Request` / `HX-Target` / `HX-Trigger` / `Content-Type` ヘッダを表示する
- WHEN インスペクターが更新される THE SYSTEM SHALL レスポンスの status 行 / `Content-Type` / 主フラグメント本文（HTML エスケープ済み）を表示する
- THE SYSTEM SHALL 主フラグメント文字列をレスポンス本文生成より前に捕捉し、インスペクター断片には主フラグメントのみを表示する（インスペクター自身の OOB ブロックは表示本文に含めない）
- WHEN CRUD 操作が行われる THE SYSTEM SHALL 直近 10 件の交信履歴（method / path / status）をインスペクター内に新しい順で表示する

### US-004: 訪問者ごとに状態を分離する
**As a** 同時に触る複数の学習者 **I want to** 自分の操作が他人の画面に影響しない **So that** 公開デモでも安心して試せる

**Acceptance Criteria:**
- WHEN 学習者が初めて `GET /` にアクセスする THE SYSTEM SHALL `Set-Cookie: ha_session=<uuid>` を発行する
- WHILE 有効なセッション Cookie が送られている THE SYSTEM SHALL そのセッション専用の Todos / Users / 履歴 / 採番カウンタを操作対象にする
- IF 受信リクエストにセッション Cookie が無い THEN THE SYSTEM SHALL 新規セッションを作成して Cookie を発行する

### US-005: リソースのタブを切り替える
**As a** 学習者 **I want to** Todos と Users のタブを切り替える **So that** 1 画面で両リソースを扱える

**Acceptance Criteria:**
- WHEN 学習者が Todos / Users タブを押す THE SYSTEM SHALL `GET /fragments/todos` または `/fragments/users` で左ペイン中身の HTML 断片を返す
- WHEN タブ切替の GET が行われる THE SYSTEM SHALL US-003 と同様にインスペクターを OOB で更新する

## Functional Requirements

### FR-001: 非 GET メソッドの直接発行
**Priority:** P0 / **Persona:** 学習者
WHEN HTML 要素の `hx-post` / `hx-patch` / `hx-delete` が発火する THE SYSTEM SHALL 対応する HTTP メソッドのハンドラで処理し HTML 断片を返す。
**Rationale:** 標準 HTML では出せない POST/PATCH/DELETE をボタンから直接出すことが教材の主眼。

### FR-002: Out-of-Band による複数領域同時更新
**Priority:** P0 / **Persona:** 学習者
WHEN CRUD/タブ操作のレスポンスを返す THE SYSTEM SHALL 「主フラグメント」と「`hx-swap-oob` 付きインスペクター断片」を連結した 1 レスポンスを返す。
**Rationale:** 単一レスポンスで複数領域が同期する HATEOAS の体感が核心。

### FR-003: リクエスト/レスポンス・メタ情報の捕捉
**Priority:** P0 / **Persona:** 学習者
WHEN 各ハンドラが応答を生成する THE SYSTEM SHALL Exchange（method, path, version, 選定ヘッダ, status, body）を組み立ててインスペクター描画に渡す。
**Rationale:** 生のヘッダ（特に HX-*）を見せることが「解剖」。

### FR-004: セッション別 in-memory ストア
**Priority:** P0 / **Persona:** 運営/学習者
THE SYSTEM SHALL セッション ID をキーに Todos / Users / 履歴 / 採番カウンタを in-memory（mutex 保護）で保持する。
**Rationale:** 永続化なしで訪問者ごとの独立状態を成立させる。

### FR-005: エラーの可視化
**Priority:** P1 / **Persona:** 学習者
IF 不正 id / 不在リソースが要求される THEN THE SYSTEM SHALL 404 をエラー用 HTML 断片 + インスペクター OOB で返す。
**Rationale:** 失敗時の交信も教材として見せる。

## Non-Functional Requirements
- パフォーマンス: 各リクエストはローカル/単一インスタンスで 50ms 以内に応答（DB なし・in-memory）。
- 移植性: 外部 JS 依存は HTMX（CDN）のみ。サーバは Go 標準ライブラリ（net/http, html/template）中心。
- セキュリティ: テンプレートは html/template の自動エスケープを用い、インスペクター本文はユーザー入力を必ずエスケープ表示する。
- スケーラビリティ: 単一インスタンス前提。状態は揮発（プロセス再起動で消える）であることを README に明記。
- 品質ゲート: `go vet` / `go test` / `gofmt -l` を CI（GitHub Actions, parity 用・リモートなし）で実行。
