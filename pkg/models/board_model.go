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
	CONTENT_REMOVED Status = -1 + iota
	CONTENT_NORMAL
	CONTENT_NOTICE
	CONTENT_SECRET
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
	Category Pair        `json:"category"`
	Cover    string      `json:"cover"`
	Comment  uint        `json:"comment"`
	Like     uint        `json:"like"`
	Liked    bool        `json:"liked"`
	Writer   BoardWriter `json:"writer"`
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
	TotalPostCount uint            `json:"totalPostCount"`
	Config         BoardConfig     `json:"config"`
	Posts          []BoardListItem `json:"posts"`
	BlackList      []uint          `json:"blackList"`
	IsAdmin        bool            `json:"isAdmin"`
}

// 사용자의 포인트 변경하기에 필요한 파라미터 정의
type ChangeUserPointParameter struct {
	BoardUid uint
	UserUid  uint
	Action   PointAction
}

// 게시판 관련 액션 정의
type BoardAction uint

// 게시판 관련 액션들
const (
	BOARD_ACTION_LIST BoardAction = iota
	BOARD_ACTION_VIEW
	BOARD_ACTION_COMMENT
	BOARD_ACTION_WRITE
	BOARD_ACTION_DOWNLOAD
)

// 게시판 액션들 문자로 변환
func (ba BoardAction) String() string {
	switch ba {
	case BOARD_ACTION_VIEW:
		return "view"
	case BOARD_ACTION_COMMENT:
		return "comment"
	case BOARD_ACTION_WRITE:
		return "write"
	case BOARD_ACTION_DOWNLOAD:
		return "download"
	default:
		return "list"
	}
}

// 게시판 첨부파일 구조체 정의
type BoardAttachment struct {
	Pair
	Size uint `json:"size"`
}

// 파일 기본 구조 정의
type BoardFile struct {
	Uid  uint   `json:"uid"`
	Path string `json:"path"`
}

// 썸네일 크기별 종류 정의
type BoardThumbnail struct {
	Large string `json:"large"`
	Small string `json:"small"`
}

// 게시판 첨부 이미지 구조체 정의
type BoardAttachedImage struct {
	File        BoardFile      `json:"file"`
	Thumbnail   BoardThumbnail `json:"thumbnail"`
	Exif        BoardExif      `json:"exif"`
	Description string         `json:"description"`
}

// EXIF 구조체 정의
type BoardExif struct {
	Make        string `json:"make"`
	Model       string `json:"model"`
	Aperture    uint   `json:"aperture"`
	ISO         uint   `json:"iso"`
	FocalLength uint   `json:"focalLength"`
	Exposure    uint   `json:"exposure"`
	Width       uint   `json:"width"`
	Height      uint   `json:"height"`
	Date        uint64 `json:"date"`
}

// 게시글 작성자의 최근 글/댓글에 전달할 게시판 기본 설정값 정의
type BoardBasicConfig struct {
	Id   string `json:"id"`
	Type Board  `json:"type"`
	Name string `json:"name"`
}

// 게시글 작성자의 최근 글/댓글 공통 요소 정의
type BoardWriterLatestCommon struct {
	Board     BoardBasicConfig `json:"board"`
	PostUid   uint             `json:"postUid"`
	Like      uint             `json:"like"`
	Submitted uint             `json:"submitted"`
}

// 게시글 작성자의 최근 댓글 정의
type BoardWriterLatestComment struct {
	BoardWriterLatestCommon
	Content string `json:"content"`
}

// 게시글 작성자의 최근 글 정의
type BoardWriterLatestPost struct {
	BoardWriterLatestCommon
	Comment uint   `json:"comment"`
	Title   string `json:"title"`
}

// 게시글 보기에서 공통으로 쓰이는 파라미터 정의
type BoardViewCommonParameter struct {
	BoardUid uint
	PostUid  uint
	UserUid  uint
}

// 첨부파일 다운로드 결과 정의
type BoardViewDownloadResult struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// 게시글 보기에 필요한 파라미터 정의
type BoardViewParameter struct {
	BoardViewCommonParameter
	UpdateHit bool
	Limit     uint
}

// 게시글 보기에 반환 타입 정의
type BoardViewResult struct {
	Config         BoardConfig                `json:"config"`
	Post           BoardListItem              `json:"post"`
	Images         []BoardAttachedImage       `json:"images"`
	Files          []BoardAttachment          `json:"files"`
	Tags           []Pair                     `json:"tags"`
	PrevPostUid    uint                       `json:"prevPostUid"`
	NextPostUid    uint                       `json:"nextPostUid"`
	WriterPosts    []BoardWriterLatestPost    `json:"writerPosts"`
	WriterComments []BoardWriterLatestComment `json:"writerComments"`
}

// 게시글 좋아하기에 필요한 파라미터 정의
type BoardViewLikeParameter struct {
	BoardViewCommonParameter
	Liked bool
}

// 게시글 이동에 필요한 파라미터 정의
type BoardMovePostParameter struct {
	BoardViewCommonParameter
	TargetBoardUid uint
}

// 게시글 이동에 필요한 게시판 목록 타입 정의
type BoardItem struct {
	Pair
	Info string `json:"info"`
}

// 에디터에서 게시판 설정 및 카테고리 불러오기 결과 타입 정의
type EditorConfigResult struct {
	Config  BoardConfig `json:"config"`
	IsAdmin bool        `json:"isAdmin"`
}

// 게시글에 삽입한 이미지 목록 가져오는 파라미터 정의
type EditorInsertImageParameter struct {
	BoardUid uint
	LastUid  uint
	UserUid  uint
	Bunch    uint
}

// 게시글에 삽입한 이미지 목록 반환 타입 정의
type EditorInsertImageResult struct {
	Images          []Pair `json:"images"`
	MaxImageUid     uint   `json:"maxImageUid"`
	TotalImageCount uint   `json:"totalImageCount"`
}

// 갤러리 그리드형 반환타입 정의
type GalleryGridItem struct {
	Uid     uint                 `json:"uid"`
	Like    uint                 `json:"like"`
	Liked   bool                 `json:"liked"`
	Writer  BoardWriter          `json:"writer"`
	Comment uint                 `json:"comment"`
	Title   string               `json:"title"`
	Images  []BoardAttachedImage `json:"images"`
}

// 갤러리 사진 보기 반환 타입 정의
type GalleryPhotoResult struct {
	Config BoardConfig          `json:"config"`
	Images []BoardAttachedImage `json:"images"`
}

// 갤러리 리스트 반환 타입 정의
type GalleryListResult struct {
	Config         BoardConfig       `json:"config"`
	Images         []GalleryGridItem `json:"images"`
	TotalPostCount uint              `json:"totalPostCount"`
}
