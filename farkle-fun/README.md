🎲 Go-Farkle
A lightweight, web-based implementation of the classic dice game Farkle, built with a Go backend and a vanilla JavaScript/CSS frontend.

🚀 Features
Complete Scoring Engine: Supports standard 1s and 5s, Three-of-a-Kind (House Rule), Straights (1-6), and Three Pairs.

"Hot Dice" Logic: Successfully score with all 6 dice to get a fresh set and keep your turn alive!

Interactive UI: Click to select which dice you want to keep before rolling again.

Win Condition: Race to be the first to bank 10,000 points.

RESTful API: Clean separation between game logic and the web interface.

📂 Project Structure
Plaintext
farkle-fun/
├── cmd/
│   └── server/
│       └── main.go          # Entry point & Route definitions
├── internal/
│   ├── game/
│   │   ├── engine.go        # Turn state & dice rolling
│   │   ├── scoring.go       # Scoring rules & math
│   │   └── scoring_test.go  # Unit tests for scoring patterns
│   └── api/
│       └── handlers.go      # HTTP request handling
├── static/
│   └── index.html           # Frontend UI & Game Logic
├── go.mod                   # Go module definition
└── README.md                # Project documentation
🛠️ Getting Started
Prerequisites
Go (version 1.18 or higher recommended)

Installation & Setup
Initialize the module (if you haven't already):

Bash
go mod init farkle-app
go mod tidy
Run the server:

Bash
go run cmd/server/main.go
Play the game: Open your browser and navigate to: http://localhost:8080

🧪 Testing
The scoring engine includes comprehensive unit tests to ensure "Three of a Kind" and "Straights" are calculated correctly.

Bash
go test ./internal/game -v
📜 Game Rules (Implemented)
Rolling: Roll 6 dice to start.

Scoring: You must keep at least one scoring die (1, 5, or a pattern) to roll again.

Farkle: If a roll contains no scoring dice, you lose all points accumulated in that turn.

Banking: You can stop at any time to add your Turn Score to your Total Bank.

Winning: The first player to reach 10,000 points wins the game.