package store

import (
	"testing"
)

func TestStore_SettingsRoundTrip(t *testing.T) {
	s := newTestStore(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSettings(uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("novo usuário deve ter settings vazios, got %v", got)
	}

	if err := s.SetSettings(uid, map[string]string{"contract": "clt-40", "lunch": "75"}); err != nil {
		t.Fatal(err)
	}
	got, err = s.GetSettings(uid)
	if err != nil {
		t.Fatal(err)
	}
	if got["contract"] != "clt-40" || got["lunch"] != "75" {
		t.Errorf("settings = %v", got)
	}

	// update (upsert)
	if err := s.SetSettings(uid, map[string]string{"contract": "pj"}); err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetSettings(uid)
	if got["contract"] != "pj" {
		t.Errorf("upsert falhou: %v", got)
	}
}

func TestStore_SettingsArePerUser(t *testing.T) {
	s := newTestStore(t)

	a, err := s.CreateUser("a@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateUser("b@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetSettings(a, map[string]string{"contract": "clt-40"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSettings(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("settings vazaram entre usuários: %v", got)
	}
}

func TestStore_DeleteSettings(t *testing.T) {
	s := newTestStore(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetSettings(uid, map[string]string{"contract": "pj"}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSettings(uid); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetSettings(uid)
	if len(got) != 0 {
		t.Errorf("após delete: %v", got)
	}
}
