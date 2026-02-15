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

func TestExtractDDLDatabase(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantName string
		wantOK   bool
	}{
		{
			name:     "create database",
			query:    "CREATE DATABASE test",
			wantName: "test",
			wantOK:   true,
		},
		{
			name:     "create schema",
			query:    "CREATE SCHEMA `foo`",
			wantName: "foo",
			wantOK:   true,
		},
		{
			name:     "drop database if exists",
			query:    "DROP DATABASE IF EXISTS db1",
			wantName: "db1",
			wantOK:   true,
		},
		{
			name:     "drop database if exists with backticks",
			query:    "DROP DATABASE IF EXISTS `db1`",
			wantName: "db1",
			wantOK:   true,
		},
		{
			name:   "create table",
			query:  "CREATE TABLE test",
			wantOK: false,
		},
		{
			name:   "alter database",
			query:  "ALTER DATABASE test",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := ExtractDDLDatabase(tt.query)
			if ok != tt.wantOK {
				t.Fatalf("ExtractDDLDatabase(%q) ok=%v, want %v", tt.query, ok, tt.wantOK)
			}
			if name != tt.wantName {
				t.Fatalf("ExtractDDLDatabase(%q) name=%q, want %q", tt.query, name, tt.wantName)
			}
		})
	}
}

func TestExtractDDLTables(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []TableName
	}{
		{
			name:  "create table",
			query: "CREATE TABLE test (id int)",
			expected: []TableName{
				{Table: "test"},
			},
		},
		{
			name:  "create table qualified",
			query: "CREATE TABLE `db1`.`t1` (id int)",
			expected: []TableName{
				{Schema: "db1", Table: "t1"},
			},
		},
		{
			name:  "create temporary table",
			query: "CREATE TEMPORARY TABLE t1 (id int)",
			expected: []TableName{
				{Table: "t1"},
			},
		},
		{
			name:  "alter table",
			query: "ALTER TABLE db1.t2 ADD COLUMN c int",
			expected: []TableName{
				{Schema: "db1", Table: "t2"},
			},
		},
		{
			name:  "drop table list",
			query: "DROP TABLE t1, `db1`.`t2`",
			expected: []TableName{
				{Table: "t1"},
				{Schema: "db1", Table: "t2"},
			},
		},
		{
			name:  "truncate table",
			query: "TRUNCATE TABLE db2.t3",
			expected: []TableName{
				{Schema: "db2", Table: "t3"},
			},
		},
		{
			name:  "rename table",
			query: "RENAME TABLE db1.t1 TO db1.t2, t3 TO t4",
			expected: []TableName{
				{Schema: "db1", Table: "t1"},
				{Schema: "db1", Table: "t2"},
				{Table: "t3"},
				{Table: "t4"},
			},
		},
		{
			name:     "create view",
			query:    "CREATE VIEW v AS SELECT 1",
			expected: nil,
		},
		{
			name:     "drop database",
			query:    "DROP DATABASE db1",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractDDLTables(tt.query)
			if len(got) != len(tt.expected) {
				t.Fatalf("ExtractDDLTables(%q) len=%d, want %d", tt.query, len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Fatalf("ExtractDDLTables(%q)[%d]=%+v, want %+v", tt.query, i, got[i], tt.expected[i])
				}
			}
		})
	}
}
