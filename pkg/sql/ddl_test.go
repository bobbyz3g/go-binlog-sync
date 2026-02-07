package sql

import "testing"

func TestIsCreateDatabase(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "simple create database",
			query:    "CREATE DATABASE test",
			expected: true,
		},
		{
			name:     "create database with case variation",
			query:    "create database test",
			expected: true,
		},
		{
			name:     "create database with mixed case",
			query:    "CrEaTe DaTaBaSe test",
			expected: true,
		},
		{
			name:     "create database with leading whitespace",
			query:    "  CREATE DATABASE test",
			expected: true,
		},
		{
			name:     "create database with trailing whitespace",
			query:    "CREATE DATABASE test  ",
			expected: true,
		},
		{
			name:     "create database with multiple spaces",
			query:    "CREATE    DATABASE    test",
			expected: true,
		},
		{
			name:     "create database with line comment",
			query:    "-- comment\nCREATE DATABASE test",
			expected: true,
		},
		{
			name:     "create database with block comment",
			query:    "/* comment */ CREATE DATABASE test",
			expected: true,
		},
		{
			name:     "create database with multiple comments",
			query:    "-- line comment\n/* block comment */CREATE DATABASE test",
			expected: true,
		},
		{
			name:     "create table",
			query:    "CREATE TABLE test",
			expected: false,
		},
		{
			name:     "create index",
			query:    "CREATE INDEX test",
			expected: false,
		},
		{
			name:     "drop database",
			query:    "DROP DATABASE test",
			expected: false,
		},
		{
			name:     "alter database",
			query:    "ALTER DATABASE test",
			expected: false,
		},
		{
			name:     "empty query",
			query:    "",
			expected: false,
		},
		{
			name:     "only whitespace",
			query:    "   ",
			expected: false,
		},
		{
			name:     "only comment",
			query:    "-- comment",
			expected: false,
		},
		{
			name:     "only block comment",
			query:    "/* comment */",
			expected: false,
		},
		{
			name:     "unclosed block comment",
			query:    "/* comment",
			expected: false,
		},
		{
			name:     "create only",
			query:    "CREATE",
			expected: false,
		},
		{
			name:     "create with whitespace only",
			query:    "CREATE   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCreateDatabase([]byte(tt.query))
			if result != tt.expected {
				t.Errorf("IsCreateDatabase(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}
