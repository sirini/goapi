package repositories

import "database/sql"

// 모든 리포지토리들을 관리
type Repository struct {
	Admin     AdminRepository
	Auth      AuthRepository
	Board     BoardRepository
	BoardEdit BoardEditRepository
	BoardView BoardViewRepository
	Chat      ChatRepository
	Comment   CommentRepository
	Home      HomeRepository
	Noti      NotiRepository
	User      UserRepository
}

// 모든 리포지토리를 생성
func NewRepository(db *sql.DB) *Repository {
	board := NewTsboardBoardRepository(db)
	return &Repository{
		Admin:     NewTsboardAdminRepository(db),
		Auth:      NewTsboardAuthRepository(db),
		Board:     board,
		BoardEdit: NewTsboardBoardEditRepository(db, board),
		BoardView: NewTsboardBoardViewRepository(db, board),
		Chat:      NewTsboardChatRepository(db),
		Comment:   NewTsboardCommentRepository(db, board),
		Home:      NewTsboardHomeRepository(db, board),
		Noti:      NewTsboardNotiRepository(db),
		User:      NewTsboardUserRepository(db),
	}
}
