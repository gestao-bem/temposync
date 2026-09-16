package store

import "fmt"

func (s *SQLiteStore) GetSettings(userID int64) (map[string]string, error) {
	rows, err := s.db.Query(
		"SELECT key, value FROM user_settings WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return out, nil
}

func (s *SQLiteStore) SetSettings(userID int64, kv map[string]string) error {
	for k, v := range kv {
		_, err := s.db.Exec(
			"INSERT INTO user_settings (user_id, key, value) VALUES (?, ?, ?) "+
				"ON CONFLICT(user_id, key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP",
			userID, k, v,
		)
		if err != nil {
			return fmt.Errorf("set setting %q: %w", k, err)
		}
	}
	return nil
}

func (s *SQLiteStore) DeleteSettings(userID int64) error {
	if _, err := s.db.Exec("DELETE FROM user_settings WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("delete settings: %w", err)
	}
	return nil
}
