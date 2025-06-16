package repository

import (
	"context"
	"tgbot/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository interface {
	SaveSubscription(ctx context.Context, subscription *entities.Subscription) error
	GetSubscriptionsByUser(ctx context.Context, userID int64) ([]entities.Subscription, error)
	GetAllSubscriptions(ctx context.Context) ([]entities.Subscription, error)
}

type subscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepository{pool: pool}
}

func (r *subscriptionRepository) SaveSubscription(ctx context.Context, subscription *entities.Subscription) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO subscriptions (user_id, category) VALUES ($1, $2) ON CONFLICT (user_id, category) DO NOTHING",
		subscription.UserID, subscription.Category)
	return err
}

func (r *subscriptionRepository) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]entities.Subscription, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT user_id, category FROM subscriptions WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []entities.Subscription
	for rows.Next() {
		var sub entities.Subscription
		if err := rows.Scan(&sub.UserID, &sub.Category); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

func (r *subscriptionRepository) GetAllSubscriptions(ctx context.Context) ([]entities.Subscription, error) {
	rows, err := r.pool.Query(ctx, "SELECT user_id, category FROM subscriptions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []entities.Subscription
	for rows.Next() {
		var sub entities.Subscription
		if err := rows.Scan(&sub.UserID, &sub.Category); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}
