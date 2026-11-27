package store

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.db.Close() })
	return s
}

func TestGetUserByID(t *testing.T) {
	s := newTestStore(t)
	id, err := s.CreateUser("alice", "fp-alice", "Alice", "hello world", "en")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := s.GetUserByID(id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Username != "alice" || u.DisplayName != "Alice" || u.Tagline != "hello world" {
		t.Fatalf("unexpected user: %+v", u)
	}

	// 不存在的 ID 应返回 nil 而非报错
	missing, err := s.GetUserByID(99999)
	if err != nil {
		t.Fatalf("GetUserByID missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing user, got %+v", missing)
	}
}

func TestListApprovedUsers(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.CreateUser("alice", "fp-alice", "Alice", "", "en")
	if err != nil {
		t.Fatalf("CreateUser alice: %v", err)
	}
	id2, err := s.CreateUser("bob", "fp-bob", "Bob", "", "en")
	if err != nil {
		t.Fatalf("CreateUser bob: %v", err)
	}
	id3, err := s.CreateUser("carol", "fp-carol", "", "", "en")
	if err != nil {
		t.Fatalf("CreateUser carol: %v", err)
	}

	// alice / bob 通过（approved），carol 未通过（pending）
	if err := s.RecordAttempt(id1, 30, 30, true); err != nil {
		t.Fatalf("RecordAttempt alice: %v", err)
	}
	if err := s.RecordAttempt(id2, 28, 30, true); err != nil {
		t.Fatalf("RecordAttempt bob: %v", err)
	}
	if err := s.RecordAttempt(id3, 5, 30, false); err != nil {
		t.Fatalf("RecordAttempt carol: %v", err)
	}

	users, err := s.ListApprovedUsers()
	if err != nil {
		t.Fatalf("ListApprovedUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 approved users, got %d", len(users))
	}

	names := map[string]bool{}
	for _, u := range users {
		names[u.Username] = true
	}
	if !names["alice"] || !names["bob"] {
		t.Fatalf("unexpected approved users: %v", names)
	}
	if names["carol"] {
		t.Fatalf("carol (pending) should not be listed as approved")
	}
}
