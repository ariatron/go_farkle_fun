package game

type ScoreResult struct {
	Score    int
	DiceUsed int
}

func CalculateScore(dice []int) ScoreResult {
	if len(dice) == 0 {
		return ScoreResult{0, 0}
	}

	counts := make(map[int]int)
	for _, d := range dice {
		counts[d]++
	}

	// ONLY check for 6-dice patterns if the user actually handed us 6 dice
	if len(dice) == 6 {
		// 1. Check for Straight (1-2-3-4-5-6)
		if len(counts) == 6 {
			return ScoreResult{Score: 1500, DiceUsed: 6}
		}

		// 2. Check for Three Pairs
		pairCount := 0
		for _, count := range counts {
			if count == 2 {
				pairCount++
			} else if count == 4 { 
				pairCount += 2
			} else if count == 6 {
				pairCount += 3
			}
		}
		if pairCount == 3 {
			return ScoreResult{Score: 1500, DiceUsed: 6}
		}
	}

	// 3. Standard Scoring (Three of a kind, 1s, 5s)
	totalScore := 0
	diceUsed := 0

	for die := 1; die <= 6; die++ {
		count := counts[die]
		
		// Three of a Kind (or more)
		if count >= 3 {
			if die == 1 {
				totalScore += 1000
			} else {
				totalScore += die * 100
			}
			diceUsed += 3
			count -= 3
			
			// Note: This logic handles "Three of a kind" exactly. 
			// If you want 4-of-a-kind to be worth more, you'd add that here.
		}

		// Remaining 1s and 5s
		if die == 1 {
			totalScore += count * 100
			diceUsed += count
		} else if die == 5 {
			totalScore += count * 50
			diceUsed += count
		}
	}

	return ScoreResult{Score: totalScore, DiceUsed: diceUsed}
}