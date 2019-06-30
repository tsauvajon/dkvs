package dkvs

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
)

// Storage is a generic storage that can save and retrieve values
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key, val string) error
	ReplicateTo() (*bytes.Buffer, error)
	ReplicateFrom(data io.Reader) error
}

// maps are not safe for concurrent use:
// https://blog.golang.org/go-maps-in-action#TOC_6.
type store struct {
	data map[string]string
	lock sync.RWMutex
}

// NewStore creates an in memory data store
func NewStore() Storage {
	return &store{
		data: make(map[string]string),
	}
}

func (s *store) Get(key string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return nil, errorKeyNotFound
	}
	return []byte(val), nil
}

func (s *store) Set(key, val string) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.data[key] = val
	return nil
}

func (s *store) ReplicateTo() (*bytes.Buffer, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	return buf, encoder.Encode(s.data)
}

func (s *store) ReplicateFrom(data io.Reader) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	decoder := json.NewDecoder(data)
	return decoder.Decode(&s.data)
}
