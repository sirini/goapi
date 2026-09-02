package services

import (
	"reflect"
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

type featuredBadgeRepo struct {
	requested []uint
	badges    map[uint][]models.UserBadge
}

func (r *featuredBadgeRepo) Award(models.BadgeAwardParam) (bool, error)         { return false, nil }
func (r *featuredBadgeRepo) CreateDefinition(models.BadgeDefinition) error      { return nil }
func (r *featuredBadgeRepo) FindForUser(uint, bool) ([]models.UserBadge, error) { return nil, nil }
func (r *featuredBadgeRepo) FindFeaturedForUsers(userUids []uint) (map[uint][]models.UserBadge, error) {
	r.requested = append([]uint(nil), userUids...)
	return r.badges, nil
}
func (r *featuredBadgeRepo) FindUnannouncedForUser(uint, uint) ([]models.UserBadge, error) {
	return nil, nil
}
func (r *featuredBadgeRepo) ListDefinitions() ([]models.BadgeDefinition, error) { return nil, nil }
func (r *featuredBadgeRepo) MarkAnnounced(uint, []string, uint64) error         { return nil }
func (r *featuredBadgeRepo) RecordPostOrigin(models.PostOriginParam) error      { return nil }
func (r *featuredBadgeRepo) UpdateDefinition(models.BadgeDefinition) (bool, error) {
	return false, nil
}

func TestAttachFeaturedBadgesToCommentsBatchesWriters(t *testing.T) {
	repo := &featuredBadgeRepo{badges: map[uint][]models.UserBadge{
		7: {{Key: models.BADGE_SENSTA_APP, Name: "Sensta 앱 사용자"}},
	}}
	comments := []models.CommentItem{
		{Uid: 1, Writer: models.BoardWriter{UserBasicInfo: models.UserBasicInfo{UserUid: 7}}},
		{Uid: 2, Writer: models.BoardWriter{UserBasicInfo: models.UserBasicInfo{UserUid: 7}}},
		{Uid: 3, Writer: models.BoardWriter{UserBasicInfo: models.UserBasicInfo{UserUid: 9}}},
	}

	attachFeaturedBadgesToComments(repo, comments)

	if !reflect.DeepEqual(repo.requested, []uint{7, 9}) {
		t.Fatalf("unexpected batched writer uids: %#v", repo.requested)
	}
	if got := comments[0].Writer.Badges; len(got) != 1 || got[0].Key != models.BADGE_SENSTA_APP {
		t.Fatalf("expected inline badge on first comment, got %#v", got)
	}
	if got := comments[1].Writer.Badges; len(got) != 1 {
		t.Fatalf("expected repeated writer to receive badge, got %#v", got)
	}
	if got := comments[2].Writer.Badges; len(got) != 0 {
		t.Fatalf("expected no badge for unmatched writer, got %#v", got)
	}
}
