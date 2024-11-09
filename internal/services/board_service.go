package services

import "github.com/sirini/goapi/internal/repositories"

type BoardService interface {
	GetBoardUid(id string) uint
	GetMaxUid() uint
}

type TsboardBoardService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardBoardService(repos *repositories.Repository) *TsboardBoardService {
	return &TsboardBoardService{repos: repos}
}

// 게시판 고유 번호 가져오기
func (s *TsboardBoardService) GetBoardUid(id string) uint {
	return s.repos.Board.GetBoardUidById(id)
}

// 게시글 최대 고유번호 반환
func (s *TsboardBoardService) GetMaxUid() uint {
	return s.repos.Board.GetMaxUid()
}
