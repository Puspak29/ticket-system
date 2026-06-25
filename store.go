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

func (s *Store) createTicket(ticket *Ticket){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[ticket.ID] = ticket
}

func (s *Store) getTicket(id string) (*Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[id]
	if !ok {
		return nil, ErrTicketNotFound
	}
	return t, nil
}

func (s *Store) listTicketsByUserID(userID string) []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Ticket, 0)
	for _, t := range s.tickets {
		if t.UserID == userID {
			result = append(result, t)
		}
	}

	// simple insertion sort by CreatedAt desc
	for i := 1; i < len(result); i++ {
		j := i
		for j > 0 && result[j-1].CreatedAt.Before(result[j].CreatedAt) {
			result[j-1], result[j] = result[j], result[j-1]
			j--
		}
	}

	return result
}