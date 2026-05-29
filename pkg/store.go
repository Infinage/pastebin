package paste

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"
)

type Store struct {
	data map[string]*Model
	mu   sync.RWMutex
}

func NewEmptyStore() *Store {
	return &Store{data: map[string]*Model{}}
}

func NewStore(filename string) (*Store, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load store: %w", err)
	}
	defer f.Close()

	s := NewEmptyStore()
	dec := gob.NewDecoder(f)
	if err = dec.Decode(&s.data); err != nil {
		return nil, fmt.Errorf("failed to decode from file: %w", err)
	}

	return s, nil
}

// If missing, adds the entry. If already present updates it. Returns the hashed ID
func (s *Store) Insert(content string, expiry time.Duration, visibility Visibility) string {
	id := GetID(content)
	expireAt := time.Now().Add(expiry)
	model := &Model{Id: id, Content: content, ExpireAt: expireAt, Visibility: visibility}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = model

	return id
}

// Returns model from store if present and not already expired
func (s *Store) Get(id string) (Model, bool) {
	return s.get(id, time.Now())
}

// Deletes 'id' from store and returns true. If not found, returns false
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return false
	}

	delete(s.data, id)
	return true
}

func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Removes all expired entries, returns count of cleaned entries
func (s *Store) Cleanup() int {
	return s.cleanup(time.Now())
}

// Dump contents of store to a 'gob' file
func (s *Store) Dump(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open file for save: %w", err)
	}
	defer f.Close()

	s.mu.RLock()
	defer s.mu.RUnlock()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(s.data); err != nil {
		return fmt.Errorf("failed to encode to file: %w", err)
	}

	return nil
}

// Returns a slice of all public pastes, unlisted ones are ignored
func (s *Store) ListPublic() []Model {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var res []Model
	for _, m := range s.data {
		if m.Visibility == VisibilityPublic {
			res = append(res, *m)
		}
	}
	return res
}

// ---- Helpers for Get and Cleanup for easy test mocking ---- //

func (s *Store) get(id string, now time.Time) (Model, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	model, ok := s.data[id]
	if !ok {
		return Model{}, false
	} else if now.After(model.ExpireAt) {
		return Model{}, false
	}

	return *model, true
}

func (s *Store) cleanup(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var deleteCount int
	for id, m := range s.data {
		if now.After(m.ExpireAt) {
			delete(s.data, id)
			deleteCount++
		}
	}

	return deleteCount
}
