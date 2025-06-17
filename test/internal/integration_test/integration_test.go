package integration_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"tgbot/internal/entities"
	"tgbot/internal/repository"
	"tgbot/internal/usecases"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool        *pgxpool.Pool
	sentRepo    *repository.SentArticlesRepository
	newsUsecase *usecases.NewsUsecase
	testDBName  string
)

type mockNewsService struct{}

func (m *mockNewsService) GetNewsByCategory(ctx context.Context, category string) ([]entities.Article, error) {
	return []entities.Article{
		{Title: "Title 1", URL: "http://example.com/1", PublishedAt: "2025-01-01T00:00:00Z"},
	}, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	adminDsn := "postgres://username1:password1@localhost:5432/postgres?sslmode=disable"
	adminPool, err := pgxpool.New(ctx, adminDsn)
	if err != nil {
		log.Fatalf("failed to connect to admin db: %v", err)
	}
	defer adminPool.Close()

	testDBName = fmt.Sprintf("test_db_%d", time.Now().UnixNano())

	_, err = adminPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		log.Fatalf("failed to create test db: %v", err)
	}

	testDsn := fmt.Sprintf("postgres://username1:password1@localhost:5432/%s?sslmode=disable", testDBName)
	pool, err = pgxpool.New(ctx, testDsn)
	if err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}

	_, err = pool.Exec(ctx, `
        CREATE TABLE users (id BIGINT PRIMARY KEY);
        CREATE TABLE subscriptions (
            id SERIAL PRIMARY KEY,
            user_id BIGINT REFERENCES users(id),
            category VARCHAR(50) NOT NULL,
            UNIQUE(user_id, category)
        );
        CREATE TABLE sent_articles (
            id SERIAL PRIMARY KEY,
            url VARCHAR(255) NOT NULL UNIQUE,
            category VARCHAR(50) NOT NULL,
            sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	sentRepo = repository.NewSentArticlesRepository(pool)
	newsUsecase = usecases.NewNewsUsecase(&mockNewsService{}, sentRepo)

	code := m.Run()

	pool.Close()

	_, err = adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE %s", testDBName))
	if err != nil {
		log.Printf("failed to drop test db: %v", err)
	}

	os.Exit(code)
}

func TestSaveAndCheckSentArticle(t *testing.T) {
	ctx := context.Background()

	article := &entities.Article{URL: "https://example.com/news1"}

	err := sentRepo.SaveSentArticle(ctx, article, "tech")
	if err != nil {
		t.Fatalf("SaveSentArticle failed: %v", err)
	}

	sent, err := sentRepo.IsArticleSent(ctx, article.URL)
	if err != nil {
		t.Fatalf("IsArticleSent failed: %v", err)
	}
	if !sent {
		t.Fatalf("expected article to be marked as sent")
	}
}
