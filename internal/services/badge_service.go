package services

import (
	"log"
	"time"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

func grantAchievement(
	repo repositories.BadgeRepository,
	userUid uint,
	badgeKey string,
	evidenceType string,
	evidenceUid uint,
) {
	if repo == nil || userUid < 1 {
		return
	}
	_, err := repo.Award(models.BadgeAwardParam{
		UserUid:      userUid,
		BadgeKey:     badgeKey,
		QualifiedAt:  uint64(time.Now().UnixMilli()),
		GrantSource:  "system",
		EvidenceType: evidenceType,
		EvidenceUid:  evidenceUid,
	})
	if err != nil {
		log.Printf("badge: failed to grant %s to user %d: %v", badgeKey, userUid, err)
	}
}

func loadFeaturedBadges(
	repo repositories.BadgeRepository,
	userUids []uint,
) map[uint][]models.UserBadge {
	if repo == nil || len(userUids) == 0 {
		return map[uint][]models.UserBadge{}
	}
	seen := make(map[uint]struct{}, len(userUids))
	unique := make([]uint, 0, len(userUids))
	for _, userUid := range userUids {
		if userUid < 1 {
			continue
		}
		if _, exists := seen[userUid]; exists {
			continue
		}
		seen[userUid] = struct{}{}
		unique = append(unique, userUid)
	}
	badges, err := repo.FindFeaturedForUsers(unique)
	if err != nil {
		log.Printf("badge: failed to load inline achievements: %v", err)
		return map[uint][]models.UserBadge{}
	}
	return badges
}

func (s *NuboBoardService) attachFeaturedBadges(groups ...[]models.BoardListItem) {
	userUids := make([]uint, 0)
	for _, posts := range groups {
		for _, post := range posts {
			userUids = append(userUids, post.Writer.UserUid)
		}
	}
	badges := loadFeaturedBadges(s.repos.Badge, userUids)
	for _, posts := range groups {
		for i := range posts {
			posts[i].Writer.Badges = badges[posts[i].Writer.UserUid]
		}
	}
}

func attachFeaturedBadgesToComments(repo repositories.BadgeRepository, comments []models.CommentItem) {
	userUids := make([]uint, 0, len(comments))
	for _, comment := range comments {
		userUids = append(userUids, comment.Writer.UserUid)
	}
	badges := loadFeaturedBadges(repo, userUids)
	for i := range comments {
		comments[i].Writer.Badges = badges[comments[i].Writer.UserUid]
	}
}
