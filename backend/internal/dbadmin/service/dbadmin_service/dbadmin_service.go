package dbadmin_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// allowedTables mirrors the tables from backend/migrations, ordered so that
// parent tables come before their FK children (safe insert order).
var allowedTables = []string{
	"countries",
	"regions",
	"districts",
	"schools",
	"users",
	"user_sessions",
	"events",
	"participants",
	"teams",
	"team_participants",
	"games",
	"game_participants",
	"leaderboards",
	"submissions",
	"quiz_sessions",
	"quiz_questions",
	"quiz_teams",
	"quiz_buzzes",
	"quiz_answers",
	"quiz_buttons",
	"quiz_tiebreak_teams",
}

type DbAdminService struct {
	log *slog.Logger
	db  *gorm.DB
}

func New(log *slog.Logger, db *gorm.DB) *DbAdminService {
	return &DbAdminService{
		log: log,
		db:  db,
	}
}

func (s *DbAdminService) ListTables(ctx context.Context) []string {
	return allowedTables
}

func (s *DbAdminService) Export(ctx context.Context, table string) (map[string][]map[string]any, error) {
	const op = "dbadmin_service.Export"

	tables, err := resolveTables(table)
	if err != nil {
		return nil, err
	}

	data, err := s.loadTables(ctx, op, tables)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ExportSQL renders the same data as Export as an executable PostgreSQL
// script: a transaction that truncates the involved tables and re-inserts
// the rows (same replace semantics as Import).
func (s *DbAdminService) ExportSQL(ctx context.Context, table string) (string, error) {
	const op = "dbadmin_service.ExportSQL"

	tables, err := resolveTables(table)
	if err != nil {
		return "", err
	}

	data, err := s.loadTables(ctx, op, tables)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	fmt.Fprintf(&b, "-- Tournament database export\n")
	fmt.Fprintf(&b, "-- Exported at: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "-- Tables: %s\n\n", strings.Join(tables, ", "))
	b.WriteString("BEGIN;\n\n")
	// table names come only from the whitelist, so building the SQL is safe
	fmt.Fprintf(&b, "TRUNCATE TABLE %s;\n\n", strings.Join(tables, ", "))

	for _, name := range tables {
		for _, row := range data[name] {
			cols := make([]string, 0, len(row))
			for col := range row {
				cols = append(cols, col)
			}
			sort.Strings(cols)

			vals := make([]string, len(cols))
			for i, col := range cols {
				vals[i] = sqlLiteral(row[col])
			}

			fmt.Fprintf(&b, "INSERT INTO %s (%s) VALUES (%s);\n", name, strings.Join(cols, ", "), strings.Join(vals, ", "))
		}
	}

	b.WriteString("\nCOMMIT;\n")
	return b.String(), nil
}

func (s *DbAdminService) loadTables(ctx context.Context, op string, tables []string) (map[string][]map[string]any, error) {
	data := make(map[string][]map[string]any, len(tables))
	for _, name := range tables {
		rows := make([]map[string]any, 0)
		if err := s.db.WithContext(ctx).Table(name).Find(&rows).Error; err != nil {
			s.log.Error("failed to export table", logger.Err(err), slog.String("table", name))
			return nil, errs.ErrInternalError.Wrap(fmt.Sprintf("%s: failed to export table %q", op, name))
		}
		data[name] = rows
	}
	return data, nil
}

func resolveTables(table string) ([]string, error) {
	if table == "all" {
		return allowedTables, nil
	}
	if !isAllowedTable(table) {
		return nil, errs.ErrBadRequest.Wrap(fmt.Sprintf("unknown table %q", table))
	}
	return []string{table}, nil
}

func sqlLiteral(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case string:
		return quoteSQL(t)
	case bool:
		return strconv.FormatBool(t)
	case int:
		return strconv.Itoa(t)
	case int8:
		return strconv.FormatInt(int64(t), 10)
	case int16:
		return strconv.FormatInt(int64(t), 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case time.Time:
		return quoteSQL(t.Format("2006-01-02 15:04:05.999999999-07:00"))
	case []byte:
		return quoteSQL(string(t))
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return "NULL"
		}
		return quoteSQL(string(raw))
	}
}

func quoteSQL(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// Import replaces the data of the given tables with the provided rows.
// It runs in a single transaction: first the involved tables are truncated
// (in one TRUNCATE statement, so FK references between them are resolved),
// then rows are inserted in FK-safe order. Any error rolls everything back.
func (s *DbAdminService) Import(ctx context.Context, tables map[string][]map[string]any) error {
	const op = "dbadmin_service.Import"

	if len(tables) == 0 {
		return errs.ErrBadRequest.Wrap("no tables to import")
	}

	for name := range tables {
		if !isAllowedTable(name) {
			return errs.ErrBadRequest.Wrap(fmt.Sprintf("unknown table %q", name))
		}
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		toTruncate := make([]string, 0, len(tables))
		for _, name := range allowedTables {
			if _, ok := tables[name]; ok {
				toTruncate = append(toTruncate, name)
			}
		}

		// names come only from the whitelist, so building the SQL is safe
		if err := tx.Exec("TRUNCATE TABLE " + strings.Join(toTruncate, ", ")).Error; err != nil {
			return fmt.Errorf("%s: truncate: %w", op, err)
		}

		for _, name := range allowedTables {
			rows, ok := tables[name]
			if !ok || len(rows) == 0 {
				continue
			}
			if err := tx.Table(name).Create(&rows).Error; err != nil {
				return fmt.Errorf("%s: insert into %s: %w", op, name, err)
			}
		}

		return nil
	})
	if err != nil {
		s.log.Error("import failed", logger.Err(err))
		return errs.ErrInternalError.Wrap(err.Error())
	}

	return nil
}

func isAllowedTable(name string) bool {
	for _, t := range allowedTables {
		if t == name {
			return true
		}
	}
	return false
}
