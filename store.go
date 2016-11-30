package main

import (
	"time"
)

type Store struct {
	data    map[string][]byte
	expires map[string]int64
}

var sharedInstance *Store = newStore()

func newStore() *Store {
	s := new(Store)
	s.data = map[string][]byte{}
	s.expires = map[string]int64{}
	return s
}

func GetSharedStore() *Store {
	return sharedInstance
}

func (s *Store) Set(key string, value []byte) error {
	copied := make([]byte, len(value), len(value))
	copy(copied, value)
	s.data[key] = copied
	s.expires[key] = time.Now().Unix() + 300
	return nil
}

func (s *Store) Get(key string) ([]byte, bool) {
	timelimit, ok := s.expires[key]
	if ok {
		now := time.Now().Unix()
		if now > timelimit {
			delete(s.data, key)
			delete(s.expires, key)
		} else {
			value, ok := s.data[key]
			if ok {
				s.expires[key] = now + 300
				return value, ok
			}
		}
	}
	return nil, false
}

func (s *Store) Keys() ([]string, error) {
	list := make([]string, 0, len(s.data))
	i := 0
	for key, _ := range s.data {
		_, ok := s.Get(key)
		if ok {
			list[i] = key
			i++
		}
	}
	return list, nil
}
