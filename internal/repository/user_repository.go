package repository

import (
	"context"
	"tgbot/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user *entities.User) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) SaveUser(ctx context.Context, user *entities.User) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING",
		user.ID)
	return err
}
