package copydata

import (
	"encoding/json"
	"strings"
)

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
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

func SQLToGenericType(sqlType string) string {
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
