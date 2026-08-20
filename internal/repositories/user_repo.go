package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	ApplyPointChange(param models.UpdatePointParam) error
	DeleteAccount(userUid uint) ([]string, error)
	GetReportResponse(userUid uint) string
	GetUserBlackList(userUid uint) []uint
	GetUserLevelPoint(userUid uint) (int, int)
	InsertBlackList(actionUserUid uint, targetUserUid uint) error
	RemoveBlackList(actionUserUid uint, targetUserUid uint) error
	InsertReportUser(param models.UserReportParam) error
	InsertNewUser(id string, pw string, name string) uint
	InsertUserPermission(userUid uint, perm models.UserPermissionResult) error
	InsertReportResponse(actionUserUid uint, targetUserUid uint, response string) error
	IsEmailDuplicated(id string) bool
	IsNameDuplicated(name string, userUid uint) bool
	IsBlocked(userUid uint) bool
	IsBannedByTarget(actionUserUid uint, targetUserUid uint) bool
	IsPermissionAdded(userUid uint) bool
	IsReported(actionUserUid uint, targetUserUid uint) bool
	IsUserReported(userUid uint) bool
	LoadUserPermission(userUid uint) models.UserPermissionResult
	UpdateUserInfoString(userUid uint, name string, signature string) error
	UpdateUserProfile(userUid uint, imagePath string) error
	UpdateUserPermission(userUid uint, perm models.UserPermissionResult) error
	UpdateUserBlocked(userUid uint, isBlocked bool) error
	UpdateReportResponse(userUid uint, response string) error
}

var ErrInsufficientPoint = errors.New("not enough point")

type NuboUserRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboUserRepository(db *sql.DB) *NuboUserRepository {
	return &NuboUserRepository{db: db}
}

// 사용자 신고 내용에 대한 응답 가져오기
func (r *NuboUserRepository) GetReportResponse(userUid uint) string {
	var response string
	query := fmt.Sprintf("SELECT response FROM %s%s WHERE to_uid = ? ORDER BY uid DESC LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)
	r.db.QueryRow(query, userUid).Scan(&response)
	return response
}

// 사용자가 지정한 블랙 리스트 목록 가져오기
func (r *NuboUserRepository) GetUserBlackList(userUid uint) []uint {
	items := make([]uint, 0)
	query := fmt.Sprintf("SELECT black_uid FROM %s%s WHERE user_uid = ?",
		configs.Env.Prefix, models.TABLE_USER_BLOCK)
	rows, err := r.db.Query(query, userUid)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		var block uint
		err := rows.Scan(&block)
		if err != nil {
			return items
		}
		items = append(items, block)
	}
	if err := rows.Err(); err != nil {
		return items
	}

	return items
}

// 사용자의 레벨과 보유 포인트 가져오기
func (r *NuboUserRepository) GetUserLevelPoint(userUid uint) (int, int) {
	var level, point int
	query := fmt.Sprintf("SELECT level, point FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	r.db.QueryRow(query, userUid).Scan(&level, &point)
	return level, point
}

// 다른 사용자를 내 블랙리스트에 등록하기
func (r *NuboUserRepository) InsertBlackList(actionUserUid uint, targetUserUid uint) error {
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE user_uid = ? AND black_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_BLOCK)

	var uid uint
	err := r.db.QueryRow(query, actionUserUid, targetUserUid).Scan(&uid)
	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %s%s (user_uid, black_uid) VALUES (?, ?)",
			configs.Env.Prefix, models.TABLE_USER_BLOCK)
		_, err = r.db.Exec(query, actionUserUid, targetUserUid)
	}
	return err
}

// 사용자를 내 차단 목록에서 제거한다.
func (r *NuboUserRepository) RemoveBlackList(actionUserUid uint, targetUserUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE user_uid = ? AND black_uid = ?", configs.Env.Prefix, models.TABLE_USER_BLOCK)
	_, err := r.db.Exec(query, actionUserUid, targetUserUid)
	return err
}

// 작성 콘텐츠는 보존하되 인증·개인정보·기기 연결을 원자적으로 제거한다.
func (r *NuboUserRepository) DeleteAccount(userUid uint) ([]string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var email string
	if err := tx.QueryRow(
		fmt.Sprintf("SELECT id FROM %suser WHERE uid = ? LIMIT 1", configs.Env.Prefix),
		userUid,
	).Scan(&email); err != nil {
		return nil, err
	}
	paths, err := collectAccountFilePaths(tx, userUid)
	if err != nil {
		return nil, err
	}
	postIDs := fmt.Sprintf("SELECT uid FROM %spost WHERE user_uid = ?", configs.Env.Prefix)
	commentIDs := fmt.Sprintf(
		"SELECT uid FROM %scomment WHERE user_uid = ? OR post_uid IN (%s)",
		configs.Env.Prefix,
		postIDs,
	)
	statements := []struct {
		query string
		args  []any
	}{
		{fmt.Sprintf("DELETE FROM %strade WHERE post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %scomment_like WHERE user_uid = ? OR comment_uid IN (%s)", configs.Env.Prefix, commentIDs), []any{userUid, userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %snotification WHERE to_uid = ? OR from_uid = ? OR post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid, userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %spost_like WHERE user_uid = ? OR post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %scomment WHERE uid IN (%s)", configs.Env.Prefix, commentIDs), []any{userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %sexif WHERE post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %simage_description WHERE post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %sfile_thumbnail WHERE post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %sfile WHERE post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %spost_hashtag WHERE post_uid IN (%s)", configs.Env.Prefix, postIDs), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %spoint_history WHERE user_uid = ?", configs.Env.Prefix), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %spost WHERE user_uid = ?", configs.Env.Prefix), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %simage WHERE user_uid = ?", configs.Env.Prefix), []any{userUid}},
		{fmt.Sprintf("DELETE FROM %schat WHERE to_uid = ? OR from_uid = ?", configs.Env.Prefix), []any{userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %sreport WHERE to_uid = ? OR from_uid = ?", configs.Env.Prefix), []any{userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %suser_black_list WHERE user_uid = ? OR black_uid = ?", configs.Env.Prefix), []any{userUid, userUid}},
		{fmt.Sprintf("DELETE FROM %smail_delivery WHERE recipient = ?", configs.Env.Prefix), []any{email}},
		{fmt.Sprintf("DELETE FROM %suser_verification WHERE email = ?", configs.Env.Prefix), []any{email}},
		{fmt.Sprintf("DELETE FROM %ssignup_invite WHERE email = ?", configs.Env.Prefix), []any{email}},
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement.query, statement.args...); err != nil {
			return nil, err
		}
	}
	for _, table := range []string{"push_device", "user_token", "user_permission", "user_access_log"} {
		query := fmt.Sprintf("DELETE FROM %s%s WHERE user_uid = ?", configs.Env.Prefix, table)
		if _, err := tx.Exec(query, userUid); err != nil {
			return nil, err
		}
	}
	query := fmt.Sprintf(`UPDATE %s%s SET id = '', name = '탈퇴한 사용자', password = '', profile = '',
		level = 0, point = 0, signature = '', signup = 0, signin = 0, blocked = 1 WHERE uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_USER)
	if _, err := tx.Exec(query, userUid); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return paths, nil
}

func collectAccountFilePaths(tx *sql.Tx, userUid uint) ([]string, error) {
	query := fmt.Sprintf(`SELECT f.path, COALESCE(ft.path, ''), COALESCE(ft.full_path, '')
		FROM %sfile f JOIN %spost p ON p.uid = f.post_uid
		LEFT JOIN %sfile_thumbnail ft ON ft.file_uid = f.uid WHERE p.user_uid = ?
		UNION ALL SELECT path, '', '' FROM %simage WHERE user_uid = ?`,
		configs.Env.Prefix, configs.Env.Prefix, configs.Env.Prefix, configs.Env.Prefix)
	rows, err := tx.Query(query, userUid, userUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paths := make([]string, 0)
	for rows.Next() {
		var values [3]string
		if err := rows.Scan(&values[0], &values[1], &values[2]); err != nil {
			return nil, err
		}
		for _, path := range values {
			if path != "" {
				paths = append(paths, path)
			}
		}
	}
	return paths, rows.Err()
}

// 다른 사용자를 신고하기
func (r *NuboUserRepository) InsertReportUser(param models.UserReportParam) error {
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE to_uid = ? AND from_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)

	var uid uint
	err := r.db.QueryRow(query, param.TargetUserUid, param.ActionUserUid).Scan(&uid)
	if err == sql.ErrNoRows {
		query = fmt.Sprintf(`INSERT INTO %s%s (to_uid, from_uid, request, response, timestamp, solved) 
												VALUES (?, ?, ?, ? ,? ,?)`, configs.Env.Prefix, models.TABLE_REPORT)
		r.db.Exec(query, param.TargetUserUid, param.ActionUserUid, param.Content, "", time.Now().UnixMilli(), 0)
	}
	return nil
}

// 신규 회원 등록
func (r *NuboUserRepository) InsertNewUser(id string, pw string, name string) uint {
	isDupId := r.IsEmailDuplicated(id)
	isDupName := r.IsNameDuplicated(name, 0)
	if isDupId || isDupName {
		return models.FAILED
	}
	newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return models.FAILED
	}
	query := fmt.Sprintf(`INSERT INTO %s%s 
											(id, name, password, profile, level, point, signature, signup, signin, blocked)
											VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_USER)
	result, err := r.db.Exec(query, id, name, newBcryptHash, "", 1, 100, "", time.Now().UnixMilli(), 0, 0)
	if err != nil {
		return models.FAILED
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 사용자 권한 설정값 추가하기
func (r *NuboUserRepository) InsertUserPermission(userUid uint, perm models.UserPermissionResult) error {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(user_uid, write_post, write_comment, send_chat, send_report)
												VALUES (?, ?, ?, ? ,?)`, configs.Env.Prefix, models.TABLE_USER_PERM)
	_, err := r.db.Exec(query, userUid, perm.WritePost, perm.WriteComment, perm.SendChatMessage, perm.SendReport)
	return err
}

// 신고받은 사용자에게 조치 결과 추가하기
func (r *NuboUserRepository) InsertReportResponse(actionUserUid uint, targetUserUid uint, response string) error {
	query := fmt.Sprintf(`INSERT INTO %s%s (to_uid, from_uid, request, response, timestamp, solved) 
												VALUES (?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_REPORT)
	_, err := r.db.Exec(query, targetUserUid, actionUserUid, "", response, time.Now().UnixMilli(), 1)
	return err
}

// (회원가입 시) 이메일 주소가 중복되는지 확인
func (r *NuboUserRepository) IsEmailDuplicated(id string) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE id = ?)",
		configs.Env.Prefix, models.TABLE_USER)
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// (회원가입 시) 이름이 중복되는지 확인
func (r *NuboUserRepository) IsNameDuplicated(name string, userUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE name = ? AND uid != ?)",
		configs.Env.Prefix, models.TABLE_USER)

	err := r.db.QueryRow(query, name, userUid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// 로그인이 차단되었는지 확인
func (r *NuboUserRepository) IsBlocked(userUid uint) bool {
	var blocked bool
	query := fmt.Sprintf("SELECT blocked FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	err := r.db.QueryRow(query, userUid).Scan(&blocked)
	if err != nil {
		return false
	}
	return blocked
}

// 상대방에게 차단되었는지 확인
func (r *NuboUserRepository) IsBannedByTarget(actionUserUid uint, targetUserUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE user_uid = ? AND black_uid = ?)",
		configs.Env.Prefix, models.TABLE_USER_BLOCK)
	err := r.db.QueryRow(query, targetUserUid, actionUserUid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// 사용자의 권한 정보가 등록된 게 있는지 확인
func (r *NuboUserRepository) IsPermissionAdded(userUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE user_uid = ?)", configs.Env.Prefix, models.TABLE_USER_PERM)
	err := r.db.QueryRow(query, userUid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// 현재 사용자가 대상 사용자를 이미 신고했는지 확인
func (r *NuboUserRepository) IsReported(actionUserUid uint, targetUserUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE to_uid = ? AND from_uid = ?)", configs.Env.Prefix, models.TABLE_REPORT)
	err := r.db.QueryRow(query, targetUserUid, actionUserUid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// 사용자가 받은 신고가 있는지 확인
func (r *NuboUserRepository) IsUserReported(userUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE to_uid = ?)", configs.Env.Prefix, models.TABLE_REPORT)
	err := r.db.QueryRow(query, userUid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// 사용자 권한 및 신고 받은 후 조치사항 조회
func (r *NuboUserRepository) LoadUserPermission(userUid uint) models.UserPermissionResult {
	result := models.UserPermissionResult{
		WritePost:       true,
		WriteComment:    true,
		SendChatMessage: true,
		SendReport:      true,
	}

	var writePost, writeComment, sendChat, sendReport uint8
	query := fmt.Sprintf(`SELECT write_post, write_comment, send_chat, send_report 
												FROM %s%s WHERE user_uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_USER_PERM)

	err := r.db.QueryRow(query, userUid).Scan(&writePost, &writeComment, &sendChat, &sendReport)
	if err == sql.ErrNoRows {
		return result
	}

	result.WritePost = writePost > 0
	result.WriteComment = writeComment > 0
	result.SendChatMessage = sendChat > 0
	result.SendReport = sendReport > 0
	return result
}

// 포인트 잔액과 변경 이력을 하나의 트랜잭션에서 원자적으로 반영한다.
func (r *NuboUserRepository) ApplyPointChange(param models.UpdatePointParam) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := applyPointChangeTx(tx, param); err != nil {
		return err
	}
	return tx.Commit()
}

func applyPointChangeTx(tx *sql.Tx, param models.UpdatePointParam) error {
	if param.UserUid < 1 || param.Point == 0 {
		return nil
	}
	requiredPoint := 0
	if param.Point < 0 {
		requiredPoint = -param.Point
	}
	query := fmt.Sprintf(`UPDATE %s%s SET point = point + ?
		WHERE uid = ? AND point >= ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER)
	result, err := tx.Exec(query, param.Point, param.UserUid, requiredPoint)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrInsufficientPoint
	}

	query = fmt.Sprintf(`INSERT INTO %s%s (user_uid, board_uid, action, point)
		VALUES (?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_POINT_HISTORY)
	_, err = tx.Exec(query, param.UserUid, param.BoardUid, uint(param.Action), param.Point)
	return err
}

// 사용자 이름, 서명 변경하기
func (r *NuboUserRepository) UpdateUserInfoString(userUid uint, name string, signature string) error {
	query := fmt.Sprintf("UPDATE %s%s SET name = ?, signature = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	_, err := r.db.Exec(query, name, signature, userUid)
	return err
}

// 사용자 프로필 이미지 변경하기
func (r *NuboUserRepository) UpdateUserProfile(userUid uint, imagePath string) error {
	query := fmt.Sprintf("UPDATE %s%s SET profile = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	_, err := r.db.Exec(query, imagePath, userUid)
	return err
}

// 사용자 권한 정보 변경하기
func (r *NuboUserRepository) UpdateUserPermission(userUid uint, perm models.UserPermissionResult) error {
	query := fmt.Sprintf(`UPDATE %s%s SET write_post = ?, write_comment = ?, send_chat = ?, send_report = ?
												WHERE user_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER_PERM)
	_, err := r.db.Exec(query, perm.WritePost, perm.WriteComment, perm.SendChatMessage, perm.SendReport, userUid)
	return err
}

// 사용자가 로그인 할 수 있는지 여부 업데이트하기
func (r *NuboUserRepository) UpdateUserBlocked(userUid uint, isBlocked bool) error {
	query := fmt.Sprintf("UPDATE %s%s SET blocked = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	_, err := r.db.Exec(query, isBlocked, userUid)
	return err
}

// 신고받은 사용자에게 조치 결과 업데이트 해주기
func (r *NuboUserRepository) UpdateReportResponse(userUid uint, response string) error {
	query := fmt.Sprintf("UPDATE %s%s SET response = ?, solved = ? WHERE to_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)
	_, err := r.db.Exec(query, response, 1, userUid)
	return err
}
