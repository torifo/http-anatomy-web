package store

import (
	"testing"

	"http-anatomy/internal/model"
)

func TestAddTodoMonotonicIDs(t *testing.T) {
	s := New()
	a := s.AddTodo("sess", "first")
	b := s.AddTodo("sess", "second")
	if a.ID != 1 || b.ID != 2 {
		t.Fatalf("want ids 1,2 got %d,%d", a.ID, b.ID)
	}
}

func TestToggleAndUpdateTodo(t *testing.T) {
	s := New()
	tdo := s.AddTodo("sess", "task")
	got, ok := s.ToggleTodo("sess", tdo.ID)
	if !ok || !got.Done {
		t.Fatalf("toggle failed: ok=%v done=%v", ok, got.Done)
	}
	got, ok = s.UpdateTodo("sess", tdo.ID, "renamed")
	if !ok || got.Title != "renamed" {
		t.Fatalf("update failed: ok=%v title=%q", ok, got.Title)
	}
	if _, ok := s.ToggleTodo("sess", 999); ok {
		t.Fatal("toggle on missing id should fail")
	}
}

func TestDeleteTodo(t *testing.T) {
	s := New()
	tdo := s.AddTodo("sess", "task")
	if !s.DeleteTodo("sess", tdo.ID) {
		t.Fatal("delete should succeed")
	}
	if s.DeleteTodo("sess", tdo.ID) {
		t.Fatal("second delete should fail")
	}
	if len(s.Todos("sess")) != 0 {
		t.Fatal("todos should be empty after delete")
	}
}

func TestSessionIsolation(t *testing.T) {
	s := New()
	s.AddTodo("alice", "a-task")
	s.AddUser("bob", "Bob", "bob@example.com")
	if len(s.Todos("bob")) != 0 {
		t.Fatal("bob should have no todos")
	}
	if len(s.Users("alice")) != 0 {
		t.Fatal("alice should have no users")
	}
	// IDs are per-session, so each starts at 1.
	if first := s.AddTodo("carol", "x"); first.ID != 1 {
		t.Fatalf("new session should start id at 1, got %d", first.ID)
	}
}

func TestUserCRUD(t *testing.T) {
	s := New()
	u := s.AddUser("sess", "Ada", "ada@example.com")
	if u.ID != 1 {
		t.Fatalf("want id 1 got %d", u.ID)
	}
	got, ok := s.UpdateUser("sess", u.ID, "Ada L.", "ada.l@example.com")
	if !ok || got.Name != "Ada L." || got.Email != "ada.l@example.com" {
		t.Fatalf("update user failed: %+v ok=%v", got, ok)
	}
	if !s.DeleteUser("sess", u.ID) {
		t.Fatal("delete user should succeed")
	}
}

func TestHistoryOrderAndTrim(t *testing.T) {
	s := New()
	for i := 1; i <= 12; i++ {
		s.AppendHistory("sess", model.Exchange{Status: i})
	}
	h := s.History("sess")
	if len(h) != maxHistory {
		t.Fatalf("history should be trimmed to %d, got %d", maxHistory, len(h))
	}
	// Newest first: last appended (status 12) is at index 0.
	if h[0].Status != 12 {
		t.Fatalf("newest should be first, got status %d", h[0].Status)
	}
	if h[len(h)-1].Status != 3 {
		t.Fatalf("oldest kept should be status 3, got %d", h[len(h)-1].Status)
	}
}
