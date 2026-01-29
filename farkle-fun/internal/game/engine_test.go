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