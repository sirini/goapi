package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

type pointDriver struct {
	mu           sync.Mutex
	rowsAffected int64
	execs        []pointExec
	committed    bool
	rolledBack   bool
}

type pointExec struct {
	query string
	args  []driver.NamedValue
}

type pointConn struct{ state *pointDriver }
type pointTx struct{ state *pointDriver }

func (d *pointDriver) Open(string) (driver.Conn, error)  { return &pointConn{state: d}, nil }
func (c *pointConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not supported") }
func (c *pointConn) Close() error                        { return nil }
func (c *pointConn) Begin() (driver.Tx, error)           { return &pointTx{state: c.state}, nil }
func (c *pointConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}
func (c *pointConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.execs = append(c.state.execs, pointExec{query: query, args: append([]driver.NamedValue(nil), args...)})
	if strings.Contains(query, "UPDATE") {
		return driver.RowsAffected(c.state.rowsAffected), nil
	}
	return driver.RowsAffected(1), nil
}
func (t *pointTx) Commit() error {
	t.state.committed = true
	return nil
}
func (t *pointTx) Rollback() error {
	t.state.rolledBack = true
	return nil
}

var pointDriverSequence atomic.Uint64

func openPointTestDB(t *testing.T, state *pointDriver) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("point-test-%d", pointDriverSequence.Add(1))
	sql.Register(name, state)
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestApplyPointChangeUsesConditionalUpdateAndNumericHistoryAction(t *testing.T) {
	state := &pointDriver{rowsAffected: 1}
	repo := NewNuboUserRepository(openPointTestDB(t, state))
	err := repo.ApplyPointChange(models.UpdatePointParam{
		UserUid: 7, BoardUid: 3, Action: models.POINT_ACTION_COMMENT, Point: -12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !state.committed || state.rolledBack {
		t.Fatalf("unexpected transaction state: committed=%v rolledBack=%v", state.committed, state.rolledBack)
	}
	if len(state.execs) != 2 {
		t.Fatalf("expected balance update and history insert, got %d statements", len(state.execs))
	}
	if !strings.Contains(state.execs[0].query, "point = point + ?") || !strings.Contains(state.execs[0].query, "point >= ?") {
		t.Fatalf("balance update is not atomic and conditional: %s", state.execs[0].query)
	}
	if got := state.execs[0].args[2].Value; got != int64(12) {
		t.Fatalf("required balance argument = %v, want 12", got)
	}
	if got := state.execs[1].args[2].Value; got != int64(models.POINT_ACTION_COMMENT) {
		t.Fatalf("history action argument = %v, want numeric comment action", got)
	}
}

func TestApplyPointChangeRollsBackWhenBalanceIsInsufficient(t *testing.T) {
	state := &pointDriver{rowsAffected: 0}
	repo := NewNuboUserRepository(openPointTestDB(t, state))
	err := repo.ApplyPointChange(models.UpdatePointParam{UserUid: 7, BoardUid: 3, Point: -12})
	if !errors.Is(err, ErrInsufficientPoint) {
		t.Fatalf("error = %v, want ErrInsufficientPoint", err)
	}
	if state.committed || !state.rolledBack {
		t.Fatalf("unexpected transaction state: committed=%v rolledBack=%v", state.committed, state.rolledBack)
	}
	if len(state.execs) != 1 {
		t.Fatalf("history must not be inserted after a rejected balance update; got %d statements", len(state.execs))
	}
}

var _ driver.Driver = (*pointDriver)(nil)
var _ driver.Conn = (*pointConn)(nil)
var _ driver.ConnBeginTx = (*pointConn)(nil)
var _ driver.ExecerContext = (*pointConn)(nil)
var _ driver.Tx = (*pointTx)(nil)
var _ io.Closer = (*pointConn)(nil)
