package db

import (
	"database/sql"
	"db-portal/internal/types"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// This format is commonly used for data interchange, especially for tabular data exports,
// and is easy to consume in any languages.
// It includes metadata about the fields and their types, followed by the actual data rows.
// The output is a JSON object with "fields", "types", and "rows" keys
// Supported field types: string, boolean, integer, numeric, date, time, datetime, duration, array, object, list
//
// Example output format:
/*
{
 "fields": ["id", "name", "active", "height"],
 "types": ["integer", "string", "boolean", "number"],
 "rows": [
   [1, "Alice", true, 1.72],
   [2, "Bob", false, 1.80],
 ]
}
*/

type TABfields struct {
	Fields []string `json:"fields"`
}
type TABtypes struct {
	Types []string `json:"types"`
}

func RowsToJsonTabular(rows *sql.Rows, file *os.File, dbVendor types.DBVendor) error {

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	if len(cols) != len(columnTypes) {
		return fmt.Errorf("mismatch between column names and types: %d names, %d types", len(cols), len(columnTypes))
	}
	colsLen := len(cols)

	f := TABfields{Fields: make([]string, colsLen)}
	t := TABtypes{Types: make([]string, colsLen)}
	for i, ct := range columnTypes {
		f.Fields[i] = cols[i]
		t.Types[i] = SQLToJsonType(ct.DatabaseTypeName())
	}

	rowCount := 0
	firstRow := true
	wroteMeta := false

	// Write opening brace
	if _, err := file.Write([]byte("{")); err != nil {
		return err
	}

	for rows.Next() {
		values := make([]any, colsLen)
		valuePtrs := make([]any, colsLen)
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}
		for i := range columnTypes {
			switch v := values[i].(type) {
			case []byte:
				values[i] = string(v)
			}
		}
		// On first row, infer type from value to update unknown field type, then write meta
		if rowCount == 0 {
			if err := writeMetaSection(file, f, t); err != nil {
				return err
			}
			// Now write "data":[
			if _, err := file.Write([]byte(`"rows":[`)); err != nil {
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
		for i := range t.Types {
			if t.Types[i] == "" {
				t.Types[i] = "string"
			}
		}
	}

	// If there were no rows, still need to write meta and data
	if !wroteMeta {
		if err := writeMetaSection(file, f, t); err != nil {
			return err
		}
		if _, err := file.Write([]byte(`"rows":[`)); err != nil {
			return err
		}
	}

	// Write closing
	if _, err := file.Write([]byte("]}")); err != nil {
		return err
	}

	return nil
}

// Helper to write the meta section
func writeMetaSection(file *os.File, f TABfields, t TABtypes) error {
	fieldsBytes, err := json.Marshal(f.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}
	typesBytes, err := json.Marshal(t.Types)
	if err != nil {
		return fmt.Errorf("failed to marshal types: %w", err)
	}
	// Join both JSON objects with a comma
	buf := append([]byte(`"fields":`), fieldsBytes...)
	buf = append(buf, []byte(`,"types":`)...)
	buf = append(buf, typesBytes...)
	buf = append(buf, ',')
	_, err = file.Write(buf)
	return err
}

// normalizeSQLType simplifies SQL types for matching.
func normalizeSQLType(sqlType string) string {
	s := strings.ToUpper(string(sqlType))
	s = strings.TrimSpace(s)

	// Remove content in parentheses
	if idx := strings.Index(s, "("); idx != -1 {
		s = s[:idx]
	}

	// Keep first member (e.g. "UNSIGNED INT" becomes "INT")
	fields := strings.Fields(s)
	if len(fields) > 0 {
		s = fields[0]
	}

	// remove extra prefixes
	prefixes := []string{"TINY", "SMALL", "MEDIUM", "BIG", "FIXED"}
	for _, prefix := range prefixes {
		if after, ok := strings.CutPrefix(s, prefix); ok {
			s = after
		} else if after0, ok0 := strings.CutPrefix(s, prefix); ok0 {
			s = after0
		}
	}

	// Remove extra trailing numbers (e.g. "INT4" becomes "INT", "FLOAT32" becomes "FLOAT")
	i := len(s)
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	s = s[:i]

	// Remove trailing "Z" (e.g. "TIMESTAMPZ" becomes "TIMESTAMP")
	s = strings.TrimSuffix(s, "Z")

	return s
}

func SQLToJsonType(sqlType string) string {
	switch normalizeSQLType(sqlType) {
	case "TEXT", "VARCHAR", "CHAR", "NVARCHAR", "STRING", "BINARY", "VARBINARY", "BLOB", "UUID":
		return "string"

	case "BOOL", "BOOLEAN", "BIT":
		return "boolean"
	case "INTEGER", "SERIAL", "INT":
		return "integer"
	case "NUMERIC", "DECIMAL", "REAL", "DOUBLE PRECISION", "FLOAT":
		return "numeric"

	case "DATE":
		return "date"
	case "TIME":
		return "time"
	case "TIMESTAMP", "DATETIME":
		return "datetime"
	case "INTERVAL":
		return "duration"

	case "ARRAY", "TEXT[]", "INTEGER[]", "_TEXT", "_INTEGER":
		return "array"
	case "JSON", "JSONB":
		return "object"
	case "ENUM":
		return "list"

	default:
		return "string"
	}
}
