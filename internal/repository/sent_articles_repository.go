package repository

import (
	"context"
	"tgbot/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SentArticlesRepository struct {
	pool *pgxpool.Pool
}

func NewSentArticlesRepository(pool *pgxpool.Pool) *SentArticlesRepository {
	return &SentArticlesRepository{pool: pool}
}

func (r *SentArticlesRepository) SaveSentArticle(ctx context.Context, article *entities.Article, category string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO sent_articles (url, category) VALUES ($1, $2) ON CONFLICT (url) DO NOTHING",
		article.URL, category)
	return err
}

func (r *SentArticlesRepository) IsArticleSent(ctx context.Context, url string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM sent_articles WHERE url = $1)",
		url).Scan(&exists)
	return exists, err
}
