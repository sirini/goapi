package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type AdminService interface {
	AddBoardCategory(boardUid uint, name string) uint
	ChangeGroupAdmin(groupUid uint, newAdminUid uint) error
	ChangeGroupId(param models.AdminGroupChangeParam) error
	CreateNewBoard(param models.AdminBoardCreateParam) (uint, error)
	CreateNewGroup(newGroupId string) (models.AdminGroupConfig, error)
	CreateNewUser(param models.AdminUserCreateParam) (uint, error)
	GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetBoardList(groupUid uint) ([]models.AdminGroupBoardItem, error)
	GetDashboardUploadUsage(path string) uint64
	GetDashboardItems(bunch uint) models.AdminDashboardItem
	GetDashboardStatistics(bunch uint) models.AdminDashboardStatisticResult
	GetExistBoardIds(boardId string, bunch uint) []models.Triple
	GetExistGroupIds(groupId string, bunch uint) []models.Pair
	GetGroupConfig(groupId string) models.AdminGroupConfig
	GetGroupList() []models.AdminGroupConfig
	GetSearchedComments(param models.AdminLatestParam) []models.AdminLatestComment
	GetSearchedPosts(param models.AdminLatestParam) []models.AdminLatestPost
	GetSearchedReports(param models.AdminReportSearchParam) []models.AdminReportItem
	GetUserList(param models.AdminUserParam) models.AdminUserListResult
	GetUserInfo(userUid uint) models.AdminUserInfo
	ModifyExistBoard(param models.AdminBoardModifyParam) error
	ModifyUserAccount(param models.AdminUserModifyParam) error
	RemoveBoardCategory(boardUid uint, catUid uint) error
	RemoveBoard(boardUid uint) error
	RemoveComment(commentUid uint) error
	RemoveGroup(groupUid uint) error
	RemovePost(postUid uint) error
	RemoveUser(userUid uint) error
}

type NuboAdminService struct {
	repos       *repositories.Repository
	userService *NuboUserService
}

// 리포지토리 묶음 주입받기
func NewNuboAdminService(repos *repositories.Repository, userService *NuboUserService) *NuboAdminService {
	return &NuboAdminService{repos: repos, userService: userService}
}

// 카테고리 추가하기
func (s *NuboAdminService) AddBoardCategory(boardUid uint, name string) uint {
	if isDup := s.repos.Admin.IsAddedCategory(boardUid, name); isDup {
		return models.FAILED
	}

	insertId := s.repos.Admin.InsertCategory(boardUid, name)
	return insertId
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
func (s *NuboAdminService) CreateNewBoard(param models.AdminBoardCreateParam) (uint, error) {
	if isAdded := s.repos.Admin.IsAdded(models.TABLE_BOARD, param.Id); isAdded {
		return 0, fmt.Errorf("already added")
	}

	boardUid := s.repos.Admin.CreateBoard(param)
	if boardUid < 1 {
		return 0, fmt.Errorf("failed to create a new board")
	}

	var cats []string
	if len(param.Categories) > 3 {
		cats = strings.Split(param.Categories, ",")
	} else {
		cats = []string{"qna", "news", "humor"}
	}
	s.repos.Admin.CreateCategories(boardUid, cats)
	return boardUid, nil
}

// 새 그룹 만들기
func (s *NuboAdminService) CreateNewGroup(newGroupId string) (models.AdminGroupConfig, error) {
	result := models.AdminGroupConfig{}
	if isAdded := s.repos.Admin.IsAdded(models.TABLE_GROUP, newGroupId); isAdded {
		return result, fmt.Errorf("already added")
	}

	groupUid := s.repos.Admin.CreateGroup(newGroupId)
	manager := s.repos.Admin.FindWriterByUid(models.CREATE_GROUP_ADMIN)
	result = models.AdminGroupConfig{
		Uid:     groupUid,
		Count:   0,
		Manager: manager,
		Id:      newGroupId,
	}
	return result, nil
}

// 새 사용자 계정 만들기
func (s *NuboAdminService) CreateNewUser(param models.AdminUserCreateParam) (uint, error) {
	if isDupId := s.repos.User.IsEmailDuplicated(param.Id); isDupId {
		return models.FAILED, fmt.Errorf("duplicated id")
	}
	if isDupName := s.repos.User.IsNameDuplicated(param.Name, 0); isDupName {
		return models.FAILED, fmt.Errorf("duplicated name")
	}
	newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(param.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.FAILED, err
	}
	param.Password = string(newBcryptHash)

	newUserUid := s.repos.Admin.CreateUser(param)
	if newUserUid < 2 {
		return models.FAILED, fmt.Errorf("failed to create an account for %s (%s)", param.Id, param.Name)
	}

	if param.Profile != nil {
		s.userService.ChangeUserProfile(newUserUid, param.Profile, "")
	}
	return newUserUid, nil
}

// 게시판 관리자 후보 목록 가져오기
func (s *NuboAdminService) GetBoardAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error) {
	return s.repos.Admin.GetAdminCandidates(name, bunch)
}

// 그룹 소속 게시판들의 목록(및 간단 통계) 가져오기
func (s *NuboAdminService) GetBoardList(groupUid uint) ([]models.AdminGroupBoardItem, error) {
	return s.repos.Admin.GetBoardList(groupUid)
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
func (s *NuboAdminService) GetSearchedReports(param models.AdminReportSearchParam) []models.AdminReportItem {
	return s.repos.Admin.GetReportList(param)
}

// (검색된) 사용자 목록 가져오기
func (s *NuboAdminService) GetUserList(param models.AdminUserParam) models.AdminUserListResult {
	item := s.repos.Admin.GetUserList(param)
	total := s.repos.Admin.GetTotalUserCount()

	return models.AdminUserListResult{
		Item:  item,
		Total: total,
	}
}

// 사용자 정보 가져오기
func (s *NuboAdminService) GetUserInfo(userUid uint) models.AdminUserInfo {
	return s.repos.Admin.GetUserInfo(userUid)
}

// 게시판 설정 수정하기
func (s *NuboAdminService) ModifyExistBoard(param models.AdminBoardModifyParam) error {
	boardUid := s.repos.Board.GetBoardUidById(param.Id)
	oldCats := s.repos.Admin.GetOldCategories(boardUid)

	// 이전에 쓴 분류명이 없어진 경우 삭제 처리
	for _, oldCat := range oldCats {
		if !strings.Contains(param.Categories, oldCat.Name) {
			if err := s.RemoveBoardCategory(boardUid, oldCat.Uid); err != nil {
				return err
			}
		}
	}

	// 새로 추가된 분류명이 생겼을 경우 추가 (중복은 무시)
	newCats := strings.Split(param.Categories, ",")
	for _, newCat := range newCats {
		s.AddBoardCategory(boardUid, newCat)
	}

	err := s.repos.Admin.ModifyBoard(param)
	return err
}

// 사용자 정보 수정하기
func (s *NuboAdminService) ModifyUserAccount(param models.AdminUserModifyParam) error {
	if isDupName := s.repos.User.IsNameDuplicated(param.Name, param.UserUid); isDupName {
		return fmt.Errorf("duplicated name")
	}

	param.Password = strings.TrimSpace(param.Password)
	if len(param.Password) > 0 {
		newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(param.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		param.Password = string(newBcryptHash)
	}

	if err := s.repos.Admin.ModifyUser(param); err != nil {
		return err
	}

	if param.Profile != nil {
		s.userService.ChangeUserProfile(param.UserUid, param.Profile, param.OldProfile)
	}
	return nil
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
	// 첨부파일, 썸네일 등 삭제
	attaches := s.repos.Admin.GetRemoveFilePaths(boardUid)
	for _, path := range attaches {
		os.Remove("." + path)
	}

	// 본문/댓글에 삽입한 이미지 파일 삭제
	images := s.repos.Admin.GetRemoveImagePaths(boardUid)
	for _, path := range images {
		os.Remove("." + path)
	}

	if err := s.repos.Admin.RemoveBoardCategories(boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveFileRecords(boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveImageRecords(boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemovePostHashtag(boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveLikeStatus(models.TABLE_COMMENT_LIKE, boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveLikeStatus(models.TABLE_POST_LIKE, boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveContentPermanently(models.TABLE_COMMENT, boardUid); err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveContentPermanently(models.TABLE_POST, boardUid); err != nil {
		return err
	}
	return s.repos.Admin.RemoveBoard(boardUid)
}

// 댓글 삭제하기
func (s *NuboAdminService) RemoveComment(commentUid uint) error {
	return s.repos.Comment.RemoveComment(commentUid)
}

// 그룹 삭제하기 (기본 그룹은 삭제 불가)
func (s *NuboAdminService) RemoveGroup(groupUid uint) error {
	DEFAULT_GROUP := uint(1)
	if groupUid == DEFAULT_GROUP {
		return fmt.Errorf("default group is not able to remove")
	}
	if err := s.repos.Admin.UpdateGroupUid(DEFAULT_GROUP, groupUid); err != nil {
		return err
	}
	return s.repos.Admin.RemoveGroup(groupUid)
}

// 게시글 삭제하기
func (s *NuboAdminService) RemovePost(postUid uint) error {
	return s.repos.BoardView.RemovePost(postUid)
}

// 사용자 삭제하기
func (s *NuboAdminService) RemoveUser(userUid uint) error {
	return s.repos.Admin.RemoveUser(userUid)
}
