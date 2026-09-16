package store

import (
	"testing"
	"time"
)

func TestStore_RequestLifecycle(t *testing.T) {
	s := newTestStore(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	id, err := s.CreateRequest(uid, "folga", day, 480, "folga planejada")
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("id = 0")
	}
	list, err := s.ListRequests(uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Kind != "folga" || list[0].Minutes != 480 || list[0].Status != "pendente" {
		t.Fatalf("requests = %+v", list)
	}
	if err := s.DecideRequest(uid, id, "aprovada"); err != nil {
		t.Fatal(err)
	}
	list, _ = s.ListRequests(uid)
	if list[0].Status != "aprovada" || !list[0].DecidedAt.Valid {
		t.Errorf("decide não aplicou: %+v", list[0])
	}
	// não permite re-decidir
	if err := s.DecideRequest(uid, id, "cancelada"); err == nil {
		t.Errorf("re-decidir deveria falhar")
	}
}

func TestStore_RequestScopedToOwner(t *testing.T) {
	s := newTestStore(t)

	a, _ := s.CreateUser("a@example.com", "hash")
	b, _ := s.CreateUser("b@example.com", "hash")
	id, err := s.CreateRequest(a, "ajuste", time.Now().UTC(), 495, "esqueci")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DecideRequest(b, id, "aprovada"); err == nil {
		t.Errorf("usuário b não pode decidir request de a")
	}
	list, _ := s.ListRequests(b)
	if len(list) != 0 {
		t.Errorf("requests de a vazaram para b: %+v", list)
	}
}

func TestStore_CreateFolgaPunchWithMinutes(t *testing.T) {
	s := newTestStore(t)

	uid, _ := s.CreateUser("lucas@example.com", "hash")
	day := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	id, err := s.CreateFolga(uid, day, 480, "folga aprovada")
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("id = 0")
	}
	punches, err := s.ListPunches(uid, day.Add(-time.Hour), day.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(punches) != 1 || punches[0].Kind != "folga" || punches[0].Minutes != 480 {
		t.Fatalf("punches = %+v", punches)
	}
}
