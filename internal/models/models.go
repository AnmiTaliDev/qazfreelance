// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package models

import "time"

const (
	SubmissionTypeOrder  = "order"
	SubmissionTypeResume = "resume"

	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusConflict  = "conflict"
	StatusWithdrawn = "withdrawn"

	DecisionApprove = "approve"
	DecisionReject  = "reject"

	ModeStream = "stream"
	ModeList   = "list"
)

type User struct {
	ID         int64
	TelegramID int64
	Language   string
	CreatedAt  time.Time
}

type Submission struct {
	ID               int64
	UserID           int64
	Type             string
	Title            string
	Description      string
	Contact          string
	PhotoFileID      string
	Status           string
	ChannelMessageID int64
	CreatedAt        time.Time
}

type ModeratorDecision struct {
	ID           int64
	SubmissionID int64
	ModeratorID  int64
	Decision     string
	CreatedAt    time.Time
}

type ModeratorSettings struct {
	ModeratorID int64
	Mode        string
}
