package services

import (
	"fmt"
	"log"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type CommentService interface {
	Like(param models.CommentLikeParam) error
	SetReaction(param models.CommentReactionParam) (models.ReactionState, error)
	List(param models.CommentListParam) (models.CommentListResult, error)
	Modify(param models.CommentModifyParam) error
	Remove(param models.CommentRemoveParam) error
	Reply(param models.CommentReplyParam) (uint, error)
	Write(param models.CommentWriteParam) (uint, error)
}

type NuboCommentService struct {
	repos         *repositories.Repository
	mailer        utils.Mailer
	notifications *notificationPublisher
}

// 리포지토리 묶음 주입받기
func NewNuboCommentService(repos *repositories.Repository) *NuboCommentService {
	return newNuboCommentService(repos, utils.NewResendMailer())
}

func newNuboCommentService(repos *repositories.Repository, mailer utils.Mailer) *NuboCommentService {
	return &NuboCommentService{
		repos:         repos,
		mailer:        mailer,
		notifications: newNotificationPublisher(repos, disabledPushSender{}),
	}
}

// 댓글에 좋아요 클릭하기
func (s *NuboCommentService) Like(param models.CommentLikeParam) error {
	if !s.repos.Comment.IsCommentInBoard(param.CommentUid, param.BoardUid) {
		return fmt.Errorf("comment does not belong to this board")
	}
	// liked:false가 접근 검사 전에 조기 성공하지 않도록 자격을 먼저 확인한다.
	if err := s.checkCommentReactionAccess(param.CommentUid, param.BoardUid, param.UserUid); err != nil {
		return err
	}
	if !param.Liked {
		// 현재 종류가 like일 때만 취소하고, 다른 리액션은 건드리지 않는다.
		if s.repos.Comment.GetCommentUserReaction(param.CommentUid, param.UserUid) != models.REACTION_LIKE {
			return nil
		}
	}
	_, err := s.SetReaction(models.CommentReactionParam{
		BoardUid: param.BoardUid, CommentUid: param.CommentUid, UserUid: param.UserUid,
		Reaction: models.REACTIONS.LIKE, ReactionIsNull: !param.Liked,
	})
	return err
}

// 댓글에 다중 리액션 남기기
func (s *NuboCommentService) SetReaction(param models.CommentReactionParam) (models.ReactionState, error) {
	if !s.repos.Comment.IsCommentInBoard(param.CommentUid, param.BoardUid) {
		return models.ReactionState{}, fmt.Errorf("comment does not belong to this board")
	}
	if err := s.checkCommentReactionAccess(param.CommentUid, param.BoardUid, param.UserUid); err != nil {
		return models.ReactionState{}, err
	}
	code := models.REACTION_NONE
	if param.ReactionIsNull {
		// 취소(null)는 활성 행을 0으로 바꾼다. 취소할 상태가 없으면 무변경으로 성공한다.
		if s.repos.Comment.GetCommentUserReaction(param.CommentUid, param.UserUid) == models.REACTION_NONE {
			return s.repos.Comment.GetCommentReactionState(param.CommentUid, param.UserUid), nil
		}
	} else {
		var err error
		code, err = models.ParseReaction(param.Reaction)
		if err != nil {
			return models.ReactionState{}, err
		}
	}
	param.ReactionCode = code
	if code == models.REACTION_LIKE {
		postUid, targetUserUid := s.repos.Comment.FindPostUserUidByUid(param.CommentUid)
		param.TargetUserUid = targetUserUid
		param.Notify = param.UserUid != targetUserUid
		changed, err := s.repos.Comment.SetCommentReaction(param)
		if err != nil {
			return models.ReactionState{}, err
		}
		state := s.repos.Comment.GetCommentReactionState(param.CommentUid, param.UserUid)
		if changed && param.Notify {
			s.notifications.Save(models.InsertNotificationParam{
				ActionUserUid: param.UserUid,
				TargetUserUid: targetUserUid,
				NotiType:      models.NOTI_LIKE_COMMENT,
				PostUid:       postUid,
				CommentUid:    param.CommentUid,
			}, true)
		}
		return state, nil
	}
	if _, err := s.repos.Comment.SetCommentReaction(param); err != nil {
		return models.ReactionState{}, err
	}
	return s.repos.Comment.GetCommentReactionState(param.CommentUid, param.UserUid), nil
}

// 댓글 리액션은 삭제된 댓글(답글 자리만 남은 경우 포함)과 열람 자격이 없는 부모 게시글을 거부한다.
// 포인트는 차감하지 않고 열람 가능 여부만 확인한다.
func (s *NuboCommentService) checkCommentReactionAccess(commentUid uint, boardUid uint, userUid uint) error {
	if status := s.repos.Comment.GetCommentStatus(commentUid); status == models.CONTENT_REMOVED {
		return fmt.Errorf("comment has been removed")
	}
	postUid, _ := s.repos.Comment.FindPostUserUidByUid(commentUid)
	status := s.repos.Comment.GetPostStatus(postUid)
	if postUid < 1 || status == models.CONTENT_REMOVED {
		return fmt.Errorf("post is not available")
	}
	if isBanned := s.repos.BoardView.CheckBannedByWriter(postUid, userUid); isBanned {
		return fmt.Errorf("you have been blocked by writer")
	}
	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_VIEW)
	if userLv < needLv {
		return fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return fmt.Errorf("not enough point")
	}
	if status == models.CONTENT_SECRET {
		isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
		isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
		if !isAdmin && !isWriter {
			return fmt.Errorf("you have no permission to react to this comment")
		}
	}
	return nil
}

// 댓글 목록 가져오기
func (s *NuboCommentService) List(param models.CommentListParam) (models.CommentListResult, error) {
	result := models.CommentListResult{}
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return result, fmt.Errorf("post does not belong to this board")
	}
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
	attachFeaturedBadgesToComments(s.repos.Badge, comments)
	result.Comments = comments
	return result, nil
}

// 기존 댓글 수정하기
func (s *NuboCommentService) Modify(param models.CommentModifyParam) error {
	if !s.repos.Comment.IsCommentInPost(param.ModifyTargetUid, param.PostUid, param.BoardUid) {
		return fmt.Errorf("comment does not belong to this post")
	}
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
	if !s.repos.Comment.IsCommentInBoard(param.RemoveTargetUid, param.BoardUid) {
		return fmt.Errorf("comment does not belong to this board")
	}
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
	if !s.repos.Comment.IsCommentInPost(param.ReplyTargetUid, param.PostUid, param.BoardUid) {
		return models.FAILED, fmt.Errorf("reply target does not belong to this post")
	}
	return s.write(param.CommentWriteParam, param.ReplyTargetUid)
}

// 새로운 댓글 작성하기
func (s *NuboCommentService) Write(param models.CommentWriteParam) (uint, error) {
	return s.write(param, 0)
}

func (s *NuboCommentService) write(param models.CommentWriteParam, replyUid uint) (uint, error) {
	if !s.repos.BoardView.IsPostInBoard(param.PostUid, param.BoardUid) {
		return models.FAILED, fmt.Errorf("post does not belong to this board")
	}
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
	insertId, err := s.repos.Comment.InsertComment(param, replyUid, models.UpdatePointParam{
		UserUid:  param.UserUid,
		BoardUid: param.BoardUid,
		Action:   models.POINT_ACTION_COMMENT,
		Point:    needPt,
	})
	if err != nil {
		return models.FAILED, err
	}
	grantAchievement(s.repos.Badge, param.UserUid, models.BADGE_FIRST_COMMENT, "comment", insertId)

	targetUserUid := s.repos.Comment.GetPostWriterUid(param.PostUid)
	if param.UserUid != targetUserUid {
		s.notifications.Save(models.InsertNotificationParam{
			ActionUserUid: param.UserUid,
			TargetUserUid: targetUserUid,
			NotiType:      models.NOTI_LEAVE_COMMENT,
			PostUid:       param.PostUid,
			CommentUid:    insertId,
		}, false)

		if s.mailer.Configured() {
			writerInfo := s.repos.Auth.FindMyInfoByUid(targetUserUid)
			commenterInfo := s.repos.Admin.FindWriterByUid(param.UserUid)
			config := s.repos.Board.GetBoardConfig(param.BoardUid)
			commentURL := fmt.Sprintf("%s/board/%s/%d", siteURL(), config.Id, param.PostUid)
			excerpt := utils.CutString(utils.PlainText(param.Content), 240)
			html, text, renderErr := templates.RenderTransactionalMail(templates.MailContent{
				SiteName:    configs.Env.Title,
				SiteURL:     siteURL(),
				Preheader:   fmt.Sprintf("%s님이 내 글에 댓글을 남겼습니다.", utils.Unescape(commenterInfo.Name)),
				Label:       "New comment",
				Heading:     "내 글에 새 댓글이 달렸습니다",
				Greeting:    fmt.Sprintf("안녕하세요, %s님.", utils.Unescape(writerInfo.Name)),
				Body:        fmt.Sprintf("%s님이 %s 게시판의 내 글에 댓글을 남겼습니다.", utils.Unescape(commenterInfo.Name), utils.Unescape(config.Name)),
				Highlight:   excerpt,
				ActionLabel: "댓글 확인하기",
				ActionURL:   commentURL,
				Notice:      "이 알림은 내 글에 새 댓글이 등록되어 발송되었습니다.",
			})
			if renderErr != nil {
				log.Printf("mail: failed to render comment notification %d: %v", insertId, renderErr)
			} else {
				go func() {
					delivery, err := s.mailer.Send(models.MailMessage{
						To:             writerInfo.Id,
						Subject:        fmt.Sprintf("[%s] 내 글에 새 댓글이 달렸습니다", configs.Env.Title),
						HTML:           html,
						Text:           text,
						IdempotencyKey: fmt.Sprintf("comment-notification/%d", insertId),
						Tags:           map[string]string{"type": "comment-notification"},
					})
					if err != nil {
						log.Printf("mail: comment notification %d delivery failed: %v", insertId, err)
						return
					}
					log.Printf("mail: comment notification %d accepted by %s as %s", insertId, delivery.Provider, delivery.MessageID)
				}()
			}
		}
	}

	return insertId, nil
}
