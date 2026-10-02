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
