# イベント用シードデータ

学園祭などのイベント開催時のみ使用する、臨時のノード・エッジ・建物情報を格納するディレクトリ。

## ディレクトリ構成（予定）

```
db/seeds/events/<年>/<イベント名>/
├── buildings.csv
├── nodes.csv
└── edges.csv
```

例: `db/seeds/events/2026/gakusai/nodes.csv`

## 現状

ディレクトリ構成のみ用意している段階で、読み込み側（`internal/infra/loader`）の対応や、
デフォルトマップ（[db/seeds/default/](../default/)）との統合方法は未実装。詳細な残タスクは
[docs/todo.md](../../../docs/todo.md) の「8. イベントなどの特殊マップの適用」を参照。
