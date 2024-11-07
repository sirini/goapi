package repositories

import "database/sql"

// 모든 리포지토리들을 관리
type Repository struct {
	Auth AuthRepository
	Chat ChatRepository
	Home HomeRepository
	Noti NotiRepository
	User UserRepository
}

// 모든 리포지토리를 생성
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Auth: NewTsboardAuthRepository(db),
		Chat: NewTsboardChatRepository(db),
		Home: NewTsboardHomeRepository(db),
		Noti: NewTsboardNotiRepository(db),
		User: NewTsboardUserRepository(db),
	}
}
