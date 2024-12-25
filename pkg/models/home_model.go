package models

import "html/template"

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
type HomePostParameter struct {
	SinceUid uint
	Bunch    uint
	Option   Search
	Keyword  string
	UserUid  uint
	BoardUid uint
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

// 사이트맵 구조체 정의
type HomeSitemapURL struct {
	Loc        string
	LastMod    string
	ChangeFreq string
	Priority   string
}

// SEO 메인화면에 출력할 구조체 정의
type HomeMainPage struct {
	PageTitle string
	PageUrl   string
	Version   string
	Articles  []HomeMainArticle
}

// SEO 메인화면에 보여줄 article 구조체 정의
type HomeMainArticle struct {
	Cover    string
	Content  template.HTML
	Comments []HomeMainComment
	Date     string
	Hashtags []Pair
	Like     uint
	Name     string
	Title    string
	Url      string
}

// SEO 메인화면에 보여줄 댓글 구조체 정의
type HomeMainComment struct {
	Content template.HTML
	Date    string
	Like    uint
	Name    string
}
