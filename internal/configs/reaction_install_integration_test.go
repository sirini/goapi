package configs

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// 실제 MySQL/MariaDB에서 새 설치와 기존 설치 이행을 검증한다. 전용 테스트 DB만 허용한다.
// NUBO_REACTION_INSTALL_TEST_DSN='root@tcp(127.0.0.1:3306)/goapi_reaction_install_test' go test ./internal/configs -run TestReactionInstallMySQL -v
func TestReactionInstallMySQL(t *testing.T) {
	dsn := os.Getenv("NUBO_REACTION_INSTALL_TEST_DSN")
	if dsn == "" {
		t.Skip("set NUBO_REACTION_INSTALL_TEST_DSN for the disposable MySQL integration database")
	}
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil || parsed.DBName != "goapi_reaction_install_test" {
		t.Fatal("integration DSN must use the dedicated goapi_reaction_install_test database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	admin := AdminInfo{Id: "admin@example.com", Pw: "password"}

	t.Run("fresh install creates reaction tables and is repeatable", func(t *testing.T) {
		prefix := fmt.Sprintf("fresh_%d_", time.Now().UnixNano())
		t.Cleanup(func() { dropPrefixedTables(t, db, prefix) })
		for round := 1; round <= 2; round++ {
			if err := BootstrapDatabase(db, prefix, admin); err != nil {
				t.Fatalf("install round %d: %v", round, err)
			}
		}
		for _, spec := range []struct {
			table     string
			uniqueKey string
			column    string
			expected  string
		}{
			{prefix + "post_like", "uq_post_like_post_user", "post_uid", prefix + "post"},
			{prefix + "comment_like", "uq_comment_like_comment_user", "comment_uid", prefix + "comment"},
		} {
			assertColumn(t, db, spec.table, "reaction_type")
			assertColumn(t, db, spec.table, "uid")
			assertIndex(t, db, spec.table, "PRIMARY")
			assertIndex(t, db, spec.table, spec.uniqueKey)
			assertForeignKey(t, db, spec.table, spec.column, spec.expected)
		}
	})

	t.Run("legacy likes migrate once and installs are repeatable", func(t *testing.T) {
		prefix := fmt.Sprintf("legacy_%d_", time.Now().UnixNano())
		t.Cleanup(func() { dropPrefixedTables(t, db, prefix) })

		// 현대 스키마로 설치한 뒤 구 버전 형태로 후퇴시켜 기존 설치를 흉내낸다.
		if err := BootstrapDatabase(db, prefix, admin); err != nil {
			t.Fatalf("seeding modern install: %v", err)
		}
		boardUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %sboard ORDER BY uid LIMIT 1", prefix)).(int64)
		writerUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %suser ORDER BY uid LIMIT 1", prefix)).(int64)
		exec(t, db, fmt.Sprintf("INSERT INTO %suser (id, name, password) VALUES ('legacy-user-2', 'legacy-user-2', 'x')", prefix))
		secondUserUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %suser WHERE id = 'legacy-user-2'", prefix)).(int64)
		exec(t, db, fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, 'legacy category')", prefix), boardUid)
		categoryUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %sboard_category ORDER BY uid LIMIT 1", prefix)).(int64)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost (board_uid, user_uid, category_uid, title) VALUES (?, ?, ?, 'legacy post one')", prefix), boardUid, writerUid, categoryUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost (board_uid, user_uid, category_uid, title) VALUES (?, ?, ?, 'legacy post two')", prefix), boardUid, writerUid, categoryUid)
		postOne := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %spost ORDER BY uid LIMIT 1", prefix)).(int64)
		postTwo := mustScalar(t, db, fmt.Sprintf("SELECT MAX(uid) FROM %spost", prefix)).(int64)
		exec(t, db, fmt.Sprintf("INSERT INTO %scomment (board_uid, post_uid, user_uid, content) VALUES (?, ?, ?, 'legacy comment')", prefix), boardUid, postOne, writerUid)
		commentOne := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %scomment ORDER BY uid LIMIT 1", prefix)).(int64)
		for _, table := range []struct{ name, target string }{
			{"post_like", "post"}, {"comment_like", "comment"},
		} {
			exec(t, db, fmt.Sprintf("ALTER TABLE %s%s DROP INDEX uq_%s_like_%s_user, DROP INDEX idx_%s_like_reaction, DROP COLUMN uid, DROP COLUMN reaction_type", prefix, table.name, table.target, table.target, table.target))
		}

		// 중복·취소·동률 행을 심는다. (board, target, user, liked, timestamp)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost_like (board_uid, post_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 1, 100)", prefix), boardUid, postOne, writerUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost_like (board_uid, post_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 0, 200)", prefix), boardUid, postOne, writerUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost_like (board_uid, post_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 1, 100)", prefix), boardUid, postOne, secondUserUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost_like (board_uid, post_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 1, 100)", prefix), boardUid, postOne, secondUserUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost_like (board_uid, post_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 0, 50)", prefix), boardUid, postTwo, writerUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %scomment_like (board_uid, comment_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 1, 150)", prefix), boardUid, commentOne, writerUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %scomment_like (board_uid, comment_uid, user_uid, liked, timestamp) VALUES (?, ?, ?, 1, 300)", prefix), boardUid, commentOne, writerUid)

		for round := 1; round <= 2; round++ {
			if err := BootstrapDatabase(db, prefix, admin); err != nil {
				t.Fatalf("migration round %d: %v", round, err)
			}
		}

		// 최신 timestamp, 동률은 큰 uid 한 행만 남고 liked=(reaction_type=1) 불변식을 지킨다.
		assertQueryRows(t, db, fmt.Sprintf(
			`SELECT user_uid, liked, reaction_type FROM %spost_like WHERE post_uid = ?`, prefix), postOne,
			"post 1 rows", []map[string]any{{"user_uid": writerUid, "liked": 0, "reaction_type": 0}, {"user_uid": secondUserUid, "liked": 1, "reaction_type": 1}})
		assertQueryRows(t, db, fmt.Sprintf(
			`SELECT user_uid, liked, reaction_type FROM %spost_like WHERE post_uid = ?`, prefix), postTwo,
			"post 2 rows", []map[string]any{{"user_uid": writerUid, "liked": 0, "reaction_type": 0}})
		assertQueryRows(t, db, fmt.Sprintf(
			`SELECT liked, reaction_type FROM %scomment_like WHERE comment_uid = ?`, prefix), commentOne,
			"comment 1 rows", []map[string]any{{"liked": 1, "reaction_type": 1}})
		assertIndex(t, db, prefix+"post_like", "uq_post_like_post_user")
		assertIndex(t, db, prefix+"comment_like", "uq_comment_like_comment_user")
		assertForeignKey(t, db, prefix+"post_like", "post_uid", prefix+"post")
		assertForeignKey(t, db, prefix+"comment_like", "comment_uid", prefix+"comment")
		assertCount(t, db, fmt.Sprintf(
			`SELECT COUNT(*) FROM %spost_like WHERE liked <> IF(reaction_type = 1, 1, 0)`, prefix), 0,
			"liked=(reaction_type=1) invariant on post_like")
		assertCount(t, db, fmt.Sprintf(
			`SELECT COUNT(*) FROM %scomment_like WHERE liked <> IF(reaction_type = 1, 1, 0)`, prefix), 0,
			"liked=(reaction_type=1) invariant on comment_like")
	})

	t.Run("legacy two-level comments migrate to thread columns", func(t *testing.T) {
		prefix := fmt.Sprintf("thread_%d_", time.Now().UnixNano())
		t.Cleanup(func() { dropPrefixedTables(t, db, prefix) })
		if err := BootstrapDatabase(db, prefix, admin); err != nil {
			t.Fatalf("seeding modern install: %v", err)
		}
		boardUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %sboard ORDER BY uid LIMIT 1", prefix)).(int64)
		writerUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %suser ORDER BY uid LIMIT 1", prefix)).(int64)
		categoryUid := mustScalar(t, db, fmt.Sprintf("SELECT uid FROM %sboard_category ORDER BY uid LIMIT 1", prefix)).(int64)
		exec(t, db, fmt.Sprintf("INSERT INTO %spost (board_uid, user_uid, category_uid, title) VALUES (?, ?, ?, 'thread post')", prefix), boardUid, writerUid, categoryUid)
		postUid := mustScalar(t, db, fmt.Sprintf("SELECT MAX(uid) FROM %spost", prefix)).(int64)
		exec(t, db, fmt.Sprintf("INSERT INTO %scomment (board_uid, post_uid, user_uid, content) VALUES (?, ?, ?, 'thread root')", prefix), boardUid, postUid, writerUid)
		rootUid := mustScalar(t, db, fmt.Sprintf("SELECT MAX(uid) FROM %scomment", prefix)).(int64)
		exec(t, db, fmt.Sprintf("UPDATE %scomment SET reply_uid = ? WHERE uid = ?", prefix), rootUid, rootUid)
		exec(t, db, fmt.Sprintf("INSERT INTO %scomment (reply_uid, board_uid, post_uid, user_uid, content) VALUES (?, ?, ?, ?, 'legacy reply')", prefix), rootUid, boardUid, postUid, writerUid)
		// 구 스키마로 후퇴시킨다.
		exec(t, db, fmt.Sprintf("ALTER TABLE %scomment DROP INDEX idx_comment_parent, DROP COLUMN parent_uid, DROP COLUMN depth", prefix))

		for round := 1; round <= 2; round++ {
			if err := BootstrapDatabase(db, prefix, admin); err != nil {
				t.Fatalf("thread migration round %d: %v", round, err)
			}
		}
		assertColumn(t, db, prefix+"comment", "parent_uid")
		assertColumn(t, db, prefix+"comment", "depth")
		assertIndex(t, db, prefix+"comment", "idx_comment_parent")
		assertQueryRows(t, db, fmt.Sprintf(
			`SELECT reply_uid, parent_uid, depth FROM %scomment WHERE post_uid = ? ORDER BY uid`, prefix), postUid,
			"thread comment rows", []map[string]any{
				{"reply_uid": rootUid, "parent_uid": 0, "depth": 0},
				{"reply_uid": rootUid, "parent_uid": rootUid, "depth": 1},
			})
	})
}

func exec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %s: %v", query, err)
	}
}

func mustScalar(t *testing.T, db *sql.DB, query string) any {
	t.Helper()
	var value any
	if err := db.QueryRow(query).Scan(&value); err != nil {
		t.Fatalf("query %s: %v", query, err)
	}
	return value
}

func dropPrefixedTables(t *testing.T, db *sql.DB, prefix string) {
	t.Helper()
	rows, err := db.Query(`SELECT TABLE_NAME FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME LIKE ?`, prefix+"%")
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			t.Fatalf("scan table name: %v", err)
		}
		names = append(names, name)
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tables: %v", err)
	}
	if len(names) == 0 {
		return
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	for _, name := range names {
		if _, err := db.Exec("DROP TABLE " + name); err != nil {
			t.Fatalf("drop table %s: %v", name, err)
		}
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
}

func assertColumn(t *testing.T, db *sql.DB, table, column string) {
	t.Helper()
	var count uint
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, table, column).Scan(&count); err != nil {
		t.Fatalf("check column %s.%s: %v", table, column, err)
	}
	if count != 1 {
		t.Fatalf("column %s.%s is missing", table, column)
	}
}

func assertIndex(t *testing.T, db *sql.DB, table, index string) {
	t.Helper()
	var count uint
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?`, table, index).Scan(&count); err != nil {
		t.Fatalf("check index %s.%s: %v", table, index, err)
	}
	if count == 0 {
		t.Fatalf("index %s on %s is missing", index, table)
	}
}

func assertForeignKey(t *testing.T, db *sql.DB, table, column, expected string) {
	t.Helper()
	var referenced sql.NullString
	if err := db.QueryRow(`SELECT REFERENCED_TABLE_NAME FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ? AND REFERENCED_TABLE_NAME IS NOT NULL LIMIT 1`,
		table, column).Scan(&referenced); err != nil {
		t.Fatalf("check foreign key %s.%s: %v", table, column, err)
	}
	if !referenced.Valid || referenced.String != expected {
		t.Fatalf("foreign key %s.%s references %v, want %s", table, column, referenced.String, expected)
	}
}

func assertCount(t *testing.T, db *sql.DB, query string, want int, label string) {
	t.Helper()
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	if count != want {
		t.Fatalf("%s: got %d rows, want %d", label, count, want)
	}
}

func assertQueryRows(t *testing.T, db *sql.DB, query string, arg any, label string, want []map[string]any) {
	t.Helper()
	rows, err := db.Query(query, arg)
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	got := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			t.Fatalf("%s scan: %v", label, err)
		}
		row := make(map[string]any)
		for i, name := range columns {
			row[name] = values[i]
		}
		got = append(got, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("%s iterate: %v", label, err)
	}
	if len(got) != len(want) {
		t.Fatalf("%s: got %d rows (%v), want %d", label, len(got), got, len(want))
	}
	for i, expectedRow := range want {
		for name, expected := range expectedRow {
			actual := got[i][name]
			if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
				t.Fatalf("%s row %d: %s = %v (%T), want %v", label, i, name, actual, actual, expected)
			}
		}
	}
}
