package dbutil

import (
	"db-portal/internal/types"
	"fmt"
	"strings"
)

func DriverName(dbVendor string) (driverName string, err error) {
	Drivers := map[string]string{
		types.DBVendorClickHouse: "clickhouse",
		types.DBVendorMySQL:      "mysql",
		types.DBVendorMariaDB:    "mysql",
		types.DBVendorMSSQL:      "sqlserver",
		types.DBVendorPostgres:   "pgx",
		types.DBVendorSQLite:     "sqlite3",
	}
	driverName = Drivers[dbVendor]
	if driverName == "" {
		err = fmt.Errorf("unknown database type: %s", dbVendor)
	}
	return
}

// PlaceholderStyle returns the placeholder style for the given database vendor.
func PlaceholderStyle(dbVendor string) (placeholder string, err error) {
	Placeholders := map[string]string{
		types.DBVendorClickHouse: "?",
		types.DBVendorMySQL:      "?",
		types.DBVendorMariaDB:    "?",
		types.DBVendorMSSQL:      "@p%d",
		types.DBVendorPostgres:   "$%d",
		types.DBVendorSQLite:     "?",
	}
	placeholder = Placeholders[dbVendor]
	if placeholder == "" {
		err = fmt.Errorf("unknown database type: %s", dbVendor)
	}
	return
}

// SetPlaceholders returns a slice of placeholders for the given dbVendor and number of columns.
// For example, for Postgres: $1, $2, $3; for MySQL: ?, ?, ?
func SetPlaceholders(dbVendor string, numCols int) ([]string, error) {
	placeholderStyle, err := PlaceholderStyle(dbVendor)
	if err != nil {
		return nil, err
	}
	placeholders := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		switch placeholderStyle {
		case "?":
			placeholders[i] = "?"
		case "$%d":
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		case "@p%d":
			placeholders[i] = fmt.Sprintf("@p%d", i+1)
		default:
			return nil, fmt.Errorf("unknown placeholder style: %s for dbVendor: %s", placeholderStyle, dbVendor)
		}
	}
	return placeholders, nil
}

func SetBatchPlaceholders(dbVendor string, numCols, numRows int) ([]string, error) {
	placeholderStyle, err := PlaceholderStyle(dbVendor)
	if err != nil {
		return nil, err
	}
	placeholders := make([]string, numCols*numRows)
	for row := range numRows {
		for col := range numCols {
			idx := row*numCols + col
			switch placeholderStyle {
			case "?":
				placeholders[idx] = "?"
			case "$%d":
				placeholders[idx] = fmt.Sprintf("$%d", idx+1)
			case "@p%d":
				placeholders[idx] = fmt.Sprintf("@p%d", idx+1)
			default:
				return nil, fmt.Errorf("unknown placeholder style: %s for dbVendor: %s", placeholderStyle, dbVendor)
			}
		}
	}
	return placeholders, nil
}

// QuoteIdentifier returns the identifier (table name, column name, etc.) quoted according to the dbVendor's syntax.
func QuoteIdentifier(dbVendor, identifier string) (quoted string) {
	switch dbVendor {
	case types.DBVendorClickHouse, types.DBVendorMySQL, types.DBVendorMariaDB, types.DBVendorSQLite:
		// Escape backticks by doubling them
		identifier = strings.ReplaceAll(identifier, "`", "``")
		quoted = "`" + identifier + "`"
	case types.DBVendorMSSQL:
		// Escape closing brackets by doubling them
		identifier = strings.ReplaceAll(identifier, "]", "]]")
		quoted = "[" + identifier + "]"
	case types.DBVendorPostgres:
		// Escape double quotes by doubling them
		identifier = strings.ReplaceAll(identifier, `"`, `""`)
		quoted = `"` + identifier + `"`
	default:
	}
	return
}
