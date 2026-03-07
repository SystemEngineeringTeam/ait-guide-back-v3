# 経路探索WebAPI プロジェクト

## 概要

建物内外の経路情報を管理し、複数の重みパラメータを考慮した最適経路を提供するWebAPIシステム。

### 主要機能

- 起動時にcsvファイルからデータベースに建物・ノード・エッジ情報を投入
- 緯度経度から最寄りノードの検索
- 複数重み（距離・コスト）を適用した経路探索
- 階段有無を考慮した経路フィルタリング
- 屋内外の区別（is_indoorフラグ）による経路制御
- 建物・部屋情報の提供

### システム特性

- データ規模: 最大1000ノード、1500エッジ
- 応答性能: 経路探索100ms以内
- 更新頻度: 静的データ（初期投入後の変更は少ない）

## 技術スタック

| カテゴリ | 技術 | バージョン |
| ------- | ---- | -------- |
| 言語 | Go | 1.25+ |
| Webフレームワーク | Gin | 最新 |
| データベース | PostgreSQL | 16+ |
| 空間拡張 | PostGIS | 3.4+ |
| 経路探索 | pgRouting | 3.6+ |
| DBドライバー | pgx | v5 |
| コンテナ | Docker / Docker Compose | - |

## アーキテクチャ

ドメイン駆動設計（DDD）を意識したレイヤードアーキテクチャを採用。

### レイヤー構成

```sh
┌─────────────────────────────────────────┐
│           Interfaces（外部との境界）      │
│   HTTPハンドラー、DTO、リクエスト検証      │
├─────────────────────────────────────────┤
│           Application（ユースケース）     │
│   アプリケーションサービス、トランザクション │
├─────────────────────────────────────────┤
│           Domain（ビジネスロジック）       │
│   エンティティ、値オブジェクト、ドメインサービス │
├─────────────────────────────────────────┤
│           Infrastructure（技術詳細）      │
│   リポジトリ実装、DB接続、外部API連携      │
└─────────────────────────────────────────┘
```

### 依存関係のルール

- 上位レイヤーは下位レイヤーに依存可能
- 下位レイヤーは上位レイヤーに依存不可
- Domain層は他のレイヤーに依存しない（純粋なビジネスロジック）
- Infrastructure層はDomain層のインターフェースを実装（依存性逆転）

## ディレクトリ構成

```sh
.
├── cmd/                        # アプリケーションエントリポイント
│   └── server/
│       └── main.go            # DIコンテナ初期化、サーバー起動
│
├── internal/                   # 内部パッケージ
│   │
│   ├── domain/                # Domain層（ビジネスロジック）
│   │   ├── entity/           # エンティティ
│   │   │   ├── node.go
│   │   │   ├── edge.go
│   │   │   ├── building.go
│   │   │   └── room.go
│   │   ├── value/            # 値オブジェクト
│   │   │   ├── coordinate.go # 座標（緯度経度）
│   │   │   ├── node_type.go  # ノードタイプ
│   │   │   └── route.go      # 経路結果
│   │   ├── repository/       # リポジトリインターフェース
│   │   │   ├── node.go
│   │   │   ├── edge.go
│   │   │   └── building.go
│   │   └── service/          # ドメインサービス
│   │       └── routing.go    # 経路探索ロジック
│   │
│   ├── application/           # Application層（ユースケース）
│   │   ├── route_search.go   # 経路探索ユースケース
│   │   ├── building_info.go  # 建物情報取得ユースケース
│   │   └── room_info.go      # 部屋情報取得ユースケース
│   │
│   ├── infrastructure/        # Infrastructure層（技術詳細）
│   │   ├── postgres/         # PostgreSQL実装
│   │   │   ├── connection.go # DB接続管理
│   │   │   ├── node_repo.go  # ノードリポジトリ実装
│   │   │   ├── edge_repo.go  # エッジリポジトリ実装
│   │   │   └── routing.go    # pgRouting呼び出し
│   │   ├── loader/           # データローダー
│   │   │   └── csv.go        # CSV読み込み・DB投入
│   │   └── config/           # 設定管理
│   │       └── config.go
│   │
│   └── interfaces/            # Interfaces層（外部との境界）
│       ├── handler/          # HTTPハンドラー
│       │   ├── route.go
│       │   ├── building.go
│       │   └── room.go
│       ├── dto/              # Data Transfer Object
│       │   ├── request/
│       │   └── response/
│       ├── middleware/       # ミドルウェア
│       │   └── error.go
│       └── router/           # ルーティング設定
│           └── router.go
│
├── pkg/                        # 外部公開可能なパッケージ
│   └── errors/               # カスタムエラー
│
├── db/                         # DBマイグレーション・シード
│   ├── migrations/
│   └── seeds/                # CSVシードデータ
│
├── docs/                       # ドキュメント
├── docker/                     # Docker関連ファイル
├── temp/                       # 一時ファイル（コミット対象外）
├── go.mod
├── go.sum
└── docker-compose.yml
```

### 各レイヤーの責務

| レイヤー | 責務 | 依存先 |
| ------- | ---- | ------ |
| Domain | エンティティ、ビジネスルール、リポジトリIF定義 | なし |
| Application | ユースケース実行、トランザクション管理 | Domain |
| Infrastructure | DB接続、リポジトリ実装、CSV読み込み | Domain |
| Interfaces | HTTP処理、リクエスト/レスポンス変換 | Application, Domain |

## 開発コマンド

```bash
# 依存関係のインストール
go mod tidy

# 開発サーバー起動
go run cmd/server/main.go

# テスト実行
go test ./...

# テスト（カバレッジ付き）
go test -cover ./...

# ビルド
go build -o bin/server cmd/server/main.go

# リンター実行
golangci-lint run

# Docker環境起動
docker compose up -d

# DBマイグレーション
go run cmd/migrate/main.go up
```

## コーディング規約

### 一般

- Go標準のコーディング規約に従う
- `gofmt` / `goimports` でフォーマット
- `golangci-lint` でリント

### 命名規則

- パッケージ名: 小文字、単一単語
- 変数名: キャメルケース（`nodeID`, `buildingName`）
- エクスポート: 大文字始まり
- 定数: キャメルケースまたは ALL_CAPS

### エラーハンドリング

- エラーは必ずチェック
- `errors.Wrap` でコンテキスト追加
- ログは構造化ログ（JSON形式）

### データベースアクセス

- `pgx` を直接使用
- SQLインジェクション対策: プレースホルダ必須
- トランザクションは `pgx.Tx` を使用

## 関連ドキュメント

- [DBスキーマ](../docs/schema.md)
- [API仕様](../docs/api.md)
- [ドメインルール](../docs/domain-rules.md)
- [システム仕様書](../temp/system_specification.md)

## セキュリティ

- SQLインジェクション対策: パラメータバインディング必須
- 入力バリデーション: 全APIエンドポイントで実施
- HTTPS通信
- API Key認証（オプション）
