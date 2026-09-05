package services

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"sync"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

func containsImage(files []*multipart.FileHeader) bool {
	for _, file := range files {
		if file != nil && utils.IsImage(file.Filename) {
			return true
		}
	}
	return false
}

type BoardService interface {
	Download(boardUid uint, fileUid uint, userUid uint) (models.BoardViewDownloadResult, error)
	GetOriginalImage(boardUid uint, fileUid uint, userUid uint) (models.BoardOriginalImageResult, error)
	GetBoardConfig(boardUid uint) models.BoardConfig
	GetBoardList(boardUid uint, userUid uint) ([]models.BoardItem, error)
	GetBoardUid(id string) uint
	GetEditorConfig(boardUid uint, userUid uint) models.EditorConfigResult
	GetInsertedImages(param models.EditorInsertImageParam) (models.EditorInsertImageResult, error)
	GetLatestUserContents(userUid uint, limit uint) models.BoardWriterLatestContent
	GetListItem(param models.BoardListParam) (models.BoardListResult, error)
	GetMaxUid() uint
	GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error)
	GetStudio(param models.BoardStudioParam) (models.BoardStudioResult, error)
	GetSuggestionTags(input string, bunch uint) []models.EditorTagItem
	GetSuggestionTitles(input string, bunch uint) []string
	GetThumbnailImage(fileUid uint, userUid uint) (string, error)
	GetViewItem(param models.BoardViewParam) (models.BoardViewResult, error)
	LikeThisPost(param models.BoardViewLikeParam) error
	LoadPost(boardUid uint, postUid uint, userUid uint) (models.EditorLoadPostResult, error)
	ModifyPost(param models.EditorModifyParam) error
	MovePost(param models.BoardMovePostParam) error
	RemoveAttachedFile(param models.EditorRemoveAttachedParam) error
	RemoveInsertedImage(imageUid uint, userUid uint)
	RemovePost(boardUid uint, postUid uint, userUid uint) error
	SaveAttachments(param models.EditorSaveAttachedParam) error
	SaveTags(boardUid uint, postUid uint, tags []string) error
	SaveThumbnail(fileUid uint, postUid uint, path string) models.BoardThumbnail
	UploadInsertImage(boardUid uint, userUid uint, images []*multipart.FileHeader) ([]string, error)
	WritePost(param models.EditorWriteParam) (uint, error)
}

// 원본 이미지는 첨부 다운로드 권한이 아니라 게시물 보기 권한을 따른다.
// 실제 저장 경로는 핸들러 밖으로 노출하지 않고 토큰 기반 스트리밍에만 사용한다.
func (s *NuboBoardService) GetOriginalImage(boardUid uint, fileUid uint, userUid uint) (models.BoardOriginalImageResult, error) {
	result := models.BoardOriginalImageResult{}
	if !s.repos.BoardView.IsFileInBoard(fileUid, boardUid) {
		return result, fmt.Errorf("file does not belong to this board")
	}
	postUid := s.repos.BoardView.GetFilePostUid(fileUid, boardUid)
	status := s.repos.Comment.GetPostStatus(postUid)
	if postUid < 1 || status == models.CONTENT_REMOVED {
		return result, fmt.Errorf("image is not available")
	}
	if s.repos.BoardView.CheckBannedByWriter(postUid, userUid) {
		return result, fmt.Errorf("you have been blocked by writer")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_VIEW)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	if status == models.CONTENT_SECRET {
		isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
		isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
		if !isAdmin && !isWriter {
			return result, fmt.Errorf("you have no permission to view this image")
		}
	}

	file := s.repos.BoardView.GetDownloadInfo(fileUid)
	if !utils.IsImage(file.Path) {
		return result, fmt.Errorf("file is not an image")
	}
	if utils.GetFileSize(file.Path) < 1 {
		return result, fmt.Errorf("image not found")
	}
	result.Path = file.Path
	return result, nil
}

type NuboBoardService struct {
	repos                  *repositories.Repository
	notifications          *notificationPublisher
	imageDescriptionConfig configs.ImageDescriptionConfig
	imageDescriptionSlots  chan struct{}
	describeImage          func(context.Context, string, string) (utils.ImageDescriptionResult, error)
}

// 리포지토리 묶음 주입받기
func NewNuboBoardService(repos *repositories.Repository) *NuboBoardService {
	return newNuboBoardService(repos, configs.GetImageDescriptionConfig(), utils.AskImageDescription)
}

func newNuboBoardService(
	repos *repositories.Repository,
	config configs.ImageDescriptionConfig,
	describeImage func(context.Context, string, string) (utils.ImageDescriptionResult, error),
) *NuboBoardService {
	return &NuboBoardService{
		repos:                  repos,
		notifications:          newNotificationPublisher(repos, disabledPushSender{}),
		imageDescriptionConfig: config,
		imageDescriptionSlots:  make(chan struct{}, config.MaxConcurrent),
		describeImage:          describeImage,
	}
}

func (s *NuboBoardService) imageDescriptionCandidates(files []*multipart.FileHeader) map[*multipart.FileHeader]struct{} {
	candidates := make(map[*multipart.FileHeader]struct{})
	if !s.imageDescriptionConfig.Enabled || s.imageDescriptionConfig.MaxPerPost < 1 {
		return candidates
	}
	for _, file := range files {
		if utils.IsImage(file.Filename) {
			candidates[file] = struct{}{}
			if len(candidates) == s.imageDescriptionConfig.MaxPerPost {
				break
			}
		}
	}
	return candidates
}

func (s *NuboBoardService) requestImageDescription(ctx context.Context, path string) (utils.ImageDescriptionResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case s.imageDescriptionSlots <- struct{}{}:
		defer func() { <-s.imageDescriptionSlots }()
	case <-ctx.Done():
		return utils.ImageDescriptionResult{}, ctx.Err()
	}
	return s.describeImage(ctx, path, s.imageDescriptionConfig.Model)
}

// 다운로드에 필요한 정보 반환
func (s *NuboBoardService) Download(boardUid uint, fileUid uint, userUid uint) (models.BoardViewDownloadResult, error) {
	var result models.BoardViewDownloadResult
	if !s.repos.BoardView.IsFileInBoard(fileUid, boardUid) {
		return result, fmt.Errorf("file does not belong to this board")
	}
	postUid := s.repos.BoardView.GetFilePostUid(fileUid, boardUid)
	status := s.repos.Comment.GetPostStatus(postUid)
	if postUid < 1 || status == models.CONTENT_REMOVED {
		return result, fmt.Errorf("file is not available")
	}
	if status == models.CONTENT_SECRET {
		isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
		isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
		if !isAdmin && !isWriter {
			return result, fmt.Errorf("you have no permission to download this file")
		}
	}
	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_DOWNLOAD)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	result = s.repos.BoardView.GetDownloadInfo(fileUid)
	fileSize := utils.GetFileSize(result.Path)
	if fileSize < 1 {
		return result, fmt.Errorf("file not found")
	}

	if err := applyPointChange(s.repos.User, models.UpdatePointParam{
		UserUid:  userUid,
		BoardUid: boardUid,
		Action:   models.POINT_ACTION_DOWNLOAD,
		Point:    needPt,
	}); err != nil {
		return result, err
	}
	return result, nil
}

// 게시판 고유 번호 가져오기
func (s *NuboBoardService) GetBoardUid(id string) uint {
	return s.repos.Board.GetBoardUidById(id)
}

// 게시글 최대 고유번호 반환
func (s *NuboBoardService) GetMaxUid() uint {
	return s.repos.Board.GetMaxUid(models.TABLE_POST)
}

// 게시판 설정값 가져오기
func (s *NuboBoardService) GetBoardConfig(boardUid uint) models.BoardConfig {
	return s.repos.Board.GetBoardConfig(boardUid)
}

// 게시글 이동할 대상 게시판 목록 가져오기
func (s *NuboBoardService) GetBoardList(boardUid uint, userUid uint) ([]models.BoardItem, error) {
	if !s.repos.Auth.CheckPermissionByUid(userUid, boardUid) {
		return nil, fmt.Errorf("unauthorized access")
	}
	if s.repos.Board.GetBoardConfig(boardUid).Type == models.BOARD_TRADE {
		return nil, fmt.Errorf("trade posts cannot be moved through the generic board endpoint")
	}

	targets := make([]models.BoardItem, 0)
	for _, board := range s.repos.BoardView.GetAllBoards() {
		if board.Uid == boardUid || board.Type == models.BOARD_TRADE {
			continue
		}
		if s.repos.Auth.CheckPermissionByUid(userUid, board.Uid) {
			targets = append(targets, board)
		}
	}
	return targets, nil
}

// 게시판 설정 및 카테고리, 관리자 여부 반환
func (s *NuboBoardService) GetEditorConfig(boardUid uint, userUid uint) models.EditorConfigResult {
	return models.EditorConfigResult{
		Config:     s.repos.Board.GetBoardConfig(boardUid),
		IsAdmin:    s.repos.Auth.CheckPermissionByUid(userUid, boardUid),
		Categories: s.repos.Board.GetBoardCategories(boardUid),
	}
}

// 게시글에 내가 삽입한 이미지 목록들 가져오기
func (s *NuboBoardService) GetInsertedImages(param models.EditorInsertImageParam) (models.EditorInsertImageResult, error) {
	result := models.EditorInsertImageResult{}
	userLv, _ := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, _ := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}

	images, err := s.repos.BoardEdit.GetInsertedImages(param)
	if err != nil {
		return result, err
	}

	maxImageUid, err := s.repos.BoardEdit.GetMaxImageUid(param.BoardUid, param.UserUid)
	if err != nil {
		return result, err
	}
	totalImageCount, err := s.repos.BoardEdit.GetTotalImageCount(param.BoardUid, param.UserUid)
	if err != nil {
		return result, err
	}
	result = models.EditorInsertImageResult{
		Images:          images,
		MaxImageUid:     maxImageUid,
		TotalImageCount: totalImageCount,
	}
	return result, nil
}

// 사용자의 최근 활동(글, 댓글) 가져오기
func (s *NuboBoardService) GetLatestUserContents(userUid uint, limit uint) models.BoardWriterLatestContent {
	posts, _ := s.repos.BoardView.GetWriterLatestPost(userUid, limit)
	comments, _ := s.repos.BoardView.GetWriterLatestComment(userUid, limit)
	return models.BoardWriterLatestContent{
		Posts:    posts,
		Comments: comments,
	}
}

// 게시판 목록글들 가져오기
func (s *NuboBoardService) GetListItem(param models.BoardListParam) (models.BoardListResult, error) {
	posts := make([]models.BoardListItem, 0)
	var err error

	result := models.BoardListResult{}
	notices, err := s.repos.Board.GetNoticePosts(param.BoardUid, param.UserUid)
	if err != nil {
		return result, err
	}

	param.NoticeCount = uint(len(notices))
	totalPostCount := s.repos.Board.GetTotalCount(param)
	posts, err = s.repos.Board.FindPosts(param)
	if err != nil {
		return result, err
	}
	s.attachFeaturedBadges(notices, posts)

	result = models.BoardListResult{
		TotalPostCount: totalPostCount,
		Config:         s.repos.Board.GetBoardConfig(param.BoardUid),
		Notices:        notices,
		Posts:          posts,
		BlackList:      s.repos.User.GetUserBlackList(param.UserUid),
		IsAdmin:        s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid),
	}
	return result, nil
}

// 최근 사용된 해시태그 가져오기
func (s *NuboBoardService) GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error) {
	return s.repos.Board.GetRecentTags(boardUid, limit)
}

// 사용자와 게시판으로 범위를 명시한 작품 스튜디오를 반환한다.
func (s *NuboBoardService) GetStudio(param models.BoardStudioParam) (models.BoardStudioResult, error) {
	return s.repos.Board.GetStudio(param)
}

// 유사 제목들 가져오기
func (s *NuboBoardService) GetSuggestionTitles(input string, bunch uint) []string {
	titles, _ := s.repos.BoardEdit.GetSuggestionTitles(input, bunch)
	return titles
}

// 추천할 태그 목록들 가져오기
func (s *NuboBoardService) GetSuggestionTags(input string, bunch uint) []models.EditorTagItem {
	tags, _ := s.repos.BoardEdit.GetSuggestionTags(input, bunch)
	return tags
}

// 글 수정 화면에서 기존에 첨부한 이미지의 썸네일 가져오기

func (s *NuboBoardService) GetThumbnailImage(fileUid uint, userUid uint) (string, error) {
	boardUid, postUid := s.repos.BoardView.GetFileOwnership(fileUid)
	if boardUid < 1 || postUid < 1 {
		return "", fmt.Errorf("file not found")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
	isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
	if !isAdmin && !isWriter {
		return "", fmt.Errorf("you have no permission to preview this file")
	}
	return s.repos.BoardEdit.FindAttachedThumbnailImageByUid(fileUid)
}

// 게시글 가져오기
func (s *NuboBoardService) GetViewItem(param models.BoardViewParam) (models.BoardViewResult, error) {
	result := models.BoardViewResult{}
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return result, fmt.Errorf("post does not belong to this board")
	}
	if isBanned := s.repos.BoardView.CheckBannedByWriter(param.PostUid, param.UserUid); isBanned {
		return result, fmt.Errorf("you have been blocked by writer")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_VIEW)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	post, err := s.repos.BoardView.GetPostItem(param.PostUid, param.UserUid)
	if err != nil {
		return result, err
	}
	postItems := []models.BoardListItem{post}
	s.attachFeaturedBadges(postItems)
	post = postItems[0]

	config := s.repos.Board.GetBoardConfig(param.BoardUid)
	result.Config = config
	result.IsAdmin = s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	result.Post = post
	result.Files = make([]models.BoardAttachment, 0)
	result.Images = make([]models.BoardAttachedImage, 0)

	if config.Level.Download <= userLv {
		files, err := s.repos.BoardView.GetAttachments(param.PostUid)
		if err != nil {
			return result, err
		}
		result.Files = files
	}

	images, err := s.repos.BoardView.GetAttachedImages(param.PostUid)
	if err != nil {
		return result, err
	}
	result.Images = images

	if param.NeedUpdateHit {
		s.repos.BoardView.UpdatePostHit(param.PostUid)
	}

	if post.Status == models.CONTENT_SECRET {
		isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
		isWriter := post.Writer.UserUid == param.UserUid

		if !isAdmin && !isWriter {
			result.Post.Title = "A Secret Post"
			result.Post.Content = "Unauthorized access: secret post"
			result.Files = make([]models.BoardAttachment, 0)
			result.Images = make([]models.BoardAttachedImage, 0)
		}
	}

	result.Tags = s.repos.BoardView.GetTags(param.PostUid)
	result.PrevPostUid = s.repos.BoardView.GetPrevPostUid(param.BoardUid, param.PostUid)
	result.NextPostUid = s.repos.BoardView.GetNextPostUid(param.BoardUid, param.PostUid)
	result.WriterPosts, _ = s.repos.BoardView.GetWriterLatestPost(post.Writer.UserUid, param.LatestLimit)
	result.WriterComments, _ = s.repos.BoardView.GetWriterLatestComment(post.Writer.UserUid, param.LatestLimit)
	if err := applyPointChange(s.repos.User, models.UpdatePointParam{
		UserUid:  param.UserUid,
		BoardUid: param.BoardUid,
		Action:   models.POINT_ACTION_VIEW,
		Point:    needPt,
	}); err != nil {
		return models.BoardViewResult{}, err
	}
	return result, nil
}

// 글 작성자에게 차단당했는지 확인
func (s *NuboBoardService) IsBannedByWriter(postUid uint, viewerUid uint) bool {
	return s.repos.BoardView.CheckBannedByWriter(postUid, viewerUid)
}

// 게시글에 좋아요 클릭
func (s *NuboBoardService) LikeThisPost(param models.BoardViewLikeParam) error {
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return fmt.Errorf("post does not belong to this board")
	}
	if isLiked := s.repos.BoardView.IsLikedPost(param.PostUid, param.UserUid); isLiked {
		s.repos.BoardView.UpdateLikePost(param)
	} else {
		s.repos.BoardView.InsertLikePost(param)
	}
	if param.Liked {
		targetUserUid := s.repos.Comment.GetPostWriterUid(param.PostUid)
		s.notifications.Save(models.InsertNotificationParam{
			ActionUserUid: param.UserUid,
			TargetUserUid: targetUserUid,
			NotiType:      models.NOTI_LIKE_POST,
			PostUid:       param.PostUid,
		}, true)
	}
	return nil
}

// 게시글 수정 시 기존 정보들 가져오기
func (s *NuboBoardService) LoadPost(boardUid uint, postUid uint, userUid uint) (models.EditorLoadPostResult, error) {
	result := models.EditorLoadPostResult{}
	if !s.repos.BoardView.IsPostInBoard(postUid, boardUid) {
		return result, fmt.Errorf("post does not belong to this board")
	}
	post, err := s.repos.BoardView.GetPostItem(postUid, userUid)
	if err != nil {
		return result, err
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
	if !isAdmin && !isAuthor {
		return result, fmt.Errorf("you have no permission to edit this post")
	}

	files, err := s.repos.BoardView.GetAttachments(postUid)
	if err != nil {
		return result, err
	}
	tags := s.repos.BoardView.GetTags(postUid)

	result.Post = post
	result.Files = files
	result.Tags = tags
	return result, nil
}

// 게시글 이동하기
func (s *NuboBoardService) MovePost(param models.BoardMovePostParam) error {
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return fmt.Errorf("post does not belong to this board")
	}
	if param.TargetBoardUid == param.BoardUid {
		return nil
	}
	sourceConfig := s.repos.Board.GetBoardConfig(param.BoardUid)
	targetConfig := s.repos.Board.GetBoardConfig(param.TargetBoardUid)
	if sourceConfig.Type == models.BOARD_TRADE || targetConfig.Type == models.BOARD_TRADE {
		return fmt.Errorf("trade posts cannot be moved through the generic board endpoint")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	if !isAdmin {
		return fmt.Errorf("unauthorized access")
	}
	if !s.repos.BoardView.BoardExists(param.TargetBoardUid) {
		return fmt.Errorf("target board does not exist")
	}
	if !s.repos.Auth.CheckPermissionByUid(param.UserUid, param.TargetBoardUid) {
		return fmt.Errorf("you have no permission to move posts into the target board")
	}
	return s.repos.BoardView.MovePost(param.TargetBoardUid, param.PostUid)
}

// 게시글 수정하기
func (s *NuboBoardService) ModifyPost(param models.EditorModifyParam) error {
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return fmt.Errorf("post does not belong to this board")
	}
	if s.repos.Board.GetBoardConfig(param.BoardUid).Type == models.BOARD_TRADE {
		return fmt.Errorf("trade posts must be modified through the trade endpoint")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("only the author can edit this post")
	}

	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return fmt.Errorf("you have no permission to edit post")
	}

	if param.IsNotice {
		if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
			param.IsNotice = false
		}
	}
	s.repos.BoardView.RemovePostTags(param.PostUid)
	err := s.repos.BoardEdit.UpdatePost(param)
	if err != nil {
		return err
	}

	err = s.SaveTags(param.BoardUid, param.PostUid, param.Tags)
	if err != nil {
		return err
	}

	return s.SaveAttachments(models.EditorSaveAttachedParam{
		Context:  param.Context,
		BoardUid: param.BoardUid,
		PostUid:  param.PostUid,
		Files:    param.Files,
	})
}

// 게시글 수정 시 첨부했던 파일 삭제하기
func (s *NuboBoardService) RemoveAttachedFile(param models.EditorRemoveAttachedParam) error {
	if !s.repos.BoardView.IsFileInPost(param.FileUid, param.PostUid, param.BoardUid) {
		return fmt.Errorf("file does not belong to this post")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("you have no permission to remove this file")
	}

	filePath, err := s.repos.BoardEdit.FindAttachedPathByUid(param.FileUid)
	if err != nil {
		return err
	}
	removes := s.repos.BoardView.RemoveAttachedFile(param.FileUid, filePath)

	for _, target := range removes {
		_ = utils.RemoveUploadFile(target)
	}
	return nil
}

// 게시글에 삽입한 이미지 삭제하기
func (s *NuboBoardService) RemoveInsertedImage(imageUid uint, userUid uint) {
	removePath, err := s.repos.BoardEdit.RemoveInsertedImage(imageUid, userUid)
	if err != nil {
		return
	}
	if len(removePath) > 0 {
		_ = utils.RemoveUploadFile(removePath)
	}
}

// 게시글 삭제하기
func (s *NuboBoardService) RemovePost(boardUid uint, postUid uint, userUid uint) error {
	if !s.repos.BoardView.IsPostInBoard(postUid, boardUid) {
		return fmt.Errorf("post does not belong to this board")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("you have no permission to remove this post")
	}

	if err := s.repos.BoardView.RemovePost(postUid); err != nil {
		return err
	}
	s.repos.BoardView.RemoveComments(postUid)
	s.repos.BoardView.RemovePostTags(postUid)
	removes := s.repos.BoardView.RemoveAttachments(postUid)

	for _, path := range removes {
		_ = utils.RemoveUploadFile(path)
	}
	return nil
}

// 첨부파일들을 저장하기
func (s *NuboBoardService) SaveAttachments(param models.EditorSaveAttachedParam) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	type savedAttachment struct {
		fileUid   uint
		filePath  string
		extraPath []string
	}
	var saved []savedAttachment
	descriptionCandidates := s.imageDescriptionCandidates(param.Files)

	for _, file := range param.Files {
		wg.Add(1)

		go func(f *multipart.FileHeader) {
			defer wg.Done()

			savedPath, err := utils.SaveAttachmentFile(f)
			if err != nil {
				if savedPath != "" {
					_ = os.Remove(savedPath)
				}
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()

				return
			}
			publicSavedPath, err := utils.PublicUploadPath(savedPath)
			if err != nil {
				_ = os.Remove(savedPath)
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			fileUid, err := s.repos.BoardEdit.InsertFile(models.EditorSaveFileParam{
				BoardUid: param.BoardUid,
				PostUid:  param.PostUid,
				Name:     utils.CutString(f.Filename, 100),
				Path:     publicSavedPath,
			})
			if err != nil {
				_ = os.Remove(savedPath)
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()

				return
			}

			if utils.IsImage(f.Filename) {
				thumb, err := utils.SaveThumbnailImage(savedPath)
				if err != nil {
					mu.Lock()
					saved = append(saved, savedAttachment{fileUid: fileUid, filePath: publicSavedPath})
					errors = append(errors, err)
					mu.Unlock()

					return
				}

				publicLarge, largePathErr := utils.PublicUploadPath(thumb.Large)
				publicSmall, smallPathErr := utils.PublicUploadPath(thumb.Small)
				if largePathErr != nil || smallPathErr != nil {
					if largePathErr != nil {
						err = largePathErr
					} else {
						err = smallPathErr
					}
					mu.Lock()
					saved = append(saved, savedAttachment{
						fileUid: fileUid, filePath: publicSavedPath,
						extraPath: []string{thumb.Small, thumb.Large},
					})
					errors = append(errors, err)
					mu.Unlock()
					return
				}

				if err := s.repos.BoardEdit.InsertFileThumbnail(models.EditorSaveThumbnailParam{
					BoardThumbnail: models.BoardThumbnail{
						Large: publicLarge,
						Small: publicSmall,
					},
					FileUid: fileUid,
					PostUid: param.PostUid,
				}); err != nil {
					mu.Lock()
					saved = append(saved, savedAttachment{
						fileUid: fileUid, filePath: publicSavedPath,
						extraPath: []string{thumb.Small, thumb.Large},
					})
					errors = append(errors, err)
					mu.Unlock()
					return
				}
				exif := utils.ExtractExif(savedPath)
				if err := s.repos.BoardEdit.InsertExif(fileUid, param.PostUid, exif); err != nil {
					mu.Lock()
					saved = append(saved, savedAttachment{
						fileUid: fileUid, filePath: publicSavedPath,
						extraPath: []string{thumb.Small, thumb.Large},
					})
					errors = append(errors, err)
					mu.Unlock()
					return
				}

				if _, shouldDescribe := descriptionCandidates[f]; shouldDescribe {
					result, descriptionErr := s.requestImageDescription(param.Context, thumb.Small)
					if descriptionErr != nil {
						log.Printf("ai: image description failed post_uid=%d file_uid=%d model=%s: %v", param.PostUid, fileUid, s.imageDescriptionConfig.Model, descriptionErr)
					} else if insertErr := s.repos.BoardEdit.InsertImageDescription(fileUid, param.PostUid, result.Description); insertErr != nil {
						log.Printf("ai: image description storage failed post_uid=%d file_uid=%d model=%s: %v", param.PostUid, fileUid, result.Model, insertErr)
					} else {
						log.Printf("ai: image description generated post_uid=%d file_uid=%d model=%s input_tokens=%d output_tokens=%d", param.PostUid, fileUid, result.Model, result.InputTokens, result.OutputTokens)
					}
				}

				mu.Lock()
				saved = append(saved, savedAttachment{
					fileUid: fileUid, filePath: publicSavedPath,
					extraPath: []string{thumb.Small, thumb.Large},
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			saved = append(saved, savedAttachment{fileUid: fileUid, filePath: publicSavedPath})
			mu.Unlock()
		}(file)
	}
	wg.Wait()

	if len(errors) > 0 {
		for _, attachment := range saved {
			for _, path := range s.repos.BoardView.RemoveAttachedFile(attachment.fileUid, attachment.filePath) {
				_ = utils.RemoveUploadFile(path)
			}
			for _, path := range attachment.extraPath {
				_ = os.Remove(path)
			}
		}
		return errors[0]
	}

	return nil
}

// 해시태그들 저장하기
func (s *NuboBoardService) SaveTags(boardUid uint, postUid uint, tags []string) error {
	for _, tag := range tags {
		tidyTag := utils.Purify(tag)
		if len(tidyTag) < 2 {
			continue
		}

		var hashtagUid uint
		hashtagUid = s.repos.BoardEdit.FindTagUidByName(tag)
		if hashtagUid > 0 {
			err := s.repos.BoardEdit.UpdateTag(hashtagUid)
			if err != nil {
				return err
			}
		} else {
			uid, err := s.repos.BoardEdit.InsertTag(boardUid, postUid, tag)
			if err != nil {
				return err
			}
			hashtagUid = uid
		}

		err := s.repos.BoardEdit.InsertPostHashtag(boardUid, postUid, hashtagUid)
		if err != nil {
			return err
		}
	}
	return nil
}

// 썸네일 이미지 생성 및 저장하기
func (s *NuboBoardService) SaveThumbnail(fileUid uint, postUid uint, path string) models.BoardThumbnail {
	thumb, err := utils.SaveThumbnailImage(path)
	if err != nil {
		return thumb
	}
	publicSmall, smallPathErr := utils.PublicUploadPath(thumb.Small)
	publicLarge, largePathErr := utils.PublicUploadPath(thumb.Large)
	if smallPathErr != nil || largePathErr != nil {
		return models.BoardThumbnail{}
	}
	s.repos.BoardEdit.InsertFileThumbnail(models.EditorSaveThumbnailParam{
		BoardThumbnail: models.BoardThumbnail{
			Small: publicSmall,
			Large: publicLarge,
		},
		FileUid: fileUid,
		PostUid: postUid,
	})
	return thumb
}

// 게시글에 삽입할 이미지 파일 업로드 처리하기
func (s *NuboBoardService) UploadInsertImage(boardUid uint, userUid uint, images []*multipart.FileHeader) ([]string, error) {
	imagePaths := make([]string, 0)
	if hasPerm := s.repos.Auth.CheckPermissionForAction(userUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return imagePaths, fmt.Errorf("you have no permission to write a new post")
	}

	hasPerm, err := s.repos.BoardEdit.CheckWriterForBlog(boardUid, userUid)
	if err != nil {
		return imagePaths, err
	}
	if !hasPerm {
		return imagePaths, fmt.Errorf("only blog owner can write a new post")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return imagePaths, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return imagePaths, fmt.Errorf("not enough point")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	tempPaths := make([]string, 0)
	errors := make([]error, 0)

	for _, header := range images {
		wg.Add(1)

		go func(h *multipart.FileHeader) {
			defer wg.Done()

			file, err := h.Open()
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			defer file.Close()

			tempPath, err := utils.SaveUploadedFile(file, h.Filename)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}

			imagePath, err := utils.SaveInsertImage(tempPath)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()

				os.Remove(tempPath)
				return
			}

			mu.Lock()
			tempPaths = append(tempPaths, tempPath)
			publicImagePath, pathErr := utils.PublicUploadPath(imagePath)
			if pathErr != nil {
				errors = append(errors, pathErr)
				_ = os.Remove(imagePath)
				mu.Unlock()
				return
			}
			imagePaths = append(imagePaths, publicImagePath)
			mu.Unlock()

		}(header)
	}

	wg.Wait()

	if len(errors) > 0 {
		for _, tempPath := range tempPaths {
			_ = os.Remove(tempPath)
		}
		for _, imagePath := range imagePaths {
			_ = utils.RemoveUploadFile(imagePath)
		}
		return nil, errors[0]
	}

	if err := s.repos.BoardEdit.InsertImagePaths(boardUid, userUid, imagePaths); err != nil {
		for _, imagePath := range imagePaths {
			_ = utils.RemoveUploadFile(imagePath)
		}
		return nil, err
	}

	for _, tempPath := range tempPaths {
		_ = os.Remove(tempPath)
	}
	return imagePaths, nil
}

// 새 게시글 작성하기
func (s *NuboBoardService) WritePost(param models.EditorWriteParam) (uint, error) {
	if s.repos.Board.GetBoardConfig(param.BoardUid).Type == models.BOARD_TRADE {
		return models.FAILED, fmt.Errorf("trade posts must be written through the trade endpoint")
	}
	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return models.FAILED, fmt.Errorf("you have no permission to write a new post")
	}
	hasPerm, err := s.repos.BoardEdit.CheckWriterForBlog(param.BoardUid, param.UserUid)
	if err != nil {
		return models.FAILED, err
	}
	if !hasPerm {
		return models.FAILED, fmt.Errorf("only blog owner can write a new post")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return models.FAILED, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return models.FAILED, fmt.Errorf("not enough point")
	}
	if param.IsNotice {
		if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
			param.IsNotice = false
		}
	}

	postUid, err := s.repos.BoardEdit.InsertPost(param, models.UpdatePointParam{
		UserUid:  param.UserUid,
		BoardUid: param.BoardUid,
		Action:   models.POINT_ACTION_WRITE,
		Point:    needPt,
	})
	if err != nil {
		return postUid, err
	}
	if err := s.SaveTags(param.BoardUid, postUid, param.Tags); err != nil {
		log.Printf("board: failed to save tags for post %d: %v", postUid, err)
	}
	attachmentsSaved := true
	if err := s.SaveAttachments(models.EditorSaveAttachedParam{
		Context:  param.Context,
		BoardUid: param.BoardUid,
		PostUid:  postUid,
		Files:    param.Files,
	}); err != nil {
		attachmentsSaved = false
		log.Printf("board: failed to save attachments for post %d: %v", postUid, err)
	}
	grantAchievement(s.repos.Badge, param.UserUid, models.BADGE_FIRST_POST, "post", postUid)

	if models.IsSenstaClient(param.ClientKey) && attachmentsSaved && containsImage(param.Files) && s.repos.Badge != nil {
		now := uint64(time.Now().UnixMilli())
		if err := s.repos.Badge.RecordPostOrigin(models.PostOriginParam{
			PostUid: postUid, ClientKey: param.ClientKey, AppVersion: param.AppVersion, RecordedAt: now,
		}); err != nil {
			log.Printf("badge: failed to record post %d origin: %v", postUid, err)
		} else {
			grantAchievement(s.repos.Badge, param.UserUid, models.BADGE_SENSTA_APP, "post", postUid)
		}
	}
	return postUid, nil
}
