package repositories

import (
	"reflect"
	"strings"
	"testing"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

func TestUserListFilterUsesPlaceholders(t *testing.T) {
	query, args := userListFilter(models.AdminUserParam{
		AdminLatestParam: models.AdminLatestParam{Option: models.SEARCH_USER_NAME, Keyword: "%' OR 1=1 --"},
		IsBlocked:        true,
	})

	if query != "id != '' AND blocked = ? AND name LIKE ?" {
		t.Fatalf("unexpected filter query: %s", query)
	}
	want := []any{true, "%%' OR 1=1 --%"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected filter args: %#v", args)
	}
}

func TestRemoveBoardDataUsesForeignKeySafeTransactionOrder(t *testing.T) {
	state := &pointDriver{rowsAffected: 1}
	repo := NewNuboAdminRepository(openPointTestDB(t, state))
	if err := repo.RemoveBoardData(19); err != nil {
		t.Fatal(err)
	}
	if !state.committed || state.rolledBack {
		t.Fatalf("unexpected transaction state: committed=%v rolledBack=%v", state.committed, state.rolledBack)
	}

	positions := map[string]int{}
	for i, exec := range state.execs {
		for _, table := range []string{"file_thumbnail", "file", "comment_like", "comment", "post", "point_history", "board_category", "board"} {
			if strings.Contains(exec.query, "DELETE FROM "+configs.Env.Prefix+table+" ") {
				positions[table] = i
			}
		}
		if len(exec.args) != 1 || exec.args[0].Value != int64(19) {
			t.Fatalf("statement %d did not bind board uid: %#v", i, exec.args)
		}
	}
	assertBefore := func(parent, child string) {
		t.Helper()
		if positions[child] >= positions[parent] {
			t.Fatalf("%s must be deleted before %s: %#v", child, parent, positions)
		}
	}
	assertBefore("file", "file_thumbnail")
	assertBefore("comment", "comment_like")
	assertBefore("post", "comment")
	assertBefore("board_category", "post")
	assertBefore("board", "point_history")
	assertBefore("board", "board_category")
}

func TestUserListFilterAppliesLevelAndActiveStatus(t *testing.T) {
	query, args := userListFilter(models.AdminUserParam{
		AdminLatestParam: models.AdminLatestParam{Option: models.SEARCH_USER_LEVEL, Keyword: "7"},
	})

	if query != "id != '' AND blocked = ? AND level = ?" {
		t.Fatalf("unexpected filter query: %s", query)
	}
	want := []any{false, "7"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected filter args: %#v", args)
	}
}
