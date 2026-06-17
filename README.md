# HTMX解剖室(http-anatomy-web)

<!-- tech-stack:start (auto-generated) -->
<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/HTMX-3D72D7?style=for-the-badge&logo=htmx&logoColor=white" alt="HTMX">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/GHCR-181717?style=for-the-badge&logo=github&logoColor=white" alt="GHCR">
  <img src="https://img.shields.io/badge/NGINX_Proxy-009639?style=for-the-badge&logo=nginx&logoColor=white" alt="NGINX Proxy">
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
- **デプロイ**: Docker(distroless)イメージを GHCR に発行 → 自前 VPS で pull
- **CI**: GitHub Actions(`gofmt --check` / `go vet` / `go test`、および GHCR への build & push)

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
| `GET` | `/fragments/todos` , `/fragments/users` | タブ中身 + インスペクター OOB。`?q=&filter=all\|active\|done` で検索・絞込 |
| `GET` | `/theme/toggle` | テーマ切替(Cookie `ha_theme`)→ `HX-Redirect` でリロード |
| `POST` | `/api/todos` , `/api/users` | 追加 → 新行 + OOB。重複タイトルは 409 |
| `PUT` | `/api/todos/{id}` | **全置換**(title+done をまとめて・冪等) |
| `PATCH` | `/api/todos/{id}` | **部分更新**(`title` あり=改名 / なし=完了トグル) |
| `PATCH` | `/api/users/{id}` | 改名・メール更新 |
| `DELETE` | `/api/todos/{id}` , `/api/users/{id}` | 削除(空断片)+ OOB |

不正 id / 不在リソースは 404、必須値欠落は 422、重複は 409 を **HTML 断片**で返す。
htmx 既定は 4xx/5xx を swap しないため、`htmx-config` の `responseHandling` を上書きして
エラー断片とインスペクター OOB がブラウザでも表示されるようにしている。

## 学べる HTTP / HTMX 概念
- **HTTP メソッド**: GET / POST / PUT(全置換・冪等) / PATCH(部分更新) / DELETE
- **ステータス**: 201 / 200 / 404 / 409(重複) / 422(検証)、エラー応答の swap 設定
- **レスポンスヘッダ**: `HX-Trigger`(トースト発火) / `HX-Redirect`(クライアントリダイレクト)をインスペクターで可視化
- **HTMX パターン**: OOB swap、swap 戦略(beforeend / outerHTML / innerHTML)、
  `hx-trigger`(`keyup changed delay:300ms`)、`hx-include`、`hx-confirm`、`hx-indicator`
- **状態の置き場所**: Cookie + サーバレンダリング(テーマ・セッション)。クライアント側の状態管理は持たない

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

## デプロイ(GHCR + 自前 VPS)
バックエンド(常駐 Go サーバ + サーバ側状態保持)のため **GitHub Pages では公開できない**。
イメージを GHCR に発行し、VPS で pull して常駐させる。VPS 既存の共有リバースプロキシ
(`nginxproxy/nginx-proxy` + `acme-companion`、外部ネット `global-proxy-network`)に
`VIRTUAL_HOST` で吊るすだけで自動ルーティング・自動 TLS される。

1. **イメージ発行(自動)**: `main` への push で GitHub Actions が
   `ghcr.io/torifo/http-anatomy-web:latest`(と short-sha タグ)を build & push。
2. **VPS 配置**: 設置先(例 `/home/ubuntu/Web/http-anatomy`)に `docker-compose.prod.yml` と
   `.env`(`.env.example` 参照: `APP_HOST` / `IMAGE_TAG`)を置く。
3. **デプロイ**:
   ```sh
   docker compose -f docker-compose.prod.yml --env-file .env pull
   docker compose -f docker-compose.prod.yml --env-file .env up -d
   ```

DB なし・単一コンテナ。状態は in-memory(揮発)なので、コンテナ再起動でセッションは消える。

## 範囲外
永続化(DB)/ 認証 / セッション失効(TTL)/ ユーザー間共有 / リアルタイム配信 / AI。

## ライセンス
個人プロジェクト — 閲覧・実行・学習は自由。
