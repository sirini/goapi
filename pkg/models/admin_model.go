package models

import (
	"sync"
	"time"
)

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

// 게시판 설정 반환값 정의
type AdminBoardResult struct {
	Config BoardConfig `json:"config"`
	Groups []Pair      `json:"groups"`
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

// 대시보드에서 볼 업로드 사용량 캐시
type UploadUsageCache struct {
	mu          sync.RWMutex
	TotalBytes  uint64
	LastUpdated time.Time
}

// 사용량 구조체 노출
var AdminUploadUsage = &UploadUsageCache{}

// 업로드 사용량 업데이트
func (s *UploadUsageCache) Update(bytes uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalBytes = bytes
	s.LastUpdated = time.Now()
}

// 캐시 읽기 함수
func (s *UploadUsageCache) Get() (uint64, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TotalBytes, s.LastUpdated
}

// 캐시 만료 여부 확인
func (s *UploadUsageCache) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastUpdated.IsZero() || time.Since(s.LastUpdated) > 1*time.Hour
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

// 그룹 관리화면 게시판 (및 통계) 목록 반환값 정의
type AdminGroupBoardItem struct {
	AdminGroupConfig
	Id    string                `json:"id"`
	Type  Board                 `json:"type"`
	Name  string                `json:"name"`
	Info  string                `json:"info"`
	Total AdminGroupBoardStatus `json:"total"`
}

// 게시판 별 간단 통계 반환값 정의
type AdminGroupBoardStatus struct {
	Post    uint `json:"post"`
	Comment uint `json:"comment"`
	File    uint `json:"file"`
	Image   uint `json:"image"`
}

// 그룹 설정 및 소속 게시판들 정보 반환값 정의
type AdminGroupListResult struct {
	Config AdminGroupConfig      `json:"config"`
	Boards []AdminGroupBoardItem `json:"boards"`
}

// 그룹 관리화면 일반 설정들 반환값 정의
type AdminGroupConfig struct {
	Uid     uint        `json:"uid"`
	Id      string      `json:"id"`
	Count   uint        `json:"count"`
	Manager BoardWriter `json:"manager"`
}

// 최근 (댓)글 출력에 필요한 공통 반환값 정의
type AdminLatestCommon struct {
	Uid    uint        `json:"uid"`
	Id     string      `json:"id"`
	Type   Board       `json:"type"`
	Like   uint        `json:"like"`
	Date   uint64      `json:"date"`
	Status Status      `json:"status"`
	Writer BoardWriter `json:"writer"`
}

// 최근 댓글 반환값 정의
type AdminLatestComment struct {
	AdminLatestCommon
	Content string `json:"content"`
	PostUid uint   `json:"postUid"`
}

// 최근 댓글 및 max uid 반환값 정의
type AdminLatestCommentResult struct {
	Comments []AdminLatestComment `json:"comments"`
	MaxUid   uint                 `json:"maxUid"`
}

// (댓)글 검색하기에 필요한 파라미터 정의
type AdminLatestParam struct {
	Page    uint   `json:"page"`
	Limit   uint   `json:"limit"`
	Option  Search `json:"option"`
	Keyword string `json:"keyword"`
}

// 신고 목록 검색하기에 필요한 파라미터 정의
type AdminReportParam struct {
	AdminLatestParam
	IsSolved bool
}

// 최근 글 반환값 정의
type AdminLatestPost struct {
	AdminLatestCommon
	Title   string `json:"title"`
	Comment uint   `json:"comment"`
	Hit     uint   `json:"hit"`
}

// 최근 글 및 max uid 반환값 정의
type AdminLatestPostResult struct {
	Posts  []AdminLatestPost `json:"posts"`
	MaxUid uint              `json:"maxUid"`
}

// 신고 목록 반환값 정의
type AdminReportItem struct {
	Uid      uint        `json:"uid"`
	To       BoardWriter `json:"to"`
	From     BoardWriter `json:"from"`
	Request  string      `json:"request"`
	Response string      `json:"response"`
	Date     uint64      `json:"date"`
}

// 신고 목록 및 max uid 반환값 정의
type AdminReportResult struct {
	Reports []AdminReportItem `json:"reports"`
	MaxUid  uint              `json:"maxUid"`
}

// 사용자 목록 검색하기에 필요한 파라미터 정의
type AdminUserParam struct {
	AdminLatestParam
	IsBlocked bool
}

// 사용자 목록 검색하기 반환값 정의
type AdminUserItem struct {
	UserBasicInfo
	Id     string `json:"id"`
	Level  uint   `json:"level"`
	Point  uint   `json:"point"`
	Signup uint64 `json:"signup"`
}

// 사용자 목록 검색 결과 및 max uid 반환값 정의
type AdminUserItemResult struct {
	User   []AdminUserItem `json:"user"`
	MaxUid uint            `json:"maxUid"`
}

// 사용자 정보 반환값 정의
type AdminUserInfo struct {
	BoardWriter
	Id    string `json:"id"`
	Level uint   `json:"level"`
	Point uint   `json:"point"`
}
