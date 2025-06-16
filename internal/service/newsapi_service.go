package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"tgbot/internal/entities"
)

type NewsAPIService struct {
	apiKey string
}

func NewNewsAPIService(apiKey string) *NewsAPIService {
	return &NewsAPIService{apiKey: apiKey}
}

func (s *NewsAPIService) GetNewsByCategory(ctx context.Context, category string) ([]entities.Article, error) {
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?category=%s&apiKey=%s", category, s.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Status   string `json:"status"`
		Articles []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			PublishedAt string `json:"publishedAt"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Status != "ok" {
		return nil, fmt.Errorf("API error: status %s", response.Status)
	}

	articles := make([]entities.Article, 0, len(response.Articles))
	for _, a := range response.Articles {
		articles = append(articles, entities.Article{
			Title:       a.Title,
			Description: a.Description,
			URL:         a.URL,
			PublishedAt: a.PublishedAt,
		})
	}

	return articles, nil
}
