package models

// 대시보드 통계 추출 시 필요한 컬럼 타입 정의
type StatisticColumn uint8

// 통계용 컬럼 타입 목록
const (
	COLUMN_TIMESTAMP StatisticColumn = iota
	COLUMN_SIGNUP
	COLUMN_SUBMITTED
)

func (s StatisticColumn) String() string {
	switch s {
	case COLUMN_SIGNUP:
		return "signup"
	case COLUMN_SUBMITTED:
		return "submitted"
	default:
		return "timestamp"
	}
}

// 게시판 생성 시 기본값 정의
const (
	CREATE_BOARD_ADMIN       = 1
	CREATE_BOARD_TYPE        = 0 /* board */
	CREATE_BOARD_NAME        = "board name"
	CREATE_BOARD_INFO        = "description for this board"
	CREATE_BOARD_ROWS        = 15
	CREATE_BOARD_WIDTH       = 1000
	CREATE_BOARD_USE_CAT     = 1
	CREATE_BOARD_LV_LIST     = 0
	CREATE_BOARD_LV_VIEW     = 0
	CREATE_BOARD_LV_WRITE    = 1 /* 0 is not allowed */
	CREATE_BOARD_LV_COMMENT  = 1 /* 0 is not allowed */
	CREATE_BOARD_LV_DOWNLOAD = 1 /* 0 is not allowed */
	CREATE_BOARD_PT_VIEW     = 0
	CREATE_BOARD_PT_WRITE    = 5
	CREATE_BOARD_PT_COMMENT  = 2
	CREATE_BOARD_PT_DOWNLOAD = -10
)

// 그룹 생성 시 기본값 정의
const CREATE_GROUP_ADMIN = 1

// 게시판 레벨 제한 반환값 정의
type AdminBoardLevelPolicy struct {
	Uid   uint             `json:"uid"`
	Admin BoardWriter      `json:"admin"`
	Level BoardActionLevel `json:"level"`
}

// 게시판 포인트 정책 반환값 정의
type AdminBoardPointPolicy struct {
	Uid uint `json:"uid"`
	BoardActionPoint
}

// 게시판 생성하기 시 반환값 정의
type AdminCreateBoardResult struct {
	Uid     uint   `json:"uid"`
	Type    Board  `json:"type"`
	Name    string `json:"name"`
	Info    string `json:"info"`
	Manager Pair   `json:"manager"`
}

// 대시보드 아이템(그룹, 게시판, 회원 최신순 목록) 반환값 정의
type AdminDashboardItem struct {
	Groups  []Pair        `json:"groups"`
	Boards  []Pair        `json:"boards"`
	Members []BoardWriter `json:"members"`
}

// 대시보드 최근 (댓)글 목록 반환값 정의
type AdminDashboardLatestContent struct {
	AdminDashboardReport
	Id   string `json:"id"`
	Type Board  `json:"type"`
}

// 대시보드 최근 신고 목록 반환값 정의
type AdminDashboardReport struct {
	Uid     uint        `json:"uid"`
	Content string      `json:"content"`
	Writer  BoardWriter `json:"writer"`
}

// 대시보드 최근 (댓)글, 신고 목록 최신순 반환값 정의
type AdminDashboardLatest struct {
	Posts    []AdminDashboardLatestContent `json:"posts"`
	Comments []AdminDashboardLatestContent `json:"comments"`
	Reports  []AdminDashboardReport        `json:"reports"`
}

// 대시보드 최근 통계들 반환값 정의
type AdminDashboardStatisticResult struct {
	Visit  AdminDashboardStatistic `json:"visit"`
	Member AdminDashboardStatistic `json:"member"`
	Post   AdminDashboardStatistic `json:"post"`
	Reply  AdminDashboardStatistic `json:"reply"`
	File   AdminDashboardStatistic `json:"file"`
	Image  AdminDashboardStatistic `json:"image"`
}

// 대시보드 최근 통계 반환값 정의
type AdminDashboardStatistic struct {
	History []AdminDashboardStatus `json:"history"`
	Total   uint                   `json:"total"`
}

// 대시보드 일자별 데이터 반환값 정의
type AdminDashboardStatus struct {
	Date  uint64 `json:"date"`
	Visit uint   `json:"visit"`
}

// 그룹 관리화면 일반 설정들 반환값 정의
type AdminGroupConfig struct {
	Uid     uint        `json:"uid"`
	Count   uint        `json:"count"`
	Manager BoardWriter `json:"manager"`
}

// 그룹 목록용 반환값 정의
type AdminGroupItem struct {
	AdminGroupConfig
	Id string `json:"id"`
}
