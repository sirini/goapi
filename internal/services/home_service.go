package services

import (
	"fmt"
	"html/template"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type HomeService interface {
	AddVisitorLog(userUid uint)
	GetBoardIDsForSitemap() []models.HomeSitemapURL
	GetLatestPosts(param models.HomePostParameter) ([]models.BoardHomePostItem, error)
	GetSidebarLinks() ([]models.HomeSidebarGroupResult, error)
	LoadMainPage(bunch uint) ([]models.HomeMainArticle, error)
}

type TsboardHomeService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardHomeService(repos *repositories.Repository) *TsboardHomeService {
	return &TsboardHomeService{repos: repos}
}

// 방문자 접속 기록하기
func (s *TsboardHomeService) AddVisitorLog(userUid uint) {
	s.repos.Home.InsertVisitorLog(userUid)
}

// 사이트맵에서 보여줄 게시판 경로 목록 반환하기
func (s *TsboardHomeService) GetBoardIDsForSitemap() []models.HomeSitemapURL {
	items := make([]models.HomeSitemapURL, 0)
	ids := s.repos.Home.GetBoardIDs()

	for _, id := range ids {
		item := models.HomeSitemapURL{
			Loc:        fmt.Sprintf("%s/board/%s/page/1", configs.Env.URL, id),
			LastMod:    time.Now().Format("2006-01-02"),
			ChangeFreq: "daily",
			Priority:   "0.5",
		}
		items = append(items, item)
	}
	return items
}

// 지정된 게시글 번호 이하의 최근글들 가져오기
func (s *TsboardHomeService) GetLatestPosts(param models.HomePostParameter) ([]models.BoardHomePostItem, error) {
	var items []models.BoardHomePostItem
	var (
		posts []models.HomePostItem
		err   error
	)

	if len(param.Keyword) < 2 {
		posts, err = s.repos.Home.GetLatestPosts(param)
	} else {
		switch param.Option {
		case models.SEARCH_TAG:
			posts, err = s.repos.Home.FindLatestPostsByTag(param)
		case models.SEARCH_CATEGORY:
		case models.SEARCH_WRITER:
			posts, err = s.repos.Home.FindLatestPostsByUserUidCatUid(param)
		default:
			posts, err = s.repos.Home.FindLatestPostsByTitleContent(param)
		}
	}
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		settings := s.repos.Home.GetBoardBasicSettings(post.BoardUid)
		if len(settings.Id) < 2 {
			continue
		}

		item := models.BoardHomePostItem{}
		item.Uid = post.Uid
		item.Title = post.Title
		item.Content = post.Content
		item.Submitted = post.Submitted
		item.Modified = post.Modified
		item.Hit = post.Hit
		item.Status = post.Status

		item.Id = settings.Id
		item.Type = settings.Type
		item.UseCategory = settings.UseCategory

		item.Category = s.repos.Board.GetCategoryByUid(post.CategoryUid)
		item.Cover = s.repos.Board.GetCoverImage(post.Uid)
		item.Comment = s.repos.Board.GetCountByTable(models.TABLE_COMMENT, post.Uid)
		item.Writer = s.repos.Board.GetWriterInfo(post.UserUid)
		item.Like = s.repos.Board.GetCountByTable(models.TABLE_POST_LIKE, post.Uid)
		item.Liked = s.repos.Board.CheckLikedPost(post.Uid, param.UserUid)

		items = append(items, item)
	}
	return items, nil
}

// 사이드바 그룹/게시판들 목록 가져오기
func (s *TsboardHomeService) GetSidebarLinks() ([]models.HomeSidebarGroupResult, error) {
	return s.repos.Home.GetGroupBoardLinks()
}

// SEO 메인 페이지 가져오기
func (s *TsboardHomeService) LoadMainPage(bunch uint) ([]models.HomeMainArticle, error) {
	var articles []models.HomeMainArticle
	posts, err := s.GetLatestPosts(models.HomePostParameter{
		SinceUid: s.repos.Board.GetMaxUid() + 1,
		Bunch:    bunch,
		Option:   models.SEARCH_NONE,
		Keyword:  "",
		UserUid:  0,
		BoardUid: 0,
	})

	if err != nil {
		return articles, err
	}

	for _, post := range posts {
		article := models.HomeMainArticle{}
		article.Cover = fmt.Sprintf("%s%s", configs.Env.URL, post.Cover)
		article.Content = template.HTML(utils.Unescape(post.Content))
		article.Date = utils.ConvTimestamp(post.Submitted)
		article.Like = post.Like
		article.Name = post.Writer.Name
		article.Title = utils.Unescape(post.Title)
		article.Url = fmt.Sprintf("%s/%s/%s/%d", configs.Env.URL, post.Type.String(), post.Id, post.Uid)
		article.Hashtags = s.repos.BoardView.GetTags(post.Uid)

		comments, err := s.repos.Comment.GetComments(models.CommentListParameter{
			BoardUid:  0,
			PostUid:   post.Uid,
			UserUid:   0,
			Page:      1,
			Bunch:     bunch,
			SinceUid:  s.repos.Comment.GetMaxUid() + 1,
			Direction: models.PAGE_NEXT,
		})
		if err != nil {
			continue
		}

		for _, comment := range comments {
			item := models.HomeMainComment{
				Content: template.HTML(comment.Content),
				Date:    utils.ConvTimestamp(comment.Submitted),
				Like:    comment.Like,
				Name:    comment.Writer.Name,
			}
			article.Comments = append(article.Comments, item)
		}
		articles = append(articles, article)
	}
	return articles, nil
}
