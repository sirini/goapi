package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/pkg/models"
)

func studioSummaryRows(postCount, photoCount, viewCount, likeCount, commentCount uint64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"post_count", "photo_count", "view_count", "like_count", "comment_count",
	}).AddRow(postCount, photoCount, viewCount, likeCount, commentCount)
}

func expectStudioSummary(mock sqlmock.Sqlmock, boardUid uint, userUid uint, rows *sqlmock.Rows) {
	mock.ExpectQuery(`(?s)SELECT.*COUNT\(\*\).*SUM\(studio\.image_count\).*SUM\(studio\.hit\).*SUM\(studio\.like_count\).*SUM\(studio\.comment_count\).*FROM nubo_post AS p.*p\.board_uid = \? AND p\.user_uid = \? AND p\.status IN \(\?, \?\)`).
		WithArgs(boardUid, userUid, models.CONTENT_NORMAL, models.CONTENT_SECRET).
		WillReturnRows(rows)
}

func TestGetStudioAggregatesAndPagesOnlyRequestedUserPosts(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	expectStudioSummary(mock, 7, 11, studioSummaryRows(3, 5, 61, 4, 3))
	mock.ExpectQuery(`(?s)SELECT.*SELECT ft\.path FROM nubo_file_thumbnail AS ft.*SELECT COUNT\(\*\) FROM nubo_file AS f.*image_thumb\.file_uid = f\.uid.*SELECT COUNT\(\*\) FROM nubo_post_like AS pl.*pl\.liked = 1.*SELECT COUNT\(\*\) FROM nubo_comment AS c.*c\.status != -1.*FROM nubo_post AS p.*p\.board_uid = \? AND p\.user_uid = \? AND p\.status IN \(\?, \?\).*ORDER BY p\.submitted DESC, p\.uid DESC.*LIMIT \? OFFSET \?`).
		WithArgs(uint(7), uint(11), models.CONTENT_NORMAL, models.CONTENT_SECRET, uint(2), uint64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"uid", "title", "cover", "submitted", "modified", "status", "image_count", "hit", "like_count", "comment_count",
		}).
			AddRow(19, "Secret work", "/upload/thumbnails/2026/08/tcover.webp", 3000, 3100, models.CONTENT_SECRET, 3, 40, 2, 1).
			AddRow(18, "Normal work", "/srv/nubo/upload/attachments/original.jpg", 2000, 2100, models.CONTENT_NORMAL, 2, 21, 2, 2))

	result, err := repo.GetStudio(models.BoardStudioParam{
		BoardUid: 7,
		UserUid:  11,
		Page:     1,
		Limit:    2,
		Sort:     models.BOARD_STUDIO_SORT_RECENT,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary != (models.BoardStudioSummary{PostCount: 3, PhotoCount: 5, ViewCount: 61, LikeCount: 4, CommentCount: 3}) {
		t.Fatalf("unexpected summary: %+v", result.Summary)
	}
	if result.Posts.Page != 1 || result.Posts.Limit != 2 || result.Posts.TotalCount != 3 || !result.Posts.HasNext {
		t.Fatalf("unexpected page metadata: %+v", result.Posts)
	}
	if len(result.Posts.Items) != 2 || result.Posts.Items[0].Status != models.CONTENT_SECRET {
		t.Fatalf("unexpected studio items: %+v", result.Posts.Items)
	}
	if result.Posts.Items[0].Cover != "/upload/thumbnails/2026/08/tcover.webp" {
		t.Fatalf("preview cover was changed: %q", result.Posts.Items[0].Cover)
	}
	if result.Posts.Items[1].Cover != "" {
		t.Fatalf("internal/original path was exposed: %q", result.Posts.Items[1].Cover)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetStudioReturnsZeroSummaryAndEmptyArray(t *testing.T) {
	withBoardRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboBoardRepository(db)

	expectStudioSummary(mock, 7, 22, studioSummaryRows(0, 0, 0, 0, 0))
	mock.ExpectQuery(`(?s)FROM nubo_post AS p.*ORDER BY p\.submitted DESC, p\.uid DESC.*LIMIT \? OFFSET \?`).
		WithArgs(uint(7), uint(22), models.CONTENT_NORMAL, models.CONTENT_SECRET, uint(20), uint64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"uid", "title", "cover", "submitted", "modified", "status", "image_count", "hit", "like_count", "comment_count",
		}))

	result, err := repo.GetStudio(models.BoardStudioParam{
		BoardUid: 7,
		UserUid:  22,
		Page:     1,
		Limit:    20,
		Sort:     models.BOARD_STUDIO_SORT_RECENT,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary != (models.BoardStudioSummary{}) || len(result.Posts.Items) != 0 || result.Posts.Items == nil || result.Posts.HasNext {
		t.Fatalf("unexpected empty studio result: %+v", result)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"items":[]`) {
		t.Fatalf("empty items were not encoded as an array: %s", encoded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStudioOrderClauseUsesWhitelistedTieBreakers(t *testing.T) {
	for _, test := range []struct {
		sort models.BoardStudioSort
		want string
	}{
		{models.BOARD_STUDIO_SORT_RECENT, "p.submitted DESC, p.uid DESC"},
		{models.BOARD_STUDIO_SORT_VIEWS, "p.hit DESC, p.uid DESC"},
		{models.BOARD_STUDIO_SORT_LIKES, "like_count DESC, p.uid DESC"},
		{models.BOARD_STUDIO_SORT_COMMENTS, "comment_count DESC, p.uid DESC"},
	} {
		t.Run(string(test.sort), func(t *testing.T) {
			got, err := studioOrderClause(test.sort)
			if err != nil || got != test.want {
				t.Fatalf("studioOrderClause(%q) = %q, %v; want %q", test.sort, got, err, test.want)
			}
		})
	}
	if _, err := studioOrderClause("uid DESC; DROP TABLE post"); err == nil {
		t.Fatal("unknown studio sort was accepted")
	}
}

func TestStudioCoverPathOnlyAllowsPublicPreviewThumbnails(t *testing.T) {
	for _, test := range []struct {
		value string
		want  string
	}{
		{"/upload/thumbnails/2026/08/tphoto.webp", "/upload/thumbnails/2026/08/tphoto.webp"},
		{"./upload/thumbnails/2026/08/tphoto.webp", "/upload/thumbnails/2026/08/tphoto.webp"},
		{"/upload/thumbnails/../attachments/original.jpg", ""},
		{"/upload/attachments/original.jpg", ""},
		{"/var/lib/nubo/upload/thumbnails/tphoto.webp", ""},
	} {
		if got := studioCoverPath(test.value); got != test.want {
			t.Errorf("studioCoverPath(%q) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestGetStudioPropagatesRepositoryErrors(t *testing.T) {
	withBoardRepositoryTestPrefix(t)

	t.Run("summary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(errors.New("summary failed"))

		_, err = NewNuboBoardRepository(db).GetStudio(models.BoardStudioParam{
			BoardUid: 7, UserUid: 11, Page: 1, Limit: 20, Sort: models.BOARD_STUDIO_SORT_RECENT,
		})
		if err == nil || !strings.Contains(err.Error(), "summary failed") {
			t.Fatalf("summary error = %v", err)
		}
	})

	t.Run("posts", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		expectStudioSummary(mock, 7, 11, studioSummaryRows(1, 1, 1, 1, 1))
		mock.ExpectQuery(`(?s)FROM nubo_post AS p.*ORDER BY p\.submitted DESC, p\.uid DESC`).
			WithArgs(uint(7), uint(11), models.CONTENT_NORMAL, models.CONTENT_SECRET, uint(20), uint64(0)).
			WillReturnError(sql.ErrConnDone)

		_, err = NewNuboBoardRepository(db).GetStudio(models.BoardStudioParam{
			BoardUid: 7, UserUid: 11, Page: 1, Limit: 20, Sort: models.BOARD_STUDIO_SORT_RECENT,
		})
		if !errors.Is(err, sql.ErrConnDone) {
			t.Fatalf("posts error = %v, want sql.ErrConnDone", err)
		}
	})
}
