package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

// RowsToJsonCompact writes the result set from rows to the given file in ClickHouse JSONCompact format.
// The output includes a "meta" section with column names and types, a "data" array with each row as a JSON array,
// and summary fields such as "rows", "rows_before_limit_at_least", and "statistics".
// Column types are inferred from the database driver and the first row of data.
// []byte values are converted to strings for JSON compatibility.
// The function streams data directly to the file, making it suitable for large result sets.
// see https://clickhouse.com/docs/interfaces/formats/JSONCompact
// beta. needs some extensive tests
func RowsToJsonCompact(rows *sql.Rows, file *os.File) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	colTypes := make([]string, len(cols))
	for i, ct := range columnTypes {
		// ScanType may return an empty interface for unsupported drivers
		colTypes[i] = scanTypeToClickhouseType(ct.ScanType())
	}

	rowCount := 0
	firstRow := true
	wroteMeta := false

	// Write opening brace
	if _, err := file.Write([]byte("{")); err != nil {
		return err
	}

	for rows.Next() {
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}
		for i := range cols {
			switch v := values[i].(type) {
			case []byte:
				values[i] = string(v)
			}
		}
		// On first row, update colTypes and write meta
		if rowCount == 0 {
			for i, v := range values {
				colTypes[i] = goTypeToClickhouseType(v)
			}
			if err := writeMetaSection(file, cols, colTypes); err != nil {
				return err
			}
			// Now write "data":[
			if _, err := file.Write([]byte(`"data":[`)); err != nil {
				return err
			}
			wroteMeta = true
		}
		rowBytes, err := json.Marshal(values)
		if err != nil {
			return err
		}
		if !firstRow {
			if _, err := file.Write([]byte(",")); err != nil {
				return err
			}
		}
		if _, err := file.Write(rowBytes); err != nil {
			return err
		}
		firstRow = false
		rowCount++
	}

	// If there were no rows, still need to write meta and data
	if !wroteMeta {
		if err := writeMetaSection(file, cols, colTypes); err != nil {
			return err
		}
		if _, err := file.Write([]byte(`"data":[`)); err != nil {
			return err
		}
	}

	if _, err := file.Write([]byte("],")); err != nil {
		return err
	}

	// Write the rest of the JSON
	tail := fmt.Sprintf(`"rows":%d,"rows_before_limit_at_least":%d,"statistics":{"elapsed":0.0,"rows_read":%d,"bytes_read":0}}`, rowCount, rowCount, rowCount)
	if _, err := file.Write([]byte(tail)); err != nil {
		return err
	}

	return nil
}

// Helper to map reflect.Type to ClickHouse type (used for initial colTypes)
func scanTypeToClickhouseType(t reflect.Type) string {
	if t == nil {
		return "String"
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "Int64"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "UInt64"
	case reflect.Float32, reflect.Float64:
		return "Float64"
	case reflect.Bool:
		return "Bool"
	case reflect.String:
		return "String"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "String" // []byte
		}
		return "Array(String)"
	case reflect.Struct:
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "DateTime"
		}
	}
	return "String"
}

// goTypeToClickhouseType maps Go types to ClickHouse types (basic mapping)
func goTypeToClickhouseType(v any) string {
	switch v := v.(type) {
	case nil:
		return "Nullable(String)"
	case int, int8, int16, int32, int64:
		return "Int64"
	case uint, uint8, uint16, uint32, uint64:
		return "UInt64"
	case float32, float64:
		return "Float64"
	case bool:
		return "Bool"
	case string:
		return "String"
	case time.Time:
		return "DateTime"
	case []byte:
		return "String"
	case []any:
		return "Array(String)"
	default:
		// Try to detect slice/array
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return "Array(String)"
		}
		return "String"
	}
}

// Helper to write the meta section
func writeMetaSection(file *os.File, cols, colTypes []string) error {
	metaBuf := []byte(`"meta":[`)
	for i, col := range cols {
		if i > 0 {
			metaBuf = append(metaBuf, ',')
		}
		meta := map[string]string{"name": col, "type": colTypes[i]}
		b, _ := json.Marshal(meta)
		metaBuf = append(metaBuf, b...)
	}
	metaBuf = append(metaBuf, "],"...)
	_, err := file.Write(metaBuf)
	return err
}
