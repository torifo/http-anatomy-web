# http-anatomy

**Webの裏側を解剖する。HTMLだけで画面の複数箇所を完全同期するHATEOAS体感教材。**

任意の HTML 要素から GET 以外の HTTP メソッド（POST / PATCH / DELETE）を直接発行できる
HTMX の本質を、視覚的に「解剖（Anatomy）」して学ぶインタラクティブ教材です。
JavaScript による状態管理を持たず、サーバが返す **HTML 断片だけ**で画面を駆動します。
さらに **Out-of-Band (OOB) swap** により、単一の HTTP レスポンスで画面内の
まったく異なる 2 領域（左の操作対象と右のインスペクター）を同時に書き換えます。

## 画面

左右 2 ペイン構成。

- **左ペイン（リソース管理）**: Todo / Users の一覧。追加・編集・完了トグル・削除。
- **右ペイン（HTTP インスペクター）**: 直前の操作で実際に飛んだ
  **生のリクエスト（method / path / HX-* ヘッダ）** と
  **サーバーの生レスポンス（status / Content-Type / 返却 HTML 断片）**、
  および直近 10 件の交信履歴をリアルタイム表示。

## スタック

- **言語/ランタイム**: Go（標準ライブラリ中心 — `net/http`, `html/template`, `go:embed`）
- **フロント**: HTMX（CDN）のみ。独自 JS の状態管理なし
- **状態**: セッション別 in-memory ストア（`sync.Mutex`）
- **デプロイ**: 単一バイナリ / Docker（distroless）/ Fly.io 単一インスタンス

> ⚠️ **状態は揮発します。** Todo/Users はメモリ上にのみ保持され、
> プロセスを再起動するとすべて消えます（永続化なし）。

## 開発

```sh
go test ./...                 # 全テスト
go vet ./... && gofmt -l .    # 静的チェック（CI と同じ）
go run .                      # http://localhost:8080
PORT=9000 go run .            # ポート変更
```

起動時に localhost と LAN の URL を表示します。
同一ネットワークの別端末からもアクセスでき、**訪問者ごとに独立したセッション**
（Cookie `ha_session`）で状態が分離されます。2 タブで開けば別セッションとして試せます。

## 主要ルート

| メソッド | パス | 内容 |
|---|---|---|
| GET | `/` | 2 ペインのページ全体（初回に Cookie 発行） |
| GET | `/fragments/todos`, `/fragments/users` | 左ペインのタブ中身 + インスペクター OOB |
| POST | `/api/todos`, `/api/users` | 追加 → 新行 + インスペクター OOB |
| PATCH | `/api/todos/{id}` | `title` あり=改名 / なし=完了トグル |
| PATCH | `/api/users/{id}` | 改名・メール更新 |
| DELETE | `/api/todos/{id}`, `/api/users/{id}` | 削除（空断片）+ インスペクター OOB |

不正 id / 不在リソースは 404、必須値欠落は 422 を **HTML 断片**で返し、
失敗した交信もインスペクターに表示します。

## OOB の仕組み（教材の核心）

各 CRUD レスポンスは 2 部構成で返ります。

1. **主フラグメント** — `hx-target` が指す要素を差し替える HTML（例: 新しい Todo 行）
2. **インスペクター断片** — `<div id="http-inspector" hx-swap-oob="true">…</div>`

サーバは主フラグメントの文字列を**先に生成して捕捉**し、その内容を
エスケープしてインスペクター断片に埋め込みます（インスペクター自身の OOB ブロックは
表示本文に含めないため、再帰しません）。これにより
「1 リクエストで左の行と右のインスペクターが同時に書き換わる」HATEOAS を体感できます。

## デプロイ（Fly.io）

```sh
docker build -t http-anatomy .
fly launch --copy-config --now    # 初回
fly deploy                        # 以降
```

状態が揮発するため `auto_stop_machines = "off"` / `min_machines_running = 1`
で常時 1 台を維持しています。

## プロジェクト構成

```
main.go                       起動・PORT・localhost/LAN URL 表示
internal/model/               Todo / User / Exchange（ドメイン型・依存なし）
internal/store/               セッション別 in-memory ストア（mutex 保護）
internal/inspector/           リクエスト+主フラグメント → Exchange 整形
internal/web/                 ルーティング・Cookie・CRUD ハンドラ・テンプレート
.kiro/specs/http-anatomy/     SDD spec（requirements / design / tasks）
```

## 範囲外

永続化（DB）/ 認証 / セッション失効（TTL）/ ユーザー間共有 / リアルタイム配信。
