package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Store struct {
	name  string
	file  string
	mu    sync.RWMutex
	items []map[string]interface{}
	nextID int
}

func New(name string) *Store {
	s := &Store{
		name:   name,
		file:   name + ".json",
		nextID: 1,
	}
	s.loadFromDisk()
	return s
}

func (s *Store) Save(data interface{}) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	var item map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		item = v
	default:
		item = map[string]interface{}{"value": data}
	}

	item["id"] = fmt.Sprintf("%d", s.nextID)
	s.nextID++
	s.items = append(s.items, item)
	s.writeToDisk()
	return item
}

func (s *Store) Load() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]interface{}, len(s.items))
	for i, item := range s.items {
		result[i] = item
	}
	return result
}

func (s *Store) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, item := range s.items {
		if fmt.Sprintf("%v", item["id"]) == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			s.writeToDisk()
			return true
		}
	}
	return false
}

func (s *Store) loadFromDisk() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return
	}
	var saved struct {
		NextID int                      `json:"next_id"`
		Items  []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}
	s.items = saved.Items
	s.nextID = saved.NextID
}

func (s *Store) writeToDisk() {
	saved := struct {
		NextID int                      `json:"next_id"`
		Items  []map[string]interface{} `json:"items"`
	}{
		NextID: s.nextID,
		Items:  s.items,
	}
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(s.file, data, 0644)
}
