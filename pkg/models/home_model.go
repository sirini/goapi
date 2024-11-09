package models

// 버전 응답 구조체
type HomeVisitResult struct {
	Success         bool   `json:"success"`
	OfficialWebsite string `json:"officialWebsite"`
	Version         string `json:"version"`
	License         string `json:"license"`
	Github          string `json:"github"`
}

// 게시판 타입 정의
type Board uint8

// 게시판 타입 목록
const (
	BOARD_DEFAULT Board = iota
	BOARD_GALLERY
	BOARD_BLOG
	BOARD_SHOP
)

// 홈 사이드바에 출력할 게시판 목록 형태 정의
type HomeSidebarBoardResult struct {
	Id   string `json:"id"`
	Type Board  `json:"type"`
	Name string `json:"name"`
	Info string `json:"info"`
}

// 홈 사이드바에 출력할 그룹 목록 형태 정의
type HomeSidebarGroupResult struct {
	Group  uint                      `json:"group"`
	Boards []*HomeSidebarBoardResult `json:"boards"`
}
