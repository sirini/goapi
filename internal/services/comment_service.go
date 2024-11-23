package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type CommentService interface {
	LoadComments(param models.CommentListParameter) (models.CommentListResult, error)
}

type TsboardCommentService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardCommentService(repos *repositories.Repository) *TsboardCommentService {
	return &TsboardCommentService{repos: repos}
}

// 댓글 목록 가져오기
func (s *TsboardCommentService) LoadComments(param models.CommentListParameter) (models.CommentListResult, error) {
	result := models.CommentListResult{}
	userLv, _ := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, _ := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_VIEW)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}

	status := s.repos.Comment.GetPostStatus(param.PostUid)
	if status == models.CONTENT_SECRET {
		isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
		isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
		if !isAdmin && !isAuthor {
			return result, fmt.Errorf("you have no permission to read comments on this post")
		}
	}
	if status == models.CONTENT_REMOVED {
		return result, fmt.Errorf("post has been removed")
	}

	if param.SinceUid < 1 {
		param.SinceUid = s.repos.Comment.GetMaxUid() + 1
	}

	result.BoardUid = param.BoardUid
	result.SinceUid = param.SinceUid
	result.TotalCommentCount = s.repos.Board.GetCountByTable(models.TABLE_COMMENT, param.PostUid)
	comments, err := s.repos.Comment.GetComments(param)
	if err != nil {
		return result, err
	}
	result.Comments = comments
	return result, nil
}
