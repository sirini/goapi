package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type AdminService interface {
	AddBoardCategory(boardUid uint, name string) uint
	UpdateBoardSetting(boardUid uint, column string, value string)
	RemoveBoardCategory(boardUid uint, catUid uint)
}

type TsboardAdminService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardAdminService(repos *repositories.Repository) *TsboardAdminService {
	return &TsboardAdminService{repos: repos}
}

// 카테고리 추가하기 (추가하면 카테고리를 사용하는 것으로 업데이트)
func (s *TsboardAdminService) AddBoardCategory(boardUid uint, name string) uint {
	if isDup := s.repos.Admin.IsAddedCategory(boardUid, name); isDup {
		return models.FAILED
	}

	insertId := s.repos.Admin.InsertCategory(boardUid, name)
	s.repos.Admin.UpdateBoardSetting(boardUid, "use_category", "1")
	return insertId
}

// 게시판 설정 변경하기
func (s *TsboardAdminService) UpdateBoardSetting(boardUid uint, column string, value string) {
	s.repos.Admin.UpdateBoardSetting(boardUid, column, value)
}

// 카테고리 삭제하기
func (s *TsboardAdminService) RemoveBoardCategory(boardUid uint, catUid uint) {
	if isValid := s.repos.Admin.CheckCategoryInBoard(boardUid, catUid); !isValid {
		return
	}

	s.repos.Admin.RemoveCategory(boardUid, catUid)
	defCatUid := s.repos.Admin.GetLowestCategoryUid(boardUid)
	s.repos.Admin.UpdatePostCategory(boardUid, catUid, defCatUid)
}
