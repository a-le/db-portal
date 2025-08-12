package config

import (
	"db-portal/internal/types"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Struct for commands.yaml
type CommandsConfig map[string]struct {
	Clickhouse string `yaml:"clickhouse"`
	Firebird   string `yaml:"firebird"`
	Mssql      string `yaml:"mssql"`
	Mysql      string `yaml:"mysql"`
	Postgresql string `yaml:"postgresql"`
	Sqlite3    string `yaml:"sqlite3"`
}

// Build SQL command string and args
func (c CommandsConfig) Command(name string, dbVendor string, identifiers []string) (qry string, args []any, err error) {
	cmd, ok := c[name]
	if !ok {
		err = fmt.Errorf("command %s could not be found", name)
		return
	}
	switch dbVendor {
	case types.DBVendorClickHouse:
		qry = cmd.Clickhouse
	case types.DBVendorMySQL, types.DBVendorMariaDB:
		qry = cmd.Mysql
	case types.DBVendorMSSQL:
		qry = cmd.Mssql
	case types.DBVendorPostgres:
		qry = cmd.Postgresql
	case types.DBVendorSQLite:
		qry = cmd.Sqlite3
	default:
		err = fmt.Errorf("unsupported db vendor: %s", dbVendor)
		return
	}

	// Replace %s with identifiers
	if strings.Contains(qry, "%s") {
		for _, identifier := range identifiers {
			if !isValidSQLIdentifier(identifier) {
				err = fmt.Errorf("invalid identifier %s", identifier)
				return
			}
		}
		args := make([]any, len(identifiers))
		for i := range identifiers {
			args[i] = identifiers[i]
		}
		qry = fmt.Sprintf(qry, args...)
	} else {
		for _, identifier := range identifiers {
			args = append(args, identifier)
		}
	}

	return
}

// Regular expression for a valid SQL identifier
// ^         : Start of the string
// [_\p{L}]  : First character must be an underscore (_) or any Unicode letter (\p{L})
// [_\p{L}\p{N}$]* : Subsequent characters can be underscores (_), Unicode letters (\p{L}), Unicode digits (\p{N}), or dollar signs ($)
// $         : End of the string
var validSQLIdentifier = regexp.MustCompile(`^[_\p{L}][_\p{L}\p{N}$]*$`)

// isValidSQLIdentifier checks if a given string is a valid SQL identifier
func isValidSQLIdentifier(identifier string) bool {
	// Check the length constraint (63 bytes maximum)
	if utf8.RuneCountInString(identifier) > 63 {
		return false
	}
	return validSQLIdentifier.MatchString(identifier)
}
