package models

// 가장 기본적인 서버 응답
type ResponseCommon struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Result  interface{} `json:"result"`
}

// 게시판 테이블 정의
type Table string

// 게시판 테이블 이름들 정리
const (
	TABLE_BOARD         Table = "board"
	TABLE_BOARD_CAT     Table = "board_category"
	TABLE_CHAT          Table = "chat"
	TABLE_COMMENT       Table = "comment"
	TABLE_COMMENT_LIKE  Table = "comment_like"
	TABLE_EXIF          Table = "exif"
	TABLE_FILE          Table = "file"
	TABLE_FILE_THUMB    Table = "file_thumbnail"
	TABLE_GROUP         Table = "group"
	TABLE_HASHTAG       Table = "hashtag"
	TABLE_IMAGE         Table = "image"
	TABLE_IMAGE_DESC    Table = "image_description"
	TABLE_NOTI          Table = "notification"
	TABLE_POINT_HISTORY Table = "point_history"
	TABLE_POST          Table = "post"
	TABLE_POST_HASHTAG  Table = "post_hashtag"
	TABLE_POST_LIKE     Table = "post_like"
	TABLE_REPORT        Table = "report"
	TABLE_USER          Table = "user"
	TABLE_USER_ACCESS   Table = "user_access_log"
	TABLE_USER_BLOCK    Table = "user_black_list"
	TABLE_USER_PERM     Table = "user_permission"
	TABLE_USER_TOKEN    Table = "user_token"
	TABLE_USER_VERIFY   Table = "user_verification"
)
