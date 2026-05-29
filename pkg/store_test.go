package paste

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	t.Run("ID should only be based on content", func(t *testing.T) {
		s := NewEmptyStore()
		id1 := s.Insert("abc", time.Minute*10, VisibilityPublic)
		id2 := s.Insert("abc", time.Minute*15, VisibilityPublic)
		id3 := s.Insert("abc", time.Minute*10, VisibilityUnlisted)

		if id1 != id2 || id1 != id3 {
			t.Error("Returned ID should only be on the basis of content")
		}

		if s.Size() != 1 {
			t.Error("Unchanged content should not create new entries")
		}
	})

	t.Run("Lookup IDs", func(t *testing.T) {
		s := NewEmptyStore()
		s.Insert("abc", time.Minute*10, VisibilityPublic)

		if _, ok := s.Get(GetID("abc")); !ok {
			t.Error("Expected 'abc' to be found from prev test insert")
		}

		if _, ok := s.Get(""); ok {
			t.Error("Incorrectly returned entry for empty string id")
		}
	})

	t.Run("Delete with IDs", func(t *testing.T) {
		s := NewEmptyStore()
		s.Insert("abc", time.Minute*10, VisibilityPublic)

		id := GetID("abc")
		if _, ok := s.Get(id); !ok {
			t.Fatal("Entry 'abc' missing, expected to be present")
		}

		if !s.Delete(id) {
			t.Error("Deleting a present entry returned false")
		}

		if _, ok := s.Get(id); ok {
			t.Error("Delete didn't delete the entry from store")
		}
	})

	t.Run("Read after expiry", func(t *testing.T) {
		s := NewEmptyStore()
		id := s.Insert("abc", time.Minute*10, VisibilityPublic)

		futureTime := time.Now().Add(time.Minute * 11)
		if _, ok := s.get(id, futureTime); ok {
			t.Error("Expected entry to be hidden after expiry time")
		}

		if s.Size() != 1 {
			t.Error("Get should hide expired entries, but not delete them")
		}
	})

	t.Run("Cleanup removes only expired entries", func(t *testing.T) {
		s := NewEmptyStore()
		s.Insert("short_lived", time.Minute*10, VisibilityPublic)
		s.Insert("long_lived", time.Minute*30, VisibilityPublic)

		futureTime := time.Now().Add(time.Minute * 15)
		if deleted := s.cleanup(futureTime); deleted != 1 {
			t.Errorf("Expected 1 entry to be cleaned up, got %d", deleted)
		}

		if s.Size() != 1 {
			t.Errorf("Expected store size to be 1, got %d", s.Size())
		}

		if _, ok := s.get(GetID("long_lived"), futureTime); !ok {
			t.Error("Expected 'long_lived' to survive the cleanup")
		}
	})

	t.Run("Load and dump test", func(t *testing.T) {
		tmpDir := t.TempDir()
		savepath := filepath.Join(tmpDir, "pastebin.gob")

		_, err := NewStore("does/not/exist")
		if err == nil {
			t.Error("Exepcted missing file load to fail")
		}

		s := NewEmptyStore()
		id1 := s.Insert("abc", time.Minute*10, VisibilityPublic)
		id2 := s.Insert("xyz", time.Minute*30, VisibilityUnlisted)

		if err := s.Dump(savepath); err != nil {
			t.Fatalf("Dumping contents failed, %v", err)
		} else if s, err = NewStore(savepath); err != nil {
			t.Fatalf("Loading contents failed, %v", err)
		} else if size := s.Size(); size != 2 {
			t.Errorf("Expected 2 entries, got %v", size)
		}

		futureTime := time.Now().Add(time.Minute * 15)
		if _, ok := s.get(id1, futureTime); ok {
			t.Error("Expected get 'abc' to return false, expiry not persisted across writes")
		} else if m, ok := s.get(id2, futureTime); !ok || m.Content != "xyz" || m.Visibility != VisibilityUnlisted {
			t.Error("Fields not persisted across writes")
		}
	})

	t.Run("ListPublic only returns public pastes", func(t *testing.T) {
		s := NewEmptyStore()
		s.Insert("xyz", time.Minute*10, VisibilityUnlisted)
		pid := s.Insert("abc", time.Minute*10, VisibilityPublic)

		publicPastes := s.ListPublic()
		if size := len(publicPastes); size != 1 {
			t.Errorf("Expected exactly 1 entry from ListPublic, found %v", size)
		} else if pst := publicPastes[0]; pst.Id != pid || pst.Visibility != VisibilityPublic {
			t.Error("Returned entry doesn't match expected")
		}
	})
}
