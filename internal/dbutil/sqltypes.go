// Cross-database type mappings and normalization
// for schema generation across vendors.
package dbutil

import (
	"db-portal/internal/types"
	"strings"
)

// Cross-vendor usage example, ClickHouse UInt8 -> Postgres:
// canonical := CanonicalType(types.DBVendorClickHouse, "UInt8")
// postgresType := VendorType(types.DBVendorPostgres, canonical)
// fmt.Println(postgresType) // BOOLEAN

// Forward mapping: canonical -> vendor -> type
// canonical type list:
// - int, bigint, smallint
// - boolean
// - decimal, float
// - char, varchar, text
// - binary
// - date, time, datetime
// - uuid
var CanonicalTypeToVendorType = map[string]map[string]string{
	"int": {
		types.DBVendorMSSQL:      "INT",
		types.DBVendorMySQL:      "INT",
		types.DBVendorMariaDB:    "INT",
		types.DBVendorSQLite:     "INTEGER",
		types.DBVendorPostgres:   "INTEGER",
		types.DBVendorClickHouse: "Int32",
	},
	"bigint": {
		types.DBVendorMSSQL:      "BIGINT",
		types.DBVendorMySQL:      "BIGINT",
		types.DBVendorMariaDB:    "BIGINT",
		types.DBVendorSQLite:     "INTEGER",
		types.DBVendorPostgres:   "BIGINT",
		types.DBVendorClickHouse: "Int64",
	},
	"smallint": {
		types.DBVendorMSSQL:      "SMALLINT",
		types.DBVendorMySQL:      "SMALLINT",
		types.DBVendorMariaDB:    "SMALLINT",
		types.DBVendorSQLite:     "INTEGER",
		types.DBVendorPostgres:   "SMALLINT",
		types.DBVendorClickHouse: "Int16",
	},
	"boolean": {
		types.DBVendorMSSQL:      "BIT",
		types.DBVendorMySQL:      "TINYINT(1)",
		types.DBVendorMariaDB:    "TINYINT(1)",
		types.DBVendorSQLite:     "INTEGER",
		types.DBVendorPostgres:   "BOOLEAN",
		types.DBVendorClickHouse: "UInt8",
	},
	"decimal": {
		types.DBVendorMSSQL:      "DECIMAL(18,2)",
		types.DBVendorMySQL:      "DECIMAL(18,2)",
		types.DBVendorMariaDB:    "DECIMAL(18,2)",
		types.DBVendorSQLite:     "NUMERIC",
		types.DBVendorPostgres:   "NUMERIC(18,2)",
		types.DBVendorClickHouse: "Decimal(18,2)",
	},
	"float": {
		types.DBVendorMSSQL:      "FLOAT",
		types.DBVendorMySQL:      "FLOAT",
		types.DBVendorMariaDB:    "FLOAT",
		types.DBVendorSQLite:     "REAL",
		types.DBVendorPostgres:   "DOUBLE PRECISION",
		types.DBVendorClickHouse: "Float64",
	},
	"char": {
		types.DBVendorMSSQL:      "CHAR(1)",
		types.DBVendorMySQL:      "CHAR(1)",
		types.DBVendorMariaDB:    "CHAR(1)",
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "CHAR(1)",
		types.DBVendorClickHouse: "FixedString(1)",
	},
	"varchar": {
		types.DBVendorMSSQL:      "VARCHAR(255)",
		types.DBVendorMySQL:      "VARCHAR(255)",
		types.DBVendorMariaDB:    "VARCHAR(255)",
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "VARCHAR(255)",
		types.DBVendorClickHouse: "String",
	},
	"text": {
		types.DBVendorMSSQL:      "TEXT",
		types.DBVendorMySQL:      "TEXT",
		types.DBVendorMariaDB:    "TEXT",
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "TEXT",
		types.DBVendorClickHouse: "String",
	},
	"binary": {
		types.DBVendorMSSQL:      "VARBINARY(MAX)",
		types.DBVendorMySQL:      "BLOB",
		types.DBVendorMariaDB:    "BLOB",
		types.DBVendorSQLite:     "BLOB",
		types.DBVendorPostgres:   "BYTEA",
		types.DBVendorClickHouse: "String",
	},
	"date": {
		types.DBVendorMSSQL:      "DATE",
		types.DBVendorMySQL:      "DATE",
		types.DBVendorMariaDB:    "DATE",
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "DATE",
		types.DBVendorClickHouse: "Date",
	},
	"time": {
		types.DBVendorMSSQL:      "TIME",
		types.DBVendorMySQL:      "TIME",
		types.DBVendorMariaDB:    "TIME",
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "TIME",
		types.DBVendorClickHouse: "String", // ClickHouse doesn't have pure time type
	},
	"datetime": {
		types.DBVendorMSSQL:      "DATETIME2",
		types.DBVendorMySQL:      "DATETIME",
		types.DBVendorMariaDB:    "DATETIME",
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "TIMESTAMP",
		types.DBVendorClickHouse: "DateTime",
	},
	"uuid": {
		types.DBVendorMSSQL:      "UNIQUEIDENTIFIER",
		types.DBVendorMySQL:      "CHAR(36)", // no native UUID
		types.DBVendorMariaDB:    "CHAR(36)", // no native UUID
		types.DBVendorSQLite:     "TEXT",
		types.DBVendorPostgres:   "UUID",
		types.DBVendorClickHouse: "UUID",
	},
}

func VendorType(DBVendor string, canonical string) string {
	return CanonicalTypeToVendorType[canonical][DBVendor]
}

// Inverse mapping: vendor -> normalized vendor-type -> canonical type
//var InverseSQLTypeMatrix = buildInverse(CanonicalTypeToVendorType)

// Aliases for normalization: vendor-type → canonical type
var typeAliases = map[string]string{
	// Integer family
	"INT":     "int",
	"INTEGER": "int",
	"INT4":    "int",
	"INT32":   "int",
	"INT8":    "bigint",
	"TINYINT": "smallint", // except MySQL/MariaDB tinyint(1)

	// Bigint
	"BIGINT": "bigint",
	"INT64":  "bigint",

	// Smallint
	"SMALLINT": "smallint",
	"INT2":     "smallint",

	// Boolean
	"BOOL":    "boolean",
	"BOOLEAN": "boolean",

	// Decimal
	"NUMERIC": "decimal",
	"DEC":     "decimal",
	"DECIMAL": "decimal",
	"MONEY":   "decimal",

	// Float
	"REAL":             "float",
	"FLOAT":            "float",
	"FLOAT4":           "float",
	"FLOAT8":           "float",
	"DOUBLE":           "float",
	"DOUBLE PRECISION": "float",

	// Strings
	"CHAR":      "char",
	"NCHAR":     "char",
	"CHARACTER": "char",
	"VARCHAR":   "varchar",
	"NVARCHAR":  "varchar",
	"TEXT":      "text",
	"CLOB":      "text",

	// Binary
	"BLOB":      "binary",
	"BYTEA":     "binary",
	"VARBINARY": "binary",

	// Dates
	"DATE":          "date",
	"TIME":          "time",
	"DATETIME":      "datetime",
	"DATETIME2":     "datetime",
	"SMALLDATETIME": "datetime",
	"TIMESTAMP":     "datetime",

	// UUID
	"UUID":             "uuid",
	"UNIQUEIDENTIFIER": "uuid",

	// --- ClickHouse Aliases ---
	"INT16": "smallint",
	//"INT32":                       "int",
	//"INT64":                       "bigint",
	"UINT8":       "boolean", // treat as bool
	"UINT16":      "int",
	"UINT32":      "bigint",
	"UINT64":      "bigint",
	"FLOAT32":     "float",
	"FLOAT64":     "float",
	"STRING":      "varchar",
	"FIXEDSTRING": "char",
	"DATE32":      "date",
	"DATETIME64":  "datetime",
	//"DATETIME":                    "datetime",
	"LOWCARDINALITY(STRING)":      "varchar",
	"LOWCARDINALITY(FIXEDSTRING)": "char",
	"NULLABLE":                    "", // unwrap
}

// Build inverse mapping from forward map
// func buildInverse(matrix map[string]map[string]string) map[string]map[string]string {
// 	inverse := make(map[string]map[string]string)
// 	for canonical, vendors := range matrix {
// 		for vendor, vtype := range vendors {
// 			if inverse[vendor] == nil {
// 				inverse[vendor] = make(map[string]string)
// 			}
// 			key := CanonicalType(vendor, vtype)
// 			inverse[vendor][key] = canonical
// 		}
// 	}
// 	return inverse
// }

// CanonicalType get a canonical type from a vendor type (=SQL type/domain).
// Example: INT32 -> int
// Returns "text" as fallback for unknown types.
func CanonicalType(vendor, s string) string {
	raw := strings.ToUpper(strings.TrimSpace(s))

	// Special case: MySQL/MariaDB TINYINT(1) -> boolean
	if (vendor == types.DBVendorMySQL || vendor == types.DBVendorMariaDB) && strings.HasPrefix(raw, "TINYINT(1") {
		return "boolean"
	}

	// Handle UNSIGNED types (fold back to signed equivalents)
	if strings.Contains(raw, "UNSIGNED") {
		raw = strings.ReplaceAll(raw, "UNSIGNED", "")
		raw = strings.TrimSpace(raw)
	}

	// Strip wrappers: Nullable(...), LowCardinality(...)
	if strings.HasPrefix(raw, "NULLABLE(") {
		raw = strings.TrimSuffix(strings.TrimPrefix(raw, "NULLABLE("), ")")
	}
	if strings.HasPrefix(raw, "LOWCARDINALITY(") {
		raw = strings.TrimSuffix(strings.TrimPrefix(raw, "LOWCARDINALITY("), ")")
	}

	// Strip parameters
	if idx := strings.Index(raw, "("); idx != -1 {
		raw = raw[:idx]
	}

	if alias, ok := typeAliases[raw]; ok && alias != "" {
		return alias
	}

	// Fallback to "text" for unknown types
	return "text"
}
