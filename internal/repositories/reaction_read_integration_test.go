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

// 실제 DB에서 목록·공지·상세·댓글 SQL의 자리표시자와 결과 스캔을 검증한다.
// NUBO_REACTION_STATE_TEST_DSN='root@tcp(127.0.0.1:3306)/goapi_reaction_state_test' go test ./internal/repositories -run TestReactionReadsMySQL -v
func TestReactionReadsMySQL(t *testing.T) {
	dsn := os.Getenv("NUBO_REACTION_STATE_TEST_DSN")
	if dsn == "" {
		t.Skip("set NUBO_REACTION_STATE_TEST_DSN for the disposable MySQL integration database")
	}
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil || parsed.DBName != "goapi_reaction_state_test" {
		t.Fatal("integration DSN must use the dedicated goapi_reaction_state_test database")
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

	prefix := fmt.Sprintf("read_%d_", time.Now().UnixNano())
	previous := configs.Env.Prefix
	configs.Env.Prefix = prefix
	t.Cleanup(func() { configs.Env.Prefix = previous })
	tables := []struct{ name, columns string }{
		{"user", "uid INT UNSIGNED PRIMARY KEY, name VARCHAR(100) NOT NULL, profile VARCHAR(100) NOT NULL"},
		{"board_category", "uid INT UNSIGNED PRIMARY KEY, name VARCHAR(100) NOT NULL"},
		{"post", "uid INT UNSIGNED PRIMARY KEY, board_uid INT UNSIGNED NOT NULL, user_uid INT UNSIGNED NOT NULL, category_uid INT UNSIGNED NOT NULL, title TEXT NOT NULL, content TEXT NOT NULL, submitted BIGINT UNSIGNED NOT NULL, modified BIGINT UNSIGNED NOT NULL, hit INT UNSIGNED NOT NULL, status TINYINT NOT NULL"},
		{"comment", "uid INT UNSIGNED PRIMARY KEY, reply_uid INT UNSIGNED NOT NULL, parent_uid INT UNSIGNED NOT NULL DEFAULT 0, depth SMALLINT UNSIGNED NOT NULL DEFAULT 0, post_uid INT UNSIGNED NOT NULL, user_uid INT UNSIGNED NOT NULL, content TEXT NOT NULL, submitted BIGINT UNSIGNED NOT NULL, modified BIGINT UNSIGNED NOT NULL, status TINYINT NOT NULL"},
		{"file_thumbnail", "post_uid INT UNSIGNED NOT NULL, path VARCHAR(255) NOT NULL"},
		{"post_like", "post_uid INT UNSIGNED NOT NULL, user_uid INT UNSIGNED NOT NULL, liked TINYINT UNSIGNED NOT NULL, reaction_type TINYINT UNSIGNED NOT NULL"},
		{"comment_like", "comment_uid INT UNSIGNED NOT NULL, user_uid INT UNSIGNED NOT NULL, liked TINYINT UNSIGNED NOT NULL, reaction_type TINYINT UNSIGNED NOT NULL"},
	}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("CREATE TABLE %s%s (%s) ENGINE=InnoDB", prefix, table.name, table.columns)); err != nil {
			t.Fatalf("create %s: %v", table.name, err)
		}
		tableName := prefix + table.name
		t.Cleanup(func() {
			if _, err := db.Exec("DROP TABLE IF EXISTS " + tableName); err != nil {
				t.Errorf("drop %s: %v", tableName, err)
			}
		})
	}
	for _, statement := range []string{
		fmt.Sprintf("INSERT INTO %suser VALUES (3, '작성자', 'profile')", prefix),
		fmt.Sprintf("INSERT INTO %sboard_category VALUES (2, '분류')", prefix),
		fmt.Sprintf("INSERT INTO %spost VALUES (11, 7, 3, 2, '일반글', '내용', 100, 100, 5, 0), (12, 7, 3, 2, '공지글', '내용', 100, 100, 3, 1)", prefix),
		fmt.Sprintf("INSERT INTO %scomment VALUES (21, 21, 0, 0, 11, 3, '댓글', 100, 100, 0)", prefix),
		fmt.Sprintf("INSERT INTO %spost_like VALUES (11, 9, 1, 1)", prefix),
		fmt.Sprintf("INSERT INTO %scomment_like VALUES (21, 9, 0, 2)", prefix),
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("insert fixture: %v", err)
		}
	}

	board := NewNuboBoardRepository(db)
	view := NewNuboBoardViewRepository(db, board)
	comment := NewNuboCommentRepository(db, board)

	posts, err := board.FindPosts(models.BoardListParam{BoardUid: 7, UserUid: 9, Limit: 12, Page: 1})
	if err != nil || len(posts) != 1 || posts[0].Uid != 11 || posts[0].MyReaction == nil || *posts[0].MyReaction != models.REACTIONS.LIKE {
		t.Fatalf("normal posts: items=%+v err=%v", posts, err)
	}
	notices, err := board.GetNoticePosts(7, 9)
	if err != nil || len(notices) != 1 || notices[0].Uid != 12 {
		t.Fatalf("notice posts: items=%+v err=%v", notices, err)
	}
	post, err := view.GetPostItem(11, 9)
	if err != nil || post.Uid != 11 || post.MyReaction == nil || *post.MyReaction != models.REACTIONS.LIKE {
		t.Fatalf("post detail: item=%+v err=%v", post, err)
	}
	post, err = view.GetPostItem(11, 0)
	if err != nil || post.Uid != 11 || post.MyReaction != nil {
		t.Fatalf("anonymous post detail: item=%+v err=%v", post, err)
	}
	comments, err := comment.GetComments(models.CommentListParam{PostUid: 11, UserUid: 9, Page: 1, Limit: 20})
	if err != nil || len(comments) != 1 || comments[0].Uid != 21 || comments[0].MyReaction == nil || *comments[0].MyReaction != models.REACTIONS.BEST {
		t.Fatalf("comments: items=%+v err=%v", comments, err)
	}
}
