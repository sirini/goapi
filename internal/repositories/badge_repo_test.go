package repositories

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

func TestBadgeAwardIsIdempotentInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	oldPrefix := configs.Env.Prefix
	configs.Env.Prefix = "nubo_"
	defer func() { configs.Env.Prefix = oldPrefix }()

	mock.ExpectExec(`INSERT IGNORE INTO nubo_user_badge`).
		WithArgs(uint(7), uint64(1000), sqlmock.AnyArg(), "system", uint(0), "post", uint(31), models.BADGE_FIRST_POST).
		WillReturnResult(sqlmock.NewResult(0, 1))

	granted, err := NewNuboBadgeRepository(db).Award(models.BadgeAwardParam{
		UserUid: 7, BadgeKey: models.BADGE_FIRST_POST, QualifiedAt: 1000,
		GrantSource: "system", EvidenceType: "post", EvidenceUid: 31,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !granted {
		t.Fatal("new achievement was not reported as granted")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindFeaturedBadgesBatchesWriters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	oldPrefix := configs.Env.Prefix
	configs.Env.Prefix = "nubo_"
	defer func() { configs.Env.Prefix = oldPrefix }()

	mock.ExpectQuery(`FROM nubo_user_badge AS ub JOIN nubo_badge_definition AS d`).
		WithArgs(uint(7), uint(9)).
		WillReturnRows(sqlmock.NewRows([]string{"user_uid", "badge_key", "name", "description", "icon_key", "qualified_at"}).
			AddRow(7, models.BADGE_SENSTA_APP, "SENSTA 앱 포토그래퍼", "SENSTA 앱으로 사진을 공유한 사용자입니다.", "aperture", 1000))

	badges, err := NewNuboBadgeRepository(db).FindFeaturedForUsers([]uint{7, 9})
	if err != nil {
		t.Fatal(err)
	}
	if len(badges[7]) != 1 || badges[7][0].Key != models.BADGE_SENSTA_APP {
		t.Fatalf("featured badges = %#v", badges)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateManualBadgeDefinition(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	oldPrefix := configs.Env.Prefix
	configs.Env.Prefix = "nubo_"
	defer func() { configs.Env.Prefix = oldPrefix }()

	definition := models.BadgeDefinition{
		Key: "manual-test", Name: "사진전 우수상", Description: "2026년 여름 사진전",
		IconKey: "trophy", ShowInline: true, SortOrder: 50, Created: 1000, Updated: 1000,
	}
	mock.ExpectExec(`INSERT INTO nubo_badge_definition`).
		WithArgs(definition.Key, definition.Name, definition.Description, definition.IconKey,
			definition.ShowInline, definition.SortOrder, definition.Created, definition.Updated).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := NewNuboBadgeRepository(db).CreateDefinition(definition); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDefinitionOnlyTargetsManualAchievements(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	oldPrefix := configs.Env.Prefix
	configs.Env.Prefix = "nubo_"
	defer func() { configs.Env.Prefix = oldPrefix }()

	definition := models.BadgeDefinition{
		Key: "manual-test", Name: "사진전 대상", Description: "수정된 설명",
		IconKey: "crown", ShowInline: true, SortOrder: 40, Updated: 2000,
	}
	mock.ExpectExec(`WHERE badge_key = \? AND rule_key = '' LIMIT 1`).
		WithArgs(definition.Name, definition.Description, definition.IconKey, definition.ShowInline,
			definition.SortOrder, definition.Updated, definition.Key).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updated, err := NewNuboBadgeRepository(db).UpdateDefinition(definition)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("manual achievement was not updated")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnannouncedAchievementsCanBeAcknowledgedByOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	oldPrefix := configs.Env.Prefix
	configs.Env.Prefix = "nubo_"
	defer func() { configs.Env.Prefix = oldPrefix }()

	mock.ExpectQuery(`ub.announced_at = 0`).
		WithArgs(uint(7), uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"badge_key", "name", "description", "icon_key", "qualified_at"}).
			AddRow("manual-summer", "여름 사진전 우수상", "좋은 사진을 공유했습니다.", "trophy", 1000))

	badges, err := NewNuboBadgeRepository(db).FindUnannouncedForUser(7, 10)
	if err != nil || len(badges) != 1 {
		t.Fatalf("unannounced badges = %#v, err = %v", badges, err)
	}

	mock.ExpectExec(`SET announced_at = \?`).
		WithArgs(uint64(2000), uint(7), "manual-summer").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := NewNuboBadgeRepository(db).MarkAnnounced(7, []string{"manual-summer"}, 2000); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
