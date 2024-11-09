package models

// 게시글 상태 정의
type Status int8

// 게시글 상태들
const (
	POST_REMOVED Status = -1 + iota
	POST_NORMAL
	POST_NOTICE
	POST_SECRET
)

// 검색 옵션 정의
type Search uint8

// 검색 옵션들
const (
	SEARCH_TITLE Search = iota
	SEARCH_CONTENT
	SEARCH_WRITER
	SEARCH_TAG
	SEARCH_CATEGORY
	SEARCH_NONE
)

func (s Search) String() string {
	switch s {
	case SEARCH_CONTENT:
		return "content"
	case SEARCH_WRITER:
		return "user_uid"
	case SEARCH_CATEGORY:
		return "category_uid"
	case SEARCH_TAG:
		return "tag"
	default:
		return "title"
	}
}

// 게시판 기본 설정값들 반환 타입 정의
type BoardBasicSettingResult struct {
	Id          string
	Type        Board
	UseCategory bool
}

// 최근 게시글 가져올 때 필요한 파라미터 정의
type BoardPostParameter struct {
	SinceUid uint
	Bunch    uint
	Option   Search
	Keyword  string
	UserUid  uint
	BoardUid uint
}

// 게시글 작성자 타입 정의
type BoardWriter struct {
	UserBasicInfo
	Signature string `json:"signature"`
}

// 게시글 공통 리턴 타입 정의
type BoardCommonPostItem struct {
	Uid       uint   `json:"uid"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Submitted uint64 `json:"submitted"`
	Modified  uint64 `json:"modified"`
	Hit       uint   `json:"hit"`
	Status    Status `json:"status"`
}

// 최근 게시글 리턴 타입 정의
type BoardPostItem struct {
	BoardCommonPostItem
	BoardUid    uint `json:"boardUid"`
	UserUid     uint `json:"userUid"`
	CategoryUid uint `json:"categoryUid"`
}

// 최근 게시글들 최종 리턴 타입 정의
type BoardFinalPostItem struct {
	BoardCommonPostItem
	Id          string       `json:"id"`
	Type        Board        `json:"type"`
	UseCategory bool         `json:"useCategory"`
	Category    string       `json:"category"`
	Cover       string       `json:"cover"`
	Comment     uint         `json:"comment"`
	Writer      *BoardWriter `json:"writer"`
	Like        uint         `json:"like"`
	Liked       bool         `json:"liked"`
}
