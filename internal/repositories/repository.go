package repositories

import "database/sql"

// 모든 리포지토리들을 관리
type Repository struct {
	UserRepo UserRepository
	AuthRepo AuthRepository
}

// 모든 리포지토리를 생성
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		UserRepo: NewMySQLUserRepository(db),
		AuthRepo: NewMySQLAuthRepository(db),
	}
}
