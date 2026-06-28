// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package bot

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"qazfreelance/internal/config"
	"qazfreelance/internal/handlers"
	"qazfreelance/internal/storage"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	handler *handlers.Handler
	logger  *slog.Logger
}

func New(cfg *config.Config, stor storage.Storage, logger *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}
	api.Debug = false

	logger.Info("authorized as bot", "username", api.Self.UserName)

	handler := handlers.New(api, stor, cfg, logger)
	return &Bot{
		api:     api,
		handler: handler,
		logger:  logger,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			b.logger.Info("bot stopped")
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			go b.handler.HandleUpdate(update)
		}
	}
}
