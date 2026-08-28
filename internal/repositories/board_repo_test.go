package repositories

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

func withBoardRepositoryTestPrefix(t *testing.T) {
	t.Helper()
	previous := configs.Env
	configs.Env.Prefix = "nubo_"
	t.Cleanup(func() { configs.Env = previous })
}

func expectMissingTag(t *testing.T, mock sqlmock.Sqlmock, keyword string) {
	t.Helper()
	query := regexp.QuoteMeta("SELECT uid FROM nubo_hashtag WHERE name = ? LIMIT 1")
	statement := mock.ExpectPrepare(query)
	statement.ExpectQuery().WithArgs(keyword).WillReturnError(sql.ErrNoRows)
}

func TestGetTagUidsAcceptsHashtagPrefix(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	query := regexp.QuoteMeta("SELECT uid FROM nubo_hashtag WHERE name = ? LIMIT 1")
	statement := mock.ExpectPrepare(query)
	statement.ExpectQuery().WithArgs("photo").WillReturnRows(
		sqlmock.NewRows([]string{"uid"}).AddRow(42),
	)
	uids, count := repo.GetTagUids("#photo")
	if uids != "'42'" || count != 1 {
		t.Fatalf("GetTagUids() = %q, %d, want %q, 1", uids, count, "'42'")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTotalCountReturnsZeroWhenTagDoesNotExist(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	expectMissingTag(t, mock, "missing")
	got := repo.GetTotalCount(models.BoardListParam{
		BoardUid: 7,
		Option:   models.SEARCH_TAG,
		Keyword:  "missing",
	})
	if got != 0 {
		t.Fatalf("GetTotalCount() = %d, want 0", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindPostsReturnsEmptyWhenTagDoesNotExist(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	expectMissingTag(t, mock, "missing")
	items, err := repo.FindPosts(models.BoardListParam{
		BoardUid: 7,
		Option:   models.SEARCH_TAG,
		Keyword:  "missing",
		Limit:    12,
		Page:     1,
	})
	if err != nil {
		t.Fatalf("FindPosts() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("FindPosts() returned %d items, want 0", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindLatestPostsByTagReturnsEmptyWhenTagDoesNotExist(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	boardRepo := NewNuboBoardRepository(db)
	homeRepo := NewNuboHomeRepository(db, boardRepo)

	expectMissingTag(t, mock, "missing")
	items, err := homeRepo.FindLatestPostsByTag(models.HomePostParam{
		Option:  models.SEARCH_TAG,
		Keyword: "missing",
		Bunch:   12,
	})
	if err != nil {
		t.Fatalf("FindLatestPostsByTag() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("FindLatestPostsByTag() returned %d items, want 0", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
