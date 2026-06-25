package main

import (
	"log"
	"net/http"
)

func main(){
	store := NewStore()
	app := NewApp(store)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", app.HealthCheckHandler)
	mux.HandleFunc("POST /auth/register", app.RegisterHandler)
	mux.HandleFunc("POST /auth/login", app.LoginHandler)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}