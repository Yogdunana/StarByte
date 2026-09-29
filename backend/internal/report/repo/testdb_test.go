package repo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const testDriverName = "report_repo_test"

var (
	testBackends sync.Map
	testDSNSeq   uint64
)

func init() {
	sql.Register(testDriverName, scriptedDriver{})
}

type queryResult struct {
	columns []string
	rows    [][]driver.Value
	err     error
}

type capturedQuery struct {
	sql  string
	args []driver.NamedValue
}

type scriptedBackend struct {
	mu      sync.Mutex
	results []queryResult
	queries []capturedQuery
}

func (b *scriptedBackend) next(query string, args []driver.NamedValue) (driver.Rows, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queries = append(b.queries, capturedQuery{sql: query, args: append([]driver.NamedValue(nil), args...)})
	if len(b.results) == 0 {
		return nil, errors.New("unexpected query")
	}
	result := b.results[0]
	b.results = b.results[1:]
	if result.err != nil {
		return nil, result.err
	}
	return &scriptedRows{columns: result.columns, rows: result.rows}, nil
}

func (b *scriptedBackend) captured(t *testing.T) []capturedQuery {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]capturedQuery(nil), b.queries...)
}

type scriptedDriver struct{}

func (scriptedDriver) Open(name string) (driver.Conn, error) {
	value, ok := testBackends.Load(name)
	if !ok {
		return nil, fmt.Errorf("unknown test DSN %q", name)
	}
	return &scriptedConn{backend: value.(*scriptedBackend)}, nil
}

type scriptedConn struct{ backend *scriptedBackend }

func (c *scriptedConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}
func (c *scriptedConn) Close() error { return nil }
func (c *scriptedConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported")
}
func (c *scriptedConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.backend.next(query, args)
}

type scriptedRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (r *scriptedRows) Columns() []string { return r.columns }
func (r *scriptedRows) Close() error      { return nil }
func (r *scriptedRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func openTestDB(t *testing.T, results ...queryResult) (*gorm.DB, *scriptedBackend) {
	t.Helper()
	dsn := fmt.Sprintf("report-%d", atomic.AddUint64(&testDSNSeq, 1))
	backend := &scriptedBackend{results: append([]queryResult(nil), results...)}
	testBackends.Store(dsn, backend)
	sqlDB, err := sql.Open(testDriverName, dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
		testBackends.Delete(dsn)
	})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, WithoutReturning: true}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	return db, backend
}

var _ driver.QueryerContext = (*scriptedConn)(nil)
var _ driver.Rows = (*scriptedRows)(nil)
