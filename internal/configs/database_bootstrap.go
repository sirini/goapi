package configs

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var tablePrefixPattern = regexp.MustCompile(`^[A-Za-z0-9_]*$`)

var baseTableNames = []string{
	"user", "user_token", "user_permission", "user_verification", "signup_invite",
	"user_access_log", "user_black_list", "report", "chat", "group", "board",
	"skin_setting", "board_category", "point_history", "post", "hashtag",
	"post_hashtag", "post_like", "comment", "comment_like", "file",
	"file_thumbnail", "image", "notification", "exif", "image_description",
	"trade", "mail_campaign", "mail_delivery", "push_device",
}

// 지정한 이름의 DB가 없으면 utf8mb4 기본값으로 생성한다.
func EnsureDatabase(db *sql.DB, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("DB_NAME이 비어 있습니다")
	}
	quoted := "`" + strings.ReplaceAll(name, "`", "``") + "`"
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + quoted +
		" CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
	return err
}

// 기본 테이블과 최초 관리자 데이터를 만든 뒤 최신 스키마까지 반영한다.
// 이미 만들어진 레코드는 보존하므로 중단 후 다시 실행할 수 있다.
func BootstrapDatabase(db *sql.DB, prefix string, admin AdminInfo) error {
	if !tablePrefixPattern.MatchString(prefix) {
		return fmt.Errorf("DB_TABLE_PREFIX에는 영문, 숫자, 밑줄만 사용할 수 있습니다")
	}
	dbInfo := DBInfo{Prefix: prefix}
	createTables(db, dbInfo)
	if err := verifyBaseTables(db, prefix); err != nil {
		return err
	}
	if err := ensureInitialRows(db, prefix, admin); err != nil {
		return err
	}
	return InstallSchema(db, prefix)
}

// 필수 테이블이 모두 실제로 생성됐는지 확인해 무시된 DDL 오류를 드러낸다.
func verifyBaseTables(db *sql.DB, prefix string) error {
	missing := make([]string, 0)
	for _, suffix := range baseTableNames {
		name := prefix + suffix
		var count uint
		err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.TABLES
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`, name).Scan(&count)
		if err != nil {
			return fmt.Errorf("테이블 확인 실패 (%s): %w", name, err)
		}
		if count == 0 {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("필수 테이블 생성 실패: %s", strings.Join(missing, ", "))
	}
	return nil
}

// 관리자·그룹·기본 게시판을 각각 조회한 뒤 없는 항목만 추가한다.
func ensureInitialRows(db *sql.DB, prefix string, admin AdminInfo) error {
	adminUID, err := ensureAdminRow(db, prefix, admin)
	if err != nil {
		return err
	}
	groupUID, err := ensureGroupRow(db, prefix, adminUID)
	if err != nil {
		return err
	}
	freeUID, err := ensureBoardRow(db, prefix, "free", "free", "write everything you want", BOARD_DEFAULT, groupUID, adminUID)
	if err != nil {
		return err
	}
	if err := ensureCategoryRows(db, prefix, freeUID, []string{"open", "qna", "news"}); err != nil {
		return err
	}
	photoUID, err := ensureBoardRow(db, prefix, "photo", "gallery", "home of photographers", BOARD_GALLERY, groupUID, adminUID)
	if err != nil {
		return err
	}
	return ensureCategoryRows(db, prefix, photoUID, []string{"daily", "landscape", "portrait"})
}

// 같은 ID의 사용자가 없을 때만 최초 관리자 계정을 추가한다.
func ensureAdminRow(db *sql.DB, prefix string, admin AdminInfo) (uint64, error) {
	admin.Id = strings.TrimSpace(admin.Id)
	var uid uint64
	err := db.QueryRow(fmt.Sprintf("SELECT uid FROM %suser WHERE id = ? ORDER BY uid LIMIT 1", prefix), admin.Id).Scan(&uid)
	if err == nil {
		return uid, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	if admin.Id == "" || len(admin.Pw) < 8 {
		return 0, fmt.Errorf("최초 관리자 생성을 위해 ADMIN_ID와 8자 이상의 ADMIN_PW가 필요합니다")
	}
	digest := sha256.Sum256([]byte(admin.Pw))
	result, err := db.Exec(fmt.Sprintf(`INSERT INTO %suser
		(id, name, password, profile, level, point, signature, signup, signin, blocked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, prefix),
		admin.Id, "Admin", hex.EncodeToString(digest[:]), "", 9, 1000, "", time.Now().UnixMilli(), 0, 0)
	if err != nil {
		return 0, err
	}
	inserted, err := result.LastInsertId()
	return uint64(inserted), err
}

// 기존 그룹을 재사용하고, 그룹이 하나도 없을 때만 기본 그룹을 생성한다.
func ensureGroupRow(db *sql.DB, prefix string, adminUID uint64) (uint64, error) {
	var uid uint64
	err := db.QueryRow(fmt.Sprintf("SELECT uid FROM %sgroup WHERE id = 'boards' ORDER BY uid LIMIT 1", prefix)).Scan(&uid)
	if err == nil {
		return uid, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = db.QueryRow(fmt.Sprintf("SELECT uid FROM %sgroup ORDER BY uid LIMIT 1", prefix)).Scan(&uid)
	if err == nil {
		return uid, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	result, err := db.Exec(fmt.Sprintf("INSERT INTO %sgroup (id, admin_uid, timestamp) VALUES (?, ?, ?)", prefix),
		"boards", adminUID, time.Now().UnixMilli())
	if err != nil {
		return 0, err
	}
	inserted, err := result.LastInsertId()
	return uint64(inserted), err
}

// ID가 같은 기본 게시판이 없을 때만 생성한다.
func ensureBoardRow(db *sql.DB, prefix, id, name, info string, boardType int, groupUID, adminUID uint64) (uint64, error) {
	var uid uint64
	err := db.QueryRow(fmt.Sprintf("SELECT uid FROM %sboard WHERE id = ? ORDER BY uid LIMIT 1", prefix), id).Scan(&uid)
	if err == nil {
		return uid, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	query := fmt.Sprintf(`INSERT INTO %sboard
		(id, group_uid, admin_uid, type, name, info, row_count, width, use_category,
		level_list, level_view, level_write, level_comment, level_download,
		point_view, point_write, point_comment, point_download)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, prefix)
	result, err := db.Exec(query, id, groupUID, adminUID, boardType, name, info,
		CREATE_BOARD_ROWS, CREATE_BOARD_WIDTH, CREATE_BOARD_USE_CAT,
		CREATE_BOARD_LV_LIST, CREATE_BOARD_LV_VIEW, CREATE_BOARD_LV_WRITE,
		CREATE_BOARD_LV_COMMENT, CREATE_BOARD_LV_DOWNLOAD, CREATE_BOARD_PT_VIEW,
		CREATE_BOARD_PT_WRITE, CREATE_BOARD_PT_COMMENT, CREATE_BOARD_PT_DOWNLOAD)
	if err != nil {
		return 0, err
	}
	inserted, err := result.LastInsertId()
	return uint64(inserted), err
}

// 게시판별 기본 분류 중 아직 없는 이름만 추가한다.
func ensureCategoryRows(db *sql.DB, prefix string, boardUID uint64, names []string) error {
	for _, name := range names {
		var count uint
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %sboard_category WHERE board_uid = ? AND name = ?", prefix),
			boardUID, name).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			if _, err := db.Exec(fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix), boardUID, name); err != nil {
				return err
			}
		}
	}
	return nil
}
