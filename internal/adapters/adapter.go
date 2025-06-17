package adapters

import tgbotapi "github.com/skinass/telegram-bot-api/v5"

type BotWrapper struct {
	Bot *tgbotapi.BotAPI
}

func (b *BotWrapper) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.Bot.Send(c)
}

func (b *BotWrapper) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return b.Bot.GetUpdatesChan(config)
}

func (b *BotWrapper) Self() tgbotapi.User {
	return b.Bot.Self
}
