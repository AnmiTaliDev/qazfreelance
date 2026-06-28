// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package handlers

import (
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"qazfreelance/internal/config"
	"qazfreelance/internal/i18n"
	"qazfreelance/internal/storage"
)

const (
	StageIdle           = "idle"
	StageSelectLanguage = "select_language"
	StageEnterTitle     = "enter_title"
	StageEnterDesc      = "enter_description"
	StageEnterContact   = "enter_contact"
	StageEnterFreeForm  = "enter_free_form"
)

type UserConversation struct {
	Stage          string
	SubmissionType string
	Title          string
	Description    string
	PhotoFileID    string
}

type Handler struct {
	Bot     *tgbotapi.BotAPI
	Storage storage.Storage
	Config  *config.Config
	Logger  *slog.Logger

	mu            sync.RWMutex
	conversations map[int64]*UserConversation
}

func New(bot *tgbotapi.BotAPI, stor storage.Storage, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		Bot:           bot,
		Storage:       stor,
		Config:        cfg,
		Logger:        logger,
		conversations: make(map[int64]*UserConversation),
	}
}

func (h *Handler) getConversation(userID int64) *UserConversation {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conv, ok := h.conversations[userID]; ok {
		return conv
	}
	return &UserConversation{Stage: StageIdle}
}

func (h *Handler) setConversation(userID int64, conv *UserConversation) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conversations[userID] = conv
}

func (h *Handler) resetConversation(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conversations, userID)
}

func (h *Handler) getUserLang(telegramID int64) string {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		return i18n.DefaultLang
	}
	return user.Language
}

func (h *Handler) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.Bot.Send(msg); err != nil {
		h.Logger.Error("send message", "chat_id", chatID, "error", err)
	}
}

func (h *Handler) sendWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := h.Bot.Send(msg); err != nil {
		h.Logger.Error("send message with keyboard", "chat_id", chatID, "error", err)
	}
}

func (h *Handler) sendWithInline(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := h.Bot.Send(msg); err != nil {
		h.Logger.Error("send message with inline keyboard", "chat_id", chatID, "error", err)
	}
}

func (h *Handler) answerCallback(callbackID, text string) {
	cb := tgbotapi.NewCallback(callbackID, text)
	if _, err := h.Bot.Request(cb); err != nil {
		h.Logger.Error("answer callback", "callback_id", callbackID, "error", err)
	}
}
