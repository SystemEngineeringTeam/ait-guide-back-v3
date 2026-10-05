# ait-guide-back-v3

## 概要

建物内外の経路情報を管理し、複数の重みパラメータを考慮した最適経路を提供するWebAPI。起動時にCSVからDB（PostgreSQL + PostGIS + pgRouting）へ建物・ノード・エッジ情報を投入し、緯度経度からの最寄りノード検索や階段有無・屋内外を考慮した経路探索を行う。

## 技術スタック

- 言語: Go 1.26+（`go.mod`では`go 1.26.1`を指定）
- Webフレームワーク: Gin
- データベース: PostgreSQL 16 + PostGIS 3.4+ + pgRouting 3.6+（Dockerイメージ: `pgrouting/pgrouting:16-3.5-main`）
- DBドライバー: pgx v5
- コンテナ: Docker / Docker Compose

## 利用環境・ツール

| ツール | バージョン / 用途 |
| --- | --- |
| Go | 1.26+（`go.mod`の`go`ディレクティブに準拠） |
| golangci-lint | v2系（CIは`golangci-lint-action@v9` + `version: latest`で都度最新を取得。Dockerイメージ内は`@latest`でインストール） |
| air | ホットリロード（Dockerイメージ内に`@latest`でインストール、開発環境のみ） |
| swag | v1.16.6固定（`make swag`でSwagger docsを再生成） |
| Docker / Docker Compose | `compose.yaml` + `compose.override.yaml`（開発）/ `compose.prod.yaml`（本番） |
| CI | GitHub Actions（`.github/workflows/ci.yml`、build/lint/testをPR時に実行） |

本番イメージ（`Dockerfile`の`production`ステージ）は`gcr.io/distroless/static-debian13:nonroot`ベース。

## ディレクトリ構成

```sh
.
├── cmd/server/main.go   # エントリポイント、DI、サーバー起動
├── internal/
│   ├── domain/           # エンティティ・リポジトリIF・ドメインサービス
│   ├── usecase/          # アプリケーションサービス
│   ├── handler/          # Ginハンドラー・ルーター
│   ├── infra/             # DB接続・リポジトリ実装・CSVローダー
│   └── config/            # 設定管理
├── db/
│   ├── migrations/        # マイグレーションSQL
│   └── seeds/              # 投入用CSV（.gitignore対象、各自で用意）
├── docs/                   # API仕様・スキーマ・ドメインルール・Swagger
├── compose.yaml            # Docker Compose（共通: DB）
├── compose.override.yaml   # Docker Compose（開発: ホットリロード）
├── compose.prod.yaml       # Docker Compose（本番）
└── Dockerfile               # マルチステージビルド（dev / builder / production）
```

詳細な設計方針は [CLAUDE.md](./claude.md) を参照。

## セットアップ

### 1. 環境変数

リポジトリルートに `.env` を作成する（`.gitignore`対象のため各自用意）。

```env
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=aitguide
DB_PASSWORD=aitguide
DB_NAME=aitguide

# Seed
SEED_DIR=db/seeds
```

### 2. シードデータ

`db/seeds/` 配下に以下のCSVを配置する（`.gitignore`対象のため各自用意、未配置のファイルはスキップされる）。

| ファイル | 投入順 | 内容 |
| --- | --- | --- |
| `buildings.csv` | 1 | 建物 |
| `nodes.csv` | 2 | 経路ノード（entrance/road/door/facility） |
| `edges.csv` | 3 | ノード間のエッジ |
| `rooms.csv` | 4 | 部屋 |
| `photos.csv` | 5 | 建物写真 |

### 3. 起動（Docker、推奨）

```bash
make dev          # db + app をホットリロードで起動
make dev-build    # イメージを再ビルドして起動
make logs         # appのログを追跡
make down         # 停止
make down-v       # 停止 + ボリューム削除
```

### 4. 起動（ローカル、DBのみDocker）

```bash
make db-up        # DBだけ起動
go mod tidy
make run          # go run cmd/server/main.go
```

起動時に `db/migrations/` のマイグレーションと `SEED_DIR` のCSVが自動投入される。

## 開発コマンド

```bash
make build        # バイナリビルド
make test         # go test ./...
make test-cover   # カバレッジ付きテスト
make fmt           # go fmt ./...
make vet            # go vet ./...
make lint            # golangci-lint run
make swag             # Swaggerドキュメント再生成
```

その他のコマンドは `make help` で一覧表示できる。

## 本番環境

```bash
make prod          # compose.yaml + compose.prod.yaml で起動
make prod-build     # イメージを再ビルドして起動
```

## 関連ドキュメント

- [DBスキーマ](./docs/schema.md)
- [API仕様](./docs/api.md)
- [ドメインルール](./docs/domain-rules.md)
- [Swagger UI](./docs/swagger)（起動後は `/swagger/index.html`）
