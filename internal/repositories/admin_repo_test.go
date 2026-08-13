package repositories

import (
	"reflect"
	"testing"

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
