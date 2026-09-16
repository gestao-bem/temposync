package store

import "testing"

func TestStore_PostsSeeded(t *testing.T) {
	s := newTestStore(t)

	posts, err := s.ListPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) < 6 {
		t.Fatalf("seed deve trazer 6 artigos, got %d", len(posts))
	}
	p, err := s.PostBySlug("tolerancia-5-minutos-clt-art-58")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title == "" || len(p.Body) < 200 || p.ReadMin == 0 {
		t.Errorf("post incompleto: %+v", p)
	}
	if _, err := s.PostBySlug("nao-existe"); err == nil {
		t.Errorf("slug inexistente deve retornar erro")
	}
}

func TestStore_SubscribeNewsletterIdempotent(t *testing.T) {
	s := newTestStore(t)

	added, err := s.SubscribeNewsletter("Alguem@Example.com ")
	if err != nil {
		t.Fatal(err)
	}
	if !added {
		t.Errorf("primeira inscrição deve inserir")
	}
	added, err = s.SubscribeNewsletter("alguem@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if added {
		t.Errorf("duplicada não deve inserir de novo")
	}
	n, err := s.CountNewsletter()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("count = %d, want 1", n)
	}
}
