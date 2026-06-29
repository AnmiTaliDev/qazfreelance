// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package storage

import (
	"database/sql"
	"fmt"
	"time"

	"qazfreelance/internal/models"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER UNIQUE NOT NULL,
    language TEXT NOT NULL DEFAULT 'en',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    contact TEXT NOT NULL DEFAULT '',
    photo_file_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS moderator_decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id INTEGER NOT NULL,
    moderator_id INTEGER NOT NULL,
    decision TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (submission_id) REFERENCES submissions(id)
);

CREATE TABLE IF NOT EXISTS moderator_settings (
    moderator_id INTEGER PRIMARY KEY,
    mode TEXT NOT NULL DEFAULT 'stream'
);
`

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLite(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("run schema migration: %w", err)
	}
	db.Exec(`ALTER TABLE submissions ADD COLUMN photo_file_id TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE submissions ADD COLUMN channel_message_id INTEGER NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE users ADD COLUMN default_sub_mode TEXT NOT NULL DEFAULT ''`)
	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) GetUser(telegramID int64) (*models.User, error) {
	row := s.db.QueryRow(
		`SELECT id, telegram_id, language, default_sub_mode, created_at FROM users WHERE telegram_id = ?`,
		telegramID,
	)
	u := &models.User{}
	var createdAt string
	err := row.Scan(&u.ID, &u.TelegramID, &u.Language, &u.DefaultSubMode, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return u, nil
}

func (s *SQLiteStorage) GetUserByID(id int64) (*models.User, error) {
	row := s.db.QueryRow(
		`SELECT id, telegram_id, language, default_sub_mode, created_at FROM users WHERE id = ?`,
		id,
	)
	u := &models.User{}
	var createdAt string
	err := row.Scan(&u.ID, &u.TelegramID, &u.Language, &u.DefaultSubMode, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return u, nil
}

func (s *SQLiteStorage) CreateUser(telegramID int64, language string) (*models.User, error) {
	res, err := s.db.Exec(
		`INSERT INTO users (telegram_id, language) VALUES (?, ?)`,
		telegramID, language,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}
	return &models.User{
		ID:         id,
		TelegramID: telegramID,
		Language:   language,
		CreatedAt:  time.Now(),
	}, nil
}

func (s *SQLiteStorage) UpdateUserLanguage(telegramID int64, language string) error {
	_, err := s.db.Exec(
		`UPDATE users SET language = ? WHERE telegram_id = ?`,
		language, telegramID,
	)
	if err != nil {
		return fmt.Errorf("update user language: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) UpdateUserDefaultSubMode(telegramID int64, mode string) error {
	_, err := s.db.Exec(
		`UPDATE users SET default_sub_mode = ? WHERE telegram_id = ?`,
		mode, telegramID,
	)
	if err != nil {
		return fmt.Errorf("update user default sub mode: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) CreateSubmission(sub *models.Submission) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO submissions (user_id, type, title, description, contact, photo_file_id, status, channel_message_id) VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		sub.UserID, sub.Type, sub.Title, sub.Description, sub.Contact, sub.PhotoFileID, sub.Status,
	)
	if err != nil {
		return 0, fmt.Errorf("create submission: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

func (s *SQLiteStorage) GetSubmission(id int64) (*models.Submission, error) {
	row := s.db.QueryRow(
		`SELECT id, user_id, type, title, description, contact, photo_file_id, status, channel_message_id, created_at FROM submissions WHERE id = ?`,
		id,
	)
	sub := &models.Submission{}
	var createdAt string
	err := row.Scan(&sub.ID, &sub.UserID, &sub.Type, &sub.Title, &sub.Description, &sub.Contact, &sub.PhotoFileID, &sub.Status, &sub.ChannelMessageID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get submission: %w", err)
	}
	sub.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return sub, nil
}

func (s *SQLiteStorage) GetSubmissionsByUserID(userID int64) ([]*models.Submission, error) {
	return s.querySubmissions(
		`SELECT id, user_id, type, title, description, contact, photo_file_id, status, channel_message_id, created_at
		 FROM submissions WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
}

func (s *SQLiteStorage) GetPendingSubmissions() ([]*models.Submission, error) {
	return s.querySubmissions(
		`SELECT id, user_id, type, title, description, contact, photo_file_id, status, channel_message_id, created_at
		 FROM submissions WHERE status = 'pending' ORDER BY created_at ASC`)
}

func (s *SQLiteStorage) GetConflictSubmissions() ([]*models.Submission, error) {
	return s.querySubmissions(
		`SELECT id, user_id, type, title, description, contact, photo_file_id, status, channel_message_id, created_at
		 FROM submissions WHERE status = 'conflict' ORDER BY created_at ASC`)
}

func (s *SQLiteStorage) querySubmissions(query string, args ...any) ([]*models.Submission, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query submissions: %w", err)
	}
	defer rows.Close()

	var subs []*models.Submission
	for rows.Next() {
		sub := &models.Submission{}
		var createdAt string
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Type, &sub.Title, &sub.Description, &sub.Contact, &sub.PhotoFileID, &sub.Status, &sub.ChannelMessageID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan submission: %w", err)
		}
		sub.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (s *SQLiteStorage) SetChannelMessageID(submissionID int64, messageID int) error {
	_, err := s.db.Exec(
		`UPDATE submissions SET channel_message_id = ? WHERE id = ?`,
		messageID, submissionID,
	)
	if err != nil {
		return fmt.Errorf("set channel message id: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) UpdateSubmissionStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE submissions SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("update submission status: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) AddDecision(decision *models.ModeratorDecision) error {
	_, err := s.db.Exec(
		`INSERT INTO moderator_decisions (submission_id, moderator_id, decision) VALUES (?, ?, ?)`,
		decision.SubmissionID, decision.ModeratorID, decision.Decision,
	)
	if err != nil {
		return fmt.Errorf("add decision: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) GetDecisions(submissionID int64) ([]*models.ModeratorDecision, error) {
	rows, err := s.db.Query(
		`SELECT id, submission_id, moderator_id, decision, created_at FROM moderator_decisions WHERE submission_id = ?`,
		submissionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get decisions: %w", err)
	}
	defer rows.Close()

	var decisions []*models.ModeratorDecision
	for rows.Next() {
		d := &models.ModeratorDecision{}
		var createdAt string
		if err := rows.Scan(&d.ID, &d.SubmissionID, &d.ModeratorID, &d.Decision, &createdAt); err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}
		d.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (s *SQLiteStorage) HasModeratorDecided(submissionID, moderatorID int64) (bool, error) {
	row := s.db.QueryRow(
		`SELECT COUNT(*) FROM moderator_decisions WHERE submission_id = ? AND moderator_id = ?`,
		submissionID, moderatorID,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("has moderator decided: %w", err)
	}
	return count > 0, nil
}

func (s *SQLiteStorage) GetModeratorMode(moderatorID int64) (string, error) {
	row := s.db.QueryRow(
		`SELECT mode FROM moderator_settings WHERE moderator_id = ?`,
		moderatorID,
	)
	var mode string
	err := row.Scan(&mode)
	if err == sql.ErrNoRows {
		return models.ModeStream, nil
	}
	if err != nil {
		return "", fmt.Errorf("get moderator mode: %w", err)
	}
	return mode, nil
}

func (s *SQLiteStorage) SetModeratorMode(moderatorID int64, mode string) error {
	_, err := s.db.Exec(
		`INSERT INTO moderator_settings (moderator_id, mode) VALUES (?, ?)
         ON CONFLICT(moderator_id) DO UPDATE SET mode = excluded.mode`,
		moderatorID, mode,
	)
	if err != nil {
		return fmt.Errorf("set moderator mode: %w", err)
	}
	return nil
}
