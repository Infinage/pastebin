package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/infinage/pastebin/internal/handlers"
)

//go:embed assets/*
var assets embed.FS

func main() {
	mux := http.NewServeMux()
	app := handlers.NewApplication(context.Background())

	mux.HandleFunc("GET /", app.HandleHome)
	mux.HandleFunc("GET /paste", app.HandleNewForm)
	mux.HandleFunc("GET /paste/{id}", app.HandleGet)
	mux.HandleFunc("POST /paste", app.HandleInsert)
	mux.HandleFunc("PUT /paste/{id}", app.HandleUpdate)
	mux.HandleFunc("DELETE /paste/{id}", app.HandleDelete)
	mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", mux)
}
