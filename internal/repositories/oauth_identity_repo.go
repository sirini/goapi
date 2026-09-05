package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

var ErrOAuthIdentityConflict = errors.New("oauth identity is already linked")
var ErrOAuthAccountLinkRequired = errors.New("an account with this email already exists")
var ErrOAuthNameConflict = errors.New("oauth display name is already in use")

type OAuthIdentityRepository interface {
	SaveNonce(provider, purpose string, userUid uint, digest string, expires int64) error
	ConsumeNonce(provider, purpose string, userUid uint, digest string, now int64) (bool, error)
	FindUser(provider, subject string) (uint, error)
	HasProvider(userUid uint, provider string) (bool, error)
	Touch(provider, subject string, now int64) error
	Link(userUid uint, provider, subject, email string, now int64) error
	CreateUserAndLink(provider, subject, email, password, name string, now int64) (uint, error)
}

type NuboOAuthIdentityRepository struct {
	db *sql.DB
}

func NewNuboOAuthIdentityRepository(db *sql.DB) *NuboOAuthIdentityRepository {
	return &NuboOAuthIdentityRepository{db: db}
}

// SaveNonce는 만료된 nonce를 정리하고 새 digest만 저장한다.
func (r *NuboOAuthIdentityRepository) SaveNonce(provider, purpose string, userUid uint, digest string, expires int64) error {
	if _, err := r.db.Exec(
		fmt.Sprintf("DELETE FROM %s%s WHERE expires < ?", configs.Env.Prefix, models.TABLE_OAUTH_NONCE),
		time.Now().UnixMilli(),
	); err != nil {
		return err
	}
	_, err := r.db.Exec(
		fmt.Sprintf("INSERT INTO %s%s (digest, provider, purpose, user_uid, expires) VALUES (?, ?, ?, ?, ?)", configs.Env.Prefix, models.TABLE_OAUTH_NONCE),
		digest, provider, purpose, userUid, expires,
	)
	return err
}

// ConsumeNonce는 목적과 사용자까지 일치하고 아직 만료되지 않은 nonce를 한 번만 삭제한다.
func (r *NuboOAuthIdentityRepository) ConsumeNonce(provider, purpose string, userUid uint, digest string, now int64) (bool, error) {
	result, err := r.db.Exec(
		fmt.Sprintf("DELETE FROM %s%s WHERE digest = ? AND provider = ? AND purpose = ? AND user_uid = ? AND expires >= ?", configs.Env.Prefix, models.TABLE_OAUTH_NONCE),
		digest, provider, purpose, userUid, now,
	)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (r *NuboOAuthIdentityRepository) FindUser(provider, subject string) (uint, error) {
	var userUid uint
	err := r.db.QueryRow(
		fmt.Sprintf("SELECT user_uid FROM %s%s WHERE provider = ? AND subject = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		provider, subject,
	).Scan(&userUid)
	return userUid, err
}

func (r *NuboOAuthIdentityRepository) HasProvider(userUid uint, provider string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE user_uid = ? AND provider = ?)", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		userUid, provider,
	).Scan(&exists)
	return exists, err
}

func (r *NuboOAuthIdentityRepository) Touch(provider, subject string, now int64) error {
	_, err := r.db.Exec(
		fmt.Sprintf("UPDATE %s%s SET last_used = ? WHERE provider = ? AND subject = ?", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		now, provider, subject,
	)
	return err
}

// Link는 현재 계정과 제공자 계정을 모두 증명한 뒤에만 연결한다.
func (r *NuboOAuthIdentityRepository) Link(userUid uint, provider, subject, email string, now int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var linkedUser uint
	err = tx.QueryRow(
		fmt.Sprintf("SELECT user_uid FROM %s%s WHERE provider = ? AND subject = ? LIMIT 1 FOR UPDATE", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		provider, subject,
	).Scan(&linkedUser)
	if err == nil {
		if linkedUser != userUid {
			return ErrOAuthIdentityConflict
		}
		_, err = tx.Exec(
			fmt.Sprintf("UPDATE %s%s SET email = ?, last_used = ? WHERE provider = ? AND subject = ?", configs.Env.Prefix, models.TABLE_USER_OAUTH),
			email, now, provider, subject,
		)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return err
	}

	var linkedSubject string
	err = tx.QueryRow(
		fmt.Sprintf("SELECT subject FROM %s%s WHERE provider = ? AND user_uid = ? LIMIT 1 FOR UPDATE", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		provider, userUid,
	).Scan(&linkedSubject)
	if err == nil {
		return ErrOAuthIdentityConflict
	}
	if err != sql.ErrNoRows {
		return err
	}

	_, err = tx.Exec(
		fmt.Sprintf("INSERT INTO %s%s (user_uid, provider, subject, email, created, last_used) VALUES (?, ?, ?, ?, ?, ?)", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		userUid, provider, subject, email, now, now,
	)
	if isDuplicateKey(err) {
		return ErrOAuthIdentityConflict
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// CreateUserAndLink는 새 사용자와 Apple subject 연결을 한 트랜잭션으로 만든다.
func (r *NuboOAuthIdentityRepository) CreateUserAndLink(provider, subject, email, password, name string, now int64) (uint, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var existing uint
	err = tx.QueryRow(
		fmt.Sprintf("SELECT user_uid FROM %s%s WHERE provider = ? AND subject = ? LIMIT 1 FOR UPDATE", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		provider, subject,
	).Scan(&existing)
	if err == nil {
		return 0, ErrOAuthIdentityConflict
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = tx.QueryRow(
		fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1 FOR UPDATE", configs.Env.Prefix, models.TABLE_USER),
		email,
	).Scan(&existing)
	if err == nil {
		return 0, ErrOAuthAccountLinkRequired
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = tx.QueryRow(
		fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1 FOR UPDATE", configs.Env.Prefix, models.TABLE_USER),
		name,
	).Scan(&existing)
	if err == nil {
		return 0, ErrOAuthNameConflict
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	result, err := tx.Exec(
		fmt.Sprintf(`INSERT INTO %s%s
			(id, name, password, profile, level, point, signature, signup, signin, blocked)
			VALUES (?, ?, ?, '', 1, 100, '', ?, 0, 0)`, configs.Env.Prefix, models.TABLE_USER),
		email, name, passwordHash, now,
	)
	if err != nil {
		return 0, err
	}
	inserted, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	userUid := uint(inserted)
	_, err = tx.Exec(
		fmt.Sprintf("INSERT INTO %s%s (user_uid, provider, subject, email, created, last_used) VALUES (?, ?, ?, ?, ?, ?)", configs.Env.Prefix, models.TABLE_USER_OAUTH),
		userUid, provider, subject, email, now, now,
	)
	if isDuplicateKey(err) {
		return 0, ErrOAuthIdentityConflict
	}
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userUid, nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
