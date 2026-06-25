package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type App struct {
	store *Store
}

func NewApp(store *Store) *App {
	return &App{
		store: store,
	}
}


// JSON response helpers
func writeJSON(w http.ResponseWriter, status int, data interface{}){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string){
	writeJSON(w, status, map[string]string{"error": message})
}

func generateID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Health check handler
func (app *App) HealthCheckHandler(w http.ResponseWriter, r *http.Request){
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Auth
type registerRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	ID string `json:"id"`
	Email string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request){
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "a valid email is required")
		return
	}

	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 character")
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	user := &User{
		ID: generateID(),
		Email: req.Email,
		PasswordHash: hash,
		CreatedAt: time.Now(),
	}

	if err := app.store.createUser(user); err != nil {
		if err == ErrUserExists {
			writeError(w, http.StatusConflict, "user already exists")
		}
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		ID: user.ID,
		Email: user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

type loginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request){
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "a valid email is required")
		return
	}

	user, err := app.store.getUserByEmail(req.Email)
	if err != nil || !checkPasswordHash(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

// Tickets
type createTicketRequest struct {
	Title string `json:"title"`
	Description string `json:"description"`
}

func (app *App) CreateTicketHandler(w http.ResponseWriter, r *http.Request){
	userID := userIDFromContext(r)

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	now := time.Now().UTC()
	ticket := &Ticket{
		ID: generateID(),
		UserID: userID,
		Title: req.Title,
		Description: req.Description,
		Status: StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	app.store.createTicket(ticket)

	writeJSON(w, http.StatusCreated, ticket)
}

func (app *App) ListTicketsHandler(w http.ResponseWriter, r *http.Request){
	userID := userIDFromContext(r)
	tickets := app.store.listTicketsByUserID(userID)
	writeJSON(w, http.StatusOK, tickets)
}

func (app *App) GetTicketHandler(w http.ResponseWriter, r *http.Request){
	userID := userIDFromContext(r)
	id := r.PathValue("id")

	ticket, err := app.store.getTicket(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}