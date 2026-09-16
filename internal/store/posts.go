package store

import (
	"fmt"
	"strings"

	"github.com/gestao-bem/temposync/internal/models"
)

func (s *SQLiteStore) ListPosts() ([]models.Post, error) {
	rows, err := s.db.Query(
		"SELECT id, slug, title, excerpt, body, category, read_min, author, published_at " +
			"FROM posts ORDER BY published_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.Body, &p.Category, &p.ReadMin, &p.Author, &p.PublishedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) PostBySlug(slug string) (models.Post, error) {
	var p models.Post
	err := s.db.QueryRow(
		"SELECT id, slug, title, excerpt, body, category, read_min, author, published_at FROM posts WHERE slug = ?",
		slug,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.Body, &p.Category, &p.ReadMin, &p.Author, &p.PublishedAt)
	if err != nil {
		return models.Post{}, fmt.Errorf("post by slug: %w", err)
	}
	return p, nil
}

// SubscribeNewsletter é idempotente: email repetido não é erro.
func (s *SQLiteStore) SubscribeNewsletter(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	result, err := s.db.Exec(
		"INSERT OR IGNORE INTO newsletter_subscribers (email) VALUES (?)",
		email,
	)
	if err != nil {
		return false, fmt.Errorf("subscribe: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("subscribe: %w", err)
	}
	return n > 0, nil
}

func (s *SQLiteStore) CountNewsletter() (int64, error) {
	var n int64
	if err := s.db.QueryRow("SELECT COUNT(*) FROM newsletter_subscribers").Scan(&n); err != nil {
		return 0, fmt.Errorf("count newsletter: %w", err)
	}
	return n, nil
}
