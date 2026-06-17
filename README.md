# http-anatomy(http-anatomy-web)

<!-- tech-stack:start (auto-generated) -->
<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/HTMX-3D72D7?style=for-the-badge&logo=htmx&logoColor=white" alt="HTMX">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/Fly.io-8B5CF6?style=for-the-badge&logo=flydotio&logoColor=white" alt="Fly.io">
  <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=githubactions&logoColor=white" alt="GitHub Actions">
</p>
<!-- tech-stack:end -->

任意の HTML 要素から GET 以外の HTTP メソッド(POST / PATCH / DELETE)を直接発行できる HTMX の本質を、
視覚的に「解剖(Anatomy)」して学ぶインタラクティブ教材。JavaScript による状態管理を持たず、
サーバが返す **HTML 断片だけ**で画面を駆動する。さらに **Out-of-Band (OOB) swap** により、
単一の HTTP レスポンスで画面内のまったく異なる 2 領域(左の操作対象と右のインスペクター)を同時に書き換える。

## コア体験
- **非 GET の直接発行**: 標準 HTML では出せない POST / PATCH / DELETE を、ボタンやフォームから直接サーバへ叩き込む。
- **OOB で 2 領域同時更新**: 1 レスポンスに「主フラグメント」と「`hx-swap-oob` 付きインスペクター」を同梱し、左右を同時に書き換える。
- **裏側の可視化**: 右ペインに実際に飛んだ生のリクエスト(method / path / HX-* ヘッダ)とレスポンス(status / Content-Type / 返却 HTML 断片)、直近 10 件の交信履歴を表示。
- **訪問者ごとに分離**: Cookie `ha_session` で訪問者ごとに独立した状態(Todos / Users / 履歴)を保持。2 タブで別セッションとして試せる。
- **JS 状態管理なし**: クライアント側の状態は持たず、サーバが返す HTML がそのまま唯一の真実。

> ⚠️ **状態は揮発する。** Todo / Users はメモリ上にのみ保持され、プロセス再起動ですべて消える(永続化なし)。

## スタック
- **言語 / ランタイム**: Go(標準ライブラリ中心 — `net/http`, `html/template`, `go:embed`)
- **フロント**: HTMX(CDN)のみ。外部 JS フレームワーク・独自状態管理なし
- **状態**: セッション別 in-memory ストア(`sync.Mutex` 保護)
- **デプロイ**: 単一バイナリ / Docker(distroless)/ Fly.io 単一インスタンス
- **CI**: GitHub Actions(`gofmt --check` / `go vet` / `go test`)

## 開発セットアップ

### 1. ツールチェーン
Go 1.22 以降(メソッド付き `ServeMux` パターンを使用)。

### 2. ビルドとテスト
```sh
go build ./...                # ビルド
go test ./...                 # 全テスト(20 件)
go vet ./... && gofmt -l .    # 静的チェック(CI と同じ)
```

### 3. 起動
```sh
go run .               # http://localhost:8080
PORT=9000 go run .     # ポート変更
```
起動時に localhost と **LAN の URL** を表示する。同一ネットワークの別端末からも開け、
**訪問者ごとに独立したセッション**(Cookie `ha_session`)で状態が分離される。

## 主要ルート
| メソッド | パス | 内容 |
|---|---|---|
| `GET` | `/` | 2 ペインのページ全体(初回に Cookie 発行) |
| `GET` | `/fragments/todos` , `/fragments/users` | 左ペインのタブ中身 + インスペクター OOB |
| `POST` | `/api/todos` , `/api/users` | 追加 → 新行 + インスペクター OOB |
| `PATCH` | `/api/todos/{id}` | `title` あり=改名 / なし=完了トグル |
| `PATCH` | `/api/users/{id}` | 改名・メール更新 |
| `DELETE` | `/api/todos/{id}` , `/api/users/{id}` | 削除(空断片)+ インスペクター OOB |

不正 id / 不在リソースは 404、必須値欠落は 422 を **HTML 断片**で返し、失敗した交信もインスペクターに表示する。

## OOB の仕組み(教材の核心)
各 CRUD レスポンスは 2 部構成で返る。
1. **主フラグメント** — `hx-target` が指す要素を差し替える HTML(例: 新しい Todo 行)
2. **インスペクター断片** — `<div id="http-inspector" hx-swap-oob="true">…</div>`

サーバは主フラグメントの文字列を**先に生成して捕捉**し、その内容をエスケープしてインスペクター断片に埋め込む
(インスペクター自身の OOB ブロックは表示本文に含めないため再帰しない)。これにより
「1 リクエストで左の行と右のインスペクターが同時に書き換わる」HATEOAS を体感できる。

## プロジェクト構成
- `main.go` … 起動・`PORT` 読取・localhost / LAN URL 表示
- `internal/model/` … `Todo` / `User` / `Exchange`(ドメイン型・依存なし)
- `internal/store/` … セッション別 in-memory ストア(mutex 保護・採番・履歴 trim)
- `internal/inspector/` … リクエスト + 主フラグメント → `Exchange` 整形
- `internal/web/` … ルーティング・Cookie・CRUD ハンドラ・テンプレート(`go:embed`)
- `.kiro/specs/http-anatomy/` … SDD spec(requirements / design / tasks)

## デプロイ(Fly.io)
```sh
docker build -t http-anatomy .
fly launch --copy-config --now    # 初回
fly deploy                        # 以降
```
状態が揮発するため `auto_stop_machines = "off"` / `min_machines_running = 1` で常時 1 台を維持する。

## 範囲外
永続化(DB)/ 認証 / セッション失効(TTL)/ ユーザー間共有 / リアルタイム配信 / AI。

## ライセンス
個人プロジェクト — 閲覧・実行・学習は自由。
