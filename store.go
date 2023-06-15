package wserve

type IStore interface {
	Push(msg string) error
	Pull() (string, error)
}

type Store struct {
}

func (s *Store) Push(msg string) error {
	return nil
}

func (s *Store) Pull() (string, error) {
	return "", nil
}
