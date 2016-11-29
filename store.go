package main

type Store struct {
	data map[string][]byte
}

var sharedInstance *Store = newStore()

func newStore() *Store {
	s := new(Store)
	s.data = map[string][]byte{}
	return s
}

func GetSharedStore() *Store {
	return sharedInstance
}

func (s *Store) Set(key string, value []byte) error {
	copied := make([]byte, len(value), len(value))
	copy(copied, value)
	s.data[key] = copied
	return nil
}

func (s *Store) Get(key string) ([]byte, bool) {
	value, ok := s.data[key]
	return value, ok
}

func (s *Store) Keys() ([]string, error) {
	list := make([]string, len(s.data))
	i := 0
	for key, _ := range s.data {
		list[i] = key
		i++
	}
	return list, nil
}
