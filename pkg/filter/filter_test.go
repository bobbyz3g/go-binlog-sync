package filter

import (
	"testing"

	pkgctx "github.com/bobbyz3g/go-binlog-sync/pkg/context"
)

func TestTableFilterAllow(t *testing.T) {
	tests := []struct {
		name   string
		cfg    pkgctx.FilterConfig
		schema string
		table  string
		want   bool
	}{
		{
			name: "empty filter allows all",
			cfg:  pkgctx.FilterConfig{},
			want: true,
		},
		{
			name: "whitelist database",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Databases: []string{"db1"},
				},
			},
			schema: "db1",
			table:  "t1",
			want:   true,
		},
		{
			name: "whitelist database blocks others",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Databases: []string{"db1"},
				},
			},
			schema: "db2",
			table:  "t1",
			want:   false,
		},
		{
			name: "whitelist table by name",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Tables: []string{"t1"},
				},
			},
			schema: "db2",
			table:  "t1",
			want:   true,
		},
		{
			name: "whitelist qualified table",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Tables: []string{"db1.t1"},
				},
			},
			schema: "db1",
			table:  "t1",
			want:   true,
		},
		{
			name: "whitelist qualified table blocks other db",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Tables: []string{"db1.t1"},
				},
			},
			schema: "db2",
			table:  "t1",
			want:   false,
		},
		{
			name: "blacklist overrides whitelist",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Databases: []string{"db1"},
				},
				Blacklist: pkgctx.FilterList{
					Tables: []string{"db1.t2"},
				},
			},
			schema: "db1",
			table:  "t2",
			want:   false,
		},
		{
			name: "blacklist database overrides whitelist table",
			cfg: pkgctx.FilterConfig{
				Whitelist: pkgctx.FilterList{
					Tables: []string{"db1.t1"},
				},
				Blacklist: pkgctx.FilterList{
					Databases: []string{"db1"},
				},
			},
			schema: "db1",
			table:  "t1",
			want:   false,
		},
		{
			name: "blacklist table by name",
			cfg: pkgctx.FilterConfig{
				Blacklist: pkgctx.FilterList{
					Tables: []string{"t1"},
				},
			},
			schema: "db2",
			table:  "t1",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewTableFilter(tt.cfg)
			if filter == nil {
				if tt.want != true {
					t.Fatalf("NewTableFilter returned nil, want filter to block")
				}
				return
			}
			got := filter.Allow(tt.schema, tt.table)
			if got != tt.want {
				t.Fatalf("Allow(%q,%q)=%v, want %v", tt.schema, tt.table, got, tt.want)
			}
		})
	}
}
