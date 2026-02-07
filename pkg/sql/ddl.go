package sql

import (
	"bytes"
	"unicode"
)

func IsCreateDatabase(query []byte) bool {
	// Trim leading whitespace
	query = bytes.TrimSpace(query)

	// Skip comments
	for len(query) > 0 {
		// Skip line comments starting with --
		if bytes.HasPrefix(query, []byte("--")) {
			// Find the end of line
			idx := bytes.IndexByte(query, '\n')
			if idx == -1 {
				// Comment goes to end of query
				return false
			}
			query = bytes.TrimSpace(query[idx+1:])
			continue
		}

		// Skip block comments /* */
		if bytes.HasPrefix(query, []byte("/*")) {
			// Find the end of block comment
			idx := bytes.Index(query, []byte("*/"))
			if idx == -1 {
				// Unclosed block comment
				return false
			}
			query = bytes.TrimSpace(query[idx+2:])
			continue
		}

		// No more comments to skip
		break
	}

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
