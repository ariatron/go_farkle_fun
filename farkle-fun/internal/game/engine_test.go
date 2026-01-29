package game

import "testing"

func TestHotDiceLogic(t *testing.T) {
	turn := NewTurn() // Starts with 6 dice

	// Scenario: Player rolls 6 dice and keeps a Straight (all 6 dice)
	keptDice := []int{1, 2, 3, 4, 5, 6}
	score, farkled := turn.ProcessRoll(keptDice)

	if farkled {
		t.Fatal("Should not have farkled with a straight")
	}
	if score != 1500 {
		t.Errorf("Expected 1500 points, got %d", score)
	}
	if turn.DiceRemaining != 6 {
		t.Errorf("HOT DICE FAIL: Expected 6 dice remaining, got %d", turn.DiceRemaining)
	}

	// Scenario: Player rolls 6 dice, keeps three 2s (3 dice used)
	turn2 := NewTurn()
	turn2.ProcessRoll([]int{2, 2, 2})
	if turn2.DiceRemaining != 3 {
		t.Errorf("Expected 3 dice remaining, got %d", turn2.DiceRemaining)
	}
}

func TestRoll(t *testing.T) {
	turn := NewTurn()

	// Test initial roll with 6 dice
	dice := turn.Roll()
	if len(dice) != 6 {
		t.Errorf("Expected 6 dice, got %d", len(dice))
	}

	for _, die := range dice {
		if die < 1 || die > 6 {
			t.Errorf("Invalid die value: %d", die)
		}
	}

	// Test roll with reduced dice remaining
	turn.DiceRemaining = 3
	dice = turn.Roll()
	if len(dice) != 3 {
		t.Errorf("Expected 3 dice, got %d", len(dice))
	}

	// Test roll with 1 die
	turn.DiceRemaining = 1
	dice = turn.Roll()
	if len(dice) != 1 {
		t.Errorf("Expected 1 die, got %d", len(dice))
	}
	if dice[0] < 1 || dice[0] > 6 {
		t.Errorf("Invalid single die value: %d", dice[0])
	}
}

func TestProcessRollSuccessfulScore(t *testing.T) {
	turn := NewTurn()

	// Test successful roll with three 1s (1000 points)
	score, farkled := turn.ProcessRoll([]int{1, 1, 1})
	if farkled {
		t.Error("Expected successful roll, but got Farkle")
	}
	if score != 1000 {
		t.Errorf("Expected 1000 points, got %d", score)
	}
	if turn.AccumulatedScore != 1000 {
		t.Errorf("Expected accumulated score 1000, got %d", turn.AccumulatedScore)
	}
	if turn.DiceRemaining != 3 {
		t.Errorf("Expected 3 dice remaining, got %d", turn.DiceRemaining)
	}
}

func TestProcessRollFarkle(t *testing.T) {
	turn := NewTurn()

	// Test Farkle (no scoring dice)
	score, farkled := turn.ProcessRoll([]int{2, 3, 4})
	if !farkled {
		t.Error("Expected Farkle, but got successful roll")
	}
	if score != 0 {
		t.Errorf("Expected 0 points on Farkle, got %d", score)
	}
	if turn.AccumulatedScore != 0 {
		t.Errorf("Expected accumulated score 0 after Farkle, got %d", turn.AccumulatedScore)
	}
	if !turn.IsGameOver {
		t.Error("Expected IsGameOver to be true")
	}
}

func TestProcessRollHotDice(t *testing.T) {
	turn := NewTurn()

	// Test hot dice: all 6 dice score (e.g., straight)
	score, farkled := turn.ProcessRoll([]int{1, 2, 3, 4, 5, 6})
	if farkled {
		t.Error("Expected successful roll with hot dice, but got Farkle")
	}
	if score != 1500 {
		t.Errorf("Expected 1500 points for straight, got %d", score)
	}
	if turn.DiceRemaining != 6 {
		t.Errorf("Expected 6 dice remaining after hot dice, got %d", turn.DiceRemaining)
	}
	if turn.AccumulatedScore != 1500 {
		t.Errorf("Expected accumulated score 1500, got %d", turn.AccumulatedScore)
	}
}

func TestProcessRollPartialScore(t *testing.T) {
	turn := NewTurn()

	// Test keeping only some dice (3 out of 6)
	score, farkled := turn.ProcessRoll([]int{5, 5, 5})
	if farkled {
		t.Error("Expected successful roll, but got Farkle")
	}
	if score != 500 {
		t.Errorf("Expected 500 points for three 5s, got %d", score)
	}
	if turn.DiceRemaining != 3 {
		t.Errorf("Expected 3 dice remaining, got %d", turn.DiceRemaining)
	}
}

func TestProcessRollAccumulatedScore(t *testing.T) {
	turn := NewTurn()

	// First roll: three 5s (500 points)
	_, farkled1 := turn.ProcessRoll([]int{5, 5, 5})
	if farkled1 {
		t.Error("First roll should not farkle")
	}
	if turn.AccumulatedScore != 500 {
		t.Errorf("After first roll, expected 500, got %d", turn.AccumulatedScore)
	}

	// Second roll: one 1 (100 points)
	_, farkled2 := turn.ProcessRoll([]int{1})
	if farkled2 {
		t.Error("Second roll should not farkle")
	}
	if turn.AccumulatedScore != 600 {
		t.Errorf("After second roll, expected 600, got %d", turn.AccumulatedScore)
	}
}

func TestHotDiceReset(t *testing.T) {
	turn := NewTurn()

	// Roll 1: Three pairs (all 6 dice used) = hot dice
	score, _ := turn.ProcessRoll([]int{2, 2, 4, 4, 6, 6})
	if turn.DiceRemaining != 6 {
		t.Errorf("After hot dice, expected 6 remaining, got %d", turn.DiceRemaining)
	}

	// Roll 2: Use 3 of the 6
	score2, _ := turn.ProcessRoll([]int{1, 1, 1})
	if turn.DiceRemaining != 3 {
		t.Errorf("After using 3 of 6, expected 3 remaining, got %d", turn.DiceRemaining)
	}

	// Roll 3: Use all 3 remaining = hot dice again
	score3, _ := turn.ProcessRoll([]int{5, 5, 5})
	if turn.DiceRemaining != 6 {
		t.Errorf("After second hot dice, expected 6 remaining, got %d", turn.DiceRemaining)
	}
	if turn.AccumulatedScore != score+score2+score3 {
		t.Errorf("Expected total of %d, got %d", score+score2+score3, turn.AccumulatedScore)
	}
}

func TestNewTurnInitialization(t *testing.T) {
	turn := NewTurn()

	if turn.AccumulatedScore != 0 {
		t.Errorf("New turn should have 0 accumulated score, got %d", turn.AccumulatedScore)
	}
	if turn.DiceRemaining != 6 {
		t.Errorf("New turn should have 6 dice remaining, got %d", turn.DiceRemaining)
	}
	if turn.IsGameOver {
		t.Error("New turn should not be game over")
	}
}
