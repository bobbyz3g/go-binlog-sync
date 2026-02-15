package filter

import (
	"strings"

	pkgctx "github.com/bobbyz3g/go-binlog-sync/pkg/context"
)

// TableFilter decides whether a schema/table should be synced.
type TableFilter struct {
	whitelistDBs    map[string]struct{}
	whitelistTables map[string]struct{}
	blacklistDBs    map[string]struct{}
	blacklistTables map[string]struct{}
}

func NewTableFilter(cfg pkgctx.FilterConfig) *TableFilter {
	f := &TableFilter{
		whitelistDBs:    make(map[string]struct{}),
		whitelistTables: make(map[string]struct{}),
		blacklistDBs:    make(map[string]struct{}),
		blacklistTables: make(map[string]struct{}),
	}
	addList(cfg.Whitelist, f.whitelistDBs, f.whitelistTables)
	addList(cfg.Blacklist, f.blacklistDBs, f.blacklistTables)

	if len(f.whitelistDBs) == 0 && len(f.whitelistTables) == 0 &&
		len(f.blacklistDBs) == 0 && len(f.blacklistTables) == 0 {
		return nil
	}
	return f
}

func addList(list pkgctx.FilterList, dbs map[string]struct{}, tables map[string]struct{}) {
	for _, db := range list.Databases {
		name := strings.TrimSpace(db)
		if name == "" {
			continue
		}
		dbs[name] = struct{}{}
	}
	for _, table := range list.Tables {
		name := strings.TrimSpace(table)
		if name == "" {
			continue
		}
		tables[name] = struct{}{}
	}
}

// Allow reports whether the schema/table should be synced.
// Blacklist entries always win over whitelist entries.
func (f *TableFilter) Allow(schema, table string) bool {
	if f == nil {
		return true
	}
	schema = strings.TrimSpace(schema)
	table = strings.TrimSpace(table)

	if f.isBlacklisted(schema, table) {
		return false
	}
	if !f.hasWhitelist() {
		return true
	}
	return f.isWhitelisted(schema, table)
}

func (f *TableFilter) hasWhitelist() bool {
	return len(f.whitelistDBs) > 0 || len(f.whitelistTables) > 0
}

func (f *TableFilter) isBlacklisted(schema, table string) bool {
	if schema != "" {
		if _, ok := f.blacklistDBs[schema]; ok {
			return true
		}
	}
	if table != "" {
		if _, ok := f.blacklistTables[table]; ok {
			return true
		}
	}
	if schema != "" && table != "" {
		if _, ok := f.blacklistTables[schema+"."+table]; ok {
			return true
		}
	}
	return false
}

func (f *TableFilter) isWhitelisted(schema, table string) bool {
	if schema != "" {
		if _, ok := f.whitelistDBs[schema]; ok {
			return true
		}
	}
	if table != "" {
		if _, ok := f.whitelistTables[table]; ok {
			return true
		}
	}
	if schema != "" && table != "" {
		if _, ok := f.whitelistTables[schema+"."+table]; ok {
			return true
		}
	}
	return false
}
