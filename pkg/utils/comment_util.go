package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirini/goapi/pkg/models"
)

// 새 댓글 및 답글 작성 시 파라미터 체크
func CheckCommentParameters(r *http.Request) (models.CommentWriteParameter, error) {
	result := models.CommentWriteParameter{}
	actionUserUid := GetUserUidFromToken(r)
	boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
	if err != nil {
		return result, err
	}
	postUid, err := strconv.ParseUint(r.FormValue("postUid"), 10, 32)
	if err != nil {
		return result, err
	}
	content := Sanitize(r.FormValue("content"))
	if len(content) < 2 {
		return result, fmt.Errorf("invalid content, too short")
	}

	result = models.CommentWriteParameter{
		BoardUid: uint(boardUid),
		PostUid:  uint(postUid),
		UserUid:  actionUserUid,
		Content:  content,
	}
	return result, nil
}
