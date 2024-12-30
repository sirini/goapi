package models

// 첨부 파일 및 이미지 썸네일 반환 타입 정의
type SyncImageItem struct {
	Uid   uint      `json:"uid"`
	File  string    `json:"file"`
	Name  string    `json:"name"`
	Thumb string    `json:"thumb"`
	Full  string    `json:"full"`
	Desc  string    `json:"desc"`
	Exif  BoardExif `json:"exif"`
}

// 동기화 시킬 데이터 결과 타입 정의
type SyncPostItem struct {
	Id        string          `json:"id"`
	No        uint            `json:"no"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Submitted uint64          `json:"submitted"`
	Name      string          `json:"name"`
	Tags      []string        `json:"tags"`
	Images    []SyncImageItem `json:"images"`
}
