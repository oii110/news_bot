package main

import (
	"context"
	"log"
	"tgbot/internal/adapters"
	"tgbot/internal/config"
	"tgbot/internal/repository"
	"tgbot/internal/service"
	"tgbot/internal/usecases"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Не удалось загрузить конфигурацию:", err)
	}

	ctx := context.Background()
	postgresRepo, err := repository.NewPostgresRepository(ctx, 3, cfg.Storage)
	if err != nil {
		log.Fatal("Не удалось подключиться к PostgreSQL:", err)
	}

	userRepo := repository.NewUserRepository(postgresRepo.Conn())
	subRepo := repository.NewSubscriptionRepository(postgresRepo.Conn())
	sentArticlesRepo := repository.NewSentArticlesRepository(postgresRepo.Conn())

	newsService := service.NewNewsAPIService(cfg.Bot.AuthKey)
	subscriptionUsecase := usecases.NewSubscriptionUsecase(userRepo, subRepo)
	newsUsecase := usecases.NewNewsUsecase(newsService, sentArticlesRepo)

	categories := []string{"technology", "business", "science", "health", "entertainment"}

	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Fatal("Ошибка при создании Telegram-бота:", err)
	}

	wrappedBot := &adapters.BotWrapper{Bot: bot}
	botUsecase := usecases.NewBotUsecase(wrappedBot, subscriptionUsecase, newsUsecase, categories)
	botUsecase.StartBot(ctx)
}
