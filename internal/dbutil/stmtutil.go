package dbutil

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

func stmtClean(sql string) string {
	inBlockComment := false
	lines := strings.SplitSeq(sql, "\n")
	var cleaned []string

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove line comments
		if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
			continue
		}

		for {
			if inBlockComment {
				end := strings.Index(line, "*/")
				if end == -1 {
					line = "" // discard line inside block comment
					break
				}
				line = line[end+2:]
				inBlockComment = false
				line = strings.TrimSpace(line)
			}

			if strings.HasPrefix(line, "/*") {
				end := strings.Index(line, "*/")
				if end == -1 {
					inBlockComment = true
					line = "" // discard the rest of the line
					break
				}
				line = line[end+2:]
				line = strings.TrimSpace(line)
				continue
			}

			// Remove inline -- or // comments after code
			if idx := strings.Index(line, "--"); idx != -1 {
				line = line[:idx]
			}
			if idx := strings.Index(line, "//"); idx != -1 {
				line = line[:idx]
			}

			break
		}
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, " ")
}

// stmtCmd returns the first word (SQL command)
func stmtCmd(sql string) string {
	for i, r := range sql {
		if unicode.IsSpace(r) {
			return sql[:i]
		}
	}
	return sql
}

type stmtInfo struct {
	Cmd  string // Top level SQL keyword : update|insert|delete|select|create|drop|alter|show...
	Type string // Identified type of statement : "query"|"not-query"
}

func StmtInfo(sql string, dbVendor string) (infos stmtInfo) {
	sql = stmtClean(strings.ToLower(sql))
	infos.Cmd = stmtCmd(sql)

	if slices.Contains([]string{"insert", "update", "delete", "drop", "alter", "create"}, infos.Cmd) {
		var r *regexp.Regexp
		if dbVendor == "mssql" {
			r = regexp.MustCompile(`[\s](output)[\s]`)
		} else {
			r = regexp.MustCompile(`[\s](returning)[\s]`)
		}
		if r.MatchString(sql) {
			infos.Type = "query"
		} else {
			infos.Type = "non-query"
		}
	} else {
		infos.Type = "query"
	}

	return
}
