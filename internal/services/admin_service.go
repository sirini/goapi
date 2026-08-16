package services

import (
	"fmt"
	"io/fs"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
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
	GetMailStatus() models.MailStatus
	GetMailDeliveries(param models.MailDeliveryListParam) (models.MailDeliveryListResult, error)
	GetMailCampaign(uid uint) (models.MailCampaign, error)
	GetMailCampaigns(limit uint) (models.MailCampaignListResult, error)
	PreviewMailCampaign(param models.MailCampaignPreviewParam) (models.MailCampaignPreviewResult, error)
	SaveMailCampaign(param models.MailCampaignSaveParam) (models.MailCampaign, error)
	SendMailCampaignTest(uid uint) error
	PrepareMailCampaign(uid uint) (models.MailCampaign, error)
	SendMailCampaign(uid uint) (models.MailCampaign, error)
	GetSearchedComments(param models.AdminLatestParam) []models.AdminLatestComment
	GetSearchedPosts(param models.AdminLatestParam) []models.AdminLatestPost
	GetSearchedReports(param models.AdminReportSearchParam) models.AdminReportListResult
	GetUserList(param models.AdminUserParam) models.AdminUserListResult
	GetUserInfo(userUid uint) models.AdminUserInfo
	GetSkinSettings() models.SkinSettings
	SetSkinSetting(param models.AdminSkinSettingParam) error
	ResolveReport(param models.AdminReportResolveParam) error
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
	mailer      utils.Mailer
	marketing   utils.MarketingMailer
}

// 리포지토리 묶음 주입받기
func NewNuboAdminService(repos *repositories.Repository, userService *NuboUserService) *NuboAdminService {
	mailer := utils.NewResendMailer()
	return newNuboAdminService(repos, userService, mailer, mailer)
}

func newNuboAdminService(repos *repositories.Repository, userService *NuboUserService, mailer utils.Mailer, marketing utils.MarketingMailer) *NuboAdminService {
	return &NuboAdminService{repos: repos, userService: userService, mailer: mailer, marketing: marketing}
}

func (s *NuboAdminService) GetMailStatus() models.MailStatus {
	return s.mailer.Status()
}

func (s *NuboAdminService) GetMailDeliveries(param models.MailDeliveryListParam) (models.MailDeliveryListResult, error) {
	since := uint64(time.Now().Add(-30 * 24 * time.Hour).UnixMilli())
	return s.repos.MailDelivery.ListDeliveries(param, since)
}

func (s *NuboAdminService) PreviewMailCampaign(param models.MailCampaignPreviewParam) (models.MailCampaignPreviewResult, error) {
	result := models.MailCampaignPreviewResult{}
	if err := validateMailCampaignContent(param.Subject, param.Markdown); err != nil {
		return result, err
	}
	html, text, err := templates.RenderMarketingMail(configs.Env.Title, siteURL(), strings.TrimSpace(param.Subject), param.Markdown, true)
	if err != nil {
		return result, err
	}
	result.HTML, result.Text = html, text
	return result, nil
}

func (s *NuboAdminService) SaveMailCampaign(param models.MailCampaignSaveParam) (models.MailCampaign, error) {
	if err := validateMailCampaignContent(param.Subject, param.Markdown); err != nil {
		return models.MailCampaign{}, err
	}
	param.Subject = strings.TrimSpace(param.Subject)
	if param.Uid == 0 {
		uid, err := s.repos.MailCampaign.CreateCampaign(param)
		if err != nil {
			return models.MailCampaign{}, err
		}
		param.Uid = uid
	} else if err := s.repos.MailCampaign.UpdateCampaign(param); err != nil {
		return models.MailCampaign{}, err
	}
	return s.repos.MailCampaign.GetCampaign(param.Uid)
}

func (s *NuboAdminService) GetMailCampaigns(limit uint) (models.MailCampaignListResult, error) {
	return s.repos.MailCampaign.ListCampaigns(limit)
}

func (s *NuboAdminService) GetMailCampaign(uid uint) (models.MailCampaign, error) {
	campaign, err := s.repos.MailCampaign.GetCampaign(uid)
	if err != nil {
		return campaign, err
	}
	if campaign.Status == models.MailCampaignSending {
		if campaign.ResendBroadcastId == "" {
			_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignReady, "발송 재개가 필요합니다")
			return s.repos.MailCampaign.GetCampaign(uid)
		}
		status, statusErr := s.marketing.GetBroadcastStatus(campaign.ResendBroadcastId)
		if statusErr != nil {
			return campaign, statusErr
		}
		switch status {
		case "draft":
			_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignReady, "발송 재개가 필요합니다")
		case "queued", "scheduled", "sent":
			_ = s.repos.MailCampaign.SetCampaignSent(uid, campaign.ResendBroadcastId)
		default:
			return campaign, fmt.Errorf("Resend returned an unknown broadcast status: %s", status)
		}
		return s.repos.MailCampaign.GetCampaign(uid)
	}
	if campaign.Status != models.MailCampaignSyncing || campaign.ResendImportId == "" {
		return campaign, nil
	}
	status, err := s.marketing.GetImportStatus(campaign.ResendImportId)
	if err != nil {
		return campaign, err
	}
	switch status.Status {
	case "completed":
		if status.Failed > 0 {
			errorText := fmt.Sprintf("Resend contact import failed for %d recipient(s)", status.Failed)
			_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignFailed, errorText)
		} else {
			_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignReady, "")
		}
	case "failed":
		_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignFailed, "Resend contact import failed")
	}
	return s.repos.MailCampaign.GetCampaign(uid)
}

func (s *NuboAdminService) SendMailCampaignTest(uid uint) error {
	if !s.mailer.Configured() {
		return ErrMailNotConfigured
	}
	campaign, err := s.repos.MailCampaign.GetCampaign(uid)
	if err != nil {
		return err
	}
	admin := s.repos.Admin.GetUserInfo(1)
	if _, err := mail.ParseAddress(admin.Id); err != nil {
		return fmt.Errorf("administrator email address is invalid")
	}
	html, text, err := templates.RenderMarketingMail(configs.Env.Title, siteURL(), campaign.Subject, campaign.Markdown, true)
	if err != nil {
		return err
	}
	_, err = s.mailer.Send(models.MailMessage{
		To: admin.Id, Subject: "[테스트] " + campaign.Subject, HTML: html, Text: text,
		IdempotencyKey: fmt.Sprintf("campaign-test/%d/%d", uid, time.Now().UnixNano()),
		Tags:           map[string]string{"type": "campaign-test"},
	})
	return err
}

func (s *NuboAdminService) PrepareMailCampaign(uid uint) (models.MailCampaign, error) {
	if !s.marketing.Configured() {
		return models.MailCampaign{}, ErrMailNotConfigured
	}
	campaign, err := s.repos.MailCampaign.GetCampaign(uid)
	if err != nil {
		return campaign, err
	}
	if campaign.Status != models.MailCampaignDraft && campaign.Status != models.MailCampaignFailed {
		return campaign, fmt.Errorf("campaign cannot be prepared in its current state")
	}
	recipients, err := s.repos.MailCampaign.GetActiveMailRecipients()
	if err != nil {
		return campaign, err
	}
	valid := make([]models.MailRecipient, 0, len(recipients))
	for _, recipient := range recipients {
		address, parseErr := mail.ParseAddress(strings.TrimSpace(recipient.Email))
		if parseErr == nil && address.Address == strings.TrimSpace(recipient.Email) {
			recipient.Email = address.Address
			valid = append(valid, recipient)
		}
	}
	if len(valid) == 0 {
		return campaign, fmt.Errorf("there are no active members with valid email addresses")
	}
	if len(valid) > 1000 {
		return campaign, fmt.Errorf("the Resend free marketing tier supports up to 1,000 contacts")
	}
	segmentName := fmt.Sprintf("NUBO %s campaign %d", strings.TrimPrefix(strings.TrimPrefix(siteURL(), "https://"), "http://"), uid)
	segmentId, err := s.marketing.CreateSegment(segmentName)
	if err != nil {
		return campaign, err
	}
	importId, err := s.marketing.ImportContacts(segmentId, valid)
	if err != nil {
		return campaign, err
	}
	if err := s.repos.MailCampaign.SetCampaignImport(uid, segmentId, importId, uint(len(valid))); err != nil {
		return campaign, err
	}
	return s.repos.MailCampaign.GetCampaign(uid)
}

func (s *NuboAdminService) SendMailCampaign(uid uint) (models.MailCampaign, error) {
	campaign, err := s.GetMailCampaign(uid)
	if err != nil {
		return campaign, err
	}
	if campaign.Status != models.MailCampaignReady {
		return campaign, fmt.Errorf("campaign recipients are not ready")
	}
	html, text, err := templates.RenderMarketingMail(configs.Env.Title, siteURL(), campaign.Subject, campaign.Markdown, false)
	if err != nil {
		return campaign, err
	}
	if err := s.repos.MailCampaign.BeginCampaignSend(uid); err != nil {
		return campaign, err
	}
	broadcastId := campaign.ResendBroadcastId
	if broadcastId == "" {
		broadcastId, err = s.marketing.CreateBroadcast(models.MarketingBroadcast{
			SegmentId: campaign.ResendSegmentId,
			Name:      fmt.Sprintf("NUBO campaign %d", campaign.Uid),
			Subject:   campaign.Subject, HTML: html, Text: text,
		})
		if err != nil {
			_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignReady, mailCampaignError(err))
			return campaign, err
		}
		if err := s.repos.MailCampaign.SetCampaignBroadcast(uid, broadcastId); err != nil {
			_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignReady, mailCampaignError(err))
			return campaign, err
		}
	}
	if err := s.marketing.SendBroadcast(broadcastId); err != nil {
		_ = s.repos.MailCampaign.SetCampaignStatus(uid, models.MailCampaignReady, mailCampaignError(err))
		return campaign, err
	}
	if err := s.repos.MailCampaign.SetCampaignSent(uid, broadcastId); err != nil {
		return campaign, err
	}
	return s.repos.MailCampaign.GetCampaign(uid)
}

func validateMailCampaignContent(subject, markdown string) error {
	if len([]rune(strings.TrimSpace(subject))) < 2 || len([]rune(strings.TrimSpace(subject))) > 200 {
		return fmt.Errorf("mail subject must be between 2 and 200 characters")
	}
	if len(strings.TrimSpace(markdown)) < 2 || len(markdown) > 200000 {
		return fmt.Errorf("mail content must be between 2 and 200,000 bytes")
	}
	return nil
}

func mailCampaignError(err error) string {
	message := err.Error()
	if len(message) > 1000 {
		return message[:1000]
	}
	return message
}

func (s *NuboAdminService) GetSkinSettings() models.SkinSettings {
	return s.repos.Admin.GetSkinSettings()
}

func (s *NuboAdminService) SetSkinSetting(param models.AdminSkinSettingParam) error {
	validType := map[string]bool{"layout": true, "home": true, "admin": true, "login": true, "profile": true, "privacy": true, "error": true}
	if !validType[param.Type] || len(param.SkinKey) < 3 || len(param.SkinKey) > 80 {
		return fmt.Errorf("invalid skin setting")
	}
	for _, char := range param.SkinKey {
		if !(char == '-' || char == '_' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9') {
			return fmt.Errorf("invalid skin key")
		}
	}
	return s.repos.Admin.SetSkinSetting(param)
}

func (s *NuboAdminService) ResolveReport(param models.AdminReportResolveParam) error {
	if param.ReportUid < 1 || len(strings.TrimSpace(param.Response)) < 2 {
		return fmt.Errorf("invalid report resolution")
	}
	return s.repos.Admin.ResolveReport(param)
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
	if param.Type > models.BOARD_TRADE {
		return 0, fmt.Errorf("invalid board type")
	}
	if param.Type == models.BOARD_TRADE && (param.SkinKey == "" || param.SkinKey == "nubo-basic-board") {
		param.SkinKey = "nubo-basic-trade"
	}
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
	} else if param.Type == models.BOARD_TRADE {
		cats = []string{"기타"}
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
func (s *NuboAdminService) GetSearchedReports(param models.AdminReportSearchParam) models.AdminReportListResult {
	return s.repos.Admin.GetReportList(param)
}

// (검색된) 사용자 목록 가져오기
func (s *NuboAdminService) GetUserList(param models.AdminUserParam) models.AdminUserListResult {
	item := s.repos.Admin.GetUserList(param)
	total := s.repos.Admin.GetTotalUserCount(param)

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
	if param.Type > models.BOARD_TRADE {
		return fmt.Errorf("invalid board type")
	}
	if param.Type == models.BOARD_TRADE && (param.SkinKey == "" || param.SkinKey == "nubo-basic-board") {
		param.SkinKey = "nubo-basic-trade"
	}
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
		return s.userService.ChangeUserProfile(param.UserUid, param.Profile, param.OldProfile)
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
	paths, err := s.repos.Admin.GetBoardRemovalPaths(boardUid)
	if err != nil {
		return err
	}
	if err := s.repos.Admin.RemoveBoardData(boardUid); err != nil {
		return err
	}
	for _, path := range paths {
		_ = os.Remove("." + path)
	}
	return nil
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
