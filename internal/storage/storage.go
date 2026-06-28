// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package storage

import "qazfreelance/internal/models"

type Storage interface {
	GetUser(telegramID int64) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	CreateUser(telegramID int64, language string) (*models.User, error)
	UpdateUserLanguage(telegramID int64, language string) error

	CreateSubmission(sub *models.Submission) (int64, error)
	GetSubmission(id int64) (*models.Submission, error)
	GetSubmissionsByUserID(userID int64) ([]*models.Submission, error)
	GetPendingSubmissions() ([]*models.Submission, error)
	GetConflictSubmissions() ([]*models.Submission, error)
	UpdateSubmissionStatus(id int64, status string) error
	SetChannelMessageID(submissionID int64, messageID int) error

	AddDecision(decision *models.ModeratorDecision) error
	GetDecisions(submissionID int64) ([]*models.ModeratorDecision, error)
	HasModeratorDecided(submissionID, moderatorID int64) (bool, error)

	GetModeratorMode(moderatorID int64) (string, error)
	SetModeratorMode(moderatorID int64, mode string) error

	Close() error
}
