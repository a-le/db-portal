package config

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Struct for commands.yaml
type CommandsConfig map[string]struct {
	Firebird   string `yaml:"firebird"`
	Mssql      string `yaml:"mssql"`
	Mysql      string `yaml:"mysql"`
	Postgresql string `yaml:"postgresql"`
	Sqlite3    string `yaml:"sqlite3"`
}

// Build SQL command string and args
func (c CommandsConfig) Command(name string, dbType string, identifiers []string) (command string, args []any, err error) {
	if _, ok := c[name]; !ok {
		err = fmt.Errorf("command %s could not be found", name)
		return
	}
	switch dbType {
	case "firebird":
		command = c[name].Firebird
	case "mysql":
		command = c[name].Mysql
	case "mssql":
		command = c[name].Mssql
	case "postgresql":
		command = c[name].Postgresql
	case "sqlite3":
		command = c[name].Sqlite3

	}

	// Replace %s with identifiers
	if strings.Contains(command, "%s") {
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
		command = fmt.Sprintf(command, args...)
	} else {
		for _, identifier := range identifiers {
			args = append(args, identifier)
		}
	}

	return
}

// isValidSQLIdentifier checks if a given string is a valid SQL identifier
func isValidSQLIdentifier(identifier string) bool {
	// Check the length constraint (63 bytes maximum)
	if utf8.RuneCountInString(identifier) > 63 {
		return false
	}

	// Regular expression for a valid SQL identifier
	// ^         : Start of the string
	// [_\p{L}]  : First character must be an underscore (_) or any Unicode letter (\p{L})
	// [_\p{L}\p{N}$]* : Subsequent characters can be underscores (_), Unicode letters (\p{L}), Unicode digits (\p{N}), or dollar signs ($)
	// $         : End of the string
	var validSQLIdentifier = regexp.MustCompile(`^[_\p{L}][_\p{L}\p{N}$]*$`)

	return validSQLIdentifier.MatchString(identifier)
}
