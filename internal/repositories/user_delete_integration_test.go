package repositories

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
)

// 실제 MySQL에서 삭제 SQL과 트랜잭션을 검증한다. 전용 테스트 DB만 허용한다.
// NUBO_ACCOUNT_DELETE_TEST_DSN='root@tcp(127.0.0.1:13367)/goapi_account_delete_test' go test ./internal/repositories -run TestDeleteAccountMySQL -v
func TestDeleteAccountMySQL(t *testing.T) {
	dsn := os.Getenv("NUBO_ACCOUNT_DELETE_TEST_DSN")
	if dsn == "" {
		t.Skip("set NUBO_ACCOUNT_DELETE_TEST_DSN for the disposable MySQL integration database")
	}
	config, err := mysql.ParseDSN(dsn)
	if err != nil || config.DBName != "goapi_account_delete_test" {
		t.Fatal("integration DSN must use the dedicated goapi_account_delete_test database")
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

	for _, scenario := range []string{"empty_account", "populated_account", "rollback_after_content_deletion"} {
		t.Run(scenario, func(t *testing.T) {
			prefix := fmt.Sprintf("delete_test_%d_", time.Now().UnixNano())
			old := configs.Env.Prefix
			configs.Env.Prefix = prefix
			t.Cleanup(func() { configs.Env.Prefix = old })
			exec := func(query string, args ...any) {
				t.Helper()
				if _, err := db.Exec(strings.ReplaceAll(query, "{p}", prefix), args...); err != nil {
					t.Fatalf("fixture SQL: %v", err)
				}
			}
			// 리포지토리가 사용하는 열과 삭제 순서를 검증할 외래 키를 구성한다.
			tables := []struct{ name, columns string }{
				{"user", "uid INT UNSIGNED PRIMARY KEY, id VARCHAR(100), name VARCHAR(100), password VARCHAR(100), profile VARCHAR(100), level INT, point INT, signature TEXT, signup BIGINT, signin BIGINT, blocked INT"},
				{"post", "uid INT UNSIGNED PRIMARY KEY, user_uid INT UNSIGNED, FOREIGN KEY (user_uid) REFERENCES {p}user(uid)"},
				{"comment", "uid INT UNSIGNED PRIMARY KEY, user_uid INT UNSIGNED, post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"comment_like", "user_uid INT UNSIGNED, comment_uid INT UNSIGNED, FOREIGN KEY (comment_uid) REFERENCES {p}comment(uid)"},
				{"notification", "to_uid INT UNSIGNED, from_uid INT UNSIGNED, post_uid INT UNSIGNED"},
				{"post_like", "user_uid INT UNSIGNED, post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"file", "uid INT UNSIGNED PRIMARY KEY, post_uid INT UNSIGNED, path VARCHAR(100), FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"file_thumbnail", "file_uid INT UNSIGNED, post_uid INT UNSIGNED, path VARCHAR(100), full_path VARCHAR(100), FOREIGN KEY (file_uid) REFERENCES {p}file(uid)"},
				{"exif", "post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"image_description", "post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"trade", "post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"post_hashtag", "post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid)"},
				{"post_origin", "post_uid INT UNSIGNED, FOREIGN KEY (post_uid) REFERENCES {p}post(uid) ON DELETE CASCADE"},
				{"image", "user_uid INT UNSIGNED, path VARCHAR(100)"},
				{"chat", "to_uid INT UNSIGNED, from_uid INT UNSIGNED"},
				{"report", "to_uid INT UNSIGNED, from_uid INT UNSIGNED"},
				{"user_black_list", "user_uid INT UNSIGNED, black_uid INT UNSIGNED"},
				{"mail_delivery", "recipient VARCHAR(100)"},
				{"user_verification", "email VARCHAR(100)"},
				{"signup_invite", "email VARCHAR(100)"},
			}
			for _, name := range []string{"point_history", "push_device", "user_token", "user_permission", "user_access_log", "user_oauth_identity", "oauth_nonce"} {
				tables = append(tables, struct{ name, columns string }{name, "user_uid INT UNSIGNED"})
			}
			for _, table := range tables {
				exec("CREATE TABLE {p}" + table.name + " (" + table.columns + ") ENGINE=InnoDB")
				name := prefix + table.name
				t.Cleanup(func() {
					if _, err := db.Exec("DROP TABLE " + name); err != nil {
						t.Errorf("cleanup %s: %v", name, err)
					}
				})
			}
			exec("INSERT INTO {p}user VALUES (7, 'delete@example.test', 'Delete', 'hash', 'profile', 1, 100, 'bio', 10, 20, 0), (9, 'keep@example.test', 'Keep', 'other', '', 1, 100, '', 10, 20, 0)")
			exec("INSERT INTO {p}post VALUES (91, 9)")
			exec("INSERT INTO {p}comment VALUES (912, 9, 91)")
			exec("INSERT INTO {p}comment_like VALUES (9, 912)")
			if scenario != "empty_account" {
				exec("INSERT INTO {p}post VALUES (71, 7)")
				// 본인의 다른 글에 쓴 댓글과 삭제할 글에 달린 다른 사람의 댓글을 포함한다.
				exec("INSERT INTO {p}comment VALUES (711, 9, 71), (911, 7, 91)")
				exec("INSERT INTO {p}comment_like VALUES (9, 711), (9, 911), (7, 912)")
				exec("INSERT INTO {p}notification VALUES (9, 9, 71), (7, 9, 91), (9, 7, 91), (9, 9, 91)")
				exec("INSERT INTO {p}post_like VALUES (9, 71), (7, 91), (9, 91)")
				exec("INSERT INTO {p}file VALUES (710, 71, 'deleted.jpg'), (910, 91, 'kept.jpg')")
				exec("INSERT INTO {p}file_thumbnail VALUES (710, 71, 'thumb.webp', 'full.webp'), (910, 91, 'keep-thumb.webp', 'keep-full.webp')")
				for _, name := range []string{"exif", "image_description", "trade", "post_hashtag", "post_origin"} {
					exec("INSERT INTO {p}" + name + " VALUES (71), (91)")
				}
				exec("INSERT INTO {p}image VALUES (7, 'editor.jpg'), (9, 'keep-editor.jpg')")
				for _, name := range []string{"chat", "report", "user_black_list"} {
					exec("INSERT INTO {p}" + name + " VALUES (7, 9), (9, 7), (9, 9)")
				}
				for _, name := range []string{"mail_delivery", "user_verification", "signup_invite"} {
					exec("INSERT INTO {p}" + name + " VALUES ('delete@example.test'), ('keep@example.test')")
				}
				for _, name := range []string{"point_history", "push_device", "user_token", "user_permission", "user_access_log", "user_oauth_identity", "oauth_nonce"} {
					exec("INSERT INTO {p}" + name + " VALUES (7), (9)")
				}
			}
			counts := func() map[string]int {
				t.Helper()
				result := make(map[string]int)
				for _, table := range tables {
					var n int
					if err := db.QueryRow("SELECT COUNT(*) FROM " + prefix + table.name).Scan(&n); err != nil {
						t.Fatal(err)
					}
					result[table.name] = n
				}
				return result
			}
			before := counts()
			if scenario == "rollback_after_content_deletion" {
				exec("CREATE TRIGGER {p}reject_update BEFORE UPDATE ON {p}user FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'test rollback'")
			}
			paths, err := NewNuboUserRepository(db).DeleteAccount(7)
			if scenario == "rollback_after_content_deletion" {
				if err == nil || !strings.Contains(err.Error(), "test rollback") || len(paths) != 0 {
					t.Fatalf("failed transaction returned paths=%v, err=%v", paths, err)
				}
				if after := counts(); !reflect.DeepEqual(before, after) {
					t.Fatalf("rollback changed rows: before=%v after=%v", before, after)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "populated_account" {
				sort.Strings(paths)
				if !reflect.DeepEqual(paths, []string{"deleted.jpg", "editor.jpg", "full.webp", "thumb.webp"}) {
					t.Fatalf("file cleanup scope = %v", paths)
				}
				for table, n := range counts() {
					want := 1
					if table == "user" {
						want = 2
					}
					if n != want {
						t.Errorf("%s retained %d rows, want %d", table, n, want)
					}
				}
			} else if len(paths) != 0 || !reflect.DeepEqual(before, counts()) {
				t.Fatal("empty account deletion affected unrelated content")
			}
			var id, name, password, profile, signature string
			var level, point, signup, signin, blocked int
			if err := db.QueryRow("SELECT id, name, password, profile, signature, level, point, signup, signin, blocked FROM "+prefix+"user WHERE uid = 7").Scan(&id, &name, &password, &profile, &signature, &level, &point, &signup, &signin, &blocked); err != nil {
				t.Fatal(err)
			}
			if id != "" || name != "탈퇴한 사용자" || password != "" || profile != "" || signature != "" || level != 0 || point != 0 || signup != 0 || signin != 0 || blocked != 1 {
				t.Fatal("deleted account retained credentials/profile or remained active")
			}
			var survivor int
			if err := db.QueryRow("SELECT COUNT(*) FROM " + prefix + "comment WHERE uid = 912 AND user_uid = 9 AND post_uid = 91").Scan(&survivor); err != nil || survivor != 1 {
				t.Fatalf("unrelated comment was not preserved: %d, %v", survivor, err)
			}
		})
	}
}
