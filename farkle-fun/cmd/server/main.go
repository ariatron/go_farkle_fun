package main

import (
	"farkle-app/internal/api"
	"fmt"
	"net/http"
)

func main() {
	// API Routes
	http.HandleFunc("/api/roll", api.RollHandler)
	http.HandleFunc("/api/bank", api.BankHandler)
	http.HandleFunc("/api/reset", api.ResetHandler)
	http.HandleFunc("/api/set-player-name", api.SetPlayerNameHandler)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("🎲 Farkle Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
