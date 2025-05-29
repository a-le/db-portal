package db

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"
)

// RowsToJson writes the result set from rows as a JSON array of objects to the given file.
// Each row is encoded as a JSON object with column names as keys.
// []byte values are converted to strings for JSON compatibility.
// The output is a valid JSON array, suitable for large result sets as it streams rows directly to the file.
func RowsToJson(rows *sql.Rows, file *os.File) (err error) {
	cols, err := rows.Columns()
	if err != nil {
		return
	}

	encoder := json.NewEncoder(file)
	_, err = file.Write([]byte("["))
	if err != nil {
		return err
	}

	first := true
	for rows.Next() {
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			break
		}

		rowMap := make(map[string]any, len(cols))
		for i, c := range cols {
			switch v := values[i].(type) {
			case []byte:
				rowMap[c] = string(v)
			default:
				rowMap[c] = v
			}
		}

		if !first {
			if _, err := file.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		if err := encoder.Encode(rowMap); err != nil {
			return err
		}
	}

	if !first {
		file.Seek(-1, io.SeekCurrent) // Remove the newline added by encoder.
	}
	_, err = file.Write([]byte("]")) // Close the JSON array

	return
}
