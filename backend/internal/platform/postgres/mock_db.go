package postgres

import (
	"context"
	"database/sql"
)

// MockDBTX provides a controllable DBTX implementation for testing query strings and parameter binding.
type MockDBTX struct {
	ExecFunc     func(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryFunc    func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowFunc func(ctx context.Context, query string, args ...any) *sql.Row
}

func (m *MockDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return mockResult{}, nil
}

func (m *MockDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}

func (m *MockDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, query, args...)
	}
	// Return empty row that scans as ErrNoRows
	db, _ := sql.Open("pgx", "")
	return db.QueryRowContext(ctx, "SELECT 1 WHERE 1=0")
}

type mockResult struct {
	rowsAffected int64
}

func (r mockResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r mockResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
