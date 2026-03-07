# データベーススキーマ

## 設計方針

- **屋内外の統合管理**: 単一の `nodes` と `edges` テーブルで管理、`is_indoor` フラグで区別
- **FK制約なし**: 柔軟なデータ投入、パフォーマンス優先
- **CHECK制約**: 最低限のデータ検証
- **空間インデックス**: 近傍検索の高速化
- **建物単位の管理**: 屋内ノードは `building_id` で所属建物を明示

## テーブル一覧

| テーブル名  | 用途       | 主要カラム                                                                                    |
| ----------- | ---------- | --------------------------------------------------------------------------------------------- |
| `buildings` | 建物情報   | id, name, description                                                                         |
| `nodes`     | 経路ノード | id, node_id, geom(Point), node_type, building_id, floor                                       |
| `edges`     | 経路エッジ | id, node_id_from, node_id_target, distance, cost, level, has_stairs, is_accessible, is_indoor |
| `rooms`     | 部屋情報   | id, room_id, building_id, node_id, name, description                                          |
| `photos`    | 建物写真   | id, building_id, url, caption, display_order                                                  |

---

## テーブル詳細

### buildings（建物）

| カラム名    | 型           | NULL     | 説明     |
| ----------- | ------------ | -------- | -------- |
| id          | SERIAL       | NOT NULL | 主キー   |
| name        | VARCHAR(200) | NOT NULL | 建物名   |
| description | TEXT         | NULL     | 建物説明 |
| affiliation | VARCHAR(100) | NULL     | 所属     |

**制約:**

- PRIMARY KEY (id)

**インデックス:** なし（小規模データのため）

**備考:** 建物の入口ノードは `nodes` テーブルで `node_type='entrance'` および `building_id` で特定

---

### nodes（経路ノード）

| カラム名    | 型                    | NULL     | 説明                         |
| ----------- | --------------------- | -------- | ---------------------------- |
| id          | SERIAL                | NOT NULL | 主キー（pgRouting用）        |
| node_id     | VARCHAR(50)           | NOT NULL | アプリケーション用ID         |
| geom        | GEOMETRY(Point, 4326) | NOT NULL | 地理座標（緯度経度）         |
| node_type   | VARCHAR(50)           | NOT NULL | ノード種別                   |
| building_id | INTEGER               | NULL     | 所属建物ID（屋内ノードのみ） |
| floor       | INTEGER               | NULL     | 階数（屋内ノードのみ）       |

**node_type の値:**

- `entrance`: 建物入口
- `road`: 道路/廊下上のノード
- `door`: 部屋のドア

**制約:**

- PRIMARY KEY (id)
- UNIQUE (node_id)
- CHECK (node*id ~ '^[a-z0-9*]+$')
- CHECK (node_type IN ('entrance', 'road', 'door'))
- CHECK (floor IS NULL OR (floor >= -10 AND floor <= 100))

**インデックス:**

- `idx_nodes_geom` (GIST on geom) - 空間検索用
- `idx_nodes_building` (building_id) - 建物内ノード検索用
- `idx_nodes_type` (node_type) - タイプ検索用

**備考:**

- 屋外ノード: `building_id` と `floor` は NULL
- 屋内ノード: `building_id` が設定され、`floor` で階数を管理

---

### edges（経路エッジ）

| カラム名       | 型          | NULL     | 説明                            |
| -------------- | ----------- | -------- | ------------------------------- |
| id             | SERIAL      | NOT NULL | 主キー                          |
| node_id_from   | VARCHAR(50) | NOT NULL | 開始ノードID                    |
| node_id_target | VARCHAR(50) | NOT NULL | 終了ノードID                    |
| distance       | FLOAT       | NOT NULL | 距離（メートル）                |
| cost           | FLOAT       | NULL     | コスト（任意単位）              |
| level          | INTEGER     | NOT NULL | 経路の主要度 1(大通り)〜5(脇道) |
| has_stairs     | BOOLEAN     | NOT NULL | 階段を含むか                    |
| is_accessible  | BOOLEAN     | NOT NULL | バリアフリーか                  |
| is_indoor      | BOOLEAN     | NOT NULL | 屋内経路か                      |

**制約:**

- PRIMARY KEY (id)
- CHECK (node*id_from ~ '^[a-z0-9*]+$')
- CHECK (node*id_target ~ '^[a-z0-9*]+$')
- CHECK (node_id_from != node_id_target)
- CHECK (distance >= 0)
- CHECK (cost IS NULL OR cost >= 0)

**インデックス:**

- `idx_edges_from` (node_id_from)
- `idx_edges_target` (node_id_target)
- `idx_edges_indoor` (is_indoor) - 屋内外フィルタ用

**備考:**

- `is_indoor = FALSE`: 屋外経路（建物間移動）
- `is_indoor = TRUE`: 屋内経路（建物内移動）
- エレベーター情報は `has_stairs = FALSE` + 経路名などで判別

---

### rooms（部屋）

| カラム名    | 型           | NULL     | 説明                         |
| ----------- | ------------ | -------- | ---------------------------- |
| id          | SERIAL       | NOT NULL | 主キー                       |
| room_id     | VARCHAR(50)  | NOT NULL | 部屋ID                       |
| building_id | INTEGER      | NOT NULL | 所属建物ID                   |
| node_id     | VARCHAR(50)  | NULL     | ドアノードID (nodes.node_id) |
| name        | VARCHAR(200) | NULL     | 部屋名                       |
| description | TEXT         | NULL     | 部屋説明                     |
| floor       | INTEGER      | NULL     | 階数                         |
| capacity    | INTEGER      | NULL     | 収容人数                     |
| room_type   | VARCHAR(50)  | NULL     | 部屋タイプ                   |

**room_type の値例:**

- `classroom`: 教室
- `office`: オフィス
- `lab`: 実験室
- `lecture_hall`: 講堂

**制約:**

- PRIMARY KEY (id)
- UNIQUE (room_id)
- CHECK (room*id ~ '^[a-z0-9*]+$')
- CHECK (node*id IS NULL OR node_id ~ '^[a-z0-9*]+$')
- CHECK (capacity IS NULL OR capacity >= 0)
- CHECK (floor IS NULL OR (floor >= -10 AND floor <= 100))

**インデックス:**

- `idx_rooms_building` (building_id)
- `idx_rooms_node` (node_id)
- `idx_rooms_type` (room_type)

---

### photos（写真）

| カラム名      | 型      | NULL     | 説明         |
| ------------- | ------- | -------- | ------------ |
| id            | SERIAL  | NOT NULL | 主キー       |
| building_id   | INTEGER | NOT NULL | 対象建物ID   |
| url           | TEXT    | NOT NULL | 写真URL      |
| caption       | TEXT    | NULL     | キャプション |
| display_order | INTEGER | NOT NULL | 表示順序     |

**制約:**

- PRIMARY KEY (id)
- CHECK (url ~ '^https?://')
- CHECK (display_order >= 0)

**インデックス:**

- `idx_photos_building` (building_id)
- `idx_photos_order` (building_id, display_order)

---

## pgRouting用ビュー

### edges_for_routing

経路探索用のビュー。`node_id`（文字列）を整数IDに変換。

**カラム:**

- id, source, target, distance, cost, has_stairs, is_accessible, is_indoor
- node_id_from, node_id_target (デバッグ用)
- building_id_from, building_id_target (屋内経路の建物判定用)

**ビュー定義:**

```sql
CREATE OR REPLACE VIEW edges_for_routing AS
SELECT
    e.id,
    n1.id as source,
    n2.id as target,
    e.distance,
    COALESCE(e.cost, 0) as cost,
    e.has_stairs,
    e.is_accessible,
    e.is_indoor,
    e.node_id_from,
    e.node_id_target,
    n1.building_id as building_id_from,
    n2.building_id as building_id_target
FROM edges e
JOIN nodes n1 ON e.node_id_from = n1.node_id
JOIN nodes n2 ON e.node_id_target = n2.node_id;
```

**使用例:**

```sql
-- 屋外のみの経路探索
SELECT * FROM pgr_dijkstra(
    'SELECT id, source, target, distance as cost
     FROM edges_for_routing WHERE is_indoor = FALSE',
    start_id, end_id, FALSE
);

-- 特定建物内の経路探索
SELECT * FROM pgr_dijkstra(
    'SELECT id, source, target, distance as cost
     FROM edges_for_routing
     WHERE is_indoor = TRUE
       AND building_id_from = 1
       AND building_id_target = 1',
    start_id, end_id, FALSE
);
```

---

## 整合性チェック関数

FK制約を使用しない代わりに、PL/pgSQL関数でデータ整合性を検証する。

### check_edges_integrity()

エッジの整合性を検証:

- エッジが参照するノード（from/target）が存在するか

### check_buildings_integrity()

建物の整合性を検証:

- 屋内ノードの `building_id` が存在するか

### check_rooms_integrity()

部屋の整合性を検証:

- 部屋が参照する建物が存在するか
- 部屋が参照するドアノードが存在するか

### check_all_integrity()

全体の整合性を一括検証:

- 上記すべてのチェックを実行
- エラーがあれば詳細を返却

**チェック項目一覧:**

1. エッジが参照するノード（from/target）が存在するか
2. 部屋が参照する建物・ドアノードが存在するか
3. 屋内ノードの building_id が存在するか
4. 屋内エッジの両端ノードが同じ building_id に属するか

---

## 座標系

**SRID 4326 (WGS84):**

- 世界測地系
- GPS座標系と同じ
- 緯度: -90 〜 90
- 経度: -180 〜 180

**距離計算:**

- 平面距離: `ST_Distance(geom1, geom2)` - 度単位
- 測地線距離: `ST_Distance(geom1::geography, geom2::geography)` - メートル単位
