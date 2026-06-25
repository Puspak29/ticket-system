package main

import "time"

const (
	StatusOpen = "open"
	StatusInProgress = "in_progress"
	StatusClosed = "closed"
)

type User struct {
	ID string `json:"id"`
	Email string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Ticket struct {
	ID string `json:"id"`
	UserID string `json:"user_id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}