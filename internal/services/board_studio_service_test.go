package services

import (
	"errors"
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type studioBoardRepo struct {
	repositories.BoardRepository
	param  models.BoardStudioParam
	result models.BoardStudioResult
	err    error
}

func (r *studioBoardRepo) GetStudio(param models.BoardStudioParam) (models.BoardStudioResult, error) {
	r.param = param
	return r.result, r.err
}

func TestBoardServiceDelegatesStudioQueryWithoutChangingUserScope(t *testing.T) {
	want := models.BoardStudioResult{Summary: models.BoardStudioSummary{PostCount: 2}}
	repo := &studioBoardRepo{result: want}
	service := NewNuboBoardService(&repositories.Repository{Board: repo})
	param := models.BoardStudioParam{
		BoardUid: 7,
		UserUid:  19,
		Page:     2,
		Limit:    10,
		Sort:     models.BOARD_STUDIO_SORT_LIKES,
	}

	got, err := service.GetStudio(param)
	if err != nil || got.Summary.PostCount != 2 || repo.param != param {
		t.Fatalf("GetStudio() = %+v, %v; repository param = %+v", got, err, repo.param)
	}

	repo.err = errors.New("database unavailable")
	if _, err := service.GetStudio(param); err == nil || err.Error() != "database unavailable" {
		t.Fatalf("repository error = %v", err)
	}
}
