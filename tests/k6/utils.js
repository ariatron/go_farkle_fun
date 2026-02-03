// K6 Shared Utilities for Farkle Game Testing

import { check, sleep } from 'k6';
import http from 'k6/http';

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Headers for JSON requests
export const JSON_HEADERS = {
  'Content-Type': 'application/json',
};

// Generate random integer between min and max (inclusive)
export function randomIntBetween(min, max) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}

// Common checks for successful API responses
export function checkSuccess(response, endpoint) {
  return check(response, {
    [`${endpoint}: status is 200`]: (r) => r.status === 200,
    [`${endpoint}: has response body`]: (r) => r.body.length > 0,
    [`${endpoint}: response is JSON`]: (r) => {
      try {
        JSON.parse(r.body);
        return true;
      } catch (e) {
        return false;
      }
    },
  });
}

// Reset the game state
export function resetGame() {
  const response = http.get(`${BASE_URL}/api/reset`);
  return JSON.parse(response.body);
}

// Set player name
export function setPlayerName(name) {
  const payload = JSON.stringify({ player_name: name });
  const response = http.post(
    `${BASE_URL}/api/set-player-name`,
    payload,
    { headers: JSON_HEADERS }
  );
  return JSON.parse(response.body);
}

// Roll dice
export function rollDice(diceToKeep = []) {
  const payload = JSON.stringify({ dice_to_keep: diceToKeep });
  const response = http.post(
    `${BASE_URL}/api/roll`,
    payload,
    { headers: JSON_HEADERS }
  );
  return JSON.parse(response.body);
}

// Bank points
export function bankPoints() {
  const response = http.get(`${BASE_URL}/api/bank`);
  return JSON.parse(response.body);
}

// Simulate a realistic turn with think time
export function playRealisticTurn() {
  // Roll dice
  let state = rollDice();
  sleep(0.5); // Think time

  // If not farkled, do 1-3 more rolls
  const numRolls = Math.floor(Math.random() * 3) + 1;

  for (let i = 0; i < numRolls && !state.turn.is_game_over; i++) {
    // Randomly decide what dice to keep
    const diceToKeep = selectRandomScoringDice(state.last_roll);

    if (diceToKeep.length > 0) {
      state = rollDice(diceToKeep);
      sleep(0.3 + Math.random() * 0.7); // Think time between 0.3-1s
    } else {
      break;
    }
  }

  // Bank if we have points and didn't farkle
  if (state.turn.accumulated_score > 0 && !state.turn.is_game_over) {
    state = bankPoints();
    sleep(0.2); // Brief pause after banking
  }

  return state;
}

// Simple logic to select some scoring dice (1s and 5s)
export function selectRandomScoringDice(dice) {
  const scoringDice = [];

  for (const die of dice) {
    if (die === 1 || die === 5) {
      // 70% chance to keep scoring dice
      if (Math.random() < 0.7) {
        scoringDice.push(die);
      }
    }
  }

  return scoringDice;
}

// Play until game is won or max turns reached
export function playFullGame(maxTurns = 100) {
  resetGame();
  let turns = 0;
  let state;

  while (turns < maxTurns) {
    state = playRealisticTurn();
    turns++;

    if (state.winner) {
      break;
    }
  }

  return { turns, winner: state?.winner || false, finalScore: state?.total_bank || 0 };
}

// Thresholds helper
export function getCommonThresholds() {
  return {
    http_req_duration: ['p(95)<200', 'p(99)<500'], // 95% under 200ms, 99% under 500ms
    http_req_failed: ['rate<0.01'], // Less than 1% failure rate
  };
}
