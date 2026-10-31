package repositories

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

// 실제 MySQL에서 리액션 상태 전이(0→종류→다른 종류→0), 반복 요청, 동시 요청을 검증한다.
// NUBO_REACTION_STATE_TEST_DSN='root@tcp(127.0.0.1:3306)/goapi_reaction_state_test' go test ./internal/repositories -run TestReactionStateMySQL -v
func TestReactionStateMySQL(t *testing.T) {
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

	prefix := fmt.Sprintf("state_%d_", time.Now().UnixNano())
	old := configs.Env.Prefix
	configs.Env.Prefix = prefix
	t.Cleanup(func() { configs.Env.Prefix = old })
	createReactionStateTables(t, db, prefix)
	t.Cleanup(func() { dropReactionStateTables(t, db, prefix) })

	boardRepo := NewNuboBoardRepository(db)
	viewRepo := NewNuboBoardViewRepository(db, boardRepo)
	commentRepo := NewNuboCommentRepository(db, boardRepo)

	const boardUid, postUid, commentUid, userUid = uint(1), uint(10), uint(20), uint(30)

	t.Run("post transitions keep one row and liked projection", func(t *testing.T) {
		changed, err := viewRepo.SetPostReaction(models.BoardReactionParam{
			BoardUid: boardUid, PostUid: postUid, UserUid: userUid,
			Reaction: models.REACTIONS.LIKE, ReactionCode: models.REACTION_LIKE,
		})
		if err != nil || !changed {
			t.Fatalf("first like: changed=%v err=%v", changed, err)
		}
		assertPostState(t, viewRepo, postUid, userUid, [10]uint{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, models.REACTIONS.LIKE)
		stamp := likeTimestamp(t, db, prefix+"post_like", "post_uid", postUid, userUid)

		// 같은 상태 재설정: 무변경, timestamp 유지
		changed, err = viewRepo.SetPostReaction(models.BoardReactionParam{
			BoardUid: boardUid, PostUid: postUid, UserUid: userUid,
			Reaction: models.REACTIONS.LIKE, ReactionCode: models.REACTION_LIKE,
		})
		if err != nil || changed {
			t.Fatalf("repeat like: changed=%v err=%v", changed, err)
		}
		if got := likeTimestamp(t, db, prefix+"post_like", "post_uid", postUid, userUid); got != stamp {
			t.Fatalf("repeat like changed timestamp: %d -> %d", stamp, got)
		}

		// 다른 종류로 교체
		changed, err = viewRepo.SetPostReaction(models.BoardReactionParam{
			BoardUid: boardUid, PostUid: postUid, UserUid: userUid,
			Reaction: models.REACTIONS.BEST, ReactionCode: models.REACTION_BEST,
		})
		if err != nil || !changed {
			t.Fatalf("switch to best: changed=%v err=%v", changed, err)
		}
		assertPostState(t, viewRepo, postUid, userUid, [10]uint{0, 1, 0, 0, 0, 0, 0, 0, 0, 0}, models.REACTIONS.BEST)

		// 확장 종류로 교체
		changed, err = viewRepo.SetPostReaction(models.BoardReactionParam{
			BoardUid: boardUid, PostUid: postUid, UserUid: userUid,
			Reaction: models.REACTIONS.LAUGH, ReactionCode: models.REACTION_LAUGH,
		})
		if err != nil || !changed {
			t.Fatalf("switch to laugh: changed=%v err=%v", changed, err)
		}
		assertPostState(t, viewRepo, postUid, userUid, [10]uint{0, 0, 0, 0, 1, 0, 0, 0, 0, 0}, models.REACTIONS.LAUGH)

		// 취소(null)
		changed, err = viewRepo.SetPostReaction(models.BoardReactionParam{
			BoardUid: boardUid, PostUid: postUid, UserUid: userUid,
			ReactionCode: models.REACTION_NONE,
		})
		if err != nil || !changed {
			t.Fatalf("cancel: changed=%v err=%v", changed, err)
		}
		assertPostState(t, viewRepo, postUid, userUid, [10]uint{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "")
		if count := likeRowCount(t, db, prefix+"post_like", "post_uid", postUid, userUid); count != 1 {
			t.Fatalf("rows after cancel = %d, want 1", count)
		}

		// 취소 상태에서 다시 취소: 무변경
		changed, err = viewRepo.SetPostReaction(models.BoardReactionParam{
			BoardUid: boardUid, PostUid: postUid, UserUid: userUid,
			ReactionCode: models.REACTION_NONE,
		})
		if err != nil || changed {
			t.Fatalf("cancel again: changed=%v err=%v", changed, err)
		}
	})

	t.Run("comment transitions keep one row and liked projection", func(t *testing.T) {
		changed, err := commentRepo.SetCommentReaction(models.CommentReactionParam{
			BoardUid: boardUid, CommentUid: commentUid, UserUid: userUid,
			Reaction: models.REACTIONS.LIKE, ReactionCode: models.REACTION_LIKE,
		})
		if err != nil || !changed {
			t.Fatalf("first like: changed=%v err=%v", changed, err)
		}
		state, err := commentRepo.GetCommentReactionState(commentUid, userUid)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if state.Reactions.Like != 1 || state.MyReaction == nil || *state.MyReaction != models.REACTIONS.LIKE {
			t.Fatalf("unexpected comment state: %+v", state)
		}
		if _, err := commentRepo.SetCommentReaction(models.CommentReactionParam{
			BoardUid: boardUid, CommentUid: commentUid, UserUid: userUid,
			ReactionCode: models.REACTION_NONE,
		}); err != nil {
			t.Fatalf("cancel: %v", err)
		}
		state, err = commentRepo.GetCommentReactionState(commentUid, userUid)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if state.Reactions.Like != 0 || state.MyReaction != nil {
			t.Fatalf("unexpected comment state after cancel: %+v", state)
		}
		if count := likeRowCount(t, db, prefix+"comment_like", "comment_uid", commentUid, userUid); count != 1 {
			t.Fatalf("comment rows = %d, want 1", count)
		}
	})

	t.Run("concurrent requests leave one row with a valid state", func(t *testing.T) {
		codes := append(models.ReactionCodeList, models.REACTION_NONE)
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				code := codes[i%len(codes)]
				reaction := ""
				if code != models.REACTION_NONE {
					reaction = code.APIValue()
				}
				_, err := viewRepo.SetPostReaction(models.BoardReactionParam{
					BoardUid: boardUid, PostUid: postUid + 1, UserUid: userUid,
					Reaction: reaction, ReactionCode: code,
				})
				if err != nil {
					t.Errorf("concurrent set %d: %v", i, err)
				}
			}(i)
		}
		wg.Wait()
		if count := likeRowCount(t, db, prefix+"post_like", "post_uid", postUid+1, userUid); count != 1 {
			t.Fatalf("rows after concurrent requests = %d, want 1", count)
		}
		var liked, reactionType int
		if err := db.QueryRow(fmt.Sprintf("SELECT liked, reaction_type FROM %spost_like WHERE post_uid = ? AND user_uid = ?",
			prefix), postUid+1, userUid).Scan(&liked, &reactionType); err != nil {
			t.Fatalf("read final row: %v", err)
		}
		if liked != func() int {
			if models.ReactionType(reactionType) == models.REACTION_LIKE {
				return 1
			}
			return 0
		}() {
			t.Fatalf("liked=%d is inconsistent with reaction_type=%d", liked, reactionType)
		}
	})
}

func createReactionStateTables(t *testing.T, db *sql.DB, prefix string) {
	t.Helper()
	exec := func(query string) {
		t.Helper()
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("fixture %s: %v", query, err)
		}
	}
	for _, table := range []string{"post_like", "comment_like"} {
		column := "post_uid"
		if table == "comment_like" {
			column = "comment_uid"
		}
		exec(fmt.Sprintf(`CREATE TABLE %s%s (
			uid BIGINT UNSIGNED NOT NULL auto_increment,
			board_uid INT UNSIGNED NOT NULL DEFAULT 0,
			%s INT UNSIGNED NOT NULL DEFAULT 0,
			user_uid INT UNSIGNED NOT NULL DEFAULT 0,
			liked TINYINT UNSIGNED NOT NULL DEFAULT 0,
			reaction_type TINYINT UNSIGNED NOT NULL DEFAULT 0,
			timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
			PRIMARY KEY (uid),
			UNIQUE KEY uq (%s, user_uid),
			KEY idx_reaction (%s, reaction_type)
		) ENGINE=InnoDB`, prefix, table, column, column, column))
	}
}

func dropReactionStateTables(t *testing.T, db *sql.DB, prefix string) {
	t.Helper()
	for _, table := range []string{prefix + "post_like", prefix + "comment_like"} {
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Fatalf("drop %s: %v", table, err)
		}
	}
}

func assertPostState(t *testing.T, viewRepo BoardViewRepository, postUid uint, userUid uint, want [10]uint, myReaction models.Reaction) {
	t.Helper()
	state, err := viewRepo.GetPostReactionState(postUid, userUid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	got := [10]uint{state.Reactions.Like, state.Reactions.Best, state.Reactions.Facepalm, state.Reactions.Hmm,
		state.Reactions.Laugh, state.Reactions.Celebrate, state.Reactions.Fire, state.Reactions.Support, state.Reactions.Sad, state.Reactions.Eyes}
	if got != want {
		t.Fatalf("reactions = %v, want %v", got, want)
	}
	if myReaction == "" {
		if state.MyReaction != nil {
			t.Fatalf("myReaction = %q, want nil", *state.MyReaction)
		}
	} else if state.MyReaction == nil || *state.MyReaction != myReaction {
		t.Fatalf("myReaction = %v, want %q", state.MyReaction, myReaction)
	}
	if liked := viewRepo.IsLikedPost(postUid, userUid); liked != (want[0] > 0) {
		t.Fatalf("IsLikedPost = %v, want %v", liked, want[0] > 0)
	}
}

func likeRowCount(t *testing.T, db *sql.DB, table, column string, targetUid uint, userUid uint) int {
	t.Helper()
	var count int
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND user_uid = ?", table, column),
		targetUid, userUid).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return count
}

func likeTimestamp(t *testing.T, db *sql.DB, table, column string, targetUid uint, userUid uint) int64 {
	t.Helper()
	var stamp int64
	if err := db.QueryRow(fmt.Sprintf("SELECT timestamp FROM %s WHERE %s = ? AND user_uid = ?", table, column),
		targetUid, userUid).Scan(&stamp); err != nil {
		t.Fatalf("read timestamp: %v", err)
	}
	return stamp
}
