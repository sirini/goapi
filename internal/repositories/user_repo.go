package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type UserRepository interface {
	GetReportResponse(userUid uint) string
	GetUserBlackList(userUid uint) []uint
	GetUserLevelPoint(userUid uint) (int, int)
	InsertBlackList(actionUserUid uint, targetUserUid uint)
	InsertReportUser(actionUserUid uint, targetUserUid uint, report string)
	InsertNewUser(id string, pw string, name string) uint
	InsertUserPermission(userUid uint, perm models.UserPermissionResult)
	InsertReportResponse(actionUserUid uint, targetUserUid uint, response string)
	IsEmailDuplicated(id string) bool
	IsNameDuplicated(name string) bool
	IsBlocked(userUid uint) bool
	IsBannedByTarget(actionUserUid uint, targetUserUid uint) bool
	IsPermissionAdded(userUid uint) bool
	IsUserReported(userUid uint) bool
	LoadUserPermission(userUid uint) models.UserPermissionResult
	UpdatePassword(userUid uint, password string)
	UpdatePointHistory(param models.UpdatePointParameter)
	UpdateUserInfoString(userUid uint, name string, signature string)
	UpdateUserProfile(userUid uint, imagePath string)
	UpdateUserPermission(userUid uint, perm models.UserPermissionResult)
	UpdateUserPoint(userUid uint, updatedPoint uint)
	UpdateUserBlocked(userUid uint, isBlocked bool)
	UpdateReportResponse(userUid uint, response string)
}

type TsboardUserRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardUserRepository(db *sql.DB) *TsboardUserRepository {
	return &TsboardUserRepository{db: db}
}

// 사용자 신고 내용에 대한 응답 가져오기
func (r *TsboardUserRepository) GetReportResponse(userUid uint) string {
	var response string
	query := fmt.Sprintf("SELECT response FROM %s%s WHERE to_uid = ? ORDER BY uid DESC LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return response
	}
	defer stmt.Close()

	stmt.QueryRow(userUid).Scan(&response)
	return response
}

// 사용자가 지정한 블랙 리스트 목록 가져오기
func (r *TsboardUserRepository) GetUserBlackList(userUid uint) []uint {
	var blocks []uint
	query := fmt.Sprintf("SELECT black_uid FROM %s%s WHERE user_uid = ?",
		configs.Env.Prefix, models.TABLE_USER_BLOCK)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return blocks
	}
	defer stmt.Close()

	rows, err := stmt.Query(userUid)
	if err != nil {
		return blocks
	}
	defer rows.Close()

	for rows.Next() {
		var block uint
		err := rows.Scan(&block)
		if err != nil {
			return blocks
		}
		blocks = append(blocks, block)
	}
	return blocks
}

// 사용자의 레벨과 보유 포인트 가져오기
func (r *TsboardUserRepository) GetUserLevelPoint(userUid uint) (int, int) {
	var level, point int
	query := fmt.Sprintf("SELECT level, point FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return 0, 0
	}
	defer stmt.Close()

	stmt.QueryRow(userUid).Scan(&level, &point)
	return level, point
}

// 다른 사용자를 내 블랙리스트에 등록하기
func (r *TsboardUserRepository) InsertBlackList(actionUserUid uint, targetUserUid uint) {
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE user_uid = ? AND black_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_BLOCK)
	stmtSelect, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmtSelect.Close()

	var uid uint
	err = stmtSelect.QueryRow(actionUserUid, targetUserUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %s%s (user_uid, black_uid) VALUES (?, ?)",
			configs.Env.Prefix, models.TABLE_USER_BLOCK)
		stmtInsert, err := r.db.Prepare(query)
		if err != nil {
			return
		}
		defer stmtInsert.Close()
		stmtInsert.Exec(actionUserUid, targetUserUid)
	}
}

// 다른 사용자를 신고하기
func (r *TsboardUserRepository) InsertReportUser(actionUserUid uint, targetUserUid uint, report string) {
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE to_uid = ? AND from_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)
	stmtSelect, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmtSelect.Close()

	var uid uint
	err = stmtSelect.QueryRow(targetUserUid, actionUserUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf(`INSERT INTO %s%s (to_uid, from_uid, request, response, timestamp, solved) 
												VALUES (?, ?, ?, ? ,? ,?)`, configs.Env.Prefix, models.TABLE_REPORT)
		stmtInsert, err := r.db.Prepare(query)
		if err != nil {
			return
		}
		defer stmtInsert.Close()
		stmtInsert.Exec(targetUserUid, actionUserUid, report, "", time.Now().UnixMilli(), 0)
	}
}

// 신규 회원 등록
func (r *TsboardUserRepository) InsertNewUser(id string, pw string, name string) uint {
	isDupId := r.IsEmailDuplicated(id)
	isDupName := r.IsNameDuplicated(name)
	if isDupId || isDupName {
		return models.FAILED
	}

	query := fmt.Sprintf(`INSERT INTO %s%s 
											(id, name, password, profile, level, point, signature, signup, signin, blocked)
											VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return models.FAILED
	}
	defer stmt.Close()

	result, err := stmt.Exec(id, name, pw, "", 1, 100, "", time.Now().UnixMilli(), 0, 0)
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
func (r *TsboardUserRepository) InsertUserPermission(userUid uint, perm models.UserPermissionResult) {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(user_uid, ACTION_WRITE_POST, ACTION_WRITE_COMMENT, ACTION_SEND_CHAT, ACTION_SEND_REPORT)
												VALUES (?, ?, ?, ? ,?)`, configs.Env.Prefix, models.TABLE_USER_PERM)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(userUid, perm.WritePost, perm.WriteComment, perm.SendChatMessage, perm.SendReport)
}

// 신고받은 사용자에게 조치 결과 추가하기
func (r *TsboardUserRepository) InsertReportResponse(actionUserUid uint, targetUserUid uint, response string) {
	query := fmt.Sprintf(`INSERT INTO %s%s (to_uid, from_uid, request, response, timestamp, solved) 
												VALUES (?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_REPORT)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(targetUserUid, actionUserUid, "", response, time.Now().UnixMilli(), 1)
}

// (회원가입 시) 이메일 주소가 중복되는지 확인
func (r *TsboardUserRepository) IsEmailDuplicated(id string) bool {
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	var uid uint
	stmt.QueryRow(id).Scan(&uid)
	return uid > 0
}

// (회원가입 시) 이름이 중복되는지 확인
func (r *TsboardUserRepository) IsNameDuplicated(name string) bool {
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	var uid uint
	stmt.QueryRow(name).Scan(&uid)
	return uid > 0
}

// 로그인이 차단되었는지 확인
func (r *TsboardUserRepository) IsBlocked(userUid uint) bool {
	var blocked uint8
	query := fmt.Sprintf("SELECT blocked FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	stmt.QueryRow(userUid).Scan(&blocked)
	return blocked > 0
}

// 상대방에게 차단되었는지 확인
func (r *TsboardUserRepository) IsBannedByTarget(actionUserUid uint, targetUserUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE user_uid = ? AND black_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_BLOCK)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	stmt.QueryRow(targetUserUid, actionUserUid).Scan(&uid)
	return uid > 0
}

// 사용자의 권한 정보가 등록된 게 있는지 확인
func (r *TsboardUserRepository) IsPermissionAdded(userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_PERM)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	stmt.QueryRow(userUid).Scan(&uid)
	return uid > 0
}

// 사용자가 받은 신고가 있는지 확인
func (r *TsboardUserRepository) IsUserReported(userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE to_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	stmt.QueryRow(userUid).Scan(&uid)
	return uid > 0
}

// 사용자 권한 및 신고 받은 후 조치사항 조회
func (r *TsboardUserRepository) LoadUserPermission(userUid uint) models.UserPermissionResult {
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
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return result
	}
	defer stmt.Close()

	err = stmt.QueryRow(userUid).Scan(&writePost, &writeComment, &sendChat, &sendReport)
	if err == sql.ErrNoRows {
		return result
	}

	result.WritePost = writePost > 0
	result.WriteComment = writeComment > 0
	result.SendChatMessage = sendChat > 0
	result.SendReport = sendReport > 0
	return result
}

// 사용자 비밀번호 변경하기
func (r *TsboardUserRepository) UpdatePassword(userUid uint, pw string) {
	query := fmt.Sprintf("UPDATE %s%s SET password = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(pw, userUid)
}

// 사용자의 포인트 변경 이력 업데이트
func (r *TsboardUserRepository) UpdatePointHistory(param models.UpdatePointParameter) {
	query := fmt.Sprintf(`INSERT INTO %s%s (user_uid, board_uid, action, point) 
												VALUES (?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_POINT_HISTORY)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(param.UserUid, param.BoardUid, param.Action.String(), param.Point)
}

// 사용자 이름, 서명 변경하기
func (r *TsboardUserRepository) UpdateUserInfoString(userUid uint, name string, signature string) {
	query := fmt.Sprintf("UPDATE %s%s SET name = ?, signature = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(name, signature, userUid)
}

// 사용자 프로필 이미지 변경하기
func (r *TsboardUserRepository) UpdateUserProfile(userUid uint, imagePath string) {
	query := fmt.Sprintf("UPDATE %s%s SET profile = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(imagePath, userUid)
}

// 사용자 권한 정보 변경하기
func (r *TsboardUserRepository) UpdateUserPermission(userUid uint, perm models.UserPermissionResult) {
	query := fmt.Sprintf(`UPDATE %s%s SET write_post = ?, write_comment = ?, send_chat = ?, send_report = ?
												WHERE user_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER_PERM)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(perm.WritePost, perm.WriteComment, perm.SendChatMessage, perm.SendReport, userUid)
}

// 사용자 포인트 변경하기
func (r *TsboardUserRepository) UpdateUserPoint(userUid uint, updatedPoint uint) {
	query := fmt.Sprintf("UPDATE %s%s SET point = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(updatedPoint, userUid)
}

// 사용자가 로그인 할 수 있는지 여부 업데이트하기
func (r *TsboardUserRepository) UpdateUserBlocked(userUid uint, isBlocked bool) {
	query := fmt.Sprintf("UPDATE %s%s SET blocked = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(isBlocked, userUid)
}

// 신고받은 사용자에게 조치 결과 업데이트 해주기
func (r *TsboardUserRepository) UpdateReportResponse(userUid uint, response string) {
	query := fmt.Sprintf("UPDATE %s%s SET response = ?, solved = ? WHERE to_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_REPORT)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	stmt.Exec(response, 1, userUid)
}
