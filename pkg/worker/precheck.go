package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
)

func (w *Worker) precheckSource(ctx context.Context) error {
	w.lg.Info("source precheck started")

	if strings.TrimSpace(w.source.Host) == "" {
		return errors.New("source host is empty")
	}
	port := w.source.Port
	if port == 0 {
		port = 3306
	}
	addr := fmt.Sprintf("%s:%d", w.source.Host, port)
	conn, err := client.ConnectWithContext(ctx, addr, w.source.User, w.source.Password, "", 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect source: %w", err)
	}
	defer conn.Close()

	varNames := []string{"log_bin", "binlog_format", "binlog_row_image", "server_id"}
	switch strings.ToLower(strings.TrimSpace(w.source.Flavor)) {
	case "", "mysql":
		if w.source.GTIDEnabled {
			varNames = append(varNames, "gtid_mode", "enforce_gtid_consistency")
		}
	case "mariadb":
		if w.source.GTIDEnabled {
			varNames = append(varNames, "gtid_strict_mode")
		}
	default:
		return errors.New("invalid flavor")
	}

	query := buildShowVariablesQuery(varNames)
	result, err := conn.Execute(query)
	if err != nil {
		return fmt.Errorf("query source variables: %w", err)
	}
	if result != nil {
		defer result.Close()
	}
	vars, err := readVariableMap(result)
	if err != nil {
		return err
	}

	if err := requireTruthy(vars, "log_bin"); err != nil {
		return err
	}
	if err := requireValue(vars, "binlog_format", "ROW"); err != nil {
		return err
	}
	if rowImage, ok := vars["binlog_row_image"]; ok && strings.ToUpper(strings.TrimSpace(rowImage)) != "FULL" {
		w.lg.Warn("binlog_row_image is not FULL", slog.String("value", rowImage))
	}
	if serverID, ok := vars["server_id"]; ok && strings.TrimSpace(serverID) == "0" {
		w.lg.Warn("server_id is 0, binlog might not be enabled", slog.String("value", serverID))
	}

	if w.source.GTIDEnabled {
		switch strings.ToLower(strings.TrimSpace(w.source.Flavor)) {
		case "", "mysql":
			if err := requireGTIDMode(vars, "gtid_mode"); err != nil {
				return err
			}
			if err := requireTruthy(vars, "enforce_gtid_consistency"); err != nil {
				return err
			}
		case "mariadb":
			if err := requireTruthy(vars, "gtid_strict_mode"); err != nil {
				return err
			}
		}
	}

	w.lg.Info("source precheck passed")
	return nil
}

func buildShowVariablesQuery(names []string) string {
	parts := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("'%s'", trimmed))
	}
	return fmt.Sprintf("SHOW VARIABLES WHERE Variable_name IN (%s)", strings.Join(parts, ","))
}

func readVariableMap(result *mysql.Result) (map[string]string, error) {
	if result == nil || result.Resultset == nil {
		return nil, errors.New("empty resultset when reading variables")
	}
	rs := result.Resultset
	vars := make(map[string]string, rs.RowNumber())
	for i := 0; i < rs.RowNumber(); i++ {
		name, err := rs.GetValue(i, 0)
		if err != nil {
			return nil, fmt.Errorf("read variable name: %w", err)
		}
		value, err := rs.GetValue(i, 1)
		if err != nil {
			return nil, fmt.Errorf("read variable value: %w", err)
		}
		nameStr, err := valueToString(name)
		if err != nil {
			return nil, fmt.Errorf("read variable name: %w", err)
		}
		valStr, err := valueToString(value)
		if err != nil {
			return nil, fmt.Errorf("read variable value: %w", err)
		}
		vars[strings.ToLower(nameStr)] = valStr
	}
	return vars, nil
}

func valueToString(v interface{}) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case []byte:
		return string(t), nil
	default:
		return "", fmt.Errorf("unexpected value type %T", v)
	}
}

func requireTruthy(vars map[string]string, key string) error {
	value, ok := vars[key]
	if !ok {
		return fmt.Errorf("precheck failed: %s not found", key)
	}
	if !isTruthy(value) {
		return fmt.Errorf("precheck failed: %s=%s (expected ON)", key, value)
	}
	return nil
}

func requireValue(vars map[string]string, key string, expected string) error {
	value, ok := vars[key]
	if !ok {
		return fmt.Errorf("precheck failed: %s not found", key)
	}
	if strings.ToUpper(strings.TrimSpace(value)) != strings.ToUpper(strings.TrimSpace(expected)) {
		return fmt.Errorf("precheck failed: %s=%s (expected %s)", key, value, expected)
	}
	return nil
}

func requireGTIDMode(vars map[string]string, key string) error {
	value, ok := vars[key]
	if !ok {
		return fmt.Errorf("precheck failed: %s not found", key)
	}
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized != "ON" && normalized != "ON_PERMISSIVE" {
		return fmt.Errorf("precheck failed: %s=%s (expected ON)", key, value)
	}
	return nil
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "on", "yes", "true", "enabled":
		return true
	default:
		return false
	}
}
