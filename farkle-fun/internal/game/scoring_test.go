package game

import "testing"

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		dice     []int
		expScore int
		expUsed  int
	}{
		{"Single 1 and 5", []int{1, 5, 2, 2, 3, 3}, 150, 2},
		{"Three of a Kind (2s)", []int{2, 2, 2, 3, 4, 6}, 200, 3},
		{"Three of a Kind (1s)", []int{1, 1, 1, 3, 4, 6}, 1000, 3},
		{"Straight (1-6)", []int{1, 2, 3, 4, 5, 6}, 1500, 6},
		{"Three Pairs", []int{2, 2, 4, 4, 6, 6}, 1500, 6},
		{"Four of a Kind (as Two Pairs)", []int{3, 3, 3, 3, 5, 5}, 1500, 6},
		{"Farkle Roll", []int{2, 3, 4, 6, 2, 3}, 0, 0},
		{"Mixed Scoring", []int{1, 1, 1, 5, 4, 3}, 1050, 4}, // Three 1s + one 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CalculateScore(tt.dice)
			if res.Score != tt.expScore || res.DiceUsed != tt.expUsed {
				t.Errorf("%s: got (%d, %d), want (%d, %d)", 
					tt.name, res.Score, res.DiceUsed, tt.expScore, tt.expUsed)
			}
		})
	}
}