package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
)

// 새 댓글 및 답글 작성 시 파라미터 체크
func CheckCommentParameters(c fiber.Ctx) (models.CommentWriteParameter, error) {
	result := models.CommentWriteParameter{}
	actionUserUid := ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return result, err
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return result, err
	}

	content := Sanitize(c.FormValue("content"))
	content = CutString(content, 9999)
	if len(content) < 2 {
		return result, fmt.Errorf("invalid content, too short")
	}

	result = models.CommentWriteParameter{
		BoardUid: uint(boardUid),
		PostUid:  uint(postUid),
		UserUid:  uint(actionUserUid),
		Content:  content,
	}
	return result, nil
}
