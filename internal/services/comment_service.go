package services

import (
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type CommentService interface {
	Like(param models.CommentLikeParam)
	List(param models.CommentListParam) (models.CommentListResult, error)
	Modify(param models.CommentModifyParam) error
	Remove(param models.CommentRemoveParam) error
	Reply(param models.CommentReplyParam) (uint, error)
	Write(param models.CommentWriteParam) (uint, error)
}

type NuboCommentService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboCommentService(repos *repositories.Repository) *NuboCommentService {
	return &NuboCommentService{repos: repos}
}

// 댓글에 좋아요 클릭하기
func (s *NuboCommentService) Like(param models.CommentLikeParam) {
	if isLiked := s.repos.Comment.IsLikedComment(param.CommentUid, param.UserUid); !isLiked {
		s.repos.Comment.InsertLikeComment(param)

		postUid, targetUserUid := s.repos.Comment.FindPostUserUidByUid(param.CommentUid)
		if param.UserUid != targetUserUid {
			s.repos.Noti.InsertNotification(models.InsertNotificationParam{
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
func (s *NuboCommentService) List(param models.CommentListParam) (models.CommentListResult, error) {
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

	result.BoardUid = param.BoardUid
	result.TotalCommentCount = s.repos.Board.GetCommentCount(param.PostUid)
	comments, err := s.repos.Comment.GetComments(param)
	if err != nil {
		return result, err
	}
	result.Comments = comments
	return result, nil
}

// 기존 댓글 수정하기
func (s *NuboCommentService) Modify(param models.CommentModifyParam) error {
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_COMMENT, param.ModifyTargetUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("you have no permission to edit this comment")
	}
	s.repos.Comment.UpdateComment(param.ModifyTargetUid, param.Content)
	return nil
}

// 댓글 삭제하기
func (s *NuboCommentService) Remove(param models.CommentRemoveParam) error {
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_COMMENT, param.RemoveTargetUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("you have no permission to remove this comment")
	}

	if hasReply := s.repos.Comment.HasReplyComment(param.RemoveTargetUid); hasReply {
		s.repos.Comment.UpdateComment(param.RemoveTargetUid, "(deleted)")
	} else {
		s.repos.Comment.RemoveComment(param.RemoveTargetUid)
	}
	return nil
}

// 새로운 답글 작성하기
func (s *NuboCommentService) Reply(param models.CommentReplyParam) (uint, error) {
	insertId, err := s.Write(param.CommentWriteParam)
	if err != nil {
		return models.FAILED, err
	}
	s.repos.Comment.UpdateReplyUid(insertId, param.ReplyTargetUid)
	return insertId, nil
}

// 새로운 댓글 작성하기
func (s *NuboCommentService) Write(param models.CommentWriteParam) (uint, error) {
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
	s.repos.User.UpdatePointHistory(models.UpdatePointParam{
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
		s.repos.Noti.InsertNotification(models.InsertNotificationParam{
			ActionUserUid: param.UserUid,
			TargetUserUid: targetUserUid,
			NotiType:      models.NOTI_LEAVE_COMMENT,
			PostUid:       param.PostUid,
			CommentUid:    insertId,
		})

		if len(configs.Env.GmailAppPassword) > 0 {
			go func() {
				writerInfo := s.repos.Auth.FindMyInfoByUid(targetUserUid)
				commenterInfo := s.repos.Admin.FindWriterByUid(param.UserUid)
				config := s.repos.Board.GetBoardConfig(param.BoardUid)

				body := strings.ReplaceAll(templates.NoticeCommentBody, "{{Host}}", configs.Env.URL)
				body = strings.ReplaceAll(body, "{{Name}}", utils.Unescape(writerInfo.Name))
				body = strings.ReplaceAll(body, "{{Commenter}}", utils.Unescape(commenterInfo.Name))
				body = strings.ReplaceAll(body, "{{Comment}}", param.Content)
				body = strings.ReplaceAll(body, "{{Link}}", fmt.Sprintf("%s%s/board/%s/view/%d", configs.Env.URL, configs.Env.URLPrefix, config.Id, param.PostUid))
				body = strings.ReplaceAll(body, "{{From}}", configs.Env.GmailID)
				subject := fmt.Sprintf("[%s] %s has just commented on your post!", config.Name, commenterInfo.Name)
				from := fmt.Sprintf("Admin <noreply@%s>", param.Hostname)
				utils.SendMail(writerInfo.Id, from, subject, body)
			}()
		}
	}

	return insertId, nil
}
