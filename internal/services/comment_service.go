package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type CommentService interface {
	LikeComment(param models.CommentLikeParameter)
	LoadComments(param models.CommentListParameter) (models.CommentListResult, error)
	WriteComment(param models.CommentWriteParameter) (uint, error)
}

type TsboardCommentService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardCommentService(repos *repositories.Repository) *TsboardCommentService {
	return &TsboardCommentService{repos: repos}
}

// 댓글에 좋아요 클릭하기
func (s *TsboardCommentService) LikeComment(param models.CommentLikeParameter) {
	if isLiked := s.repos.Comment.IsLikedComment(param.CommentUid, param.UserUid); !isLiked {
		s.repos.Comment.InsertLikeComment(param)

		postUid, targetUserUid := s.repos.Comment.FindPostUserUidByUid(param.CommentUid)
		if param.UserUid != targetUserUid {
			s.repos.Noti.InsertNotification(models.InsertNotificationParameter{
				ActionUserUid: param.UserUid,
				TargetUserUid: targetUserUid,
				NotiType:      models.NOTI_LIKE_COMMENT,
				PostUid:       postUid,
				CommentUid:    param.CommentUid,
			})
		}
	} else {
		s.repos.Comment.UpdateLikeComment(param)
	}
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

// 새로운 댓글 작성하기
func (s *TsboardCommentService) WriteComment(param models.CommentWriteParameter) (uint, error) {
	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_COMMENT); !hasPerm {
		return models.FAILED, fmt.Errorf("you have no permission to write a comment")
	}
	if isBanned := s.repos.BoardView.CheckBannedByWriter(param.PostUid, param.UserUid); isBanned {
		return models.FAILED, fmt.Errorf("you have been blocked by writer")
	}
	if status := s.repos.Comment.GetPostStatus(param.PostUid); status == models.CONTENT_REMOVED {
		return models.FAILED, fmt.Errorf("leaving a comment on a removed post is not allowed")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_COMMENT)
	if userLv < needLv {
		return models.FAILED, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return models.FAILED, fmt.Errorf("not enough point")
	}
	s.repos.User.UpdateUserPoint(param.UserUid, uint(userPt+needPt))
	s.repos.User.UpdatePointHistory(models.UpdatePointParameter{
		UserUid:  param.UserUid,
		BoardUid: param.BoardUid,
		Action:   models.POINT_ACTION_COMMENT,
		Point:    needPt,
	})

	insertId, err := s.repos.Comment.InsertComment(param)
	if err != nil {
		return models.FAILED, err
	}
	s.repos.Comment.UpdateReplyUid(insertId, insertId)

	targetUserUid := s.repos.Comment.GetPostWriterUid(param.PostUid)
	if param.UserUid != targetUserUid {
		s.repos.Noti.InsertNotification(models.InsertNotificationParameter{
			ActionUserUid: param.UserUid,
			TargetUserUid: targetUserUid,
			NotiType:      models.NOTI_LEAVE_COMMENT,
			PostUid:       param.PostUid,
			CommentUid:    insertId,
		})
	}
	return insertId, nil
}
