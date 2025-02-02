package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"news-bot/internal/utils"
)

func ViewCmdStart() utils.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if _, err := bot.Send(tgbotapi.NewMessage(update.FromChat().ID, "Test message")); err != nil {
			return err
		}

		return nil
	}
}
