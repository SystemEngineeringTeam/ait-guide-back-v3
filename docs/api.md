# API仕様

## 概要

経路探索WebAPIのエンドポイント仕様。

## エンドポイント一覧

| メソッド | パス | 説明 |
| ------ | ---- | ---- |
| GET | `/api/health` | APIの健全性を確認 |
| GET | `/api/get/route` | 経路探索（旧バージョン） |
| GET | `/api/routes/search/{building_id}` | 経路探索（新バージョン） |
| GET | `/api/buildings` | 建物一覧取得 |
| GET | `/api/buildings/{building_id}` | 建物情報取得 |
| GET | `/api/rooms` | 部屋一覧取得 |
| GET | `/api/rooms/{room_id}` | 部屋情報取得 |
| GET | `/api/nodes` | ノード一覧取得 |

---

## ヘルスチェック

### リクエスト

```sh
GET /api/health
```

### レスポンス

```json
{
  "status": "ok",
  "timestamp": "2024-06-01T12:00:00Z"
}
```

## 経路探索（旧バージョン）

旧バージョンとの互換性のためのエンドポイント。距離のみを重みとして使用。

### リクエスト

```sh
GET /api/get/route
```

**クエリパラメータ:**

| パラメータ | 型 | 必須 | デフォルト | 説明 |
| -------- | -- | ---- | -------- | ---- |
| lat | float | No | 35.181531 | 現在地の緯度 |
| lng | float | No | 137.109509 | 現在地の経度 |
| end | string | No | 1 | 目的地のID（nodes.node_id に対応） |

### レスポンス

```json
{
    "route": [
        {"lat": 35.1846295800688, "lng": 137.11079120636},
        {"lat": 35.1846076583848, "lng": 137.110667824745},
        {"lat": 35.1846909607524, "lng": 137.110683917999},
        {"lat": 35.18474795706, "lng": 137.110544443131}
    ]
}
```

---

## 経路探索（新バージョン）

複数の重みパラメータとフィルタリングオプションを使用した経路探索。

### リクエスト

```sh
GET /api/routes/search/{building_id}
```

**パスパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
| -------- | -- | --- | ---- |
| building_id | int | Yes | 建物ID |

**クエリパラメータ:**

| パラメータ | 型 | 必須 | デフォルト | 説明 |
| -------- | -- | ---- | -------- | ---- |
| lat | float | No | 35.181531 | 現在地の緯度 |
| lng | float | No | 137.109509 | 現在地の経度 |
| level | int | No | 3 | 使用する道の主要度（1〜5） |
| stairs | bool | No | true | 階段を含む経路を許可するか |
| accessible | bool | No | false | バリアフリー経路を使用するか |
| indoor | bool | No | false | 屋内優先か |

### レスポンス

```json
{
  "success": true,
  "data": {
    "route": [
      {"lat": 35.1846295800688, "lng": 137.11079120636, "is_indoor": false},
      {"lat": 35.1846076583848, "lng": 137.110667824745, "is_indoor": false},
      {"lat": 35.1846909607524, "lng": 137.110683917999, "is_indoor": true}
    ],
    "total_distance": 150.5,
    "total_cost": 120.0
  }
}
```

---

## 建物一覧取得

### リクエスト

```sh
GET /api/buildings
```

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
| -------- | -- | ---- | ---- |
| limit | int | No | 返却する建物の最大数（デフォルト100） |
| affiliation | string | No | 所属 |

### レスポンス

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "総合技術研究所",
    },
    {
      "id": 2,
      "name": "情報科学研究所",
    }
  ]
}
```

## 建物情報取得

### リクエスト

```sh
GET /api/buildings/{building_id}
```

**パスパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
| ---------- | --- | ---- | ---- |
| building_id | int | Yes | 建物ID |

### レスポンス

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "総合技術研究所",
    "description": "先端の実験施設が揃い、産学官が連携した様々な研究が行われている研究所です。",
    "photos": [
      {
        "url": "https://example.com/photo1.jpg",
        "caption": "正面入り口"
      }
    ]
  }
}
```

**備考:** 建物の入口ノードは `/api/nodes?building_id={id}&node_type=entrance` で取得可能

---

## 部屋一覧取得

### リクエスト

```sh
GET /api/rooms
```

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
| -------- | -- | ---- | ---- |
| building_id | int | No | 建物IDでフィルタリング |
| floor | int | No | 階数でフィルタリング |
| room_type | string | No | 部屋タイプでフィルタリング（classroom, meeting_roomなど） |

### レスポンス

```json
{
  "success": true,
  "data": [
    {
      "room_id": "room_101",
      "name": "Room 101",
      "building_id": 1,
      "floor": 1,
      "capacity": 50,
      "room_type": "classroom",
    },
    {
      "room_id": "room_102",
      "name": "Room 102",
      "building_id": 1,
      "floor": 1,
      "capacity": 10,
      "room_type": "meeting_room",
    }
  ]
}
```

## 部屋情報取得

### リクエスト

```sh
GET /api/rooms/{room_id}
```

**パスパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
| -------- | -- | ---- | ---- |
| room_id | string | Yes | 部屋ID |

### レスポンス

```json
{
  "success": true,
  "data": {
    "room_id": "room_101",
    "name": "Room 101",
    "description": "Large classroom",
    "building_id": 1,
    "floor": 1,
    "capacity": 50,
    "room_type": "classroom",
    "door_node_id": "indoor_door_room_101"
  }
}
```

---

### レスポンス形式

#### 成功時

```json
{
  "success": true,
  "data": {...}
}
```

#### エラー時

```json
{
  "success": false,
  "error": {
    "code": "NODE_NOT_FOUND",
    "message": "指定されたノードが見つかりません",
    "details": {...}
  }
}
```

## ノード情報取得

nodesテーブルからノード情報を取得するエンドポイント。
クエリパラメータでフィルタリング可能。
デフォルトではノードの一覧を返す

### リクエスト

```sh
GET /api/nodes
```

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
| -------- | -- | ---- | ---- |
| node_id | string | No | ノードIDでフィルタリング |
| node_type | string | No | ノードタイプでフィルタリング（entrance, road, door） |
| floor | int | No | 階数でフィルタリング |
| limit | int | No | 返却するノードの最大数（デフォルト100） |

---

## エラーコード一覧

| コード | HTTPステータス | 説明 |
| ----- | ------------ | ---- |
| `NODE_NOT_FOUND` | 404 | ノードが存在しない |
| `NO_ROUTE_FOUND` | 404 | 経路が見つからない |
| `INVALID_PARAMETERS` | 400 | パラメータが不正 |
| `BUILDING_NOT_FOUND` | 404 | 建物が存在しない |
| `ROOM_NOT_FOUND` | 404 | 部屋が存在しない |

---

## HTTPステータスコード

| ステータス | 説明 |
| -------- | ---- |
| 200 | 成功 |
| 400 | リクエストパラメータが不正 |
| 404 | リソースが見つからない |
| 500 | サーバー内部エラー |

---

## 経路探索の処理フロー

```sh
1. クライアント: 開始座標(緯度経度) + 目的の建物ID + オプション
   ↓
2. API: 最寄りノード検索（PostGIS空間検索）
   ↓
3. API: 目的地ノード特定（建物入口ノードを使用）
  - 目的の建物IDを持つentranceノードを検索
   ↓
4. API: オプションに応じてエッジフィルタリング
  - level: 指定以上の主要度のエッジを選択 重要度：1(高) ~ 5(低)
  - stairs: 階段の有無でエッジを除外
  - accessible: バリアフリーでないエッジを除外
   ↓
5. API: pgRoutingで経路計算（A*法）
  - 重みは距離 + コスト（オプションで調整可能）
  - indoorオプションから屋内優先の経路制御
   ↓
6. API: 結果を返却（各ノードのis_indoor情報含む）
   ↓
7. クライアント: 地図上に経路を表示（屋内外を色分け可能）
```
