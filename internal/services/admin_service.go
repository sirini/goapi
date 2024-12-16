package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type AdminService interface {
	AddBoardCategory(boardUid uint, name string) uint
	ChangeBoardAdmin(boardUid uint, newAdminUid uint) error
	ChangeBoardLevelPolicy(boardUid uint, level models.BoardActionLevel) error
	ChangeBoardPointPolicy(boardUid uint, point models.BoardActionPoint) error
	GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetBoardLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error)
	GetBoardPointPolicy(boardUid uint) (models.AdminBoardPointPolicy, error)
	GetDashboardItems(bunch uint) models.AdminDashboardItem
	GetDashboardLatests(bunch uint) models.AdminDashboardLatest
	RemoveBoardCategory(boardUid uint, catUid uint)
	UpdateBoardSetting(boardUid uint, column string, value string)
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

// 게시판 관리자 변경하기
func (s *TsboardAdminService) ChangeBoardAdmin(boardUid uint, newAdminUid uint) error {
	if isBlocked := s.repos.User.IsBlocked(newAdminUid); isBlocked {
		return fmt.Errorf("blocked user is not able to be an administrator")
	}
	return s.repos.Admin.UpdateBoardAdmin(boardUid, newAdminUid)
}

// 게시판 레벨 제한값 변경하기
func (s *TsboardAdminService) ChangeBoardLevelPolicy(boardUid uint, level models.BoardActionLevel) error {
	return s.repos.Admin.UpdateLevelPolicy(boardUid, level)
}

// 게시판 포인트 정책 변경하기
func (s *TsboardAdminService) ChangeBoardPointPolicy(boardUid uint, point models.BoardActionPoint) error {
	return s.repos.Admin.UpdatePointPolicy(boardUid, point)
}

// 게시판 관리자 후보 목록 가져오기
func (s *TsboardAdminService) GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error) {
	return s.repos.Admin.GetAdminCandidates(name, bunch)
}

// 게시판의 레벨 제한값 가져오기
func (s *TsboardAdminService) GetBoardLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error) {
	perm, err := s.repos.Admin.GetLevelPolicy(boardUid)
	if err != nil {
		return models.AdminBoardLevelPolicy{}, err
	}

	if perm.Admin.UserUid < 1 {
		return models.AdminBoardLevelPolicy{}, fmt.Errorf("unable to find an administrator's uid")
	}

	admin := s.repos.Board.GetWriterInfo(perm.Admin.UserUid)
	perm.Admin = admin
	return perm, nil
}

// 게시판 포인트 정책값 가져오기
func (s *TsboardAdminService) GetBoardPointPolicy(boardUid uint) (models.AdminBoardPointPolicy, error) {
	result := models.AdminBoardPointPolicy{}
	point, err := s.repos.Admin.GetPointPolicy(boardUid)
	if err != nil {
		return result, err
	}

	result.Uid = boardUid
	result.BoardActionPoint = point
	return result, nil
}

// 대시보드용 그룹, 게시판, 회원 목록 가져오기
func (s *TsboardAdminService) GetDashboardItems(bunch uint) models.AdminDashboardItem {
	groups := s.repos.Admin.GetGroupBoardList(models.TABLE_GROUP, bunch)
	boards := s.repos.Admin.GetGroupBoardList(models.TABLE_BOARD, bunch)
	members := s.repos.Admin.GetMemberList(bunch)

	result := models.AdminDashboardItem{
		Groups:  groups,
		Boards:  boards,
		Members: members,
	}
	return result
}

// 대시보드용 최근 (댓)글, 신고 목록 가져오기
func (s *TsboardAdminService) GetDashboardLatests(bunch uint) models.AdminDashboardLatest {
	//
	//
	// TODO
	//
	//
	return models.AdminDashboardLatest{}
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

// 게시판 설정 변경하기
func (s *TsboardAdminService) UpdateBoardSetting(boardUid uint, column string, value string) {
	s.repos.Admin.UpdateBoardSetting(boardUid, column, value)
}
