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
	Sync      SyncRepository
	Trade     TradeRepository
	User      UserRepository
}

// 모든 리포지토리를 생성
func NewRepository(db *sql.DB) *Repository {
	board := NewNuboBoardRepository(db)
	return &Repository{
		Admin:     NewNuboAdminRepository(db),
		Auth:      NewNuboAuthRepository(db),
		Board:     board,
		BoardEdit: NewNuboBoardEditRepository(db, board),
		BoardView: NewNuboBoardViewRepository(db, board),
		Chat:      NewNuboChatRepository(db),
		Comment:   NewNuboCommentRepository(db, board),
		Home:      NewNuboHomeRepository(db, board),
		Noti:      NewNuboNotiRepository(db),
		Sync:      NewNuboSyncRepository(db),
		Trade:     NewNuboTradeRepository(db),
		User:      NewNuboUserRepository(db),
	}
}
