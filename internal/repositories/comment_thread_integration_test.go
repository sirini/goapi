package repositories

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

// 실제 DB에서 다층 댓글 스키마 이행(백필), 다층 쓰기·읽기, 직계 자식 기반 삭제 판정을 검증한다.
// NUBO_COMMENT_THREAD_TEST_DSN='root@tcp(127.0.0.1:3306)/goapi_comment_thread_test' go test ./internal/repositories -run TestCommentThreadMySQL -v
func TestCommentThreadMySQL(t *testing.T) {
	dsn := os.Getenv("NUBO_COMMENT_THREAD_TEST_DSN")
	if dsn == "" {
		t.Skip("set NUBO_COMMENT_THREAD_TEST_DSN for the disposable MySQL integration database")
	}
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil || parsed.DBName != "goapi_comment_thread_test" {
		t.Fatal("integration DSN must use the dedicated goapi_comment_thread_test database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	prefix := fmt.Sprintf("thread_%d_", time.Now().UnixNano())
	previous := configs.Env.Prefix
	configs.Env.Prefix = prefix
	t.Cleanup(func() { configs.Env.Prefix = previous })
	t.Cleanup(func() { dropCommentThreadTables(t, db, prefix) })

	// 2단계 시절 스키마(신규 컬럼 없음)로 시작해 이행이 컬럼 추가·백필을 수행하는지 확인한다.
	createLegacyCommentTables(t, db, prefix)
	writeParam := models.CommentWriteParam{BoardUid: 1, PostUid: 10, UserUid: 3, Content: "내용"}
	board := NewNuboBoardRepository(db)
	commentRepo := NewNuboCommentRepository(db, board)

	t.Run("migration backfills two-level rows and is repeatable", func(t *testing.T) {
		// 신 GOAPI는 install 이전 구 스키마에 쓰지 않는다. 이행 전 데이터는 구 형식으로 넣는다.
		result, err := db.Exec(fmt.Sprintf("INSERT INTO %scomment (reply_uid, board_uid, post_uid, user_uid, content, submitted, modified, status) VALUES (0, 1, 10, 3, '구 루트', 100, 0, 0)", prefix))
		if err != nil {
			t.Fatal(err)
		}
		rootId, err := result.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		root := uint(rootId)
		if _, err := db.Exec(fmt.Sprintf("INSERT INTO %scomment (reply_uid, board_uid, post_uid, user_uid, content, submitted, modified, status) VALUES (?, 1, 10, 3, '구 답글', 100, 0, 0)", prefix), root); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(fmt.Sprintf("UPDATE %scomment SET reply_uid = ? WHERE uid = ?", prefix), root, root); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if err := runCommentThreadMigration(t, db, prefix); err != nil {
				t.Fatalf("migration %d: %v", i, err)
			}
		}
		var parentUid, depth, replyUid uint
		if err := db.QueryRow(fmt.Sprintf("SELECT parent_uid, depth, reply_uid FROM %scomment WHERE user_uid = 3 AND content = '구 답글'", prefix)).Scan(&parentUid, &depth, &replyUid); err != nil {
			t.Fatal(err)
		}
		if parentUid != root || depth != 1 || replyUid != root {
			t.Fatalf("legacy reply backfill: parent=%d depth=%d reply=%d want parent=%d depth=1 reply=%d", parentUid, depth, replyUid, root, root)
		}
		if err := db.QueryRow(fmt.Sprintf("SELECT parent_uid, depth, reply_uid FROM %scomment WHERE uid = ?", prefix), root).Scan(&parentUid, &depth, &replyUid); err != nil {
			t.Fatal(err)
		}
		if parentUid != 0 || depth != 0 || replyUid != root {
			t.Fatalf("root backfill: parent=%d depth=%d reply=%d want 0,0,%d", parentUid, depth, replyUid, root)
		}
	})

	t.Run("nested writes keep thread root and depth chain", func(t *testing.T) {
		root, err := commentRepo.InsertComment(writeParam, 0, 0, 0, models.UpdatePointParam{})
		if err != nil {
			t.Fatal(err)
		}
		rootInfo := commentRepo.GetCommentThreadInfo(root)
		if rootInfo.ReplyUid != root || rootInfo.Depth != 0 {
			t.Fatalf("root info: %+v", rootInfo)
		}
		firstReply, err := commentRepo.InsertComment(writeParam, root, root, rootInfo.Depth+1, models.UpdatePointParam{})
		if err != nil {
			t.Fatal(err)
		}
		deepReply, err := commentRepo.InsertComment(writeParam, root, firstReply, 2, models.UpdatePointParam{})
		if err != nil {
			t.Fatal(err)
		}
		info := commentRepo.GetCommentThreadInfo(firstReply)
		if info.ReplyUid != root || info.Depth != 1 {
			t.Fatalf("first reply info: %+v", info)
		}
		if deep := commentRepo.GetCommentThreadInfo(deepReply); deep.ReplyUid != root || deep.Depth != 2 {
			t.Fatalf("deep reply info: %+v", deep)
		}

		items, err := commentRepo.GetComments(models.CommentListParam{PostUid: 10, UserUid: 0, Page: 1, Limit: 50})
		if err != nil {
			t.Fatal(err)
		}
		byUid := make(map[uint]models.CommentItem)
		for _, item := range items {
			byUid[item.Uid] = item
		}
		if len(items) < 3 {
			t.Fatalf("comments = %d, want >= 3", len(items))
		}
		if got := byUid[firstReply]; got.ParentUid != root || got.ReplyUid != root || got.Depth != 1 {
			t.Fatalf("first reply item: %+v", got)
		}
		if got := byUid[deepReply]; got.ParentUid != firstReply || got.ReplyUid != root || got.Depth != 2 {
			t.Fatalf("deep reply item: %+v", got)
		}

		if !commentRepo.HasReplyComment(firstReply) {
			t.Fatal("HasReplyComment(firstReply) = false, want true for direct child")
		}
		if commentRepo.HasReplyComment(deepReply) {
			t.Fatal("HasReplyComment(deepReply) = true, want false for leaf")
		}
	})
}

func runCommentThreadMigration(t *testing.T, db *sql.DB, prefix string) error {
	t.Helper()
	// configs가 노출하는 설치 경로 대신 동일 로직을 검증한다. env_setup 함수는 내부 패키지라
	// 여기서는 컬럼·인덱스·백필 SQL을 직접 재실행한다.
	table := prefix + "comment"
	addColumn := func(name string, ddl string) error {
		var count uint
		if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, table, name).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := db.Exec(fmt.Sprintf(ddl, table)); err != nil {
				return err
			}
		}
		return nil
	}
	if err := addColumn("parent_uid", "ALTER TABLE %s ADD COLUMN parent_uid INT UNSIGNED NOT NULL DEFAULT 0 AFTER reply_uid"); err != nil {
		return err
	}
	if err := addColumn("depth", "ALTER TABLE %s ADD COLUMN depth SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER parent_uid"); err != nil {
		return err
	}
	_, err := db.Exec(fmt.Sprintf(`UPDATE %s SET
		parent_uid = IF(reply_uid = uid OR reply_uid = 0, 0, reply_uid),
		depth = IF(reply_uid = uid OR reply_uid = 0, 0, 1)
		WHERE parent_uid = 0 AND depth = 0`, table))
	return err
}

func createLegacyCommentTables(t *testing.T, db *sql.DB, prefix string) {
	t.Helper()
	for _, table := range []string{
		fmt.Sprintf(`CREATE TABLE %suser (uid INT UNSIGNED PRIMARY KEY, name VARCHAR(100) NOT NULL, profile VARCHAR(100) NOT NULL) ENGINE=InnoDB`, prefix),
		fmt.Sprintf(`CREATE TABLE %scomment_like (comment_uid INT UNSIGNED NOT NULL, user_uid INT UNSIGNED NOT NULL, liked TINYINT UNSIGNED NOT NULL, reaction_type TINYINT UNSIGNED NOT NULL) ENGINE=InnoDB`, prefix),
	} {
		if _, err := db.Exec(table); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(fmt.Sprintf("INSERT INTO %suser VALUES (3, '작성자', 'profile')", prefix)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(fmt.Sprintf(`CREATE TABLE %scomment (
		uid INT UNSIGNED NOT NULL auto_increment,
		reply_uid INT UNSIGNED NOT NULL DEFAULT 0,
		board_uid INT UNSIGNED NOT NULL DEFAULT 0,
		post_uid INT UNSIGNED NOT NULL DEFAULT 0,
		user_uid INT UNSIGNED NOT NULL DEFAULT 0,
		content VARCHAR(10000) NOT NULL DEFAULT '',
		submitted BIGINT UNSIGNED NOT NULL DEFAULT 0,
		modified BIGINT UNSIGNED NOT NULL DEFAULT 0,
		status TINYINT NOT NULL DEFAULT 0,
		PRIMARY KEY (uid),
		KEY (reply_uid),
		KEY (post_uid)
	) ENGINE=InnoDB`, prefix)); err != nil {
		t.Fatal(err)
	}
}

func dropCommentThreadTables(t *testing.T, db *sql.DB, prefix string) {
	t.Helper()
	if _, err := db.Exec("DROP TABLE IF EXISTS " + prefix + "comment"); err != nil {
		t.Errorf("drop comment: %v", err)
	}
	for _, table := range []string{prefix + "user", prefix + "comment_like"} {
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Errorf("drop %s: %v", table, err)
		}
	}
}
