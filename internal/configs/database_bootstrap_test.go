package configs

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
)

type bootstrapExecDriver struct {
	query string
}

type bootstrapExecConn struct {
	driver *bootstrapExecDriver
}

// 테스트용 연결을 열어 실행된 DDL을 기록한다.
func (driverState *bootstrapExecDriver) Open(string) (driver.Conn, error) {
	return &bootstrapExecConn{driver: driverState}, nil
}

// 이 테스트는 prepared statement를 사용하지 않는다.
func (connection *bootstrapExecConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not supported")
}

// 테스트 연결에는 해제할 자원이 없다.
func (connection *bootstrapExecConn) Close() error { return nil }

// 트랜잭션을 사용하지 않는 DDL 테스트임을 명시한다.
func (connection *bootstrapExecConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not supported")
}

// 전달된 DB 생성문을 검증할 수 있도록 저장한다.
func (connection *bootstrapExecConn) Exec(query string, _ []driver.Value) (driver.Result, error) {
	connection.driver.query = query
	return driver.RowsAffected(1), nil
}

// DB 이름을 식별자로 인용해 특수문자가 SQL로 실행되지 않게 한다.
func TestEnsureDatabaseQuotesIdentifier(t *testing.T) {
	driverState := &bootstrapExecDriver{}
	sql.Register("bootstrap-ensure-database", driverState)
	db, err := sql.Open("bootstrap-ensure-database", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := EnsureDatabase(db, "nubo`test"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(driverState.query, "`nubo``test`") {
		t.Fatalf("database identifier was not quoted safely: %s", driverState.query)
	}
}

// 테이블명에 삽입되는 prefix는 안전한 문자만 허용한다.
func TestBootstrapDatabaseRejectsUnsafePrefix(t *testing.T) {
	err := BootstrapDatabase(nil, "nubo_; DROP TABLE user", AdminInfo{})
	if err == nil || !strings.Contains(err.Error(), "DB_TABLE_PREFIX") {
		t.Fatalf("unsafe prefix error = %v", err)
	}
}

// bootstrap 검증 목록은 실제 생성·조회에 쓰는 user_black_list 이름을 따른다.
func TestBaseTableNamesUsesUserBlackListSchemaName(t *testing.T) {
	found := false
	for _, name := range baseTableNames {
		if name == "user_blacklist" {
			t.Fatal("실제 스키마에 없는 user_blacklist 이름을 검증하고 있습니다")
		}
		if name == "user_black_list" {
			found = true
		}
	}
	if !found {
		t.Fatal("user_black_list가 필수 테이블 검증 목록에 없습니다")
	}
}

func TestBaseTableNamesIncludesPushDevice(t *testing.T) {
	for _, name := range baseTableNames {
		if name == "push_device" {
			return
		}
	}
	t.Fatal("push_device가 필수 테이블 검증 목록에 없습니다")
}

var _ driver.Driver = (*bootstrapExecDriver)(nil)
var _ driver.Conn = (*bootstrapExecConn)(nil)
var _ driver.Execer = (*bootstrapExecConn)(nil)
