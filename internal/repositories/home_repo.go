package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type HomeRepository interface {
	InsertVisitorLog(userUid uint)
	LoadBoardLinks(groupUid uint) ([]*models.HomeSidebarBoardResult, error)
	LoadGroupBoardLinks() ([]*models.HomeSidebarGroupResult, error)
}

type TsboardHomeRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardHomeRepository(db *sql.DB) *TsboardHomeRepository {
	return &TsboardHomeRepository{db: db}
}

// 방문자 기록하기
func (r *TsboardHomeRepository) InsertVisitorLog(userUid uint) {
	query := fmt.Sprintf("INSERT INTO %s%s (user_uid, timestamp) VALUES (?, ?)",
		configs.Env.Prefix, models.TABLE_USER_ACCESS)
	r.db.Exec(query, userUid, time.Now().UnixMilli())
}

// 게시판 목록들 가져오기
func (r *TsboardHomeRepository) LoadBoardLinks(groupUid uint) ([]*models.HomeSidebarBoardResult, error) {
	var boards []*models.HomeSidebarBoardResult

	query := fmt.Sprintf("SELECT id, type, name, info FROM %s%s WHERE group_uid = ?",
		configs.Env.Prefix, models.TABLE_BOARD)
	rows, err := r.db.Query(query, groupUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		board := &models.HomeSidebarBoardResult{}
		if err := rows.Scan(&board.Id, &board.Type, &board.Name, &board.Info); err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return boards, nil
}

// 그룹 및 하위 게시판 목록들 가져오기
func (r *TsboardHomeRepository) LoadGroupBoardLinks() ([]*models.HomeSidebarGroupResult, error) {
	var groups []*models.HomeSidebarGroupResult
	query := fmt.Sprintf("SELECT uid, id FROM %s%s", configs.Env.Prefix, models.TABLE_GROUP)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupUid uint
	var groupId string

	for rows.Next() {
		if err := rows.Scan(&groupUid, &groupId); err != nil {
			return nil, err
		}
		boards, err := r.LoadBoardLinks(groupUid)
		if err != nil {
			return nil, err
		}

		group := &models.HomeSidebarGroupResult{}
		group.Group = groupUid
		group.Boards = boards
		groups = append(groups, group)
	}
	return groups, nil
}
