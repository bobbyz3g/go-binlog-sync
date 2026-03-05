package sql

import (
	"testing"
	"time"

	"github.com/go-mysql-org/go-mysql/replication"
)

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "column", "`column`"},
		{"with backtick", "col`umn", "`col``umn`"},
		{"multiple backticks", "col`um`n", "`col``um``n`"},
		{"empty", "", "``"},
		{"with space", "col umn", "`col umn`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestQuoteTable(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		table  string
		want   string
	}{
		{"simple", "db", "table", "`db`.`table`"},
		{"with backticks", "d`b", "tab`le", "`d``b`.`tab``le`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteTable(tt.schema, tt.table)
			if got != tt.want {
				t.Errorf("QuoteTable(%q, %q) = %q, want %q", tt.schema, tt.table, got, tt.want)
			}
		})
	}
}

func TestBuildInsertStatement(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		columns []string
		values  []any
		wantSQL string
		wantErr bool
	}{
		{
			name:    "simple insert",
			table:   "`db`.`table`",
			columns: []string{"id", "name"},
			values:  []any{1, "alice"},
			wantSQL: "INSERT INTO `db`.`table` (`id`,`name`) VALUES (?,?)",
			wantErr: false,
		},
		{
			name:    "zero columns",
			table:   "`db`.`table`",
			columns: []string{},
			values:  []any{},
			wantSQL: "",
			wantErr: true,
		},
		{
			name:    "with null",
			table:   "`db`.`table`",
			columns: []string{"id", "data"},
			values:  []any{1, nil},
			wantSQL: "INSERT INTO `db`.`table` (`id`,`data`) VALUES (?,?)",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := BuildInsertStatement(tt.table, tt.columns, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildInsertStatement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if gotSQL != tt.wantSQL {
					t.Errorf("BuildInsertStatement() SQL = %q, want %q", gotSQL, tt.wantSQL)
				}
				if len(gotArgs) != len(tt.values) {
					t.Errorf("BuildInsertStatement() args length = %d, want %d", len(gotArgs), len(tt.values))
				}
			}
		})
	}
}

func TestBuildDeleteStatement(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		columns []string
		values  []any
		wantSQL string
		wantErr bool
	}{
		{
			name:    "simple delete",
			table:   "`db`.`table`",
			columns: []string{"id"},
			values:  []any{1},
			wantSQL: "DELETE FROM `db`.`table` WHERE `id` = ?",
			wantErr: false,
		},
		{
			name:    "delete with null",
			table:   "`db`.`table`",
			columns: []string{"id", "data"},
			values:  []any{1, nil},
			wantSQL: "DELETE FROM `db`.`table` WHERE `id` = ? AND `data` IS NULL",
			wantErr: false,
		},
		{
			name:    "zero columns",
			table:   "`db`.`table`",
			columns: []string{},
			values:  []any{},
			wantSQL: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, _, err := BuildDeleteStatement(tt.table, tt.columns, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDeleteStatement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && gotSQL != tt.wantSQL {
				t.Errorf("BuildDeleteStatement() SQL = %q, want %q", gotSQL, tt.wantSQL)
			}
		})
	}
}

func TestBuildUpdateStatement(t *testing.T) {
	tests := []struct {
		name         string
		table        string
		setColumns   []string
		setValues    []any
		whereColumns []string
		whereValues  []any
		wantSQL      string
		wantErr      bool
	}{
		{
			name:         "simple update",
			table:        "`db`.`table`",
			setColumns:   []string{"name"},
			setValues:    []any{"bob"},
			whereColumns: []string{"id"},
			whereValues:  []any{1},
			wantSQL:      "UPDATE `db`.`table` SET `name` = ? WHERE `id` = ?",
			wantErr:      false,
		},
		{
			name:         "update multiple columns",
			table:        "`db`.`table`",
			setColumns:   []string{"name", "age"},
			setValues:    []any{"bob", 30},
			whereColumns: []string{"id"},
			whereValues:  []any{1},
			wantSQL:      "UPDATE `db`.`table` SET `name` = ?, `age` = ? WHERE `id` = ?",
			wantErr:      false,
		},
		{
			name:         "zero set columns",
			table:        "`db`.`table`",
			setColumns:   []string{},
			setValues:    []any{},
			whereColumns: []string{"id"},
			whereValues:  []any{1},
			wantSQL:      "",
			wantErr:      true,
		},
		{
			name:         "zero where columns",
			table:        "`db`.`table`",
			setColumns:   []string{"name"},
			setValues:    []any{"bob"},
			whereColumns: []string{},
			whereValues:  []any{},
			wantSQL:      "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, _, err := BuildUpdateStatement(tt.table, tt.setColumns, tt.setValues, tt.whereColumns, tt.whereValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildUpdateStatement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && gotSQL != tt.wantSQL {
				t.Errorf("BuildUpdateStatement() SQL = %q, want %q", gotSQL, tt.wantSQL)
			}
		})
	}
}

func TestBuildWhereClause(t *testing.T) {
	tests := []struct {
		name       string
		columns    []string
		values     []any
		wantClause string
		wantArgLen int
		wantErr    bool
	}{
		{
			name:       "simple where",
			columns:    []string{"id"},
			values:     []any{1},
			wantClause: "`id` = ?",
			wantArgLen: 1,
			wantErr:    false,
		},
		{
			name:       "multiple conditions",
			columns:    []string{"id", "name"},
			values:     []any{1, "alice"},
			wantClause: "`id` = ? AND `name` = ?",
			wantArgLen: 2,
			wantErr:    false,
		},
		{
			name:       "with null",
			columns:    []string{"id", "data"},
			values:     []any{1, nil},
			wantClause: "`id` = ? AND `data` IS NULL",
			wantArgLen: 1,
			wantErr:    false,
		},
		{
			name:       "column/value mismatch",
			columns:    []string{"id", "name"},
			values:     []any{1},
			wantClause: "",
			wantArgLen: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClause, gotArgs, err := BuildWhereClause(tt.columns, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWhereClause() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if gotClause != tt.wantClause {
					t.Errorf("BuildWhereClause() clause = %q, want %q", gotClause, tt.wantClause)
				}
				if len(gotArgs) != tt.wantArgLen {
					t.Errorf("BuildWhereClause() args length = %d, want %d", len(gotArgs), tt.wantArgLen)
				}
			}
		})
	}
}

func TestBuildSetClause(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		want    string
	}{
		{
			name:    "single column",
			columns: []string{"name"},
			want:    "`name` = ?",
		},
		{
			name:    "multiple columns",
			columns: []string{"name", "age", "email"},
			want:    "`name` = ?, `age` = ?, `email` = ?",
		},
		{
			name:    "empty",
			columns: []string{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSetClause(tt.columns)
			if got != tt.want {
				t.Errorf("BuildSetClause() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeValue(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name    string
		input   any
		want    any
		wantErr bool
	}{
		{
			name:    "string",
			input:   "test",
			want:    "test",
			wantErr: false,
		},
		{
			name:    "int",
			input:   42,
			want:    42,
			wantErr: false,
		},
		{
			name:    "time",
			input:   now,
			want:    "2024-01-15 10:30:00",
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("NormalizeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeValues(t *testing.T) {
	tests := []struct {
		name    string
		input   []any
		want    []any
		wantErr bool
	}{
		{
			name:    "mixed types",
			input:   []any{1, "test", nil},
			want:    []any{1, "test", nil},
			wantErr: false,
		},
		{
			name:    "empty",
			input:   []any{},
			want:    []any{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeValues(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(got) != len(tt.want) {
				t.Errorf("NormalizeValues() length = %d, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestPickColumns(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		row      []any
		skipped  []int
		wantCols []string
		wantVals []any
		wantErr  bool
	}{
		{
			name:     "no skipped",
			columns:  []string{"id", "name", "age"},
			row:      []any{1, "alice", 30},
			skipped:  []int{},
			wantCols: []string{"id", "name", "age"},
			wantVals: []any{1, "alice", 30},
			wantErr:  false,
		},
		{
			name:     "skip middle column",
			columns:  []string{"id", "name", "age"},
			row:      []any{1, "alice", 30},
			skipped:  []int{1},
			wantCols: []string{"id", "age"},
			wantVals: []any{1, 30},
			wantErr:  false,
		},
		{
			name:     "skip multiple",
			columns:  []string{"id", "name", "age", "email"},
			row:      []any{1, "alice", 30, "a@b.com"},
			skipped:  []int{1, 3},
			wantCols: []string{"id", "age"},
			wantVals: []any{1, 30},
			wantErr:  false,
		},
		{
			name:     "column/row mismatch",
			columns:  []string{"id", "name"},
			row:      []any{1},
			skipped:  []int{},
			wantCols: nil,
			wantVals: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCols, gotVals, err := PickColumns(tt.columns, tt.row, tt.skipped)
			if (err != nil) != tt.wantErr {
				t.Errorf("PickColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if len(gotCols) != len(tt.wantCols) {
					t.Errorf("PickColumns() cols length = %d, want %d", len(gotCols), len(tt.wantCols))
				}
				if len(gotVals) != len(tt.wantVals) {
					t.Errorf("PickColumns() vals length = %d, want %d", len(gotVals), len(tt.wantVals))
				}
				for i := range gotCols {
					if gotCols[i] != tt.wantCols[i] {
						t.Errorf("PickColumns() col[%d] = %q, want %q", i, gotCols[i], tt.wantCols[i])
					}
				}
			}
		})
	}
}

func TestSkipColumns(t *testing.T) {
	tests := []struct {
		name     string
		ev       *replication.RowsEvent
		rowIndex int
		want     []int
	}{
		{
			name: "valid index",
			ev: &replication.RowsEvent{
				SkippedColumns: [][]int{{1, 2}, {3}},
			},
			rowIndex: 0,
			want:     []int{1, 2},
		},
		{
			name: "second row",
			ev: &replication.RowsEvent{
				SkippedColumns: [][]int{{1, 2}, {3}},
			},
			rowIndex: 1,
			want:     []int{3},
		},
		{
			name: "out of bounds",
			ev: &replication.RowsEvent{
				SkippedColumns: [][]int{{1, 2}},
			},
			rowIndex: 5,
			want:     nil,
		},
		{
			name: "negative index",
			ev: &replication.RowsEvent{
				SkippedColumns: [][]int{{1, 2}},
			},
			rowIndex: -1,
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SkipColumns(tt.ev, tt.rowIndex)
			if len(got) != len(tt.want) {
				t.Errorf("SkipColumns() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SkipColumns()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}
