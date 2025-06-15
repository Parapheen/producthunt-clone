package session

import "fmt"

type SessionStorage interface {
	Create(session *Session) error
}

type Service struct {
	storage SessionStorage
}

func NewService(storage SessionStorage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Create(session *Session) error {
	err := s.storage.Create(session)
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	return nil
}
