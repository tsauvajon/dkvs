package dkvs

import "sync"

type Storage interface {
	Get(key string) ([]byte, error)
	Set(key, val string) error
}

// maps are not safe for concurrent use:
// https://blog.golang.org/go-maps-in-action#TOC_6.
type store struct {
	data map[string]string
	lock sync.RWMutex
}

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
		return nil, ERROR_KEY_NOT_FOUND
	}
	return []byte(val), nil
}

func (s *store) Set(key, val string) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.data[key] = val
	return nil
}
