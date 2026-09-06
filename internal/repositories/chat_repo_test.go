package repositories

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirini/goapi/internal/configs"
)

func TestChatListQuerySelectsCompleteLatestRow(t *testing.T) {
	query := chatListQuery("nubo_")
	if strings.Contains(query, "MAX(c.message)") || strings.Contains(query, "MAX(c.timestamp)") {
		t.Fatalf("chat list query combines values from different rows: %s", query)
	}
	if !strings.Contains(query, "latest.latest_uid = c.uid") {
		t.Fatalf("chat list query does not join the latest message row: %s", query)
	}
}

func withChatRepositoryTestPrefix(t *testing.T) {
	t.Helper()
	previous := configs.Env.Prefix
	configs.Env.Prefix = "nubo_"
	t.Cleanup(func() { configs.Env.Prefix = previous })
}

func TestLoadChatHistoryIncludesReadTimestamp(t *testing.T) {
	withChatRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repository := NewNuboChatRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(chatHistoryQuery("nubo_"))).
		WithArgs(uint(9), uint(7), uint(7), uint(9), uint(100)).
		WillReturnRows(sqlmock.NewRows([]string{"uid", "from_uid", "message", "timestamp", "read_at"}).
			AddRow(42, 7, "보낸 메시지", 1000, 1250).
			AddRow(41, 9, "받은 메시지", 900, 0))

	history, err := repository.LoadChatHistory(7, 9, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Uid != 41 || history[1].ReadAt != 1250 {
		t.Fatalf("unexpected chronological chat history: %+v", history)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkChatReadScopesUpdateToReceivedMessages(t *testing.T) {
	withChatRepositoryTestPrefix(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repository := NewNuboChatRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(chatReadQuery("nubo_"))).
		WithArgs(uint64(1500), uint(7), uint(9), uint(42)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	updated, err := repository.MarkChatRead(7, 9, 42, 1500)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 2 {
		t.Fatalf("updated rows = %d, want 2", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
