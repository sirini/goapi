package utils

import (
	"database/sql"

	"github.com/sirini/goapi/pkg/models"
)

// 홈화면에 보여줄 게시글 레코드들을 패킹해서 반환
func AppendItem(rows *sql.Rows) ([]*models.HomePostItem, error) {
	var items []*models.HomePostItem
	for rows.Next() {
		item := &models.HomePostItem{}
		err := rows.Scan(&item.Uid, &item.BoardUid, &item.UserUid, &item.CategoryUid,
			&item.Title, &item.Content, &item.Submitted, &item.Modified, &item.Hit, &item.Status)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
