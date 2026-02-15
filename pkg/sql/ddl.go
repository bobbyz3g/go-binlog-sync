package sql

import (
	"bytes"
	"strings"
	"unicode"
)

// TableName describes a parsed table identifier.
type TableName struct {
	Schema string
	Table  string
}

func IsCreateDatabase(query []byte) bool {
	// Trim leading whitespace
	query = bytes.TrimSpace(trimLeadingComments(query))

	// Extract first word
	i := 0
	for i < len(query) && !unicode.IsSpace(rune(query[i])) {
		i++
	}

	if i == 0 {
		return false
	}

	firstWord := bytes.ToUpper(query[:i])
	if !bytes.Equal(firstWord, []byte("CREATE")) {
		return false
	}

	// Skip whitespace after CREATE
	query = query[i:]
	query = bytes.TrimSpace(query)

	// Extract second word
	i = 0
	for i < len(query) && !unicode.IsSpace(rune(query[i])) {
		i++
	}

	if i == 0 {
		return false
	}

	secondWord := bytes.ToUpper(query[:i])
	return bytes.Equal(secondWord, []byte("DATABASE"))
}

// ExtractDDLDatabase returns the database name for CREATE/DROP DATABASE statements.
func ExtractDDLDatabase(query string) (string, bool) {
	tokens := ddlTokens(query)
	if len(tokens) < 3 {
		return "", false
	}
	first := strings.ToUpper(tokens[0])
	if first != "CREATE" && first != "DROP" {
		return "", false
	}
	if !isToken(tokens, 1, "DATABASE") && !isToken(tokens, 1, "SCHEMA") {
		return "", false
	}
	idx := 2
	idx = skipIfExists(tokens, idx)
	if idx >= len(tokens) {
		return "", false
	}
	name := cleanIdentifier(tokens[idx])
	if name == "" {
		return "", false
	}
	return name, true
}

// ExtractDDLTables returns table names for common DDL statements.
func ExtractDDLTables(query string) []TableName {
	tokens := ddlTokens(query)
	if len(tokens) < 2 {
		return nil
	}
	switch strings.ToUpper(tokens[0]) {
	case "CREATE":
		return extractCreateTables(tokens)
	case "ALTER":
		return extractAlterTables(tokens)
	case "DROP":
		return extractDropTables(tokens)
	case "TRUNCATE":
		return extractTruncateTables(tokens)
	case "RENAME":
		return extractRenameTables(tokens)
	default:
		return nil
	}
}

func trimLeadingComments(query []byte) []byte {
	query = bytes.TrimSpace(query)
	for len(query) > 0 {
		if bytes.HasPrefix(query, []byte("--")) {
			idx := bytes.IndexByte(query, '\n')
			if idx == -1 {
				return nil
			}
			query = bytes.TrimSpace(query[idx+1:])
			continue
		}
		if bytes.HasPrefix(query, []byte("/*")) {
			idx := bytes.Index(query, []byte("*/"))
			if idx == -1 {
				return nil
			}
			query = bytes.TrimSpace(query[idx+2:])
			continue
		}
		break
	}
	return query
}

func ddlTokens(query string) []string {
	raw := trimLeadingComments([]byte(query))
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	return strings.Fields(string(raw))
}

func isToken(tokens []string, idx int, want string) bool {
	if idx < 0 || idx >= len(tokens) {
		return false
	}
	return strings.EqualFold(tokens[idx], want)
}

func skipIfExists(tokens []string, idx int) int {
	if idx+2 < len(tokens) && isToken(tokens, idx, "IF") && isToken(tokens, idx+1, "NOT") && isToken(tokens, idx+2, "EXISTS") {
		return idx + 3
	}
	if idx+1 < len(tokens) && isToken(tokens, idx, "IF") && isToken(tokens, idx+1, "EXISTS") {
		return idx + 2
	}
	return idx
}

func extractCreateTables(tokens []string) []TableName {
	idx := 1
	if isToken(tokens, idx, "TEMPORARY") {
		idx++
	}
	if !isToken(tokens, idx, "TABLE") {
		return nil
	}
	idx++
	idx = skipIfExists(tokens, idx)
	if idx >= len(tokens) {
		return nil
	}
	name := parseTableName(tokens[idx])
	if name.Table == "" {
		return nil
	}
	return []TableName{name}
}

func extractAlterTables(tokens []string) []TableName {
	if !isToken(tokens, 1, "TABLE") {
		return nil
	}
	if len(tokens) < 3 {
		return nil
	}
	name := parseTableName(tokens[2])
	if name.Table == "" {
		return nil
	}
	return []TableName{name}
}

func extractDropTables(tokens []string) []TableName {
	idx := 1
	if isToken(tokens, idx, "TEMPORARY") {
		idx++
	}
	if !isToken(tokens, idx, "TABLE") {
		return nil
	}
	idx++
	idx = skipIfExists(tokens, idx)
	return parseTableList(tokens, idx)
}

func extractTruncateTables(tokens []string) []TableName {
	idx := 1
	if isToken(tokens, idx, "TABLE") {
		idx++
	}
	if idx >= len(tokens) {
		return nil
	}
	name := parseTableName(tokens[idx])
	if name.Table == "" {
		return nil
	}
	return []TableName{name}
}

func extractRenameTables(tokens []string) []TableName {
	if !isToken(tokens, 1, "TABLE") {
		return nil
	}
	idx := 2
	var names []TableName
	for idx < len(tokens) {
		name := parseTableName(tokens[idx])
		if name.Table == "" {
			idx++
			continue
		}
		idx++
		if idx >= len(tokens) || !isToken(tokens, idx, "TO") {
			continue
		}
		idx++
		if idx >= len(tokens) {
			break
		}
		newName := parseTableName(tokens[idx])
		if newName.Table != "" {
			names = append(names, name, newName)
		}
		idx++
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func parseTableList(tokens []string, idx int) []TableName {
	var names []TableName
	for i := idx; i < len(tokens); i++ {
		if isToken(tokens, i, "RESTRICT") || isToken(tokens, i, "CASCADE") {
			break
		}
		parts := strings.Split(tokens[i], ",")
		for _, part := range parts {
			name := parseTableName(part)
			if name.Table != "" {
				names = append(names, name)
			}
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func parseTableName(token string) TableName {
	name := cleanIdentifier(token)
	if name == "" {
		return TableName{}
	}
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		return TableName{Schema: parts[0], Table: parts[1]}
	}
	return TableName{Table: name}
}

func cleanIdentifier(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	token = strings.Trim(token, ",;")
	token = strings.TrimRight(token, "(")
	token = strings.TrimRight(token, ")")
	token = strings.Trim(token, ",;")
	token = strings.ReplaceAll(token, "`", "")
	return strings.TrimSpace(token)
}
