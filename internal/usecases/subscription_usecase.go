package usecases

import (
	"context"
	"tgbot/internal/entities"
	"tgbot/internal/repository"
)

type SubscriptionUsecase struct {
	userRepo repository.UserRepository
	subRepo  repository.SubscriptionRepository
}

func NewSubscriptionUsecase(userRepo repository.UserRepository, subRepo repository.SubscriptionRepository) *SubscriptionUsecase {
	return &SubscriptionUsecase{
		userRepo: userRepo,
		subRepo:  subRepo,
	}
}

func (u *SubscriptionUsecase) SaveSubscription(ctx context.Context, user *entities.User, subscription *entities.Subscription) error {
	if err := u.userRepo.SaveUser(ctx, user); err != nil {
		return err
	}
	return u.subRepo.SaveSubscription(ctx, subscription)
}

func (u *SubscriptionUsecase) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]string, error) {
	subscriptions, err := u.subRepo.GetSubscriptionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	categories := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		categories[i] = sub.Category
	}
	return categories, nil
}

func (u *SubscriptionUsecase) GetAllSubscriptions(ctx context.Context) ([]entities.Subscription, error) {
	return u.subRepo.GetAllSubscriptions(ctx)
}
