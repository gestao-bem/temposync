package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gestao-bem/temposync/internal/models"
)

const punchTimeFormat = "2006-01-02 15:04:05"

func (s *SQLiteStore) CreatePunch(userID int64, at time.Time, kind string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO punches (user_id, happened_at, kind) VALUES (?, ?, ?)",
		userID, at.UTC().Format(punchTimeFormat), kind,
	)
	if err != nil {
		return 0, fmt.Errorf("create punch: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) ListPunches(userID int64, from, to time.Time) ([]models.Punch, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, happened_at, kind, note, minutes, created_at FROM punches "+
			"WHERE user_id = ? AND happened_at >= ? AND happened_at < ? ORDER BY happened_at ASC",
		userID, from.UTC().Format(punchTimeFormat), to.UTC().Format(punchTimeFormat),
	)
	if err != nil {
		return nil, fmt.Errorf("list punches: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Punch
	for rows.Next() {
		var p models.Punch
		if err := rows.Scan(&p.ID, &p.UserID, &p.HappenedAt, &p.Kind, &p.Note, &p.Minutes, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan punch: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list punches: %w", err)
	}
	return out, nil
}

func (s *SQLiteStore) FindUserByID(id int64) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *SQLiteStore) DeletePunch(userID, punchID int64) error {
	result, err := s.db.Exec("DELETE FROM punches WHERE id = ? AND user_id = ?", punchID, userID)
	if err != nil {
		return fmt.Errorf("delete punch: %w", err)
	}
	if n, err := result.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("delete punch: %w", sql.ErrNoRows)
	}
	return nil
}
