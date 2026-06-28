// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"qazfreelance/internal/bot"
	"qazfreelance/internal/config"
	"qazfreelance/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	stor, err := storage.NewSQLite("qazfreelance.db")
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := stor.Close(); err != nil {
			logger.Error("close database", "error", err)
		}
	}()

	b, err := bot.New(cfg, stor, logger)
	if err != nil {
		logger.Error("create bot", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		logger.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	logger.Info("starting bot")
	if err := b.Run(ctx); err != nil {
		logger.Error("bot run error", "error", err)
		os.Exit(1)
	}
}
