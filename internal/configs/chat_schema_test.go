package configs

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnsureChatSchemaMigratesReadReceiptsAndMessageLength(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM information_schema\.COLUMNS`).
		WithArgs("nubo_chat").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`ALTER TABLE nubo_chat ADD COLUMN read_at BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER timestamp`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema\.COLUMNS`).
		WithArgs("nubo_chat").
		WillReturnRows(sqlmock.NewRows([]string{"character_maximum_length"}).AddRow(1000))
	mock.ExpectExec(`ALTER TABLE nubo_chat MODIFY COLUMN message VARCHAR\(2000\) NOT NULL DEFAULT ''`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM information_schema\.STATISTICS`).
		WithArgs("nubo_chat").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`ALTER TABLE nubo_chat ADD KEY idx_chat_recipient_sender_uid \(to_uid, from_uid, uid\)`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := ensureChatSchema(db, "nubo_"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureChatSchemaIsRepeatable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM information_schema\.COLUMNS`).
		WithArgs("nubo_chat").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema\.COLUMNS`).
		WithArgs("nubo_chat").
		WillReturnRows(sqlmock.NewRows([]string{"character_maximum_length"}).AddRow(2000))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM information_schema\.STATISTICS`).
		WithArgs("nubo_chat").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	if err := ensureChatSchema(db, "nubo_"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
