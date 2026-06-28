// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package handlers

import (
	"fmt"
	"html"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"qazfreelance/internal/i18n"
	"qazfreelance/internal/models"
)

func (h *Handler) HandleModeCommand(msg *tgbotapi.Message) {
	if !h.Config.IsModerator(msg.From.ID) {
		lang := h.getUserLang(msg.From.ID)
		h.send(msg.Chat.ID, i18n.T(lang, "unknown_command"))
		return
	}
	lang := h.getUserLang(msg.From.ID)
	h.toggleModeratorMode(msg.Chat.ID, msg.From.ID, lang)
}

func (h *Handler) showModeratorMenu(chatID int64, moderatorID int64) {
	lang := h.getUserLang(moderatorID)

	mode, err := h.Storage.GetModeratorMode(moderatorID)
	if err != nil {
		h.Logger.Error("get moderator mode for menu", "moderator_id", moderatorID, "error", err)
		mode = models.ModeStream
	}

	modeKey := "mode_current_stream"
	if mode == models.ModeList {
		modeKey = "mode_current_list"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_open_list"), "open_list"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_conflicts"), "open_conflicts"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_toggle_mode"), "toggle_mode"),
		),
	)
	text := i18n.T(lang, "moderator_menu") + "\n" + i18n.T(lang, modeKey)
	h.sendWithInline(chatID, text, keyboard)
}

func (h *Handler) sendSubmissionToModerator(moderatorID int64, sub *models.Submission) {
	lang := h.getUserLang(moderatorID)
	text := i18n.T(lang, "submission_new") + "\n\n" + h.buildSubmissionText(lang, sub)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.T(lang, "btn_approve"),
				fmt.Sprintf("approve_%d", sub.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.T(lang, "btn_reject"),
				fmt.Sprintf("reject_%d", sub.ID),
			),
		),
	)
	h.sendWithInline(moderatorID, text, keyboard)
}

func (h *Handler) handleApprove(cb *tgbotapi.CallbackQuery, moderatorID int64, submissionID int64) {
	lang := h.getUserLang(moderatorID)

	decided, err := h.Storage.HasModeratorDecided(submissionID, moderatorID)
	if err != nil {
		h.Logger.Error("check moderator decision", "error", err)
		h.answerCallback(cb.ID, "")
		return
	}
	if decided {
		h.answerCallback(cb.ID, i18n.T(lang, "already_decided"))
		return
	}

	sub, err := h.Storage.GetSubmission(submissionID)
	if err != nil || sub == nil {
		h.Logger.Error("get submission for approve", "id", submissionID, "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	if sub.Status != models.StatusPending && sub.Status != models.StatusConflict {
		h.answerCallback(cb.ID, i18n.T(lang, "already_decided"))
		return
	}

	if err := h.Storage.AddDecision(&models.ModeratorDecision{
		SubmissionID: submissionID,
		ModeratorID:  moderatorID,
		Decision:     models.DecisionApprove,
	}); err != nil {
		h.Logger.Error("add approve decision", "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	decisions, err := h.Storage.GetDecisions(submissionID)
	if err != nil {
		h.Logger.Error("get decisions after approve", "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	approvals, rejections := countDecisions(decisions)
	h.answerCallback(cb.ID, i18n.T(lang, "decision_recorded"))

	switch {
	case approvals > 0 && rejections == 0:
		if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusApproved); err != nil {
			h.Logger.Error("update submission status to approved", "error", err)
			return
		}
		h.publishToChannel(sub)
		h.notifyUser(sub, "submission_approved")

	case approvals > 0 && rejections > 0:
		if sub.Status != models.StatusConflict {
			if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusConflict); err != nil {
				h.Logger.Error("update submission status to conflict", "error", err)
				return
			}
			h.notifyModeratorsConflict(sub)
		}
	}
}

func (h *Handler) handleReject(cb *tgbotapi.CallbackQuery, moderatorID int64, submissionID int64) {
	lang := h.getUserLang(moderatorID)

	decided, err := h.Storage.HasModeratorDecided(submissionID, moderatorID)
	if err != nil {
		h.Logger.Error("check moderator decision", "error", err)
		h.answerCallback(cb.ID, "")
		return
	}
	if decided {
		h.answerCallback(cb.ID, i18n.T(lang, "already_decided"))
		return
	}

	sub, err := h.Storage.GetSubmission(submissionID)
	if err != nil || sub == nil {
		h.Logger.Error("get submission for reject", "id", submissionID, "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	if sub.Status != models.StatusPending && sub.Status != models.StatusConflict {
		h.answerCallback(cb.ID, i18n.T(lang, "already_decided"))
		return
	}

	if err := h.Storage.AddDecision(&models.ModeratorDecision{
		SubmissionID: submissionID,
		ModeratorID:  moderatorID,
		Decision:     models.DecisionReject,
	}); err != nil {
		h.Logger.Error("add reject decision", "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	decisions, err := h.Storage.GetDecisions(submissionID)
	if err != nil {
		h.Logger.Error("get decisions after reject", "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	approvals, rejections := countDecisions(decisions)
	h.answerCallback(cb.ID, i18n.T(lang, "decision_recorded"))

	switch {
	case rejections > 0 && approvals == 0:
		if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusRejected); err != nil {
			h.Logger.Error("update submission status to rejected", "error", err)
			return
		}
		h.notifyUser(sub, "submission_rejected")

	case approvals > 0 && rejections > 0:
		if sub.Status != models.StatusConflict {
			if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusConflict); err != nil {
				h.Logger.Error("update submission status to conflict", "error", err)
				return
			}
			h.notifyModeratorsConflict(sub)
		}
	}
}

func (h *Handler) handleResolve(cb *tgbotapi.CallbackQuery, moderatorID int64, submissionID int64, decision string) {
	lang := h.getUserLang(moderatorID)

	sub, err := h.Storage.GetSubmission(submissionID)
	if err != nil || sub == nil {
		h.Logger.Error("get submission for resolve", "id", submissionID, "error", err)
		h.answerCallback(cb.ID, "")
		return
	}

	if sub.Status != models.StatusConflict {
		h.answerCallback(cb.ID, i18n.T(lang, "already_decided"))
		return
	}

	h.answerCallback(cb.ID, i18n.T(lang, "decision_recorded"))

	switch decision {
	case models.DecisionApprove:
		if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusApproved); err != nil {
			h.Logger.Error("resolve conflict to approved", "error", err)
			return
		}
		h.publishToChannel(sub)
		h.notifyUser(sub, "submission_approved")

	case models.DecisionReject:
		if err := h.Storage.UpdateSubmissionStatus(submissionID, models.StatusRejected); err != nil {
			h.Logger.Error("resolve conflict to rejected", "error", err)
			return
		}
		h.notifyUser(sub, "submission_rejected")
	}
}

func (h *Handler) showPendingList(chatID int64, moderatorID int64) {
	lang := h.getUserLang(moderatorID)

	subs, err := h.Storage.GetPendingSubmissions()
	if err != nil {
		h.Logger.Error("get pending submissions", "error", err)
		return
	}
	if len(subs) == 0 {
		h.send(chatID, i18n.T(lang, "no_pending"))
		return
	}
	h.sendPendingSubmission(chatID, lang, subs[0], len(subs) > 1)
}

func (h *Handler) showNextPending(chatID int64, moderatorID int64, afterID int64) {
	lang := h.getUserLang(moderatorID)

	subs, err := h.Storage.GetPendingSubmissions()
	if err != nil {
		h.Logger.Error("get pending submissions for next", "error", err)
		return
	}

	var found bool
	for i, sub := range subs {
		if sub.ID == afterID {
			found = true
			if i+1 < len(subs) {
				h.sendPendingSubmission(chatID, lang, subs[i+1], i+2 < len(subs))
			} else {
				h.send(chatID, i18n.T(lang, "list_end"))
			}
			break
		}
	}
	if !found {
		if len(subs) == 0 {
			h.send(chatID, i18n.T(lang, "no_pending"))
		} else {
			h.sendPendingSubmission(chatID, lang, subs[0], len(subs) > 1)
		}
	}
}

func (h *Handler) sendPendingSubmission(chatID int64, lang string, sub *models.Submission, hasNext bool) {
	text := h.buildSubmissionText(lang, sub)
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.T(lang, "btn_approve"),
				fmt.Sprintf("approve_%d", sub.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.T(lang, "btn_reject"),
				fmt.Sprintf("reject_%d", sub.ID),
			),
		},
	}
	if hasNext {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.T(lang, "btn_next"),
				fmt.Sprintf("next_%d", sub.ID),
			),
		})
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.sendWithInline(chatID, text, keyboard)
}

func (h *Handler) showConflictList(chatID int64, moderatorID int64) {
	lang := h.getUserLang(moderatorID)

	subs, err := h.Storage.GetConflictSubmissions()
	if err != nil {
		h.Logger.Error("get conflict submissions", "error", err)
		return
	}
	if len(subs) == 0 {
		h.send(chatID, i18n.T(lang, "no_conflicts"))
		return
	}
	for _, sub := range subs {
		text := h.buildSubmissionText(lang, sub)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.T(lang, "btn_resolve_approve"),
					fmt.Sprintf("resolve_approve_%d", sub.ID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.T(lang, "btn_resolve_reject"),
					fmt.Sprintf("resolve_reject_%d", sub.ID),
				),
			),
		)
		h.sendWithInline(chatID, text, keyboard)
	}
}

func contactToURL(contact string) string {
	c := strings.TrimSpace(contact)
	if strings.HasPrefix(c, "@") {
		return "https://t.me/" + c[1:]
	}
	if strings.HasPrefix(c, "https://") || strings.HasPrefix(c, "http://") {
		return c
	}
	if strings.HasPrefix(c, "t.me/") {
		return "https://" + c
	}
	return ""
}

func (h *Handler) buildChannelPost(lang string, sub *models.Submission) (text string, contactURL string) {
	typeName := i18n.T(lang, "submission_type_order")
	if sub.Type == models.SubmissionTypeResume {
		typeName = i18n.T(lang, "submission_type_resume")
	}

	var b strings.Builder
	b.WriteString("<b>")
	b.WriteString(html.EscapeString(typeName))
	b.WriteString("</b>")

	if sub.Title != "" {
		b.WriteString("\n\n<b>")
		b.WriteString(html.EscapeString(sub.Title))
		b.WriteString("</b>")
	}

	if sub.Description != "" {
		b.WriteString("\n\n")
		b.WriteString(html.EscapeString(sub.Description))
	}

	contactURL = contactToURL(sub.Contact)
	if sub.Contact != "" && contactURL == "" {
		b.WriteString("\n\n")
		b.WriteString(html.EscapeString(i18n.Tf(lang, "contact_label", sub.Contact)))
	}

	return b.String(), contactURL
}

func channelKeyboard(lang string, contactLabel string, contactURL string) *tgbotapi.InlineKeyboardMarkup {
	if contactURL == "" {
		return nil
	}
	label := i18n.T(lang, "btn_contact")
	if contactLabel != "" {
		label = contactLabel
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(label, contactURL),
		),
	)
	return &kb
}

func (h *Handler) publishToChannel(sub *models.Submission) {
	user, err := h.Storage.GetUserByID(sub.UserID)
	lang := "en"
	if err != nil || user == nil {
		h.Logger.Error("get user language for channel publish", "user_id", sub.UserID, "error", err)
	} else {
		lang = user.Language
	}

	text, contactURL := h.buildChannelPost(lang, sub)
	kb := channelKeyboard(lang, sub.Contact, contactURL)

	var sentMsg tgbotapi.Message
	var sendErr error

	if sub.PhotoFileID != "" {
		photo := tgbotapi.NewPhoto(h.Config.ChannelID, tgbotapi.FileID(sub.PhotoFileID))
		photo.Caption = text
		photo.ParseMode = tgbotapi.ModeHTML
		if kb != nil {
			photo.ReplyMarkup = kb
		}
		sentMsg, sendErr = h.Bot.Send(photo)
	} else {
		msg := tgbotapi.NewMessage(h.Config.ChannelID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		if kb != nil {
			msg.ReplyMarkup = kb
		}
		sentMsg, sendErr = h.Bot.Send(msg)
	}

	if sendErr != nil {
		h.Logger.Error("publish to channel", "submission_id", sub.ID, "error", sendErr)
		return
	}

	if err := h.Storage.SetChannelMessageID(sub.ID, sentMsg.MessageID); err != nil {
		h.Logger.Error("save channel message id", "submission_id", sub.ID, "error", err)
	}
}

func (h *Handler) notifyUser(sub *models.Submission, msgKey string) {
	user, err := h.Storage.GetUserByID(sub.UserID)
	if err != nil || user == nil {
		h.Logger.Error("get user for notification", "user_id", sub.UserID, "error", err)
		return
	}
	h.send(user.TelegramID, i18n.T(user.Language, msgKey))
}

func (h *Handler) notifyModeratorsConflict(sub *models.Submission) {
	for _, modID := range h.Config.ModeratorIDs {
		lang := h.getUserLang(modID)
		text := i18n.Tf(lang, "conflict_notify", sub.ID)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, "btn_conflicts"), "open_conflicts"),
			),
		)
		h.sendWithInline(modID, text, keyboard)
	}
}

func countDecisions(decisions []*models.ModeratorDecision) (approvals, rejections int) {
	for _, d := range decisions {
		switch d.Decision {
		case models.DecisionApprove:
			approvals++
		case models.DecisionReject:
			rejections++
		}
	}
	return
}
