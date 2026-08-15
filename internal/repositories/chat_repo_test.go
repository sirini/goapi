package repositories

import (
	"strings"
	"testing"
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
