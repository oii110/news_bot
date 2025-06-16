package usecases

import (
	"context"
	"fmt"
	"log"
	"strings"
	"tgbot/internal/entities"
	"time"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

type BotUsecase struct {
	bot                 *tgbotapi.BotAPI
	subscriptionUsecase SubscriptionUsecaseInterface
	newsUsecase         NewsUsecaseInterface
	categories          []string
}

func NewBotUsecase(bot *tgbotapi.BotAPI, subUsecase SubscriptionUsecaseInterface, newsUsecase NewsUsecaseInterface, categories []string) *BotUsecase {
	return &BotUsecase{
		bot:                 bot,
		subscriptionUsecase: subUsecase,
		newsUsecase:         newsUsecase,
		categories:          categories,
	}
}

func (u *BotUsecase) StartBot(ctx context.Context) {
	u.bot.Debug = true
	log.Printf("Бот %s запущен!", u.bot.Self.UserName)

	go u.StartNewsChecker(ctx)

	updates := u.bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout: 60,
	})

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}
		u.HandleCommand(ctx, update)
	}
}

func (u *BotUsecase) StartNewsChecker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("News checker stopped")
			return
		case <-ticker.C:
			u.CheckAndSendNews(ctx)
		}
	}
}

func (u *BotUsecase) CheckAndSendNews(ctx context.Context) {
	log.Println("Checking for new news...")

	subscriptions, err := u.subscriptionUsecase.GetAllSubscriptions(ctx)
	if err != nil {
		log.Printf("Error getting subscriptions: %v", err)
		return
	}

	categoryUsers := make(map[string][]int64)
	for _, sub := range subscriptions {
		categoryUsers[sub.Category] = append(categoryUsers[sub.Category], sub.UserID)
	}

	for category, userIDs := range categoryUsers {
		articles, err := u.newsUsecase.GetNewArticles(ctx, category, 5)
		if err != nil {
			log.Printf("Error getting news for category %s: %v", category, err)
			continue
		}

		for _, userID := range userIDs {
			if len(articles) == 0 {
				msg := tgbotapi.NewMessage(userID, fmt.Sprintf("*Пока новых новостей нет* для категории %s.", category))
				msg.ParseMode = "Markdown"
				fmt.Printf("Sending no-news message to user %d: %s\n", userID, msg.Text)
				if _, err := u.bot.Send(msg); err != nil {
					log.Printf("Error sending no-news message to user %d: %v", userID, err)
				}
				continue
			}

			for _, article := range articles {
				msg := tgbotapi.NewMessage(userID, u.FormatArticle(&article))
				msg.ParseMode = "Markdown"
				fmt.Printf("Sending article to user %d: %s\n", userID, msg.Text)
				if _, err := u.bot.Send(msg); err != nil {
					log.Printf("Error sending news to user %d: %v", userID, err)
				}
			}
		}
	}
}

func (u *BotUsecase) FormatArticle(article *entities.Article) string {
	return fmt.Sprintf("*%s*\n%s\n[Read more](%s)", article.Title, article.Description, article.URL)
}

func (u *BotUsecase) HandleCommand(ctx context.Context, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	command := update.Message.Command()
	args := update.Message.CommandArguments()

	switch command {
	case "start":
		msg.Text = "Здравствуйте! Данный бот предназначен для получения новостей. Используйте /add для подписки, /news <category> для получения новостей, /mysubs для просмотра подписок, /help для справки."
	case "add":
		if args == "" {
			msg.Text = "Пожалуйста, укажите категорию (например, /add technology)."
			break
		}
		category := strings.ToLower(strings.TrimSpace(args))
		if !contains(u.categories, category) {
			msg.Text = fmt.Sprintf("Категория '%s' не поддерживается. Доступные категории: %s", category, strings.Join(u.categories, ", "))
			break
		}
		user := &entities.User{ID: update.Message.From.ID}
		subscription := &entities.Subscription{UserID: user.ID, Category: category}
		if err := u.subscriptionUsecase.SaveSubscription(ctx, user, subscription); err != nil {
			msg.Text = "Ошибка при добавлении подписки: " + err.Error()
			break
		}
		msg.Text = fmt.Sprintf("Вы успешно подписались на категорию '%s'!", category)
	case "news":
		if args == "" {
			msg.Text = "Пожалуйста, укажите категорию (например, /news technology)."
			break
		}
		category := strings.ToLower(strings.TrimSpace(args))
		articles, err := u.newsUsecase.GetNewsByCategory(ctx, category)
		if err != nil {
			msg.Text = "Ошибка при получении новостей: " + err.Error()
			break
		}
		if len(articles) == 0 {
			msg.Text = fmt.Sprintf("Нет новостей для категории '%s'.", category)
			break
		}
		msg.Text = ""
		limit := len(articles)
		if limit > 5 {
			limit = 5
		}
		for _, article := range articles[:limit] {
			msg.Text += u.FormatArticle(&article) + "\n\n"
		}
		msg.ParseMode = "Markdown"
	case "mysubs":
		subscriptions, err := u.subscriptionUsecase.GetSubscriptionsByUser(ctx, update.Message.From.ID)
		if err != nil {
			msg.Text = "Ошибка при получении подписок: " + err.Error()
			break
		}
		if len(subscriptions) == 0 {
			msg.Text = "У вас нет активных подписок."
			break
		}
		msg.Text = "Ваши подписки:\n" + strings.Join(subscriptions, "\n")
	case "help":
		msg.Text = "Доступные команды:\n/start - Начать работу\n/add <category> - Подписаться на категорию\n/news <category> - Получить новости\n/mysubs - Показать подписки\n/help - Справка"
	default:
		msg.Text = "Неизвестная команда. Используйте /help для списка команд."
	}

	if msg.Text != "" {
		fmt.Printf("Sending message to chat %d: %s\n", msg.ChatID, msg.Text)
		if _, err := u.bot.Send(msg); err != nil {
			log.Printf("Error sending message to chat %d: %v", msg.ChatID, err)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
