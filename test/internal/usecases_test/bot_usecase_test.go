package usecases_test

import (
	"context"
	"strings"
	"testing"

	"tgbot/internal/entities"
	"tgbot/internal/usecases"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBotAPI struct {
	mock.Mock
	tgbotapi.BotAPI
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	args := m.Called(config)
	return args.Get(0).(tgbotapi.UpdatesChannel)
}

func (m *MockBotAPI) Self() tgbotapi.User {
	args := m.Called()
	return args.Get(0).(tgbotapi.User)
}

type MockSubscriptionUsecase struct {
	mock.Mock
}

func (m *MockSubscriptionUsecase) SaveSubscription(ctx context.Context, user *entities.User, subscription *entities.Subscription) error {
	args := m.Called(ctx, user, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionUsecase) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockSubscriptionUsecase) GetAllSubscriptions(ctx context.Context) ([]entities.Subscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.Subscription), args.Error(1)
}

type MockNewsUsecase struct {
	mock.Mock
}

func (m *MockNewsUsecase) GetNewsByCategory(ctx context.Context, category string) ([]entities.Article, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]entities.Article), args.Error(1)
}

func (m *MockNewsUsecase) GetNewArticles(ctx context.Context, category string, maxArticles int) ([]entities.Article, error) {
	args := m.Called(ctx, category, maxArticles)
	return args.Get(0).([]entities.Article), args.Error(1)
}

func TestBotUsecase_HandleCommand(t *testing.T) {
	ctx := context.Background()
	mockBot := &MockBotAPI{}
	mockSubUsecase := &MockSubscriptionUsecase{}
	mockNewsUsecase := &MockNewsUsecase{}
	categories := []string{"technology", "business"}

	botUsecase := usecases.NewBotUsecase(mockBot, mockSubUsecase, mockNewsUsecase, categories)

	mockBot.On("GetUpdatesChan", mock.Anything).Return(make(chan tgbotapi.Update, 1)).Maybe()

	tests := []struct {
		name        string
		update      tgbotapi.Update
		expectedMsg string
		setupMocks  func()
	}{
		{
			name: "Start command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: 123},
					From:     &tgbotapi.User{ID: 123},
					Text:     "/start",
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}},
				},
			},
			expectedMsg: "Здравствуйте! Данный бот предназначен для получения новостей",
		},
		{
			name: "Add valid category",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: 123},
					From:     &tgbotapi.User{ID: 123},
					Text:     "/add technology",
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 4}},
				},
			},
			expectedMsg: "Вы успешно подписались на категорию 'technology'!",
			setupMocks: func() {
				mockSubUsecase.On("SaveSubscription", ctx, &entities.User{ID: 123}, &entities.Subscription{UserID: 123, Category: "technology"}).Return(nil)
			},
		},
		{
			name: "Add invalid category",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: 123},
					From:     &tgbotapi.User{ID: 123},
					Text:     "/add invalid",
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 4}},
				},
			},
			expectedMsg: "Категория 'invalid' не поддерживается",
		},
		{
			name: "News command with articles",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: 123},
					From:     &tgbotapi.User{ID: 123},
					Text:     "/news technology",
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}},
				},
			},
			expectedMsg: "*Test Title*\nTest Description\n[Read more](http://example.com)",
			setupMocks: func() {
				mockNewsUsecase.On("GetNewsByCategory", ctx, "technology").Return([]entities.Article{
					{
						Title:       "Test Title",
						Description: "Test Description",
						URL:         "http://example.com",
						PublishedAt: "2025-06-16T12:00:00Z",
					},
				}, nil)
			},
		},
		{
			name: "MySubs with subscriptions",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: 123},
					From:     &tgbotapi.User{ID: 123},
					Text:     "/mysubs",
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 7}},
				},
			},
			expectedMsg: "Ваши подписки:\ntechnology\nbusiness",
			setupMocks: func() {
				mockSubUsecase.On("GetSubscriptionsByUser", ctx, int64(123)).Return([]string{"technology", "business"}, nil)
			},
		},
		{
			name: "Unknown command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: 123},
					From:     &tgbotapi.User{ID: 123},
					Text:     "/unknown",
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}},
				},
			},
			expectedMsg: "Неизвестная команда",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			mockBot.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
				msg, ok := c.(tgbotapi.MessageConfig)
				if !ok {
					t.Logf("Send called with unexpected type: %T", c)
					return false
				}
				t.Logf("Send called with message: %s", msg.Text)
				return strings.Contains(msg.Text, tt.expectedMsg)
			})).Return(tgbotapi.Message{MessageID: 1}, nil).Once()

			botUsecase.HandleCommand(ctx, tt.update)

			mockBot.AssertExpectations(t)
		})
	}
}

func TestBotUsecase_SendToSubs(t *testing.T) {
	ctx := context.Background()
	mockBot := &MockBotAPI{}
	mockSubUsecase := &MockSubscriptionUsecase{}
	mockNewsUsecase := &MockNewsUsecase{}
	categories := []string{"technology"}

	botUsecase := usecases.NewBotUsecase(mockBot, mockSubUsecase, mockNewsUsecase, categories)

	mockBot.On("GetUpdatesChan", mock.Anything).Return(make(chan tgbotapi.Update, 1)).Maybe()

	t.Run("Send new articles to subscribed users", func(t *testing.T) {
		mockSubUsecase.On("GetAllSubscriptions", ctx).Return([]entities.Subscription{
			{UserID: 123, Category: "technology"},
		}, nil)
		mockNewsUsecase.On("GetNewArticles", ctx, "technology", 5).Return([]entities.Article{
			{
				Title:       "Test Title",
				Description: "Test Description",
				URL:         "http://example.com",
				PublishedAt: "2025-06-16T12:00:00Z",
			},
		}, nil)

		mockBot.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			msg, ok := c.(tgbotapi.MessageConfig)
			if !ok {
				t.Logf("Send called with unexpected type: %T", c)
				return false
			}
			t.Logf("Send called with message: %s", msg.Text)
			return strings.Contains(msg.Text, "*Test Title*\nTest Description\n[Read more](http://example.com)")
		})).Return(tgbotapi.Message{MessageID: 1}, nil).Once()

		botUsecase.CheckAndSendNews(ctx)

		mockBot.AssertExpectations(t)
	})
}

func TestBotUsecase_NoNewArticles(t *testing.T) {
	ctx := context.Background()
	mockBot := &MockBotAPI{}
	mockSubUsecase := &MockSubscriptionUsecase{}
	mockNewsUsecase := &MockNewsUsecase{}
	categories := []string{"technology"}

	botUsecase := usecases.NewBotUsecase(mockBot, mockSubUsecase, mockNewsUsecase, categories)

	mockBot.On("GetUpdatesChan", mock.Anything).Return(make(chan tgbotapi.Update, 1)).Maybe()

	t.Run("No new articles", func(t *testing.T) {
		mockSubUsecase.On("GetAllSubscriptions", ctx).Return([]entities.Subscription{
			{UserID: 123, Category: "technology"},
		}, nil)
		mockNewsUsecase.On("GetNewArticles", ctx, "technology", 5).Return([]entities.Article{}, nil)

		mockBot.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			msg, ok := c.(tgbotapi.MessageConfig)
			if !ok {
				t.Logf("Send called with unexpected type: %T", c)
				return false
			}
			t.Logf("Send called with message: %s", msg.Text)
			return strings.Contains(msg.Text, "*Пока новых новостей нет*")
		})).Return(tgbotapi.Message{MessageID: 1}, nil).Once()

		botUsecase.CheckAndSendNews(ctx)

		mockBot.AssertExpectations(t)
	})
}

func TestBotUsecase_FormatArticle(t *testing.T) {
	mockBot := &MockBotAPI{}
	mockSubUsecase := &MockSubscriptionUsecase{}
	mockNewsUsecase := &MockNewsUsecase{}
	categories := []string{"technology"}

	botUsecase := usecases.NewBotUsecase(mockBot, mockSubUsecase, mockNewsUsecase, categories)

	article := &entities.Article{
		Title:       "Test Title",
		Description: "Test Description",
		URL:         "http://example.com",
		PublishedAt: "2025-06-16T12:00:00Z",
	}

	result := botUsecase.FormatArticle(article)
	expected := "*Test Title*\nTest Description\n[Read more](http://example.com)"
	assert.Equal(t, expected, result)
}
