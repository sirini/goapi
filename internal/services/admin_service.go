package services

import (
	"fmt"
	"os"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type AdminService interface {
	AddBoardCategory(boardUid uint, name string) uint
	ChangeBoardAdmin(boardUid uint, newAdminUid uint) error
	ChangeBoardLevelPolicy(boardUid uint, level models.BoardActionLevel) error
	ChangeBoardPointPolicy(boardUid uint, point models.BoardActionPoint) error
	ChangeGroupAdmin(groupUid uint, newAdminUid uint) error
	ChangeGroupId(groupUid uint, newGroupId string) error
	CreateNewBoard(groupUid uint, newBoardId string) models.AdminCreateBoardResult
	CreateNewGroup(newGroupId string) models.AdminGroupItem
	GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetBoardLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error)
	GetBoardPointPolicy(boardUid uint) (models.AdminBoardPointPolicy, error)
	GetDashboardItems(bunch uint) models.AdminDashboardItem
	GetDashboardLatests(bunch uint) models.AdminDashboardLatest
	GetDashboardStatistics(bunch uint) models.AdminDashboardStatisticResult
	GetExistBoardIds(boardId string, bunch uint) []models.Triple
	GetExistGroupIds(groupId string, bunch uint) []models.Pair
	GetGroupConfig(groupId string) models.AdminGroupConfig
	GetGroupList() []models.AdminGroupItem
	RemoveBoardCategory(boardUid uint, catUid uint)
	RemoveBoard(boardUid uint) error
	RemoveGroup(groupUid uint) error
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
	return s.repos.Admin.UpdateGroupBoardAdmin(models.TABLE_BOARD, boardUid, newAdminUid)
}

// 게시판 레벨 제한값 변경하기
func (s *TsboardAdminService) ChangeBoardLevelPolicy(boardUid uint, level models.BoardActionLevel) error {
	return s.repos.Admin.UpdateLevelPolicy(boardUid, level)
}

// 게시판 포인트 정책 변경하기
func (s *TsboardAdminService) ChangeBoardPointPolicy(boardUid uint, point models.BoardActionPoint) error {
	return s.repos.Admin.UpdatePointPolicy(boardUid, point)
}

// 그룹 관리자 변경하기
func (s *TsboardAdminService) ChangeGroupAdmin(groupUid uint, newAdminUid uint) error {
	if isBlocked := s.repos.User.IsBlocked(newAdminUid); isBlocked {
		return fmt.Errorf("blocked user is not able to be an administrator")
	}
	return s.repos.Admin.UpdateGroupBoardAdmin(models.TABLE_GROUP, groupUid, newAdminUid)
}

// 그룹 ID 변경하기
func (s *TsboardAdminService) ChangeGroupId(groupUid uint, newGroupId string) error {
	uid, _ := s.repos.Admin.FindGroupUidAdminUidById(newGroupId)
	if uid > 0 {
		return fmt.Errorf("duplicated group id")
	}
	return s.repos.Admin.UpdateGroupId(groupUid, newGroupId)
}

// 새 게시판 만들기
func (s *TsboardAdminService) CreateNewBoard(groupUid uint, newBoardId string) models.AdminCreateBoardResult {
	result := models.AdminCreateBoardResult{}
	if isAdded := s.repos.Admin.IsAdded(models.TABLE_BOARD, newBoardId); isAdded {
		return result
	}

	boardUid := s.repos.Admin.CreateBoard(groupUid, newBoardId)
	if boardUid < 1 {
		return result
	}

	admin := s.repos.Admin.FindWriterByUid(models.CREATE_BOARD_ADMIN)
	result = models.AdminCreateBoardResult{
		Uid:  boardUid,
		Type: models.CREATE_BOARD_TYPE,
		Name: models.CREATE_BOARD_NAME,
		Info: models.CREATE_BOARD_INFO,
		Manager: models.Pair{
			Uid:  models.CREATE_BOARD_ADMIN,
			Name: admin.Name,
		},
	}

	defaultCats := []string{"free", "news", "qna", "etc"}
	s.repos.Admin.CreateDefaultCategories(boardUid, defaultCats)
	return result
}

// 새 그룹 만들기
func (s *TsboardAdminService) CreateNewGroup(newGroupId string) models.AdminGroupItem {
	groupUid := s.repos.Admin.CreateGroup(newGroupId)
	manager := s.repos.Admin.FindWriterByUid(models.CREATE_GROUP_ADMIN)
	result := models.AdminGroupItem{
		AdminGroupConfig: models.AdminGroupConfig{
			Uid:     groupUid,
			Count:   0,
			Manager: manager,
		},
		Id: newGroupId,
	}
	return result
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
	posts := s.repos.Admin.GetLatestPosts(bunch)
	comments := s.repos.Admin.GetLatestComments(bunch)
	reports := s.repos.Admin.GetLatestReports(bunch)
	result := models.AdminDashboardLatest{
		Posts:    posts,
		Comments: comments,
		Reports:  reports,
	}
	return result
}

// 대시보드용 최근 통계 가져오기
func (s *TsboardAdminService) GetDashboardStatistics(bunch uint) models.AdminDashboardStatisticResult {
	days := 7
	visit := s.repos.Admin.GetStatistic(models.TABLE_USER_ACCESS, models.COLUMN_TIMESTAMP, days)
	member := s.repos.Admin.GetStatistic(models.TABLE_USER, models.COLUMN_SIGNUP, days)
	post := s.repos.Admin.GetStatistic(models.TABLE_POST, models.COLUMN_SUBMITTED, days)
	reply := s.repos.Admin.GetStatistic(models.TABLE_COMMENT, models.COLUMN_SUBMITTED, days)
	file := s.repos.Admin.GetStatistic(models.TABLE_FILE, models.COLUMN_TIMESTAMP, days)
	image := s.repos.Admin.GetStatistic(models.TABLE_IMAGE, models.COLUMN_TIMESTAMP, days)
	result := models.AdminDashboardStatisticResult{
		Visit:  visit,
		Member: member,
		Post:   post,
		Reply:  reply,
		File:   file,
		Image:  image,
	}
	return result
}

// 게시판 아이디와 유사한 목록 가져오기
func (s *TsboardAdminService) GetExistBoardIds(boardId string, bunch uint) []models.Triple {
	return s.repos.Admin.FindBoardInfoById(boardId, bunch)
}

// 그룹 아이디와 유사한 목록 가져오기
func (s *TsboardAdminService) GetExistGroupIds(groupId string, bunch uint) []models.Pair {
	return s.repos.Admin.FindGroupUidIdById(groupId, bunch)
}

// 그룹 설정값 가져오기
func (s *TsboardAdminService) GetGroupConfig(groupId string) models.AdminGroupConfig {
	result := models.AdminGroupConfig{}
	groupUid, adminUid := s.repos.Admin.FindGroupUidAdminUidById(groupId)
	if groupUid < 1 || adminUid < 1 {
		return result
	}
	result.Uid = groupUid
	result.Manager = s.repos.Admin.FindWriterByUid(adminUid)
	result.Count = s.repos.Admin.GetTotalBoardCount(groupUid)
	return result
}

// 그룹 목록 가져오기
func (s *TsboardAdminService) GetGroupList() []models.AdminGroupItem {
	return s.repos.Admin.GetGroupList()
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

// 게시판 삭제하기
func (s *TsboardAdminService) RemoveBoard(boardUid uint) error {
	paths := s.repos.Admin.GetRemoveFilePaths(boardUid)
	for _, path := range paths {
		os.Remove("." + path)
	}

	err := s.repos.Admin.RemoveBoardCategories(boardUid)
	if err != nil {
		return err
	}
	err = s.repos.Admin.RemoveFileRecords(boardUid)
	if err != nil {
		return err
	}
	err = s.repos.Admin.UpdateStatusRemoved(models.TABLE_POST, boardUid)
	if err != nil {
		return err
	}
	err = s.repos.Admin.UpdateStatusRemoved(models.TABLE_COMMENT, boardUid)
	if err != nil {
		return err
	}
	return s.repos.Admin.RemoveBoard(boardUid)
}

// 그룹 삭제하기
func (s *TsboardAdminService) RemoveGroup(groupUid uint) error {
	groupCount := s.repos.Admin.GetTotalGroupCount()
	if groupCount < 2 {
		return fmt.Errorf("only one group is left, it cannot be removed")
	}
	defaultUid := s.repos.Admin.GetDefaultGroupUid(groupUid)
	err := s.repos.Admin.UpdateGroupUid(defaultUid, groupUid)
	if err != nil {
		return err
	}
	return s.repos.Admin.RemoveGroup(groupUid)
}

// 게시판 설정 변경하기
func (s *TsboardAdminService) UpdateBoardSetting(boardUid uint, column string, value string) {
	s.repos.Admin.UpdateBoardSetting(boardUid, column, value)
}
