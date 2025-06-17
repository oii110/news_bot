package usecases_test

import (
	"context"
	"errors"
	"testing"
	"tgbot/internal/entities"
	usage "tgbot/internal/usecases"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) SaveUser(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type mockSubscriptionRepository struct {
	mock.Mock
}

func (m *mockSubscriptionRepository) SaveSubscription(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *mockSubscriptionRepository) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]entities.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entities.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) GetAllSubscriptions(ctx context.Context) ([]entities.Subscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.Subscription), args.Error(1)
}

func TestSubscriptionUsecase_SaveSubscription(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{ID: 123}
	subscription := &entities.Subscription{UserID: 123, Category: "technology"}

	tests := []struct {
		name          string
		userRepoError error
		subRepoError  error
		expectedError bool
	}{
		{
			name:          "Success",
			userRepoError: nil,
			subRepoError:  nil,
			expectedError: false,
		},
		{
			name:          "UserRepoError",
			userRepoError: errors.New("user repo error"),
			subRepoError:  nil,
			expectedError: true,
		},
		{
			name:          "SubRepoError",
			userRepoError: nil,
			subRepoError:  errors.New("sub repo error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepository{}
			subRepo := &mockSubscriptionRepository{}
			usecase := usage.NewSubscriptionUsecase(userRepo, subRepo)

			userRepo.On("SaveUser", ctx, user).Return(tt.userRepoError)
			if tt.userRepoError == nil {
				subRepo.On("SaveSubscription", ctx, subscription).Return(tt.subRepoError)
			}

			err := usecase.SaveSubscription(ctx, user, subscription)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			userRepo.AssertExpectations(t)
			subRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUsecase_GetSubscriptionsByUser(t *testing.T) {
	ctx := context.Background()
	userID := int64(123)
	subscriptions := []entities.Subscription{
		{UserID: userID, Category: "technology"},
		{UserID: userID, Category: "business"},
	}

	tests := []struct {
		name          string
		subRepoReturn []entities.Subscription
		subRepoError  error
		expected      []string
		expectedError bool
	}{
		{
			name:          "Success",
			subRepoReturn: subscriptions,
			subRepoError:  nil,
			expected:      []string{"technology", "business"},
			expectedError: false,
		},
		{
			name:          "EmptySubscriptions",
			subRepoReturn: []entities.Subscription{},
			subRepoError:  nil,
			expected:      []string{},
			expectedError: false,
		},
		{
			name:          "Error",
			subRepoReturn: nil,
			subRepoError:  errors.New("sub repo error"),
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepository{}
			subRepo := &mockSubscriptionRepository{}
			usecase := usage.NewSubscriptionUsecase(userRepo, subRepo)

			subRepo.On("GetSubscriptionsByUser", ctx, userID).Return(tt.subRepoReturn, tt.subRepoError)

			result, err := usecase.GetSubscriptionsByUser(ctx, userID)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			subRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUsecase_GetAllSubscriptions(t *testing.T) {
	ctx := context.Background()
	subscriptions := []entities.Subscription{
		{UserID: 123, Category: "technology"},
		{UserID: 456, Category: "business"},
	}

	tests := []struct {
		name          string
		subRepoReturn []entities.Subscription
		subRepoError  error
		expected      []entities.Subscription
		expectedError bool
	}{
		{
			name:          "Success",
			subRepoReturn: subscriptions,
			subRepoError:  nil,
			expected:      subscriptions,
			expectedError: false,
		},
		{
			name:          "EmptySubscriptions",
			subRepoReturn: []entities.Subscription{},
			subRepoError:  nil,
			expected:      []entities.Subscription{},
			expectedError: false,
		},
		{
			name:          "Error",
			subRepoReturn: nil,
			subRepoError:  errors.New("sub repo error"),
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepository{}
			subRepo := &mockSubscriptionRepository{}
			usecase := usage.NewSubscriptionUsecase(userRepo, subRepo)

			subRepo.On("GetAllSubscriptions", ctx).Return(tt.subRepoReturn, tt.subRepoError)

			result, err := usecase.GetAllSubscriptions(ctx)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			subRepo.AssertExpectations(t)
		})
	}
}
