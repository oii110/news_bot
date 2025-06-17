package usecases

import (
	"context"
	"sort"
	"tgbot/internal/entities"
)

type NewsUsecase struct {
	newsService NewsServiceInterface
	sentRepo    SentArticlesRepositoryInterface
}

func NewNewsUsecase(newsService NewsServiceInterface, sentRepo SentArticlesRepositoryInterface) *NewsUsecase {
	return &NewsUsecase{
		newsService: newsService,
		sentRepo:    sentRepo,
	}
}
func (u *NewsUsecase) GetNewsByCategory(ctx context.Context, category string) ([]entities.Article, error) {
	articles, err := u.newsService.GetNewsByCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func (u *NewsUsecase) GetNewArticles(ctx context.Context, category string, maxArticles int) ([]entities.Article, error) {
	articles, err := u.newsService.GetNewsByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	var newArticles []entities.Article
	for _, article := range articles {
		sent, err := u.sentRepo.IsArticleSent(ctx, article.URL)
		if err != nil {
			continue
		}
		if !sent {
			newArticles = append(newArticles, article)
		}
	}

	sort.Slice(newArticles, func(i, j int) bool {
		return newArticles[i].PublishedAt > newArticles[j].PublishedAt
	})

	if len(newArticles) > maxArticles {
		newArticles = newArticles[:maxArticles]
	}

	for _, article := range newArticles {
		if err := u.sentRepo.SaveSentArticle(ctx, &article, category); err != nil {
			continue
		}
	}

	return newArticles, nil
}
