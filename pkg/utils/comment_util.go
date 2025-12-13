package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
)

// 새 댓글 및 답글 작성 시 파라미터 체크
func CheckCommentParams(c fiber.Ctx) (models.CommentWriteParam, error) {
	result := models.CommentWriteParam{}
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

	result = models.CommentWriteParam{
		BoardUid: uint(boardUid),
		PostUid:  uint(postUid),
		UserUid:  uint(actionUserUid),
		Content:  content,
	}
	return result, nil
}
