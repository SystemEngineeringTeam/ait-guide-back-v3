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
- `db/seeds/edges.csv`のCSVフォーマットに列を追加し、ローダー（`internal/infra/loader/csv.go`）を対応させる
- `model.RouteOption.Indoor`を「雨に濡れないルート優先」の意味に寄せるか、新たに`Covered`相当のオプションを追加するか設計し、`service.CostWeightFromOption`等のコスト計算ロジックを見直す
- `docs/domain-rules.md`・`docs/schema.md`の該当記述を更新
