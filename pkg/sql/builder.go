package sql

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

// BuildInsertStatement constructs an INSERT statement for the given table and columns.
func BuildInsertStatement(table string, columns []string, values []any) (string, []any, error) {
	if len(columns) == 0 {
		return "", nil, errors.New("insert with zero columns")
	}
	normValues, err := NormalizeValues(values)
	if err != nil {
		return "", nil, err
	}
	colNames := make([]string, len(columns))
	placeholders := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = QuoteIdentifier(col)
		placeholders[i] = "?"
	}
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(colNames, ","), strings.Join(placeholders, ","))
	return stmt, normValues, nil
}

// BuildDeleteStatement constructs a DELETE statement with a WHERE clause.
func BuildDeleteStatement(table string, columns []string, values []any) (string, []any, error) {
	if len(columns) == 0 {
		return "", nil, errors.New("delete with zero columns")
	}
	normValues, err := NormalizeValues(values)
	if err != nil {
		return "", nil, err
	}
	whereClause, whereArgs, err := BuildWhereClause(columns, normValues)
	if err != nil {
		return "", nil, err
	}
	stmt := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereClause)
	return stmt, whereArgs, nil
}

// BuildUpdateStatement constructs an UPDATE statement with SET and WHERE clauses.
func BuildUpdateStatement(table string, setColumns []string, setValues []any, whereColumns []string, whereValues []any) (string, []any, error) {
	if len(setColumns) == 0 {
		return "", nil, errors.New("update with zero set columns")
	}
	if len(whereColumns) == 0 {
		return "", nil, errors.New("update with zero where columns")
	}
	normSet, err := NormalizeValues(setValues)
	if err != nil {
		return "", nil, err
	}
	normWhere, err := NormalizeValues(whereValues)
	if err != nil {
		return "", nil, err
	}
	setClause := BuildSetClause(setColumns)
	whereClause, whereArgs, err := BuildWhereClause(whereColumns, normWhere)
	if err != nil {
		return "", nil, err
	}
	args := append(normSet, whereArgs...)
	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setClause, whereClause)
	return stmt, args, nil
}

// BuildSetClause constructs a SET clause for UPDATE statements.
func BuildSetClause(columns []string) string {
	parts := make([]string, len(columns))
	for i, col := range columns {
		parts[i] = fmt.Sprintf("%s = ?", QuoteIdentifier(col))
	}
	return strings.Join(parts, ", ")
}

// BuildWhereClause constructs a WHERE clause handling NULL values correctly.
func BuildWhereClause(columns []string, values []any) (string, []any, error) {
	if len(columns) != len(values) {
		return "", nil, fmt.Errorf("where column/value mismatch: %d vs %d", len(columns), len(values))
	}
	parts := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	for i, col := range columns {
		if values[i] == nil {
			parts = append(parts, fmt.Sprintf("%s IS NULL", QuoteIdentifier(col)))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = ?", QuoteIdentifier(col)))
		args = append(args, values[i])
	}
	return strings.Join(parts, " AND "), args, nil
}

// NormalizeValues normalizes all values in a slice.
func NormalizeValues(values []any) ([]any, error) {
	norm := make([]any, len(values))
	for i, v := range values {
		if v == nil {
			norm[i] = nil
			continue
		}
		nv, err := NormalizeValue(v)
		if err != nil {
			return nil, err
		}
		norm[i] = nv
	}
	return norm, nil
}

// NormalizeValue converts special types to database-compatible representations.
func NormalizeValue(v any) (any, error) {
	switch val := v.(type) {
	case time.Time:
		return val.Format(mysql.TimeFormat), nil
	case *replication.JsonDiff:
		return nil, fmt.Errorf("json diff update unsupported: %s", val.String())
	default:
		return v, nil
	}
}

// QuoteTable returns a fully qualified table name with both schema and table quoted.
func QuoteTable(schema, table string) string {
	return QuoteIdentifier(schema) + "." + QuoteIdentifier(table)
}

// QuoteIdentifier quotes a SQL identifier with backticks, escaping existing backticks.
func QuoteIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}

// PickColumns filters columns and values based on skipped column indices.
func PickColumns(columns []string, row []any, skipped []int) ([]string, []any, error) {
	if len(columns) != len(row) {
		return nil, nil, fmt.Errorf("column/value mismatch: %d columns vs %d values", len(columns), len(row))
	}
	skipSet := make(map[int]struct{}, len(skipped))
	for _, idx := range skipped {
		skipSet[idx] = struct{}{}
	}
	selectedCols := make([]string, 0, len(columns)-len(skipSet))
	selectedVals := make([]any, 0, len(columns)-len(skipSet))
	for i, col := range columns {
		if _, ok := skipSet[i]; ok {
			continue
		}
		selectedCols = append(selectedCols, col)
		selectedVals = append(selectedVals, row[i])
	}
	return selectedCols, selectedVals, nil
}

// SkipColumns returns the list of skipped column indices for a given row.
func SkipColumns(ev *replication.RowsEvent, rowIndex int) []int {
	if rowIndex >= 0 && rowIndex < len(ev.SkippedColumns) {
		return ev.SkippedColumns[rowIndex]
	}
	return nil
}
