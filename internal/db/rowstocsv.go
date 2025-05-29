package db

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
)

// RowsToCsv streams rows as CSV to a file at 'path'.
func RowsToCsv(rows *sql.Rows, file *os.File) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(cols); err != nil {
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

		record := make([]string, len(cols))
		for i := range cols {
			switch v := values[i].(type) {
			case nil:
				record[i] = ""
			case []byte:
				record[i] = string(v)
			case string:
				record[i] = v
			default:
				record[i] = fmt.Sprintf("%v", v)
			}
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}
