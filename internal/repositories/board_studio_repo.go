package repositories

import (
	"fmt"
	"path"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

func studioOrderClause(sort models.BoardStudioSort) (string, error) {
	switch sort {
	case models.BOARD_STUDIO_SORT_RECENT:
		return "p.submitted DESC, p.uid DESC", nil
	case models.BOARD_STUDIO_SORT_VIEWS:
		return "p.hit DESC, p.uid DESC", nil
	case models.BOARD_STUDIO_SORT_LIKES:
		return "like_count DESC, p.uid DESC", nil
	case models.BOARD_STUDIO_SORT_COMMENTS:
		return "comment_count DESC, p.uid DESC", nil
	default:
		return "", fmt.Errorf("invalid studio sort")
	}
}

// 스튜디오 cover는 기존에 생성된 작은 preview thumbnail 공개 경로만 허용한다.
func studioCoverPath(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "./upload/thumbnails/") {
		value = strings.TrimPrefix(value, ".")
	}
	if !strings.HasPrefix(value, "/upload/thumbnails/") {
		return ""
	}
	cleaned := path.Clean(value)
	if !strings.HasPrefix(cleaned, "/upload/thumbnails/") {
		return ""
	}
	return cleaned
}

func studioImageCountExpression(prefix string) string {
	return fmt.Sprintf(`(
		SELECT COUNT(*) FROM %s%s AS f
		WHERE f.post_uid = p.uid
		AND EXISTS (
			SELECT 1 FROM %s%s AS image_thumb
			WHERE image_thumb.file_uid = f.uid
		)
	)`, prefix, models.TABLE_FILE, prefix, models.TABLE_FILE_THUMB)
}

func studioLikeCountExpression(prefix string) string {
	return fmt.Sprintf(`(
		SELECT COUNT(*) FROM %s%s AS pl
		WHERE pl.post_uid = p.uid AND pl.liked = 1
	)`, prefix, models.TABLE_POST_LIKE)
}

func studioCommentCountExpression(prefix string) string {
	return fmt.Sprintf(`(
		SELECT COUNT(*) FROM %s%s AS c
		WHERE c.post_uid = p.uid AND c.status != %d
	)`, prefix, models.TABLE_COMMENT, models.CONTENT_REMOVED)
}

// GetStudio는 지정 사용자와 게시판에 한정된 작품 집계와 한 페이지를 반환한다.
// 호출자가 인증 주체를 결정하므로 공개 사용자 작품 API도 같은 쿼리를 재사용할 수 있다.
func (r *NuboBoardRepository) GetStudio(param models.BoardStudioParam) (models.BoardStudioResult, error) {
	result := models.BoardStudioResult{
		Posts: models.BoardStudioPosts{
			Page:  param.Page,
			Limit: param.Limit,
			Items: make([]models.BoardStudioPostItem, 0),
		},
	}
	if param.Page < 1 || param.Limit < 1 {
		return result, fmt.Errorf("invalid studio pagination")
	}

	orderClause, err := studioOrderClause(param.Sort)
	if err != nil {
		return result, err
	}

	// 공개 통계는 비공개 작품을 포함하지 않는다. 본인 스튜디오의 기존 범위는 유지한다.
	secondStatus := models.CONTENT_SECRET
	if param.PublicOnly {
		secondStatus = models.CONTENT_NORMAL
	}
	prefix := configs.Env.Prefix
	imageCount := studioImageCountExpression(prefix)
	likeCount := studioLikeCountExpression(prefix)
	commentCount := studioCommentCountExpression(prefix)

	summaryQuery := fmt.Sprintf(`SELECT
		COUNT(*),
		COALESCE(SUM(studio.image_count), 0),
		COALESCE(SUM(studio.hit), 0),
		COALESCE(SUM(studio.like_count), 0),
		COALESCE(SUM(studio.comment_count), 0)
	FROM (
		SELECT p.uid, p.hit,
			%s AS image_count,
			%s AS like_count,
			%s AS comment_count
		FROM %s%s AS p
		WHERE p.board_uid = ? AND p.user_uid = ? AND p.status IN (?, ?)
	) AS studio`, imageCount, likeCount, commentCount, prefix, models.TABLE_POST)

	err = r.db.QueryRow(summaryQuery,
		param.BoardUid,
		param.UserUid,
		models.CONTENT_NORMAL,
		secondStatus,
	).Scan(
		&result.Summary.PostCount,
		&result.Summary.PhotoCount,
		&result.Summary.ViewCount,
		&result.Summary.LikeCount,
		&result.Summary.CommentCount,
	)
	if err != nil {
		return result, err
	}

	if param.SummaryOnly {
		return result, nil
	}
	result.Posts.TotalCount = result.Summary.PostCount
	result.Posts.HasNext = uint64(param.Page)*uint64(param.Limit) < result.Posts.TotalCount
	offset := uint64(param.Page-1) * uint64(param.Limit)

	postsQuery := fmt.Sprintf(`SELECT
		p.uid,
		p.title,
		COALESCE((
			SELECT ft.path FROM %s%s AS ft
			WHERE ft.post_uid = p.uid
			ORDER BY ft.uid ASC LIMIT 1
		), ''),
		p.submitted,
		p.modified,
		p.status,
		%s AS image_count,
		p.hit,
		%s AS like_count,
		%s AS comment_count
	FROM %s%s AS p
	WHERE p.board_uid = ? AND p.user_uid = ? AND p.status IN (?, ?)
	ORDER BY %s
	LIMIT ? OFFSET ?`,
		prefix, models.TABLE_FILE_THUMB,
		imageCount,
		likeCount,
		commentCount,
		prefix, models.TABLE_POST,
		orderClause,
	)

	rows, err := r.db.Query(postsQuery,
		param.BoardUid,
		param.UserUid,
		models.CONTENT_NORMAL,
		secondStatus,
		param.Limit,
		offset,
	)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardStudioPostItem{}
		if err := rows.Scan(
			&item.Uid,
			&item.Title,
			&item.Cover,
			&item.Submitted,
			&item.Modified,
			&item.Status,
			&item.ImageCount,
			&item.Hit,
			&item.Like,
			&item.Comment,
		); err != nil {
			return result, err
		}
		item.Cover = studioCoverPath(item.Cover)
		result.Posts.Items = append(result.Posts.Items, item)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	return result, nil
}
