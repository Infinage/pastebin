package main

import (
	"embed"
	"fmt"

	"github.com/infinage/pastebin/internal/handlers"
)

//go:embed assets/*
var assets embed.FS

func main() {
	app := handlers.NewApplication(assets)
	port := "8080"
	fmt.Printf("Starting server on port '%v'\n", port)
	app.Serve(":" + port)
}
