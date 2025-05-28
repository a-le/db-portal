package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tealeg/xlsx/v3"
)

// RowsToXlsx converts a sql.Rows to an xlsx file
func RowsToXlsx(rows *sql.Rows, path string) (err error) {
	// Create a new Excel file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	var cols []string
	if cols, err = rows.Columns(); err != nil {
		return
	}

	// Write the column headers
	headerRow := sheet.AddRow()
	for _, colName := range cols {
		cell := headerRow.AddCell()
		cell.SetString(colName)
	}

	// Iterate over the rows from the query
	for rows.Next() {
		// Prepare a slice to hold the values from the query
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))

		// Create value pointers
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row values into value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Add a new row to the Excel sheet
		row := sheet.AddRow()

		// Loop through each column and add the value to the cell
		for _, val := range values {
			cell := row.AddCell()

			// Preserve the original type while writing to the cell
			switch v := val.(type) {
			case int64: // For integer types
				cell.SetInt64(v)
			case float64: // For float types
				cell.SetFloat(v)
			case bool: // For boolean types
				cell.SetBool(v)
			case []byte: // For text fields that are returned as []byte
				cell.SetString(string(v))
			case string: // For string types
				cell.SetString(v)
			case time.Time: // For time.Time (dates/timestamps)
				cell.SetDate(v)
			case nil: // Handle NULL values by setting an empty cell
			default: // Fallback to string representation for any other types
				cell.SetString(fmt.Sprintf("%v", val))
			}
		}
	}

	// Check for errors after row iteration
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error during row iteration: %w", err)
	}

	// Save the Excel file to the specified path
	if err := file.Save(path); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	return nil
}
