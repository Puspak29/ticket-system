package main

import (
	"errors"
	"sync"
)

var (
	ErrUserExists = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidUserData = errors.New("invalid user data")
	ErrTicketNotFound = errors.New("ticket not found")
)

type Store struct {
	mu sync.RWMutex
	usersByEmail map[string]*User
	usersByID map[string]*User
	tickets map[string]*Ticket
}

func NewStore() *Store {
	return &Store{
		usersByEmail: make(map[string]*User),
		usersByID: make(map[string]*User),
		tickets: make(map[string]*Ticket),
	}
}

func (s *Store) createUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.usersByEmail[user.Email]; exists {
		return ErrUserExists
	}
	s.usersByEmail[user.Email] = user
	s.usersByID[user.ID] = user
	return nil
}

func (s *Store) getUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}