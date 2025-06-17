package usecases_test

import (
	"context"
	"errors"
	"testing"
	"tgbot/internal/entities"
	"tgbot/internal/usecases"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockNewsAPIService struct {
	GetNewsByCategoryFunc func(ctx context.Context, category string) ([]entities.Article, error)
}

func (m *MockNewsAPIService) GetNewsByCategory(ctx context.Context, category string) ([]entities.Article, error) {
	return m.GetNewsByCategoryFunc(ctx, category)
}

type MockSentArticlesRepository struct {
	IsArticleSentFunc   func(ctx context.Context, url string) (bool, error)
	SaveSentArticleFunc func(ctx context.Context, article *entities.Article, category string) error
}

func (m *MockSentArticlesRepository) IsArticleSent(ctx context.Context, url string) (bool, error) {
	return m.IsArticleSentFunc(ctx, url)
}

func (m *MockSentArticlesRepository) SaveSentArticle(ctx context.Context, article *entities.Article, category string) error {
	return m.SaveSentArticleFunc(ctx, article, category)
}

func TestNewsUsecase_GetNewsByCategory(t *testing.T) {
	mockNews := &MockNewsAPIService{
		GetNewsByCategoryFunc: func(ctx context.Context, category string) ([]entities.Article, error) {
			return []entities.Article{
				{Title: "Title1", URL: "http://1.com"},
			}, nil
		},
	}

	usecase := usecases.NewNewsUsecase(mockNews, nil)

	articles, err := usecase.GetNewsByCategory(context.Background(), "technology")
	assert.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Equal(t, "Title1", articles[0].Title)
}

func TestNewsUsecase_GetNewsByCategory_Error(t *testing.T) {
	mockNews := &MockNewsAPIService{
		GetNewsByCategoryFunc: func(ctx context.Context, category string) ([]entities.Article, error) {
			return nil, errors.New("service error")
		},
	}

	usecase := usecases.NewNewsUsecase(mockNews, nil)

	articles, err := usecase.GetNewsByCategory(context.Background(), "science")
	assert.Error(t, err)
	assert.Nil(t, articles)
}

func TestNewsUsecase_GetNewArticles(t *testing.T) {
	mockNews := &MockNewsAPIService{
		GetNewsByCategoryFunc: func(ctx context.Context, category string) ([]entities.Article, error) {
			return []entities.Article{
				{Title: "New Article", URL: "http://new.com", PublishedAt: time.Now().Format(time.RFC3339)},
				{Title: "Old Article", URL: "http://old.com", PublishedAt: time.Now().Add(-time.Hour).Format(time.RFC3339)},
			}, nil
		},
	}

	mockRepo := &MockSentArticlesRepository{
		IsArticleSentFunc: func(ctx context.Context, url string) (bool, error) {
			if url == "http://new.com" {
				return false, nil
			}
			return true, nil
		},
		SaveSentArticleFunc: func(ctx context.Context, article *entities.Article, category string) error {
			return nil
		},
	}

	usecase := usecases.NewNewsUsecase(mockNews, mockRepo)

	articles, err := usecase.GetNewArticles(context.Background(), "technology", 5)
	assert.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Equal(t, "New Article", articles[0].Title)
}

func TestNewsUsecase_GetNewArticles_EmptyResult(t *testing.T) {
	mockNews := &MockNewsAPIService{
		GetNewsByCategoryFunc: func(ctx context.Context, category string) ([]entities.Article, error) {
			return []entities.Article{
				{Title: "Sent Article", URL: "http://sent.com"},
			}, nil
		},
	}

	mockRepo := &MockSentArticlesRepository{
		IsArticleSentFunc: func(ctx context.Context, url string) (bool, error) {
			return true, nil
		},
		SaveSentArticleFunc: func(ctx context.Context, article *entities.Article, category string) error {
			return nil
		},
	}

	usecase := usecases.NewNewsUsecase(mockNews, mockRepo)

	articles, err := usecase.GetNewArticles(context.Background(), "science", 5)
	assert.NoError(t, err)
	assert.Empty(t, articles)
}

func TestNewsUsecase_GetNewArticles_SaveErrorIgnored(t *testing.T) {
	mockNews := &MockNewsAPIService{
		GetNewsByCategoryFunc: func(ctx context.Context, category string) ([]entities.Article, error) {
			return []entities.Article{
				{Title: "New", URL: "http://new.com", PublishedAt: time.Now().Format(time.RFC3339)},
			}, nil
		},
	}

	mockRepo := &MockSentArticlesRepository{
		IsArticleSentFunc: func(ctx context.Context, url string) (bool, error) {
			return false, nil
		},
		SaveSentArticleFunc: func(ctx context.Context, article *entities.Article, category string) error {
			return errors.New("save error")
		},
	}

	usecase := usecases.NewNewsUsecase(mockNews, mockRepo)

	articles, err := usecase.GetNewArticles(context.Background(), "tech", 5)
	assert.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Equal(t, "New", articles[0].Title)
}
