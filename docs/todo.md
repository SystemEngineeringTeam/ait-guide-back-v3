# 残タスク

現時点では未着手の設計・実装課題を記録する。対応時はこのファイルから該当項目を削除する。

## 1. 経路探索APIをbuilding_idではなくbuilding.keyで指定できるようにする

**現状:**

- `/routes/search/{building_id}`・`/get/route?end=`はともに`buildings.building_no`（数値、欠番あり、CSV由来）で建物を指定する仕様
- `buildings.key`（`B1`, `AIT_PLAZA`等の安定した文字列キー）は`buildings`テーブルに存在し、`/buildings`系APIのレスポンスには含まれているが、経路探索APIの入力としては使えない

**やりたいこと:**

- 経路探索APIの建物指定を`building_no`ではなく`buildings.key`で行えるようにする
- `building_no`は欠番があり、意味を持たない内部的な採番であるのに対し、`key`は人が読んで分かる安定した識別子のため、APIの入力としてより適切

**想定する対応:**

- `nodes`/`rooms`/`photos`.`building_id`を`key`参照に置き換える（カラム追加 or 置き換え）
- `/routes/search/{building_id}`・`/routes/search/{key}`のようにパスパラメータをkeyベースに変更（後方互換が必要なら両対応も検討）
- `/get/route?end=`も同様にkey文字列を受け付けるよう変更
- フロントエンド（ait-guide-front-v3）側の呼び出し箇所の追従
- 詳細は[docs/schema.md](./schema.md#buildings建物)の`buildings`テーブルTODOも参照

## 2. 建物が指定されない場合、座標（緯度経度）を目的地として経路探索できるようにする

**現状:**

- 経路探索APIは必ず目的地として建物ID（`building_id`/`end`）を指定する必要がある
- 建物に属さない任意の地点（例: 広場の特定地点、工事中のエリアなど）を目的地にする手段がない

**やりたいこと:**

- 建物IDが指定されなかった場合、目的地の緯度経度をクエリパラメータ（例: `dest_lat`, `dest_lng`）として受け取り、その座標から最寄りのノードを目的地として経路探索できるようにする
- 現在地の最寄りノード検索（`FindNearestByLatLng`）と同様のロジックを目的地側にも使う想定

**想定する対応:**

- `/routes/search/{building_id}`を建物ID必須のパスパラメータから、建物ID or 座標のどちらかを受け付けるクエリベースの設計に見直す（または座標指定用の別エンドポイントを新設）
- `usecase.RouteUsecase`に、目的地座標から最寄りノードを解決する経路（`FindNearestByLatLng`を目的地側にも使うパス）を追加
- building_id未指定かつ座標も未指定の場合のバリデーション・エラーハンドリングを整理

## 3. 「屋内（is_indoor）」と「屋根あり（雨に濡れない）」を分けて扱う

**現状:**

- `edges.is_indoor`の1つのフラグのみで経路の性質を管理している（[docs/domain-rules.md](./domain-rules.md#屋内外経路の判定ルール)、[docs/schema.md](./schema.md)参照）
- `is_indoor = TRUE`は「建物内の移動（廊下、階段、エレベーター）」を表すが、屋外でも屋根付き通路（渡り廊下、アーケード等、雨に濡れない経路）が実際には存在し得る
- 現状の2値（屋内/屋外）では「屋外だが雨に濡れない」経路を区別できず、`indoor`オプション（雨天時に屋内優先、等の用途）で正しく考慮できない

**やりたいこと:**

- 「屋内かどうか（`is_indoor`）」と「雨に濡れないかどうか（屋根の有無）」を independent な軸として分離する
- 雨天時の経路探索オプションで、屋内ノードだけでなく屋根付き屋外経路も優先対象にできるようにする

**想定する対応:**

- `edges`テーブルに`is_covered`（または`is_roofed`）BOOLEANカラムを追加するマイグレーション
  - `is_indoor = TRUE`の行は基本的に`is_covered = TRUE`（建物内は雨に濡れない）
  - `is_indoor = FALSE`でも`is_covered = TRUE`になり得る行（屋根付き屋外通路）を区別
- `db/seeds/default/edges.csv`のCSVフォーマットに列を追加し、ローダー（`internal/infra/loader/csv.go`）を対応させる
- `model.RouteOption.Indoor`を「雨に濡れないルート優先」の意味に寄せるか、新たに`Covered`相当のオプションを追加するか設計し、`service.CostWeightFromOption`等のコスト計算ロジックを見直す
- `docs/domain-rules.md`・`docs/schema.md`の該当記述を更新

## 4. 建物情報の充実（説明、写真など）

**現状:**

- `buildings.description`・`photos`テーブルは存在するが、`db/seeds/default/buildings.csv`の`description`列は全行空欄、`photos.csv`も未投入
- `/buildings/{building_id}`は写真・説明を返せる実装になっているが、データが無いため実質機能していない

**やりたいこと:**

- 各建物の説明文・外観/内観写真を整備し、`buildings.csv`の`description`・`photos.csv`に投入する
- 建物ごとの所属（`affiliation`）も合わせて埋める

**想定する対応:**

- 建物ごとの説明文・写真素材を収集
- `db/seeds/default/buildings.csv`の`description`/`affiliation`列を埋める
- `db/seeds/default/photos.csv`（現状未作成）を新設し投入する

## 5. 部屋を含めた経路探索

**現状:**

- `rooms`テーブル・`rooms.csv`のローダー（`SeedRooms`）は実装済みだが、`db/seeds/default/rooms.csv`自体が未作成で部屋データが投入されていない
- 経路探索（`SearchRoute`/`SearchRouteLegacy`）は建物単位（`entrance`/`facility`）までしか目的地にできず、部屋（`door`ノード）を目的地にする手段がない

**やりたいこと:**

- 部屋（教室・研究室等）を目的地に指定して、建物内の該当ドアまで経路探索できるようにする

**想定する対応:**

- `db/seeds/default/rooms.csv`を作成し、各部屋と`node_id`（`door`ノード）の対応を投入する
- 経路探索APIに部屋ID（`room_id`）指定の経路を追加する（建物ID解決と同様に、room_idから対応する`door`ノードを引く`FindByRoomID`相当のリポジトリメソッドが必要）
- 建物の入口から部屋のドアまでの屋内経路（`is_indoor = TRUE`のエッジ）が正しく繋がっているか、シードデータ側の整備も必要

## 6. フロアを考慮した経路探索

**現状:**

- `nodes.floor`・`rooms.floor`カラムは存在するが、現状のシードデータは全ノードが同一フロア想定で作成されている
- 経路探索ロジック（`pgr_dijkstra`）はフロアを考慮しておらず、階段・エレベーターでのフロア間移動を表現するエッジや、フロア違いのノード間を正しく区別する仕組みがない

**やりたいこと:**

- 複数フロアを持つ建物で、フロアをまたいだ経路探索（階段・エレベーター経由）を正しく行えるようにする

**想定する対応:**

- 階段・エレベーターを表すノード/エッジの設計（同じ位置で`floor`違いのノードをペアで持ち、`has_stairs`等でつなぐ等）を検討
- マルチフロアに対応したシードデータ（`nodes.csv`/`edges.csv`）を作成
- フロア指定オプション（現在地のfloor、目的地のfloor）をAPI・経路探索ロジックに追加するか検討
- `docs/domain-rules.md`のフロアに関する記述を更新

## 7. 自販機・ロッカー情報（スキーマ変更あり）

**現状:**

- 自販機・ロッカー等の設備情報を管理するテーブルが存在しない

**やりたいこと:**

- 自販機・ロッカーなどの設備の位置・情報を管理し、検索や経路探索の目的地にできるようにする

**想定する対応:**

- 新規テーブル（例: `amenities`）の設計（種別、位置orノード紐付け、building_id、floor等）とマイグレーション追加
- 対応するシードCSV・ローダー（`internal/infra/loader/csv.go`）の追加
- 一覧取得・検索用のAPIエンドポイント追加（`internal/handler`/`internal/usecase`）
- 必要であれば経路探索の目的地としても指定できるようにする

## 8. イベントなどの特殊マップの適用

**現状:**

- `SEED_DIR`（デフォルト`db/seeds/default`）配下の単一のnodes/edges/buildings等のCSVのみを前提にしたデータ投入・経路探索になっている
- 学園祭等のイベント時に、通常と異なる通行可否・臨時設置の出店/設備などを反映した「特殊マップ」に切り替える仕組みがない
- `db/seeds/events/<年>/<イベント名>/`というディレクトリ構成のみ作成済み（[db/seeds/events/README.md](../db/seeds/events/README.md)参照）。読み込み側の実装は未着手

**やりたいこと:**

- `db/seeds/events/<年>/<イベント名>/`配下に、イベント開催時のみ使うデータ（臨時ノード・エッジ・通行止め情報等）を格納できるようにする
- アプリケーション起動時にイベント名を指定すると、そのイベント用データを反映した状態で立ち上がるようにする
- デフォルトマップとイベントマップの「つなぎ込み」（イベント用ノード/エッジをデフォルトのグラフにどう合成するか）の設計が重要な検討事項

**想定する対応:**

- イベントデータの形式を設計する（デフォルトCSVの差分として追加/上書き/除外するノード・エッジを表現する方式 等）
- `SEED_DIR`に加えて環境変数（例: `EVENT_NAME`）等でイベントディレクトリを指定できるよう`internal/config`・`cmd/server/main.go`・`internal/infra/loader`を拡張
- デフォルトマップとイベントマップのマージ戦略（起動時にDBへ両方投入するか、イベント専用のオーバーレイとして扱うか等）を設計・決定する
- 通行止め（既存エッジの無効化）・臨時ノード追加の両方を表現できるデータ構造にする
- イベント終了後にデフォルトマップへ戻す運用（再起動 or 切り替えAPI）も含めて検討する
