package repositories

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/utils"
)

func TestVerificationCodeValid(t *testing.T) {
	now := time.UnixMilli(2_000_000)
	tests := []struct {
		name      string
		timestamp int64
		want      bool
	}{
		{name: "fresh", timestamp: now.Add(-time.Minute).UnixMilli(), want: true},
		{name: "boundary", timestamp: now.Add(-verificationCodeLifetime).UnixMilli(), want: true},
		{name: "expired", timestamp: now.Add(-verificationCodeLifetime - time.Millisecond).UnixMilli(), want: false},
		{name: "future", timestamp: now.Add(time.Millisecond).UnixMilli(), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verificationCodeValid(tt.timestamp, now); got != tt.want {
				t.Fatalf("verificationCodeValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func withAuthRepositoryTestPrefix(t *testing.T) {
	t.Helper()
	previous := configs.Env
	configs.Env.Prefix = "nubo_"
	configs.Env.JWTRefreshDays = "30"
	t.Cleanup(func() { configs.Env = previous })
}

func TestConsumeVerificationCodeLocksDeletesAndRejectsReplay(t *testing.T) {
	withAuthRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboAuthRepository(db)
	selectQuery := regexp.QuoteMeta("SELECT email, code, timestamp FROM nubo_user_verification WHERE uid = ? LIMIT 1 FOR UPDATE")
	deleteQuery := regexp.QuoteMeta("DELETE FROM nubo_user_verification WHERE uid = ? LIMIT 1")

	mock.ExpectBegin()
	mock.ExpectQuery(selectQuery).WithArgs(uint(42)).WillReturnRows(
		sqlmock.NewRows([]string{"email", "code", "timestamp"}).AddRow("member@example.com", "123456", time.Now().UnixMilli()),
	)
	mock.ExpectExec(deleteQuery).WithArgs(uint(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	email, ok := repo.ConsumeVerificationCode(42, "123456", "member@example.com")
	if !ok || email != "member@example.com" {
		t.Fatalf("first consume = %q, %v", email, ok)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(selectQuery).WithArgs(uint(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	if _, ok := repo.ConsumeVerificationCode(42, "123456", "member@example.com"); ok {
		t.Fatal("consumed verification code was accepted again")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConsumeVerificationCodeRejectsDifferentEmailBeforeDelete(t *testing.T) {
	withAuthRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboAuthRepository(db)
	selectQuery := regexp.QuoteMeta("SELECT email, code, timestamp FROM nubo_user_verification WHERE uid = ? LIMIT 1 FOR UPDATE")
	mock.ExpectBegin()
	mock.ExpectQuery(selectQuery).WithArgs(uint(42)).WillReturnRows(
		sqlmock.NewRows([]string{"email", "code", "timestamp"}).AddRow("member@example.com", "123456", time.Now().UnixMilli()),
	)
	mock.ExpectRollback()
	if _, ok := repo.ConsumeVerificationCode(42, "123456", "other@example.com"); ok {
		t.Fatal("verification code was accepted for another email")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRotateRefreshTokenRequiresSingleConditionalUpdate(t *testing.T) {
	withAuthRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNuboAuthRepository(db)
	query := regexp.QuoteMeta(`UPDATE nubo_user_token SET refresh = ?, timestamp = ?
		WHERE user_uid = ? AND refresh = ? AND timestamp > ? LIMIT 1`)
	oldHash := utils.GetHashedString("old-refresh")

	mock.ExpectExec(query).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint(7), oldHash, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !repo.RotateRefreshToken(7, "old-refresh", "new-refresh") {
		t.Fatal("valid refresh token was not rotated")
	}

	mock.ExpectExec(query).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint(7), oldHash, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if repo.RotateRefreshToken(7, "old-refresh", "replayed-refresh") {
		t.Fatal("already rotated refresh token was accepted again")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
