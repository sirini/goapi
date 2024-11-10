package models

// 게시판 타입 정의
type Board uint8

// 게시판 타입 목록
const (
	BOARD_DEFAULT Board = iota
	BOARD_GALLERY
	BOARD_BLOG
	BOARD_SHOP
)

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

// 게시글 작성자 타입 정의
type BoardWriter struct {
	UserBasicInfo
	Signature string `json:"signature"`
}

// 홈화면 게시글 공통 리턴 타입 정의
type BoardCommonPostItem struct {
	Uid       uint   `json:"uid"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Submitted uint64 `json:"submitted"`
	Modified  uint64 `json:"modified"`
	Hit       uint   `json:"hit"`
	Status    Status `json:"status"`
}

// 게시글 목록보기에 추가로 필요한 리턴 타입 정의
type BoardCommonListItem struct {
	Category Pair         `json:"category"`
	Cover    string       `json:"cover"`
	Comment  uint         `json:"comment"`
	Like     uint         `json:"like"`
	Liked    bool         `json:"liked"`
	Writer   *BoardWriter `json:"writer"`
}

// 게시글 목록보기용 리턴 타입 정의
type BoardListItem struct {
	BoardCommonPostItem
	BoardCommonListItem
}

// 게시판 목록 페이징 이동 방향 정의
type Paging int8

const (
	PAGE_PREV Paging = -1
	PAGE_NEXT Paging = 1
)

// 페이징 방향 반환 시 쿼리에 사용할 문자열 반환
func (p Paging) Query() (string, string) {
	switch p {
	case PAGE_PREV:
		return ">", "ASC"
	default:
		return "<", "DESC"
	}
}

// 활동별로 필요한 포인트 정의
type BoardActionPoint struct {
	View     int `json:"view"`
	Write    int `json:"write"`
	Comment  int `json:"comment"`
	Download int `json:"download"`
}

// 활동별로 필요한 레벨 정의
type BoardActionLevel struct {
	BoardActionPoint
	List int `json:"list"`
}

// 게시판 설정 타입 정의
type BoardConfig struct {
	Uid   uint `json:"uid"`
	Admin struct {
		Group uint `json:"group"`
		Board uint `json:"board"`
	} `json:"admin"`
	Type        Board            `json:"type"`
	Name        string           `json:"name"`
	Info        string           `json:"info"`
	RowCount    uint             `json:"rowCount"`
	Width       uint             `json:"width"`
	UseCategory bool             `json:"useCategory"`
	Category    []Pair           `json:"category"`
	Level       BoardActionLevel `json:"level"`
	Point       BoardActionPoint `json:"point"`
}

// 게시글 가져오기 시 필요한 파라미터 정의
type BoardListParameter struct {
	HomePostParameter
	Page        uint
	Direction   Paging
	NoticeCount uint
}

// 게시글 목록보기 리턴 값 정의
type BoardListResult struct {
	TotalPostCount uint             `json:"totalPostCount"`
	Config         *BoardConfig     `json:"config"`
	Posts          []*BoardListItem `json:"posts"`
	BlackList      []uint           `json:"blackList"`
	IsAdmin        bool             `json:"isAdmin"`
}
