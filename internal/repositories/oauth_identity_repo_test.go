package repositories

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/internal/configs"
)

func TestConsumeOAuthNonceRequiresSingleMatchingRow(t *testing.T) {
	old := configs.Env
	configs.Env.Prefix = "nubo_"
	t.Cleanup(func() { configs.Env = old })

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboOAuthIdentityRepository(db)
	query := "DELETE FROM nubo_oauth_nonce WHERE digest = ? AND provider = ? AND purpose = ? AND user_uid = ? AND expires >= ?"
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("digest", "apple", "signin", uint(0), int64(1000)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("digest", "apple", "signin", uint(0), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	consumed, err := repo.ConsumeNonce("apple", "signin", 0, "digest", 1000)
	if err != nil || !consumed {
		t.Fatalf("first consume = %t, %v", consumed, err)
	}
	consumed, err = repo.ConsumeNonce("apple", "signin", 0, "digest", 1001)
	if err != nil || consumed {
		t.Fatalf("replayed consume = %t, %v", consumed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLinkOAuthIdentityRejectsAnotherUser(t *testing.T) {
	old := configs.Env
	configs.Env.Prefix = "nubo_"
	t.Cleanup(func() { configs.Env = old })

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_uid FROM nubo_user_oauth_identity").
		WithArgs("apple", "apple-subject").
		WillReturnRows(sqlmock.NewRows([]string{"user_uid"}).AddRow(9))
	mock.ExpectRollback()

	err = NewNuboOAuthIdentityRepository(db).Link(7, "apple", "apple-subject", "person@example.com", 1000)
	if !errors.Is(err, ErrOAuthIdentityConflict) {
		t.Fatalf("link error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindOAuthIdentityKeepsMissingIdentityDistinct(t *testing.T) {
	old := configs.Env
	configs.Env.Prefix = "nubo_"
	t.Cleanup(func() { configs.Env = old })

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT user_uid FROM nubo_user_oauth_identity").
		WithArgs("apple", "unknown").
		WillReturnError(sql.ErrNoRows)

	_, err = NewNuboOAuthIdentityRepository(db).FindUser("apple", "unknown")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing identity error = %v", err)
	}
}

func TestCreateAppleUserRequiresExplicitLinkForExistingEmail(t *testing.T) {
	old := configs.Env
	configs.Env.Prefix = "nubo_"
	t.Cleanup(func() { configs.Env = old })

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_uid FROM nubo_user_oauth_identity").
		WithArgs("apple", "new-apple-subject").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT uid FROM nubo_user WHERE id").
		WithArgs("existing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"uid"}).AddRow(7))
	mock.ExpectRollback()

	_, err = NewNuboOAuthIdentityRepository(db).CreateUserAndLink(
		"apple", "new-apple-subject", "existing@example.com", "password", "Person", 1000,
	)
	if !errors.Is(err, ErrOAuthAccountLinkRequired) {
		t.Fatalf("registration error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
