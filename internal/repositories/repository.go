package repositories

import "database/sql"

// 모든 리포지토리들을 관리
type Repository struct {
	Auth      AuthRepository
	Board     BoardRepository
	BoardView BoardViewRepository
	Chat      ChatRepository
	Home      HomeRepository
	Noti      NotiRepository
	User      UserRepository
}

// 모든 리포지토리를 생성
func NewRepository(db *sql.DB) *Repository {
	board := NewTsboardBoardRepository(db)
	return &Repository{
		Auth:      NewTsboardAuthRepository(db),
		Board:     board,
		BoardView: NewTsboardBoardViewRepository(db, board),
		Chat:      NewTsboardChatRepository(db),
		Home:      NewTsboardHomeRepository(db, board),
		Noti:      NewTsboardNotiRepository(db),
		User:      NewTsboardUserRepository(db),
	}
}
