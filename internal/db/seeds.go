package db

import (
	"github.com/gestao-bem/temposync/internal/models"
	"github.com/gestao-bem/temposync/internal/store"
)

// RunSeeds populates demo data. Safe to run multiple times.
func RunSeeds(s store.Store) error {
	// cais:recurring-seeds
	// cais:seeds
	if _, err := s.InsertContact(models.Contact{
		Name:  "Demo",
		Email: "demo@example.com",
	}); err != nil {
		return err
	}
	return nil
}
