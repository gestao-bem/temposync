package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gestao-bem/temposync/internal/models"
)

const requestTimeFormat = "2006-01-02 15:04:05"

func (s *SQLiteStore) CreateRequest(userID int64, kind string, day time.Time, minutes int, reason string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO requests (user_id, kind, day, minutes, reason) VALUES (?, ?, ?, ?, ?)",
		userID, kind, day.UTC().Format(requestTimeFormat), minutes, reason,
	)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) ListRequests(userID int64) ([]models.Request, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, kind, day, minutes, reason, status, created_at, decided_at "+
			"FROM requests WHERE user_id = ? ORDER BY id DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Request
	for rows.Next() {
		var r models.Request
		if err := rows.Scan(&r.ID, &r.UserID, &r.Kind, &r.Day, &r.Minutes, &r.Reason, &r.Status, &r.CreatedAt, &r.DecidedAt); err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	return out, nil
}

// DecideRequest só altera solicitação pendente do próprio usuário.
func (s *SQLiteStore) DecideRequest(userID, requestID int64, status string) error {
	result, err := s.db.Exec(
		"UPDATE requests SET status = ?, decided_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ? AND status = 'pendente'",
		status, requestID, userID,
	)
	if err != nil {
		return fmt.Errorf("decide request: %w", err)
	}
	if n, err := result.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("decide request: %w", sql.ErrNoRows)
	}
	return nil
}

func (s *SQLiteStore) CreateFolga(userID int64, day time.Time, minutes int, note string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO punches (user_id, happened_at, kind, note, minutes) VALUES (?, ?, 'folga', ?, ?)",
		userID, day.UTC().Format(punchTimeFormat), note, minutes,
	)
	if err != nil {
		return 0, fmt.Errorf("create folga: %w", err)
	}
	return result.LastInsertId()
}
