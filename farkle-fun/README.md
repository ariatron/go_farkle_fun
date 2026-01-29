# Go-Farkle

A lightweight, web-based implementation of the classic dice game **Farkle**, built with a **Go backend** and a **vanilla JavaScript/CSS frontend**.


## Features

- **Complete Scoring Engine**
  - Standard **1s** and **5s**
  - **Three-of-a-Kind** (House Rule)
  - **Straights (1–6)**
  - **Three Pairs**
- **Hot Dice Logic**
  - Score with all 6 dice to get a fresh set and keep your turn alive
- **Interactive UI**
  - Click dice to select which ones to keep before rolling again
- **Win Condition**
  - First player to bank **10,000 points** wins
- **RESTful API**
  - Clean separation between game logic and the web interface

### Prerequisites
- **Go 1.18+** (recommended)

### Installation & Setup

Initialize the module (if not already initialized):

```bash
go mod init farkle-app
go mod tidy
```

Run the server:
```bash
go run cmd/server/main.go
```

Run tests:
```bash
go test ./internal/game -v
```

Example Output:
```bash
go test ./internal/game -v
=== RUN   TestHotDiceLogic
--- PASS: TestHotDiceLogic (0.00s)
=== RUN   TestCalculateScore
=== RUN   TestCalculateScore/Single_1_and_5
=== RUN   TestCalculateScore/Three_of_a_Kind_(2s)
=== RUN   TestCalculateScore/Three_of_a_Kind_(1s)
=== RUN   TestCalculateScore/Straight_(1-6)
=== RUN   TestCalculateScore/Three_Pairs
=== RUN   TestCalculateScore/Four_of_a_Kind_(as_Two_Pairs)
=== RUN   TestCalculateScore/Farkle_Roll
=== RUN   TestCalculateScore/Mixed_Scoring
--- PASS: TestCalculateScore (0.00s)
    --- PASS: TestCalculateScore/Single_1_and_5 (0.00s)
    --- PASS: TestCalculateScore/Three_of_a_Kind_(2s) (0.00s)
    --- PASS: TestCalculateScore/Three_of_a_Kind_(1s) (0.00s)
    --- PASS: TestCalculateScore/Straight_(1-6) (0.00s)
    --- PASS: TestCalculateScore/Three_Pairs (0.00s)
    --- PASS: TestCalculateScore/Four_of_a_Kind_(as_Two_Pairs) (0.00s)
    --- PASS: TestCalculateScore/Farkle_Roll (0.00s)
    --- PASS: TestCalculateScore/Mixed_Scoring (0.00s)
PASS
ok      farkle-app/internal/game        (cached)
```