package repositories

import (
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

// 목록·상세 SQL에 끼워 넣는 종류별 리액션 집계 서브셀렉트를 만든다.
// targetCol은 집계 대상 컬럼(post_uid/comment_uid), targetExpr은 해당 SQL의 대상 식(예: p.uid, c.uid)이다.
func reactionCountSubselects(table models.Table, targetCol string, targetExpr string) string {
	fragments := make([]string, 0, len(models.ReactionCodeList))
	for _, code := range models.ReactionCodeList {
		fragments = append(fragments, fmt.Sprintf(
			"(SELECT COUNT(*) FROM %s%s WHERE %s = %s AND reaction_type = %d)",
			configs.Env.Prefix, table, targetCol, targetExpr, code,
		))
	}
	return strings.Join(fragments, ",\n\t\t\t")
}

// 대상 SQL 묶음(예: post_uid IN (...))에 쓰는 종류별 집계 SUM 목록을 만든다.
func reactionSumColumns(table models.Table, targetCol string) string {
	fragments := make([]string, 0, len(models.ReactionCodeList))
	for _, code := range models.ReactionCodeList {
		fragments = append(fragments, fmt.Sprintf("SUM(%s = %d)", targetCol, code))
	}
	return strings.Join(fragments, ", ")
}
