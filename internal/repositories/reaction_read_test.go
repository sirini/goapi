package repositories

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/pkg/models"
)

// 목록·공지·상세·댓글 읽기는 리액션 기록이 없는 사용자(와 0 uid의 비로그인 사용자)도
// COALESCE된 0을 받아 행을 온전히 스캔해야 한다. 인자 개수·순서 회귀도 함께 고정한다.

func reactionRowsForPosts() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"uid", "user_uid", "category_uid", "title", "content", "submitted", "modified", "hit", "status",
		"name", "profile", "category_name", "cover", "comment_count", "like_count", "liked",
		"like_c", "best_c", "facepalm_c", "hmm_c", "my_reaction",
	}).AddRow(
		uint(11), uint(3), uint(2), "제목", "내용", int64(100), int64(100), uint(5), models.CONTENT_NORMAL,
		"작성자", "프로필", "분류", "", uint(4), uint(3), 0,
		uint(3), uint(2), uint(0), uint(1), uint8(0),
	)
}

func assertPostReactionItem(t *testing.T, item models.BoardListItem) {
	t.Helper()
	if item.Reactions.Like != 3 || item.Reactions.Best != 2 ||
		item.Reactions.Facepalm != 0 || item.Reactions.Hmm != 1 {
		t.Fatalf("unexpected reactions: %+v", item.Reactions)
	}
	if item.MyReaction != nil {
		t.Fatalf("myReaction = %q, want nil for user without reaction rows", *item.MyReaction)
	}
	if item.Uid != 11 || item.Comment != 4 || item.Like != 3 || item.Liked {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestFindPostsScansReactionsForUserWithoutReactionRows(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	mock.ExpectQuery(`(?s)AS sub ON p\.uid = sub\.uid.*ORDER BY p\.uid DESC`).
		WithArgs(
			models.CONTENT_REMOVED,
			uint(9), uint(9),
			uint(7),
			models.CONTENT_NORMAL, models.CONTENT_SECRET,
			uint(12), uint(0),
		).
		WillReturnRows(reactionRowsForPosts())

	items, err := repo.FindPosts(models.BoardListParam{
		BoardUid: 7, UserUid: 9, Limit: 12, Page: 1,
	})
	if err != nil {
		t.Fatalf("FindPosts() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("FindPosts() returned %d items, want 1", len(items))
	}
	assertPostReactionItem(t, items[0])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindPostsScansReactionsForAnonymousVisitor(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	mock.ExpectQuery(`(?s)AS sub ON p\.uid = sub\.uid.*ORDER BY p\.uid DESC`).
		WithArgs(
			models.CONTENT_REMOVED,
			uint(0), uint(0),
			uint(7),
			models.CONTENT_NORMAL, models.CONTENT_SECRET,
			uint(12), uint(0),
		).
		WillReturnRows(reactionRowsForPosts())

	items, err := repo.FindPosts(models.BoardListParam{
		BoardUid: 7, UserUid: 0, Limit: 12, Page: 1,
	})
	if err != nil {
		t.Fatalf("FindPosts() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("FindPosts() returned %d items, want 1", len(items))
	}
	assertPostReactionItem(t, items[0])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetNoticePostsScansReactionsForUserWithoutReactionRows(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	mock.ExpectQuery(`(?s)board_uid = \? AND status = \?`).
		WithArgs(
			models.CONTENT_REMOVED,
			uint(9), uint(9),
			uint(7),
			models.CONTENT_NOTICE,
		).
		WillReturnRows(reactionRowsForPosts())

	items, err := repo.GetNoticePosts(7, 9)
	if err != nil {
		t.Fatalf("GetNoticePosts() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("GetNoticePosts() returned %d items, want 1", len(items))
	}
	assertPostReactionItem(t, items[0])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetPostItemScansReactionsForUserWithoutReactionRows(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardViewRepository(db, NewNuboBoardRepository(db))

	mock.ExpectQuery(`(?s)WHERE p\.uid = \? AND p\.status != \?`).
		WithArgs(
			models.CONTENT_REMOVED,
			uint(9), uint(9),
			uint(11),
			models.CONTENT_REMOVED,
		).
		WillReturnRows(reactionRowsForPosts())

	item, err := repo.GetPostItem(11, 9)
	if err != nil {
		t.Fatalf("GetPostItem() error = %v", err)
	}
	assertPostReactionItem(t, item)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetCommentsScansReactionsForUserWithoutReactionRows(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboCommentRepository(db, NewNuboBoardRepository(db))

	mock.ExpectQuery(`(?s)AS p ON c\.uid = p\.uid.*ORDER BY c\.reply_uid ASC, c\.uid ASC`).
		WithArgs(
			uint(9), uint(9),
			uint(5),
			models.CONTENT_NORMAL, models.CONTENT_SECRET,
			uint(20), uint(0),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"uid", "reply_uid", "user_uid", "content", "submitted", "modified", "status",
			"name", "profile", "like_count", "liked",
			"like_c", "best_c", "facepalm_c", "hmm_c", "my_reaction",
		}).AddRow(
			uint(21), uint(0), uint(3), "댓글", int64(100), int64(100), models.CONTENT_NORMAL,
			"작성자", "프로필", uint(2), 0,
			uint(2), uint(1), uint(0), uint(0), uint8(0),
		))

	items, err := repo.GetComments(models.CommentListParam{PostUid: 5, UserUid: 9, Limit: 20, Page: 1})
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("GetComments() returned %d items, want 1", len(items))
	}
	reactions := items[0].Reactions
	myReaction := items[0].MyReaction
	if reactions.Like != 2 || reactions.Best != 1 || reactions.Facepalm != 0 || reactions.Hmm != 0 {
		t.Fatalf("unexpected reactions: %+v", reactions)
	}
	if myReaction != nil {
		t.Fatalf("myReaction = %q, want nil for user without reaction rows", *myReaction)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
