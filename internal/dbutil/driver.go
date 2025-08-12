package dbutil

import (
	"db-portal/internal/types"
	"fmt"
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
