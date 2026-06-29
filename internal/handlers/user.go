// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"qazfreelance/internal/i18n"
	"qazfreelance/internal/models"
)

func emptyKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.Logger.Debug("callback query", "data", update.CallbackQuery.Data, "user_id", update.CallbackQuery.From.ID)
		h.handleCallback(update.CallbackQuery)
		return
	}
	if update.Message == nil {
		return
	}
	h.Logger.Debug("message", "text", update.Message.Text, "user_id", update.Message.From.ID)

	if update.Message.IsCommand() {
		h.handleCommand(update.Message)
		return
	}
	h.handleMessage(update.Message)
}

func (h *Handler) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.HandleStart(msg)
	case "mode":
		h.HandleModeCommand(msg)
	case "settings":
		h.HandleSettingsCommand(msg)
	default:
		lang := h.getUserLang(msg.From.ID)
		h.send(msg.Chat.ID, i18n.T(lang, "unknown_command"))
	}
}

func (h *Handler) HandleStart(msg *tgbotapi.Message) {
	telegramID := msg.From.ID
	user, err := h.Storage.GetUser(telegramID)
	if err != nil {
		h.Logger.Error("get user on start", "telegram_id", telegramID, "error", err)
		return
	}
	if user == nil {
		h.setConversation(telegramID, &UserConversation{Stage: StageSelectLanguage})
		h.showLanguageSelection(msg.Chat.ID)
		return
	}
	h.resetConversation(telegramID)
	h.showMainMenu(msg.Chat.ID, telegramID, user.Language)
}

func (h *Handler) showLanguageSelection(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T("kk", "btn_kk"), "lang_kk"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T("ru", "btn_ru"), "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T("en", "btn_en"), "lang_en"),
		),
	)
	text := "Kazakh / Русский / English\n" +
		i18n.T("kk", "choose_language") + "\n" +
		i18n.T("ru", "choose_language") + "\n" +
		i18n.T("en", "choose_language")
	h.sendWithInline(chatID, text, keyboard)
}

func (h *Handler) showMainMenu(chatID int64, telegramID int64, lang string) {
	rows := [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(i18n.T(lang, "btn_post_order")),
			tgbotapi.NewKeyboardButton(i18n.T(lang, "btn_advertise")),
		},
		{
			tgbotapi.NewKeyboardButton(i18n.T(lang, "btn_my_submissions")),
			tgbotapi.NewKeyboardButton(i18n.T(lang, "btn_settings")),
		},
	}
	if h.Config.IsModerator(telegramID) {
		rows = append(rows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(i18n.T(lang, "btn_moderator_menu")),
		})
	}
	keyboard := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        rows,
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
	h.sendWithKeyboard(chatID, i18n.T(lang, "main_menu"), keyboard)
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) {
	telegramID := msg.From.ID
	lang := h.getUserLang(telegramID)
	conv := h.getConversation(telegramID)
	text := strings.TrimSpace(msg.Text)

	switch conv.Stage {
	case StageSelectLanguage:
		h.showLanguageSelection(msg.Chat.ID)
		return

	case StageEnterTitle:
		if text == i18n.T(lang, "btn_cancel") {
			h.cancelSubmission(msg.Chat.ID, telegramID, lang)
			return
		}
		if len(msg.Photo) > 0 {
			h.send(msg.Chat.ID, i18n.T(lang, "photo_not_in_title"))
			return
		}
		if text == "" {
			return
		}
		conv.Title = text
		conv.Stage = StageEnterDesc
		h.setConversation(telegramID, conv)
		h.sendCancelKeyboard(msg.Chat.ID, lang, i18n.T(lang, "enter_description"))
		return

	case StageEnterDesc:
		if text == i18n.T(lang, "btn_cancel") {
			h.cancelSubmission(msg.Chat.ID, telegramID, lang)
			return
		}
		if len(msg.Photo) > 0 {
			conv.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
			conv.Description = strings.TrimSpace(msg.Caption)
		} else {
			if text == "" {
				return
			}
			conv.Description = text
		}
		conv.Stage = StageEnterContact
		h.setConversation(telegramID, conv)
		h.sendCancelKeyboard(msg.Chat.ID, lang, i18n.T(lang, "enter_contact"))
		return

	case StageEnterContact:
		if text == i18n.T(lang, "btn_cancel") {
			h.cancelSubmission(msg.Chat.ID, telegramID, lang)
			return
		}
		if len(msg.Photo) > 0 {
			h.send(msg.Chat.ID, i18n.T(lang, "photo_not_in_contact"))
			return
		}
		if text == "" {
			return
		}
		h.finalizeSubmission(msg.Chat.ID, telegramID, lang, conv, text)
		return

	case StageEnterFreeForm:
		if text == i18n.T(lang, "btn_cancel") {
			h.cancelSubmission(msg.Chat.ID, telegramID, lang)
			return
		}
		if len(msg.Photo) > 0 {
			conv.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
			conv.Description = strings.TrimSpace(msg.Caption)
		} else {
			if text == "" {
				return
			}
			conv.Description = text
		}
		h.finalizeSubmission(msg.Chat.ID, telegramID, lang, conv, "")
		return

	default:
		if text == i18n.T(lang, "btn_post_order") {
			h.startOrChooseMode(msg.Chat.ID, telegramID, lang, models.SubmissionTypeOrder)
			return
		}
		if text == i18n.T(lang, "btn_advertise") {
			h.startOrChooseMode(msg.Chat.ID, telegramID, lang, models.SubmissionTypeResume)
			return
		}
		if text == i18n.T(lang, "btn_my_submissions") {
			h.showUserSubmissions(msg.Chat.ID, telegramID, lang)
			return
		}
		if text == i18n.T(lang, "btn_settings") {
			h.sendSettings(msg.Chat.ID, telegramID, lang)
			return
		}
		if h.Config.IsModerator(telegramID) && text == i18n.T(lang, "btn_moderator_menu") {
			h.showModeratorMenu(msg.Chat.ID, telegramID)
			return
		}
		h.send(msg.Chat.ID, i18n.T(lang, "unknown_command"))
	}
}

func (h *Handler) cancelSubmission(chatID int64, telegramID int64, lang string) {
	h.resetConversation(telegramID)
	h.send(chatID, i18n.T(lang, "cancel"))
	h.showMainMenu(chatID, telegramID, lang)
}

func (h *Handler) sendCancelKeyboard(chatID int64, lang, prompt string) {
	keyboard := tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton(i18n.T(lang, "btn_cancel"))},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	h.sendWithKeyboard(chatID, prompt, keyboard)
}

func (h *Handler) HandleSettingsCommand(msg *tgbotapi.Message) {
	lang := h.getUserLang(msg.From.ID)
	h.sendSettings(msg.Chat.ID, msg.From.ID, lang)
}

func (h *Handler) settingsText(lang string, user *models.User) string {
	langName := i18n.T(user.Language, "btn_"+user.Language)

	var modeName string
	switch user.DefaultSubMode {
	case models.SubModeGuided:
		modeName = i18n.T(lang, "btn_guided")
	case models.SubModeFree:
		modeName = i18n.T(lang, "btn_free_form")
	default:
		modeName = i18n.T(lang, "settings_mode_ask")
	}

	return i18n.T(lang, "settings_header") + "\n\n" +
		i18n.Tf(lang, "settings_language", langName) + "\n" +
		i18n.Tf(lang, "settings_sub_mode", modeName)
}

func (h *Handler) settingsKeyboard(lang string, user *models.User) tgbotapi.InlineKeyboardMarkup {
	mark := func(active bool, label string) string {
		if active {
			return "> " + label
		}
		return label
	}

	langKK := mark(user.Language == "kk", i18n.T("kk", "btn_kk"))
	langRU := mark(user.Language == "ru", i18n.T("ru", "btn_ru"))
	langEN := mark(user.Language == "en", i18n.T("en", "btn_en"))

	modeGuided := mark(user.DefaultSubMode == models.SubModeGuided, i18n.T(lang, "btn_guided"))
	modeFree := mark(user.DefaultSubMode == models.SubModeFree, i18n.T(lang, "btn_free_form"))
	modeAsk := mark(user.DefaultSubMode == models.SubModeAsk, i18n.T(lang, "settings_mode_ask"))

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(langKK, "settings_lang_kk"),
			tgbotapi.NewInlineKeyboardButtonData(langRU, "settings_lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData(langEN, "settings_lang_en"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(modeGuided, "settings_mode_guided"),
			tgbotapi.NewInlineKeyboardButtonData(modeFree, "settings_mode_free"),
			tgbotapi.NewInlineKeyboardButtonData(modeAsk, "settings_mode_ask"),
		),
	)
}

func (h *Handler) sendSettings(chatID int64, telegramID int64, lang string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.Logger.Error("get user for settings", "telegram_id", telegramID, "error", err)
		return
	}
	text := h.settingsText(lang, user)
	kb := h.settingsKeyboard(lang, user)
	h.sendWithInline(chatID, text, kb)
}

func (h *Handler) editSettings(cb *tgbotapi.CallbackQuery, telegramID int64, lang string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.Logger.Error("get user for settings edit", "telegram_id", telegramID, "error", err)
		return
	}
	text := h.settingsText(lang, user)
	kb := h.settingsKeyboard(lang, user)
	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text)
	edit.ReplyMarkup = &kb
	if _, err := h.Bot.Send(edit); err != nil {
		h.Logger.Error("edit settings message", "error", err)
	}
}

func (h *Handler) handleSettingsLang(cb *tgbotapi.CallbackQuery, telegramID int64, newLang string) {
	if err := h.Storage.UpdateUserLanguage(telegramID, newLang); err != nil {
		h.Logger.Error("update language from settings", "telegram_id", telegramID, "error", err)
		h.answerCallback(cb.ID, "")
		return
	}
	h.answerCallback(cb.ID, i18n.T(newLang, "settings_saved"))
	h.editSettings(cb, telegramID, newLang)
}

func (h *Handler) handleSettingsMode(cb *tgbotapi.CallbackQuery, telegramID int64, lang string, newMode string) {
	if newMode == "ask" {
		newMode = models.SubModeAsk
	}
	if err := h.Storage.UpdateUserDefaultSubMode(telegramID, newMode); err != nil {
		h.Logger.Error("update sub mode from settings", "telegram_id", telegramID, "error", err)
		h.answerCallback(cb.ID, "")
		return
	}
	h.answerCallback(cb.ID, i18n.T(lang, "settings_saved"))
	h.editSettings(cb, telegramID, lang)
}

func (h *Handler) startOrChooseMode(chatID int64, telegramID int64, lang string, subType string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.showSubmissionModeChoice(chatID, lang, subType)
		return
	}
	switch user.DefaultSubMode {
	case models.SubModeGuided:
		h.startGuidedSubmission(chatID, telegramID, lang, subType)
	case models.SubModeFree:
		h.startFreeFormSubmission(chatID, telegramID, lang, subType)
	default:
		h.showSubmissionModeChoice(chatID, lang, subType)
	}
}

func (h *Handler) showSubmissionModeChoice(chatID int64, lang string, subType string) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_guided"), "mode_guided_"+subType),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_free_form"), "mode_free_"+subType),
		),
	)
	h.sendWithInline(chatID, i18n.T(lang, "choose_submission_mode"), keyboard)
}

func (h *Handler) startGuidedSubmission(chatID int64, telegramID int64, lang string, subType string) {
	h.setConversation(telegramID, &UserConversation{
		Stage:          StageEnterTitle,
		SubmissionType: subType,
	})
	h.sendCancelKeyboard(chatID, lang, i18n.T(lang, "enter_title"))
}

func (h *Handler) startFreeFormSubmission(chatID int64, telegramID int64, lang string, subType string) {
	h.setConversation(telegramID, &UserConversation{
		Stage:          StageEnterFreeForm,
		SubmissionType: subType,
	})
	h.sendCancelKeyboard(chatID, lang, i18n.T(lang, "enter_free_form"))
}

func (h *Handler) finalizeSubmission(chatID int64, telegramID int64, lang string, conv *UserConversation, contact string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.Logger.Error("get user for finalize", "telegram_id", telegramID, "error", err)
		return
	}

	sub := &models.Submission{
		UserID:      user.ID,
		Type:        conv.SubmissionType,
		Title:       conv.Title,
		Description: conv.Description,
		Contact:     contact,
		PhotoFileID: conv.PhotoFileID,
		Status:      models.StatusPending,
	}

	id, err := h.Storage.CreateSubmission(sub)
	if err != nil {
		h.Logger.Error("create submission", "error", err)
		return
	}
	sub.ID = id

	h.resetConversation(telegramID)
	h.send(chatID, i18n.T(lang, "submission_received"))
	h.showMainMenu(chatID, telegramID, lang)
	h.notifyModerators(sub)
}

func (h *Handler) notifyModerators(sub *models.Submission) {
	for _, modID := range h.Config.ModeratorIDs {
		mode, err := h.Storage.GetModeratorMode(modID)
		if err != nil {
			h.Logger.Error("get moderator mode", "moderator_id", modID, "error", err)
			mode = models.ModeStream
		}
		switch mode {
		case models.ModeStream:
			h.sendSubmissionToModerator(modID, sub)
		case models.ModeList:
			lang := h.getUserLang(modID)
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_open_list"), "open_list"),
				),
			)
			h.sendWithInline(modID, i18n.T(lang, "pending_notify"), keyboard)
		}
	}
}

func submissionStatusKey(status string) string {
	switch status {
	case models.StatusApproved:
		return "status_approved"
	case models.StatusRejected:
		return "status_rejected"
	case models.StatusConflict:
		return "status_conflict"
	case models.StatusWithdrawn:
		return "status_withdrawn"
	default:
		return "status_pending"
	}
}

func submissionButtonLabel(sub *models.Submission) string {
	label := sub.Title
	if label == "" {
		label = sub.Description
	}
	runes := []rune(label)
	if len(runes) > 25 {
		label = string(runes[:25]) + "..."
	}
	return fmt.Sprintf("#%d %s", sub.ID, label)
}

func (h *Handler) buildSubmissionsListView(lang string, subs []*models.Submission) (string, tgbotapi.InlineKeyboardMarkup) {
	var text strings.Builder
	text.WriteString(i18n.T(lang, "my_submissions_header"))

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, sub := range subs {
		typeName := i18n.T(lang, "submission_type_order")
		if sub.Type == models.SubmissionTypeResume {
			typeName = i18n.T(lang, "submission_type_resume")
		}
		statusName := i18n.T(lang, submissionStatusKey(sub.Status))
		text.WriteString(fmt.Sprintf("\n#%d · %s · %s", sub.ID, typeName, statusName))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				submissionButtonLabel(sub),
				fmt.Sprintf("my_sub_%d", sub.ID),
			),
		))
	}

	return text.String(), tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) showUserSubmissions(chatID int64, telegramID int64, lang string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.Logger.Error("get user for submissions list", "telegram_id", telegramID, "error", err)
		return
	}
	subs, err := h.Storage.GetSubmissionsByUserID(user.ID)
	if err != nil {
		h.Logger.Error("get user submissions", "user_id", user.ID, "error", err)
		return
	}
	if len(subs) == 0 {
		h.send(chatID, i18n.T(lang, "no_submissions"))
		return
	}
	text, keyboard := h.buildSubmissionsListView(lang, subs)
	h.sendWithInline(chatID, text, keyboard)
}

func (h *Handler) editToSubmissionsList(cb *tgbotapi.CallbackQuery, telegramID int64, lang string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.Logger.Error("get user for list edit", "telegram_id", telegramID, "error", err)
		return
	}
	subs, err := h.Storage.GetSubmissionsByUserID(user.ID)
	if err != nil {
		h.Logger.Error("get user submissions for list edit", "user_id", user.ID, "error", err)
		return
	}
	text, keyboard := h.buildSubmissionsListView(lang, subs)
	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text)
	edit.ReplyMarkup = &keyboard
	if _, err := h.Bot.Send(edit); err != nil {
		h.Logger.Error("edit submissions list message", "error", err)
	}
}

func (h *Handler) showSubmissionDetail(cb *tgbotapi.CallbackQuery, telegramID int64, submissionID int64, lang string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.answerCallback(cb.ID, "")
		return
	}
	sub, err := h.Storage.GetSubmission(submissionID)
	if err != nil || sub == nil {
		h.answerCallback(cb.ID, "")
		return
	}
	if sub.UserID != user.ID {
		h.answerCallback(cb.ID, "")
		return
	}

	typeName := i18n.T(lang, "submission_type_order")
	if sub.Type == models.SubmissionTypeResume {
		typeName = i18n.T(lang, "submission_type_resume")
	}
	statusName := i18n.T(lang, submissionStatusKey(sub.Status))

	var text strings.Builder
	text.WriteString(fmt.Sprintf("#%d — %s — %s", sub.ID, typeName, statusName))
	if sub.Title != "" {
		text.WriteString("\n\n")
		text.WriteString(sub.Title)
	}
	if sub.Description != "" {
		text.WriteString("\n\n")
		text.WriteString(sub.Description)
	}
	if sub.Contact != "" {
		text.WriteString("\n\n")
		text.WriteString(i18n.Tf(lang, "contact_label", sub.Contact))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	if sub.Status == models.StatusPending || sub.Status == models.StatusApproved {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.T(lang, "btn_withdraw"),
				fmt.Sprintf("withdraw_%d", sub.ID),
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_back"), "my_subs"),
	))
	keyboard := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}

	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text.String())
	edit.ReplyMarkup = &keyboard
	if _, err := h.Bot.Send(edit); err != nil {
		h.Logger.Error("edit submission detail message", "error", err)
	}
}

func (h *Handler) buildSubmissionText(lang string, sub *models.Submission) string {
	typeName := i18n.T(lang, "submission_type_order")
	if sub.Type == models.SubmissionTypeResume {
		typeName = i18n.T(lang, "submission_type_resume")
	}
	return i18n.Tf(lang, "submission_info",
		sub.ID, typeName, sub.Title, sub.Description, sub.Contact, sub.Status,
	)
}

func (h *Handler) handleCallback(cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	lang := h.getUserLang(userID)

	switch {
	case data == "lang_kk" || data == "lang_ru" || data == "lang_en":
		h.handleLanguageSelection(cb, userID, chatID, data)

	case strings.HasPrefix(data, "mode_guided_"):
		subType := strings.TrimPrefix(data, "mode_guided_")
		h.answerCallback(cb.ID, "")
		h.startGuidedSubmission(chatID, userID, lang, subType)

	case strings.HasPrefix(data, "mode_free_"):
		subType := strings.TrimPrefix(data, "mode_free_")
		h.answerCallback(cb.ID, "")
		h.startFreeFormSubmission(chatID, userID, lang, subType)

	case data == "settings":
		h.answerCallback(cb.ID, "")
		h.editSettings(cb, userID, lang)

	case strings.HasPrefix(data, "settings_lang_"):
		newLang := strings.TrimPrefix(data, "settings_lang_")
		h.handleSettingsLang(cb, userID, newLang)

	case strings.HasPrefix(data, "settings_mode_"):
		newMode := strings.TrimPrefix(data, "settings_mode_")
		h.handleSettingsMode(cb, userID, lang, newMode)

	case data == "my_subs":
		h.answerCallback(cb.ID, "")
		h.editToSubmissionsList(cb, userID, lang)

	case strings.HasPrefix(data, "my_sub_"):
		var subID int64
		fmt.Sscanf(strings.TrimPrefix(data, "my_sub_"), "%d", &subID)
		h.answerCallback(cb.ID, "")
		h.showSubmissionDetail(cb, userID, subID, lang)

	case strings.HasPrefix(data, "withdraw_"):
		var subID int64
		fmt.Sscanf(strings.TrimPrefix(data, "withdraw_"), "%d", &subID)
		h.handleWithdraw(cb, userID, subID, lang)

	case data == "open_list":
		if h.Config.IsModerator(userID) {
			h.answerCallback(cb.ID, "")
			h.showPendingList(chatID, userID)
		}

	case data == "open_conflicts":
		if h.Config.IsModerator(userID) {
			h.answerCallback(cb.ID, "")
			h.showConflictList(chatID, userID)
		}

	case data == "toggle_mode":
		if h.Config.IsModerator(userID) {
			h.answerCallback(cb.ID, "")
			h.toggleModeratorMode(chatID, userID, lang)
		}

	case strings.HasPrefix(data, "approve_"):
		if h.Config.IsModerator(userID) {
			var subID int64
			fmt.Sscanf(strings.TrimPrefix(data, "approve_"), "%d", &subID)
			h.handleApprove(cb, userID, subID)
		}

	case strings.HasPrefix(data, "reject_"):
		if h.Config.IsModerator(userID) {
			var subID int64
			fmt.Sscanf(strings.TrimPrefix(data, "reject_"), "%d", &subID)
			h.handleReject(cb, userID, subID)
		}

	case strings.HasPrefix(data, "resolve_approve_"):
		if h.Config.IsModerator(userID) {
			var subID int64
			fmt.Sscanf(strings.TrimPrefix(data, "resolve_approve_"), "%d", &subID)
			h.handleResolve(cb, userID, subID, models.DecisionApprove)
		}

	case strings.HasPrefix(data, "resolve_reject_"):
		if h.Config.IsModerator(userID) {
			var subID int64
			fmt.Sscanf(strings.TrimPrefix(data, "resolve_reject_"), "%d", &subID)
			h.handleResolve(cb, userID, subID, models.DecisionReject)
		}

	case strings.HasPrefix(data, "next_"):
		if h.Config.IsModerator(userID) {
			var afterID int64
			fmt.Sscanf(strings.TrimPrefix(data, "next_"), "%d", &afterID)
			h.answerCallback(cb.ID, "")
			h.showNextPending(chatID, userID, afterID)
		}

	default:
		h.answerCallback(cb.ID, i18n.T(lang, "unknown_command"))
	}
}

func (h *Handler) handleWithdraw(cb *tgbotapi.CallbackQuery, telegramID int64, submissionID int64, lang string) {
	user, err := h.Storage.GetUser(telegramID)
	if err != nil || user == nil {
		h.answerCallback(cb.ID, "")
		return
	}
	sub, err := h.Storage.GetSubmission(submissionID)
	if err != nil || sub == nil {
		h.answerCallback(cb.ID, "")
		return
	}
	if sub.UserID != user.ID {
		h.answerCallback(cb.ID, i18n.T(lang, "cannot_withdraw"))
		return
	}
	if sub.Status != models.StatusPending && sub.Status != models.StatusApproved {
		h.answerCallback(cb.ID, i18n.T(lang, "cannot_withdraw"))
		return
	}

	wasApproved := sub.Status == models.StatusApproved
	channelMsgID := sub.ChannelMessageID

	if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusWithdrawn); err != nil {
		h.Logger.Error("withdraw submission", "submission_id", submissionID, "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	h.answerCallback(cb.ID, i18n.T(lang, "submission_withdrawn"))

	if wasApproved && channelMsgID != 0 {
		notice := i18n.T(lang, "channel_withdrawn_notice")
		kb := emptyKeyboard()
		if sub.PhotoFileID != "" {
			editCaption := tgbotapi.EditMessageCaptionConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:      h.Config.ChannelID,
					MessageID:   int(channelMsgID),
					ReplyMarkup: &kb,
				},
				Caption: notice,
			}
			if _, err := h.Bot.Send(editCaption); err != nil {
				h.Logger.Error("edit channel caption on withdraw", "submission_id", submissionID, "error", err)
			}
		} else {
			edit := tgbotapi.NewEditMessageText(h.Config.ChannelID, int(channelMsgID), notice)
			edit.ReplyMarkup = &kb
			if _, err := h.Bot.Send(edit); err != nil {
				h.Logger.Error("edit channel message on withdraw", "submission_id", submissionID, "error", err)
			}
		}
	}

	h.editToSubmissionsList(cb, telegramID, lang)
}

func (h *Handler) handleLanguageSelection(cb *tgbotapi.CallbackQuery, userID int64, chatID int64, data string) {
	langMap := map[string]string{
		"lang_kk": "kk",
		"lang_ru": "ru",
		"lang_en": "en",
	}
	lang := langMap[data]
	h.answerCallback(cb.ID, "")

	conv := h.getConversation(userID)
	if conv.Stage != StageSelectLanguage {
		if err := h.Storage.UpdateUserLanguage(userID, lang); err != nil {
			h.Logger.Error("update user language", "telegram_id", userID, "error", err)
			return
		}
		h.resetConversation(userID)
		h.showMainMenu(chatID, userID, lang)
		return
	}

	_, err := h.Storage.CreateUser(userID, lang)
	if err != nil {
		h.Logger.Error("create user", "telegram_id", userID, "error", err)
		return
	}
	h.resetConversation(userID)
	h.showMainMenu(chatID, userID, lang)
}

func (h *Handler) toggleModeratorMode(chatID int64, moderatorID int64, lang string) {
	current, err := h.Storage.GetModeratorMode(moderatorID)
	if err != nil {
		h.Logger.Error("get moderator mode", "moderator_id", moderatorID, "error", err)
		return
	}
	next := models.ModeList
	if current == models.ModeList {
		next = models.ModeStream
	}
	if err := h.Storage.SetModeratorMode(moderatorID, next); err != nil {
		h.Logger.Error("set moderator mode", "moderator_id", moderatorID, "error", err)
		return
	}
	key := "mode_switched_list"
	if next == models.ModeStream {
		key = "mode_switched_stream"
	}
	h.send(chatID, i18n.T(lang, key))
	h.showModeratorMenu(chatID, moderatorID)
}
