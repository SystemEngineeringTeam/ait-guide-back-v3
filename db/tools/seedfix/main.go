// seedfix は db/seeds/<データセット名>（デフォルトのdb/seeds/default、
// 将来的にはdb/seeds/events/<年>/<イベント名>等）配下のCSVを、
// アプリ側のローダー/DBスキーマが要求する
// 形式に正規化するスタンドアロンのメンテナンスツール。
//
// アプリ本体（internal/ 以下）には一切依存せず、標準ライブラリのみで完結する。
// シードCSVは外部のマップ編集ツールから再取得・再投入されることがあり、
// 以下のような既知の不整合が再発しやすい。本ツールはそれらを機械的に修正する。
//
//   - edges.csv: 不要な先頭 "id" 列が付与されている
//     （ローダーは node_id_from,node_id_target,distance,... の7列を期待）
//   - nodes.csv: node_type に DB未対応の値（例: "passage"）が混入している
//     （DBの chk_node_type は entrance/road/door/facility のみ許可。
//     passage は road の意味で使われているため road に正規化する）
//   - nodes.csv: 行ごとに列数が異なる（末尾の building_id/floor が省略されている）
//     ため、Go標準の encoding/csv がヘッダ行と列数不一致でエラーになる
//
// 使い方:
//
//	go run ./db/tools/seedfix [-dir db/seeds/default] [-dry-run]
//
// -dry-run を指定すると、変更内容をレポートするだけでファイルは書き換えない。
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// nodeTypeAlias はDBが受け付けない node_type 値から、正規の値へのマッピング。
// db/migrations/001_create_tables.sql の chk_node_type に合わせて更新すること。
var nodeTypeAlias = map[string]string{
	"passage": "road",
}

// validNodeTypes はDBの chk_node_type が許可する値の集合。
var validNodeTypes = map[string]bool{
	"entrance": true,
	"road":     true,
	"door":     true,
	"facility": true,
}

func main() {
	dir := flag.String("dir", "db/seeds/default", "シードCSVが置かれているディレクトリ")
	dryRun := flag.Bool("dry-run", false, "ファイルを書き換えず、修正内容の確認のみ行う")
	flag.Parse()

	var anyFixed bool

	if fixed, err := fixNodesCSV(filepath.Join(*dir, "nodes.csv"), *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "nodes.csv: %v\n", err)
		os.Exit(1)
	} else if fixed {
		anyFixed = true
	}

	if fixed, err := fixEdgesCSV(filepath.Join(*dir, "edges.csv"), *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "edges.csv: %v\n", err)
		os.Exit(1)
	} else if fixed {
		anyFixed = true
	}

	if !anyFixed {
		fmt.Println("修正対象なし（既に正規化済み）")
	}
}

// readAllRows はヘッダ行を含む全行を、列数の違いを許容して読み込む。
func readAllRows(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return rows, nil
}

func writeAllRows(path string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	w := csv.NewWriter(f)
	w.UseCRLF = false
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			_ = f.Close()
			return fmt.Errorf("write: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// fixNodesCSV は nodes.csv の node_type エイリアス解決と、行ごとの列数不整合
// （ヘッダより短い行を空文字で埋める）を修正する。
func fixNodesCSV(path string, dryRun bool) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("nodes.csv: スキップ（ファイルなし: %s）\n", path)
		return false, nil
	}

	rows, err := readAllRows(path)
	if err != nil {
		return false, err
	}
	if len(rows) < 2 {
		return false, nil
	}

	header := rows[0]
	const typeCol = 3 // id,lat,lng,type,building_id
	width := len(header)

	var (
		typeFixCount int
		padFixCount  int
		unknownTypes = map[string][]int{} // 未知のnode_type -> 行番号一覧
	)

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		lineNo := i + 1

		// 列数不足を末尾空文字で補完する（building_id/floorの省略を許容するため）。
		if len(row) < width {
			padded := make([]string, width)
			copy(padded, row)
			rows[i] = padded
			row = padded
			padFixCount++
		}

		if typeCol >= len(row) {
			continue
		}
		t := row[typeCol]
		if alias, ok := nodeTypeAlias[t]; ok {
			row[typeCol] = alias
			typeFixCount++
		} else if t != "" && !validNodeTypes[row[typeCol]] {
			unknownTypes[row[typeCol]] = append(unknownTypes[row[typeCol]], lineNo)
		}
	}

	for t, lines := range unknownTypes {
		fmt.Printf("nodes.csv: 警告: 未知のnode_type %q （%d行、例: line %d）。nodeTypeAliasへの追加を検討してください\n", t, len(lines), lines[0])
	}

	if typeFixCount == 0 && padFixCount == 0 {
		fmt.Println("nodes.csv: 修正なし")
		return false, nil
	}

	fmt.Printf("nodes.csv: node_type正規化 %d件、列数補完 %d件%s\n", typeFixCount, padFixCount, dryRunSuffix(dryRun))
	if dryRun {
		return true, nil
	}
	return true, writeAllRows(path, rows)
}

// fixEdgesCSV は edges.csv に不要な先頭 "id" 列が付いている場合、取り除く。
// ローダーが期待する形式は node_id_from,node_id_target,distance,level,
// has_stairs,is_accessible,is_indoor の7列。
func fixEdgesCSV(path string, dryRun bool) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("edges.csv: スキップ（ファイルなし: %s）\n", path)
		return false, nil
	}

	rows, err := readAllRows(path)
	if err != nil {
		return false, err
	}
	if len(rows) < 1 {
		return false, nil
	}

	header := rows[0]
	if len(header) == 0 || header[0] != "id" {
		fmt.Println("edges.csv: 修正なし")
		return false, nil
	}

	stripped := make([][]string, len(rows))
	for i, row := range rows {
		if len(row) <= 1 {
			stripped[i] = row
			continue
		}
		stripped[i] = row[1:]
	}

	fmt.Printf("edges.csv: 先頭id列を削除（%d行）%s\n", len(rows)-1, dryRunSuffix(dryRun))
	if dryRun {
		return true, nil
	}
	return true, writeAllRows(path, stripped)
}

func dryRunSuffix(dryRun bool) string {
	if dryRun {
		return "（dry-run: 未書き込み）"
	}
	return ""
}
