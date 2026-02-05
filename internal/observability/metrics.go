package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP Metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Game Metrics
	GameRollsTotal    prometheus.Counter
	GameBanksTotal    prometheus.Counter
	GameFarklesTotal  prometheus.Counter
	GameWinsTotal     prometheus.Counter
	ActiveGames       prometheus.Gauge
	PointsDistribution *prometheus.HistogramVec

	// Multiplayer Metrics (only used in multi mode)
	ActiveRooms       prometheus.Gauge
	PlayersOnline     prometheus.Gauge
	RoomsCreatedTotal prometheus.Counter
	RoomsFull        prometheus.Counter
}

// AppMetrics is the global metrics instance
var AppMetrics *Metrics

// InitMetrics initializes all Prometheus metrics
func InitMetrics() {
	AppMetrics = &Metrics{
		// HTTP Metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "farkle_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "farkle_http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "farkle_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 200, 500, 1000, 2000, 5000, 10000},
			},
			[]string{"method", "endpoint"},
		),

		// Game Metrics
		GameRollsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "farkle_game_rolls_total",
				Help: "Total number of dice rolls",
			},
		),
		GameBanksTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "farkle_game_banks_total",
				Help: "Total number of points banked",
			},
		),
		GameFarklesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "farkle_game_farkles_total",
				Help: "Total number of farkles",
			},
		),
		GameWinsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "farkle_game_wins_total",
				Help: "Total number of games won",
			},
		),
		ActiveGames: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "farkle_active_games",
				Help: "Number of active games",
			},
		),
		PointsDistribution: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "farkle_points_distribution",
				Help:    "Distribution of points scored",
				Buckets: []float64{0, 50, 100, 150, 200, 300, 500, 1000, 1500, 2000, 3000},
			},
			[]string{"type"}, // type: "roll" or "bank"
		),

		// Multiplayer Metrics
		ActiveRooms: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "farkle_active_rooms",
				Help: "Number of active game rooms (multiplayer mode)",
			},
		),
		PlayersOnline: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "farkle_players_online",
				Help: "Total number of players in all rooms (multiplayer mode)",
			},
		),
		RoomsCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "farkle_rooms_created_total",
				Help: "Total number of rooms created (multiplayer mode)",
			},
		),
		RoomsFull: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "farkle_rooms_full_total",
				Help: "Total number of times a room became full (multiplayer mode)",
			},
		),
	}

	// Initialize gauges to 0
	AppMetrics.ActiveGames.Set(0)
	AppMetrics.ActiveRooms.Set(0)
	AppMetrics.PlayersOnline.Set(0)

	Logger.Info("Metrics initialized", "metrics_registered", "16")
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration float64, responseSize int) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, string(rune(statusCode))).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	m.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
}

// RecordRoll records a dice roll event
func (m *Metrics) RecordRoll(points int) {
	m.GameRollsTotal.Inc()
	m.PointsDistribution.WithLabelValues("roll").Observe(float64(points))
}

// RecordBank records a banking event
func (m *Metrics) RecordBank(points int) {
	m.GameBanksTotal.Inc()
	m.PointsDistribution.WithLabelValues("bank").Observe(float64(points))
}

// RecordFarkle records a farkle event
func (m *Metrics) RecordFarkle() {
	m.GameFarklesTotal.Inc()
}

// RecordWin records a game win event
func (m *Metrics) RecordWin() {
	m.GameWinsTotal.Inc()
}

// SetActiveGames sets the number of active games
func (m *Metrics) SetActiveGames(count int) {
	m.ActiveGames.Set(float64(count))
}
