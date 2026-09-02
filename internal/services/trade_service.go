package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type TradeService interface {
	GetList(param models.BoardListParam) (models.TradeListResult, error)
	GetView(param models.BoardViewParam) (models.TradeViewResult, error)
	LoadPost(boardUid uint, postUid uint, userUid uint) (models.TradeLoadPostResult, error)
	ModifyTradePost(param models.TradeModifyParam) error
	UpdateTradeStatus(param models.TradeStatusParam) error
	WriteTradePost(param models.TradeWriteParam) (models.TradeWriteResult, error)
}

type NuboTradeService struct {
	repos *repositories.Repository
	board BoardService
}

func NewNuboTradeService(repos *repositories.Repository, board BoardService) *NuboTradeService {
	return &NuboTradeService{repos: repos, board: board}
}

func (s *NuboTradeService) requireTradeBoard(boardUid uint) (models.BoardConfig, error) {
	config := s.repos.Board.GetBoardConfig(boardUid)
	if config.Uid < 1 || config.Type != models.BOARD_TRADE {
		return config, fmt.Errorf("board is not a trade board")
	}
	return config, nil
}

func mergeTradeItems(posts []models.BoardListItem, tradeByPost map[uint]models.TradeResult) ([]models.TradeListItem, error) {
	items := make([]models.TradeListItem, 0, len(posts))
	for _, post := range posts {
		trade, ok := tradeByPost[post.Uid]
		if !ok {
			return nil, fmt.Errorf("trade metadata not found for post %d", post.Uid)
		}
		items = append(items, models.TradeListItem{Post: post, Trade: trade})
	}
	return items, nil
}

func (s *NuboTradeService) GetList(param models.BoardListParam) (models.TradeListResult, error) {
	result := models.TradeListResult{}
	if _, err := s.requireTradeBoard(param.BoardUid); err != nil {
		return result, err
	}
	boardResult, err := s.board.GetListItem(param)
	if err != nil {
		return result, err
	}
	postUids := make([]uint, 0, len(boardResult.Notices)+len(boardResult.Posts))
	for _, post := range boardResult.Notices {
		postUids = append(postUids, post.Uid)
	}
	for _, post := range boardResult.Posts {
		postUids = append(postUids, post.Uid)
	}
	tradeByPost, err := s.repos.Trade.GetTradeItems(postUids)
	if err != nil {
		return result, err
	}
	notices, err := mergeTradeItems(boardResult.Notices, tradeByPost)
	if err != nil {
		return result, err
	}
	posts, err := mergeTradeItems(boardResult.Posts, tradeByPost)
	if err != nil {
		return result, err
	}
	return models.TradeListResult{
		TotalPostCount: boardResult.TotalPostCount,
		Config:         boardResult.Config,
		Notices:        notices,
		Posts:          posts,
		BlackList:      boardResult.BlackList,
		IsAdmin:        boardResult.IsAdmin,
	}, nil
}

func (s *NuboTradeService) GetView(param models.BoardViewParam) (models.TradeViewResult, error) {
	result := models.TradeViewResult{}
	if _, err := s.requireTradeBoard(param.BoardUid); err != nil {
		return result, err
	}
	boardResult, err := s.board.GetViewItem(param)
	if err != nil {
		return result, err
	}
	trade, err := s.repos.Trade.GetTradeItem(param.PostUid)
	if err != nil {
		return result, err
	}
	result.BoardViewResult = boardResult
	result.Trade = trade
	return result, nil
}

func (s *NuboTradeService) LoadPost(boardUid uint, postUid uint, userUid uint) (models.TradeLoadPostResult, error) {
	result := models.TradeLoadPostResult{}
	if _, err := s.requireTradeBoard(boardUid); err != nil {
		return result, err
	}
	boardResult, err := s.board.LoadPost(boardUid, postUid, userUid)
	if err != nil {
		return result, err
	}
	trade, err := s.repos.Trade.GetTradeItem(postUid)
	if err != nil {
		return result, err
	}
	result.Board = boardResult
	result.Trade = trade
	return result, nil
}

func hasProductImage(content string, files []*multipart.FileHeader) bool {
	if strings.Contains(strings.ToLower(content), "<img") {
		return true
	}
	for _, file := range files {
		if strings.HasPrefix(strings.ToLower(file.Header.Get("Content-Type")), "image/") {
			return true
		}
		switch strings.ToLower(filepath.Ext(file.Filename)) {
		case ".avif", ".gif", ".heic", ".jpeg", ".jpg", ".png", ".webp":
			return true
		}
	}
	return false
}

func (s *NuboTradeService) WriteTradePost(param models.TradeWriteParam) (models.TradeWriteResult, error) {
	result := models.TradeWriteResult{}
	if _, err := s.requireTradeBoard(param.BoardUid); err != nil {
		return result, err
	}
	if !hasProductImage(param.Content, param.Files) {
		return result, fmt.Errorf("at least one product image is required")
	}
	if !s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST) {
		return result, fmt.Errorf("you have no permission to write a new trade post")
	}
	userLv, userPt := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}
	if param.IsNotice && !s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid) {
		param.IsNotice = false
	}
	postUid, err := s.repos.Trade.InsertTradePost(param, models.UpdatePointParam{
		UserUid: param.UserUid, BoardUid: param.BoardUid, Action: models.POINT_ACTION_WRITE, Point: needPt,
	})
	if err != nil {
		return result, err
	}
	if err := s.board.SaveTags(param.BoardUid, postUid, param.Tags); err != nil {
		return result, err
	}
	if err := s.board.SaveAttachments(models.EditorSaveAttachedParam{
		Context: param.Context, BoardUid: param.BoardUid, PostUid: postUid, Files: param.Files,
	}); err != nil {
		return result, err
	}
	grantAchievement(s.repos.Badge, param.UserUid, models.BADGE_FIRST_POST, "post", postUid)
	result.PostUid = postUid
	return result, nil
}

func (s *NuboTradeService) ModifyTradePost(param models.TradeModifyParam) error {
	if _, err := s.requireTradeBoard(param.BoardUid); err != nil {
		return err
	}
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return fmt.Errorf("post does not belong to this board")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("only the author can edit this trade post")
	}
	if !s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST) {
		return fmt.Errorf("you have no permission to edit post")
	}
	existingImages, err := s.repos.BoardView.GetAttachedImages(param.PostUid)
	if err != nil {
		return err
	}
	if !hasProductImage(param.Content, param.Files) && len(existingImages) == 0 {
		return fmt.Errorf("at least one product image is required")
	}
	if param.IsNotice && !isAdmin {
		param.IsNotice = false
	}
	if err := s.repos.Trade.UpdateTradePost(param); err != nil {
		return err
	}
	s.repos.BoardView.RemovePostTags(param.PostUid)
	if err := s.board.SaveTags(param.BoardUid, param.PostUid, param.Tags); err != nil {
		return err
	}
	return s.board.SaveAttachments(models.EditorSaveAttachedParam{
		Context: param.Context, BoardUid: param.BoardUid, PostUid: param.PostUid, Files: param.Files,
	})
}

func (s *NuboTradeService) UpdateTradeStatus(param models.TradeStatusParam) error {
	if !param.Status.Valid() {
		return fmt.Errorf("invalid trade status")
	}
	if _, err := s.requireTradeBoard(param.BoardUid); err != nil {
		return err
	}
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return fmt.Errorf("post does not belong to this board")
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("only the author or board administrator can change the transaction status")
	}
	return s.repos.Trade.UpdateStatus(param.PostUid, param.Status)
}
