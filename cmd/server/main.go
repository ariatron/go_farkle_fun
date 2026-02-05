package main

import (
	"context"
	"farkle-app/internal/api"
	"farkle-app/internal/game"
	"farkle-app/internal/observability"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize observability stack
	initObservability()

	// Get game mode from environment variable
	gameMode := os.Getenv("GAME_MODE")
	if gameMode == "" {
		gameMode = "single" // Default to single-player
	}

	// Create router with middleware
	mux := http.NewServeMux()

	if gameMode == "multi" {
		// Multiplayer mode: initialize room manager and multiplayer endpoints
		observability.Logger.Info("🎮 Starting in MULTIPLAYER mode")
		fmt.Println("🎮 Game Mode: MULTIPLAYER")

		roomManager := game.NewRoomManager()

		// Start background cleanup for stale rooms
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				cleaned := roomManager.CleanupStaleRooms(30 * time.Minute)
				if cleaned > 0 {
					observability.Logger.Info("Cleaned up stale rooms", "count", cleaned)
				}
			}
		}()

		// Multiplayer API Routes
		mux.HandleFunc("/api/rooms/create", api.CreateRoomHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/join", api.JoinRoomHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/start", api.StartGameHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/state", api.GetRoomStateHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/roll", api.MultiplayerRollHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/bank", api.MultiplayerBankHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/end-turn", api.EndTurnHandler(roomManager))
		mux.HandleFunc("/api/rooms/{roomId}/leave", api.LeaveRoomHandler(roomManager))
		mux.HandleFunc("/api/rooms", api.ListRoomsHandler(roomManager))

	} else {
		// Single-player mode: original endpoints
		observability.Logger.Info("🎮 Starting in SINGLE-PLAYER mode")
		fmt.Println("🎮 Game Mode: SINGLE-PLAYER")

		// Single-player API Routes
		mux.HandleFunc("/api/roll", api.RollHandler)
		mux.HandleFunc("/api/bank", api.BankHandler)
		mux.HandleFunc("/api/reset", api.ResetHandler)
		mux.HandleFunc("/api/set-player-name", api.SetPlayerNameHandler)
	}

	// Common endpoints (available in both modes)
	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// Build middleware chain
	handler := observability.RecoveryMiddleware(
		observability.HealthCheckMiddleware(
			observability.CORSMiddleware(
				observability.ObservabilityMiddleware(mux),
			),
		),
	)

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		observability.Logger.Info("🎲 Farkle Server starting",
			"addr", "http://localhost:8080",
			"metrics", "http://localhost:8080/metrics",
			"health", "http://localhost:8080/health",
		)
		fmt.Println("🎲 Farkle Server started at http://localhost:8080")
		fmt.Println("📊 Metrics available at http://localhost:8080/metrics")
		fmt.Println("🔍 Traces will be sent to Jaeger (if running)")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	observability.Logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		observability.Logger.Error("Server forced to shutdown", "error", err)
	}

	// Shutdown tracing
	if err := observability.ShutdownTracing(ctx); err != nil {
		observability.Logger.Error("Failed to shutdown tracing", "error", err)
	}

	observability.Logger.Info("Server stopped gracefully")

	// Close log file
	observability.CloseLogger()
}

// initObservability initializes logging, metrics, and tracing
func initObservability() {
	// Initialize logger
	observability.InitLogger()

	// Initialize metrics
	observability.InitMetrics()

	// Initialize tracing (OTLP endpoint for Jaeger)
	// Default Jaeger OTLP endpoint is localhost:4318
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "localhost:4318"
	}

	if err := observability.InitTracing("farkle-game", jaegerEndpoint); err != nil {
		observability.Logger.Warn("Failed to initialize tracing (Jaeger may not be running)",
			"error", err,
			"endpoint", jaegerEndpoint,
		)
		observability.Logger.Info("App will continue without tracing. To enable, start Jaeger with: docker run -d -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:latest")
	}

	observability.Logger.Info("Observability stack initialized successfully")
}
