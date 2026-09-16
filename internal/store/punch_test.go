package store

import (
	"testing"
	"time"
)

func TestStore_CreatePunch(t *testing.T) {
	s := newTestStore(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreatePunch(uid, time.Date(2024, 10, 24, 8, 32, 10, 0, time.UTC), "entrada")
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Error("id = 0, want non-zero")
	}
}

func TestStore_ListPunchesDay(t *testing.T) {
	s := newTestStore(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2024, 10, 24, 0, 0, 0, 0, time.UTC)
	marks := []time.Time{
		base.Add(8*time.Hour + 32*time.Minute),
		base.Add(12*time.Hour + 5*time.Minute),
		base.Add(13*time.Hour + 10*time.Minute),
	}
	for _, at := range marks {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreatePunch(uid, base.AddDate(0, 0, 1), ""); err != nil {
		t.Fatal(err)
	}

	got, err := s.ListPunches(uid, base, base.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i := 1; i < len(got); i++ {
		if !got[i].HappenedAt.After(got[i-1].HappenedAt) {
			t.Errorf("punches not ordered at %d", i)
		}
	}
}

func TestStore_FindUserByID(t *testing.T) {
	s := newTestStore(t)

	id, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "lucas@example.com" {
		t.Errorf("email = %q", u.Email)
	}
}
