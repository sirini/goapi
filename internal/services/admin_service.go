package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type AdminService interface {
	AddBoardCategory(boardUid uint, name string) uint
	ChangeBoardAdmin(boardUid uint, newAdminUid uint) error
	ChangeBoardLevelPolicy(boardUid uint, level models.BoardActionLevel) error
	ChangeBoardPointPolicy(boardUid uint, point models.BoardActionPoint) error
	ChangeGroupAdmin(groupUid uint, newAdminUid uint) error
	ChangeGroupId(param models.AdminGroupChangeParam) error
	CreateNewBoard(groupUid uint, newBoardId string) models.AdminCreateBoardResult
	CreateNewGroup(newGroupId string) models.AdminGroupConfig
	GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetBoardLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error)
	GetBoardList(groupUid uint) []models.AdminGroupBoardItem
	GetBoardPointPolicy(boardUid uint) (models.AdminBoardPointPolicy, error)
	GetDashboardUploadUsage(path string) uint64
	GetDashboardItems(bunch uint) models.AdminDashboardItem
	GetDashboardStatistics(bunch uint) models.AdminDashboardStatisticResult
	GetExistBoardIds(boardId string, bunch uint) []models.Triple
	GetExistGroupIds(groupId string, bunch uint) []models.Pair
	GetGroupConfig(groupId string) models.AdminGroupConfig
	GetGroupList() []models.AdminGroupConfig
	GetSearchedComments(param models.AdminLatestParam) []models.AdminLatestComment
	GetSearchedPosts(param models.AdminLatestParam) []models.AdminLatestPost
	GetSearchedReports(param models.AdminReportParam) []models.AdminReportItem
	GetUserList(param models.AdminUserParam) []models.AdminUserItem
	GetUserInfo(userUid uint) models.AdminUserInfo
	RemoveBoardCategory(boardUid uint, catUid uint) error
	RemoveBoard(boardUid uint) error
	RemoveComment(commentUid uint) error
	RemoveGroup(groupUid uint) error
	RemovePost(postUid uint) error
	UpdateBoardSetting(boardUid uint, column string, value string) error
	UpdateUserLevelPoint(userUid uint, level uint, point uint) error
}

type NuboAdminService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboAdminService(repos *repositories.Repository) *NuboAdminService {
	return &NuboAdminService{repos: repos}
}

// 카테고리 추가하기 (추가하면 카테고리를 사용하는 것으로 업데이트)
func (s *NuboAdminService) AddBoardCategory(boardUid uint, name string) uint {
	if isDup := s.repos.Admin.IsAddedCategory(boardUid, name); isDup {
		return models.FAILED
	}

	insertId := s.repos.Admin.InsertCategory(boardUid, name)
	s.repos.Admin.UpdateBoardSetting(boardUid, "use_category", "1")
	return insertId
}

// 게시판 관리자 변경하기
func (s *NuboAdminService) ChangeBoardAdmin(boardUid uint, newAdminUid uint) error {
	if isBlocked := s.repos.User.IsBlocked(newAdminUid); isBlocked {
		return fmt.Errorf("blocked user is not able to be an administrator")
	}
	return s.repos.Admin.UpdateGroupBoardAdmin(models.TABLE_BOARD, boardUid, newAdminUid)
}

// 게시판 레벨 제한값 변경하기
func (s *NuboAdminService) ChangeBoardLevelPolicy(boardUid uint, level models.BoardActionLevel) error {
	return s.repos.Admin.UpdateLevelPolicy(boardUid, level)
}

// 게시판 포인트 정책 변경하기
func (s *NuboAdminService) ChangeBoardPointPolicy(boardUid uint, point models.BoardActionPoint) error {
	return s.repos.Admin.UpdatePointPolicy(boardUid, point)
}

// 그룹 관리자 변경하기
func (s *NuboAdminService) ChangeGroupAdmin(groupUid uint, newAdminUid uint) error {
	if isBlocked := s.repos.User.IsBlocked(newAdminUid); isBlocked {
		return fmt.Errorf("blocked user is not able to be an administrator")
	}
	return s.repos.Admin.UpdateGroupBoardAdmin(models.TABLE_GROUP, groupUid, newAdminUid)
}

// 그룹 ID 변경하기
func (s *NuboAdminService) ChangeGroupId(param models.AdminGroupChangeParam) error {
	uid, _ := s.repos.Admin.FindGroupUidAdminUidById(param.NewGroupId)
	if uid > 0 {
		return fmt.Errorf("duplicated group id")
	}
	return s.repos.Admin.UpdateGroupId(param.GroupUid, param.NewGroupId)
}

// 새 게시판 만들기
func (s *NuboAdminService) CreateNewBoard(groupUid uint, newBoardId string) models.AdminCreateBoardResult {
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
func (s *NuboAdminService) CreateNewGroup(newGroupId string) models.AdminGroupConfig {
	groupUid := s.repos.Admin.CreateGroup(newGroupId)
	manager := s.repos.Admin.FindWriterByUid(models.CREATE_GROUP_ADMIN)
	result := models.AdminGroupConfig{
		Uid:     groupUid,
		Count:   0,
		Manager: manager,
		Id:      newGroupId,
	}
	return result
}

// 게시판 관리자 후보 목록 가져오기
func (s *NuboAdminService) GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error) {
	return s.repos.Admin.GetAdminCandidates(name, bunch)
}

// 게시판의 레벨 제한값 가져오기
func (s *NuboAdminService) GetBoardLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error) {
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

// 그룹 소속 게시판들의 목록(및 간단 통계) 가져오기
func (s *NuboAdminService) GetBoardList(groupUid uint) []models.AdminGroupBoardItem {
	return s.repos.Admin.GetBoardList(groupUid)
}

// 게시판 포인트 정책값 가져오기
func (s *NuboAdminService) GetBoardPointPolicy(boardUid uint) (models.AdminBoardPointPolicy, error) {
	result := models.AdminBoardPointPolicy{}
	point, err := s.repos.Admin.GetPointPolicy(boardUid)
	if err != nil {
		return result, err
	}

	result.Uid = boardUid
	result.BoardActionPoint = point
	return result, nil
}

// 첨부파일 총 용량 가져오기
func (s *NuboAdminService) GetDashboardUploadUsage(path string) uint64 {
	if !models.AdminUploadUsage.IsExpired() {
		size, _ := models.AdminUploadUsage.Get()
		return size
	}

	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return 0
	}

	var size uint64
	filepath.WalkDir(realPath, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += uint64(info.Size())
			}
		}
		return nil
	})
	models.AdminUploadUsage.Update(size)
	return size
}

// 대시보드용 그룹, 게시판, 회원 목록 가져오기
func (s *NuboAdminService) GetDashboardItems(bunch uint) models.AdminDashboardItem {
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

// 대시보드용 최근 통계 가져오기
func (s *NuboAdminService) GetDashboardStatistics(bunch uint) models.AdminDashboardStatisticResult {
	days := int(bunch)
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
func (s *NuboAdminService) GetExistBoardIds(boardId string, bunch uint) []models.Triple {
	return s.repos.Admin.FindBoardInfoById(boardId, bunch)
}

// 그룹 아이디와 유사한 목록 가져오기
func (s *NuboAdminService) GetExistGroupIds(groupId string, bunch uint) []models.Pair {
	return s.repos.Admin.FindGroupUidIdById(groupId, bunch)
}

// 그룹 설정값 가져오기
func (s *NuboAdminService) GetGroupConfig(groupId string) models.AdminGroupConfig {
	result := models.AdminGroupConfig{}
	groupUid, adminUid := s.repos.Admin.FindGroupUidAdminUidById(groupId)
	if groupUid < 1 || adminUid < 1 {
		return result
	}
	result.Uid = groupUid
	result.Id = groupId
	result.Manager = s.repos.Admin.FindWriterByUid(adminUid)
	result.Count = s.repos.Admin.GetTotalBoardCount(groupUid)
	return result
}

// 그룹 목록 가져오기
func (s *NuboAdminService) GetGroupList() []models.AdminGroupConfig {
	return s.repos.Admin.GetGroupList()
}

// 검색된 댓글들 가져오기
func (s *NuboAdminService) GetSearchedComments(param models.AdminLatestParam) []models.AdminLatestComment {
	return s.repos.Admin.GetCommentList(param)
}

// 검색된 게시글들 가져오기
func (s *NuboAdminService) GetSearchedPosts(param models.AdminLatestParam) []models.AdminLatestPost {
	return s.repos.Admin.GetPostList(param)
}

// 검색된 신고 목록 가져오기
func (s *NuboAdminService) GetSearchedReports(param models.AdminReportParam) []models.AdminReportItem {
	return s.repos.Admin.GetReportList(param)
}

// (검색된) 사용자 목록 가져오기
func (s *NuboAdminService) GetUserList(param models.AdminUserParam) []models.AdminUserItem {
	return s.repos.Admin.GetUserList(param)
}

// 사용자 정보 가져오기
func (s *NuboAdminService) GetUserInfo(userUid uint) models.AdminUserInfo {
	return s.repos.Admin.GetUserInfo(userUid)
}

// 카테고리 삭제하기
func (s *NuboAdminService) RemoveBoardCategory(boardUid uint, catUid uint) error {
	if isValid := s.repos.Admin.CheckCategoryInBoard(boardUid, catUid); !isValid {
		return fmt.Errorf("category is not belong to this board")
	}

	err := s.repos.Admin.RemoveCategory(boardUid, catUid)
	if err != nil {
		return err
	}
	defCatUid := s.repos.Admin.GetLowestCategoryUid(boardUid)
	return s.repos.Admin.UpdatePostCategory(boardUid, catUid, defCatUid)
}

// 게시판 삭제하기
func (s *NuboAdminService) RemoveBoard(boardUid uint) error {
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

// 댓글 삭제하기
func (s *NuboAdminService) RemoveComment(commentUid uint) error {
	return s.repos.Comment.RemoveComment(commentUid)
}

// 그룹 삭제하기
func (s *NuboAdminService) RemoveGroup(groupUid uint) error {
	groupCount := s.repos.Admin.GetTotalCount(models.TABLE_GROUP)
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

// 게시글 삭제하기
func (s *NuboAdminService) RemovePost(postUid uint) error {
	return s.repos.BoardView.RemovePost(postUid)
}

// 게시판 설정 변경하기
func (s *NuboAdminService) UpdateBoardSetting(boardUid uint, column string, value string) error {
	return s.repos.Admin.UpdateBoardSetting(boardUid, column, value)
}

// 사용자의 레벨, 포인트 정보 변경하기
func (s *NuboAdminService) UpdateUserLevelPoint(userUid uint, level uint, point uint) error {
	return s.repos.Admin.UpdateUserLevelPoint(userUid, level, point)
}
