package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInviteInvalid  = errors.New("invitation is invalid, expired, revoked, used, or issued for another email")
	ErrSignupConflict = errors.New("email or name is already in use")
)

type SignupInviteRepository interface {
	CreateInvite(email, tokenHash string, createdBy uint, created, expires int64) (uint, error)
	ListInvites(limit uint) ([]models.SignupInvite, error)
	RevokeInvite(uid uint) error
	ConsumeInviteAndCreateUser(tokenHash, email, password, name string, now int64) (uint, error)
}

type NuboSignupInviteRepository struct {
	db *sql.DB
}

func NewNuboSignupInviteRepository(db *sql.DB) *NuboSignupInviteRepository {
	return &NuboSignupInviteRepository{db: db}
}

func (r *NuboSignupInviteRepository) CreateInvite(email, tokenHash string, createdBy uint, created, expires int64) (uint, error) {
	query := fmt.Sprintf(`INSERT INTO %s%s (email, token_hash, created, expires, used, revoked, created_by)
		VALUES (?, ?, ?, ?, 0, 0, ?)`, configs.Env.Prefix, models.TABLE_SIGNUP_INVITE)
	result, err := r.db.Exec(query, strings.ToLower(strings.TrimSpace(email)), tokenHash, created, expires, createdBy)
	if err != nil {
		return 0, err
	}
	uid, err := result.LastInsertId()
	return uint(uid), err
}

func (r *NuboSignupInviteRepository) ListInvites(limit uint) ([]models.SignupInvite, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	query := fmt.Sprintf(`SELECT uid, email, created, expires, used, revoked, created_by
		FROM %s%s ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_SIGNUP_INVITE)
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.SignupInvite, 0)
	for rows.Next() {
		var item models.SignupInvite
		if err := rows.Scan(&item.Uid, &item.Email, &item.Created, &item.Expires, &item.Used, &item.Revoked, &item.CreatedBy); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *NuboSignupInviteRepository) RevokeInvite(uid uint) error {
	query := fmt.Sprintf(`UPDATE %s%s SET revoked = 1 WHERE uid = ? AND used = 0`, configs.Env.Prefix, models.TABLE_SIGNUP_INVITE)
	result, err := r.db.Exec(query, uid)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrInviteInvalid
	}
	return nil
}

func (r *NuboSignupInviteRepository) ConsumeInviteAndCreateUser(tokenHash, email, password, name string, now int64) (uint, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var inviteUid uint
	var invitedEmail string
	var expires, used int64
	var revoked bool
	query := fmt.Sprintf(`SELECT uid, email, expires, used, revoked FROM %s%s
		WHERE token_hash = ? LIMIT 1 FOR UPDATE`, configs.Env.Prefix, models.TABLE_SIGNUP_INVITE)
	if err := tx.QueryRow(query, tokenHash).Scan(&inviteUid, &invitedEmail, &expires, &used, &revoked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInviteInvalid
		}
		return 0, err
	}
	if !strings.EqualFold(invitedEmail, strings.TrimSpace(email)) || revoked || used != 0 || expires < now {
		return 0, ErrInviteInvalid
	}

	var exists uint
	query = fmt.Sprintf(`SELECT COUNT(*) FROM %s%s WHERE id = ? OR name = ?`, configs.Env.Prefix, models.TABLE_USER)
	if err := tx.QueryRow(query, strings.ToLower(strings.TrimSpace(email)), name).Scan(&exists); err != nil {
		return 0, err
	}
	if exists > 0 {
		return 0, ErrSignupConflict
	}
	query = fmt.Sprintf(`INSERT INTO %s%s
		(id, name, password, profile, level, point, signature, signup, signin, blocked)
		VALUES (?, ?, ?, '', 1, 100, '', ?, 0, 0)`, configs.Env.Prefix, models.TABLE_USER)
	result, err := tx.Exec(query, strings.ToLower(strings.TrimSpace(email)), name, hash, now)
	if err != nil {
		return 0, err
	}
	userUid, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	query = fmt.Sprintf(`UPDATE %s%s SET used = ? WHERE uid = ? AND used = 0 AND revoked = 0`, configs.Env.Prefix, models.TABLE_SIGNUP_INVITE)
	result, err = tx.Exec(query, now, inviteUid)
	if err != nil {
		return 0, err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return 0, ErrInviteInvalid
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint(userUid), nil
}
