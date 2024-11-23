package models

// 댓글 목록 가져오기에 필요한 파라미터 정의
type CommentListParameter struct {
	BoardUid  uint
	PostUid   uint
	UserUid   uint
	Page      uint
	Bunch     uint
	SinceUid  uint
	Direction Paging
}

// 댓글 내용 항목 정의
type CommentItem struct {
	Uid       uint        `json:"uid"`
	ReplyUid  uint        `json:"replyUid"`
	PostUid   uint        `json:"postUid"`
	Writer    BoardWriter `json:"writer"`
	Like      uint        `json:"like"`
	Liked     bool        `json:"liked"`
	Submitted uint64      `json:"submitted"`
	Modified  uint64      `json:"modified"`
	Status    Status      `json:"status"`
	Content   string      `json:"content"`
}

// 댓글 목록 가져오기 결과 정의
type CommentListResult struct {
	BoardUid          uint          `json:"boardUid"`
	SinceUid          uint          `json:"sinceUid"`
	TotalCommentCount uint          `json:"totalCommentCount"`
	Comments          []CommentItem `json:"comments"`
}

// 댓글에 좋아요 처리에 필요한 파라미터 정의
type CommentLikeParameter struct {
	BoardUid   uint
	CommentUid uint
	UserUid    uint
	Liked      bool
}

// 새 댓글 작성하기에 필요한 파라미터 정의
type CommentWriteParameter struct {
	BoardUid uint
	PostUid  uint
	UserUid  uint
	Content  string
}
