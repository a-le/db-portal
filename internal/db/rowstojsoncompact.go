package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// RowsToJsonCompact exports the result set from rows to the specified file in ClickHouse JSONCompact format.
// The output contains a "meta" section with column names and types, a "data" array where each row is a JSON array,
// and summary fields such as "rows", "rows_before_limit_at_least", and "statistics".
// Column types are determined from the database driver when available, or inferred from the first row of data for unknown types.
// Type mapping is basic. []byte values are converted to strings for JSON compatibility.
// Data is streamed directly to the file, making this function suitable for large result sets.
// See: https://clickhouse.com/docs/interfaces/formats/JSONCompact
// Note: This is a beta implementation and requires thorough testing.
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
		if ct.ScanType() == nil {
			colTypes[i] = ""
			continue
		}
		switch ct.ScanType().String() {
		case "interface{}", "interface {}":
			colTypes[i] = ""
		default:
			colTypes[i] = goTypeToExportType(ct.ScanType().String())
		}
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
		// On first row, update unknown colTypes and write meta
		if rowCount == 0 {
			for i, v := range values {
				if colTypes[i] == "" {
					if reflect.TypeOf(v) == nil {
						colTypes[i] = ""
					} else {
						colTypes[i] = goTypeToExportType(reflect.TypeOf(v).String())
					}
				}
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

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	if rowCount == 0 {
		for i := range colTypes {
			if colTypes[i] == "" {
				colTypes[i] = "Nullable(String)"
			}
		}
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

// Helper to write the meta section
func writeMetaSection(file *os.File, cols, colTypes []string) error {
	metaBuf := []byte(`"meta":[`)
	for i, col := range cols {
		if i > 0 {
			metaBuf = append(metaBuf, ',')
		}
		meta := map[string]string{"name": col, "type": colTypes[i]}
		b, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta for column %s: %w", col, err)
		}
		metaBuf = append(metaBuf, b...)
	}
	metaBuf = append(metaBuf, "],"...)
	_, err := file.Write(metaBuf)
	return err
}

func goTypeToExportType(goType string) string {
	switch goType {
	case "bool":
		return "Nullable(Bool)"
	case "[]bool":
		return "Array(Bool)"
	case "[]byte":
		return "Nullable(String)"
	case "float32":
		return "Nullable(Float32)"
	case "[]float32":
		return "Array(Float32)"
	case "float64":
		return "Nullable(Float64)"
	case "[]float64":
		return "Array(Float64)"
	case "int":
		return "Nullable(Int64)"
	case "[]int":
		return "Array(Int64)"
	case "int16":
		return "Nullable(Int16)"
	case "[]int16":
		return "Array(Int16)"
	case "int32":
		return "Nullable(Int32)"
	case "[]int32":
		return "Array(Int32)"
	case "int64":
		return "Nullable(Int64)"
	case "[]int64":
		return "Array(Int64)"
	case "int8":
		return "Nullable(Int8)"
	case "[]int8":
		return "Array(Int8)"
	case "string":
		return "Nullable(String)"
	case "[]string":
		return "Array(String)"
	case "time.Time":
		return "Nullable(DateTime)"
	case "[]time.Time":
		return "Array(DateTime)"
	case "uint":
		return "Nullable(UInt64)"
	case "[]uint":
		return "Array(UInt64)"
	case "uint16":
		return "Nullable(UInt16)"
	case "[]uint16":
		return "Array(UInt16)"
	case "uint32":
		return "Nullable(UInt32)"
	case "[]uint32":
		return "Array(UInt32)"
	case "uint64":
		return "Nullable(UInt64)"
	case "[]uint64":
		return "Array(UInt64)"
	case "uint8":
		return "Nullable(UInt8)"
	case "[]uint8":
		return "Array(UInt8)"
	case "[]interface {}":
		return "Array(String)"
	default:
		return "Nullable(String)"
	}
}
