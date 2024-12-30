package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type SyncService interface {
	GetLatestPosts(bunch uint) []models.SyncPostItem
}

type TsboardSyncService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardSyncService(repos *repositories.Repository) *TsboardSyncService {
	return &TsboardSyncService{repos: repos}
}

// (허용된) 다른 곳에서 이 곳 게시글들을 동기화 할 수 있도록 최근 게시글들 가져오기
func (s *TsboardSyncService) GetLatestPosts(bunch uint) []models.SyncPostItem {
	items := make([]models.SyncPostItem, 0)
	maxUid := s.repos.Board.GetMaxUid(models.TABLE_POST) + 1
	posts, err := s.repos.Home.GetLatestPosts(models.HomePostParameter{
		SinceUid: maxUid,
		Bunch:    bunch,
		Option:   models.SEARCH_NONE,
		Keyword:  "",
		UserUid:  0,
		BoardUid: 0,
	})
	if err != nil {
		return items
	}

	for _, post := range posts {
		config := s.repos.Board.GetBoardConfig(post.BoardUid)
		writer := s.repos.Board.GetWriterInfo(post.UserUid)

		hashtags := s.repos.BoardView.GetTags(post.Uid)
		tags := make([]string, 0)
		for _, tag := range hashtags {
			tags = append(tags, tag.Name)
		}

		attachedImages, err := s.repos.BoardView.GetAttachedImages(post.Uid)
		if err != nil {
			return items
		}

		images := make([]models.SyncImageItem, 0)
		for _, img := range attachedImages {
			filename := s.repos.Sync.GetFileName(img.File.Uid)
			image := models.SyncImageItem{
				Uid:   img.File.Uid,
				File:  img.File.Path,
				Name:  filename,
				Thumb: img.Thumbnail.Small,
				Full:  img.Thumbnail.Large,
				Desc:  img.Description,
				Exif:  img.Exif,
			}
			images = append(images, image)
		}

		item := models.SyncPostItem{
			Id:        config.Id,
			No:        post.Uid,
			Title:     utils.Unescape(post.Title),
			Content:   utils.Unescape(post.Content),
			Submitted: post.Submitted,
			Name:      writer.Name,
			Tags:      tags,
			Images:    images,
		}
		items = append(items, item)
	}
	return items
}
