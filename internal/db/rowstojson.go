package db

import (
	"database/sql"
	"encoding/json"
)

func RowsToJson(rows *sql.Rows) (jsonData []byte, err error) {
	cols, err := rows.Columns()
	if err != nil {
		return
	}

	result := make([]map[string]any, 0)
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
			v := values[i]
			if b, ok := v.([]byte); ok {
				rowMap[c] = string(b) // Convert byte slices to strings
			} else {
				rowMap[c] = v // Keep other types as is
			}
		}
		result = append(result, rowMap)
	}

	jsonData, err = json.Marshal(result)

	return
}
