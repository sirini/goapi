package models

// 버전 응답 구조체
type HomeVisitResult struct {
	Success         bool   `json:"success"`
	OfficialWebsite string `json:"officialWebsite"`
	Version         string `json:"version"`
	License         string `json:"license"`
	Github          string `json:"github"`
}

// 홈 사이드바에 출력할 게시판 목록 형태 정의
type HomeSidebarBoardResult struct {
	Id   string `json:"id"`
	Type Board  `json:"type"`
	Name string `json:"name"`
	Info string `json:"info"`
}

// 최근 게시글 가져올 때 필요한 파라미터 정의
type HomePostParam struct {
	SinceUid uint   `json:"sinceUid"`
	Bunch    uint   `json:"bunch"`
	Option   Search `json:"option"`
	Keyword  string `json:"keyword"`
	UserUid  uint   `json:"userUid"`
	BoardUid uint   `json:"boardUid"`
}

// 홈 사이드바에 출력할 그룹 목록 형태 정의
type HomeSidebarGroupResult struct {
	Group  string                   `json:"group"`
	Boards []HomeSidebarBoardResult `json:"boards"`
}

// 최근 게시글들 최종 리턴 타입 정의
type BoardHomePostItem struct {
	BoardCommonPostItem
	BoardCommonListItem
	Id          string `json:"id"`
	Type        Board  `json:"type"`
	UseCategory bool   `json:"useCategory"`
}

// 최근 게시글들 최종 리턴 타입 및 게시판 정보 정의
type BoardHomePostResult struct {
	Items  []BoardHomePostItem `json:"items"`
	Config BoardConfig         `json:"config"`
}

// 최근 게시글 리턴 타입 정의
type HomePostItem struct {
	BoardCommonPostItem
	BoardUid    uint `json:"boardUid"`
	UserUid     uint `json:"userUid"`
	CategoryUid uint `json:"categoryUid"`
}
