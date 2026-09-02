package configs

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBadgeSchemaAndBuiltInsAreReRunnable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS nubo_badge_definition`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS nubo_user_badge`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS nubo_post_origin`).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := createBadgeTables(db, "nubo_"); err != nil {
		t.Fatal(err)
	}

	for range 3 {
		mock.ExpectExec(`INSERT IGNORE INTO nubo_badge_definition`).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	if err := seedBuiltInBadges(db, "nubo_"); err != nil {
		t.Fatal(err)
	}

	for _, badgeKey := range []string{"first-post", "first-comment"} {
		mock.ExpectQuery(`SELECT backfilled_at FROM nubo_badge_definition`).
			WithArgs(badgeKey).
			WillReturnRows(sqlmock.NewRows([]string{"backfilled_at"}).AddRow(0))
		mock.ExpectExec(`INSERT IGNORE INTO nubo_user_badge`).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(`UPDATE nubo_badge_definition SET backfilled_at`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), badgeKey).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	if err := backfillBuiltInBadges(db, "nubo_"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBadgeBackfillSkipsCompletedAchievement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT backfilled_at FROM nubo_badge_definition`).
		WithArgs("first-post").
		WillReturnRows(sqlmock.NewRows([]string{"backfilled_at"}).AddRow(1000))
	mock.ExpectQuery(`SELECT backfilled_at FROM nubo_badge_definition`).
		WithArgs("first-comment").
		WillReturnRows(sqlmock.NewRows([]string{"backfilled_at"}).AddRow(1000))
	if err := backfillBuiltInBadges(db, "nubo_"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
