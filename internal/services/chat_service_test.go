package services

import (
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type chatBlockUserRepoStub struct {
	repositories.UserRepository
	blocked map[[2]uint]bool
}

func (r chatBlockUserRepoStub) IsBannedByTarget(actionUserUid uint, targetUserUid uint) bool {
	// 실제 저장소 메서드의 의미와 같이 target이 action을 차단했는지 반환한다.
	return r.blocked[[2]uint{targetUserUid, actionUserUid}]
}

type chatRepoStub struct {
	repositories.ChatRepository
	historyCalls int
	insertCalls  int
}

func (r *chatRepoStub) LoadChatHistory(uint, uint, uint) ([]models.ChatHistory, error) {
	r.historyCalls++
	return []models.ChatHistory{{Uid: 1}}, nil
}

func (r *chatRepoStub) InsertNewChat(uint, uint, string) uint {
	r.insertCalls++
	return 1
}

func TestChatIsHiddenAndSendingFailsForEitherBlockDirection(t *testing.T) {
	for _, blocked := range []map[[2]uint]bool{
		{{7, 9}: true},
		{{9, 7}: true},
	} {
		chat := &chatRepoStub{}
		service := NewNuboChatService(&repositories.Repository{
			Chat: chat,
			User: chatBlockUserRepoStub{blocked: blocked},
		})

		history, err := service.GetChattingHistory(7, 9, 100)
		if err != nil || len(history) != 0 || chat.historyCalls != 0 {
			t.Fatalf("blocked history = (%v, %v), calls = %d", history, err, chat.historyCalls)
		}
		if uid := service.SaveChatMessage(7, 9, "차단 우회"); uid != 0 || chat.insertCalls != 0 {
			t.Fatalf("blocked send uid = %d, calls = %d", uid, chat.insertCalls)
		}
	}
}
