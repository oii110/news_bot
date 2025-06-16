package usecases

import (
	"context"
	"tgbot/internal/entities"
)

type SubscriptionUsecaseInterface interface {
	SaveSubscription(ctx context.Context, user *entities.User, subscription *entities.Subscription) error
	GetSubscriptionsByUser(ctx context.Context, userID int64) ([]string, error)
	GetAllSubscriptions(ctx context.Context) ([]entities.Subscription, error)
}

type NewsUsecaseInterface interface {
	GetNewsByCategory(ctx context.Context, category string) ([]entities.Article, error)
	GetNewArticles(ctx context.Context, category string, maxArticles int) ([]entities.Article, error)
}
