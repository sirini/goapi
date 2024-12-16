package models

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
