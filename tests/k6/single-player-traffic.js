/**
 * Single Player Traffic Generator - Sequential gameplay
 *
 * Purpose: Generate realistic single-player traffic without concurrency issues
 * VUs: 1 (sequential to avoid shared state conflicts)
 * Duration: 30 minutes
 */

import http from 'k6/http';
import { sleep, check } from 'k6';

const BASE_URL = 'http://localhost:8080';
const JSON_HEADERS = { 'Content-Type': 'application/json' };

export const options = {
  vus: 1, // Single VU to avoid concurrency with shared game state
  duration: '30m',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
  },
};

// Helper to get all scoring dice from a roll
function getScoringDice(dice) {
  // Keep all 1s and 5s for simplicity (could be more sophisticated)
  return dice.filter(d => d === 1 || d === 5);
}

export default function () {
  // Reset and start new game
  let res = http.get(`${BASE_URL}/api/reset`);
  check(res, { 'reset ok': (r) => r.status === 200 });

  const playerName = `K6Player_${Date.now()}`;
  res = http.post(
    `${BASE_URL}/api/set-player-name`,
    JSON.stringify({ player_name: playerName }),
    { headers: JSON_HEADERS }
  );
  check(res, { 'set name ok': (r) => r.status === 200 });

  sleep(0.5);

  // Play until we win or farkle too many times
  let totalTurns = 0;
  const maxTurns = 50; // Prevent infinite loops

  while (totalTurns < maxTurns) {
    totalTurns++;

    // Start a new turn
    res = http.post(
      `${BASE_URL}/api/roll`,
      JSON.stringify({ dice_to_keep: [] }),
      { headers: JSON_HEADERS }
    );

    if (res.status !== 200) {
      console.log(`Roll failed: ${res.status}`);
      break;
    }

    let state;
    try {
      state = JSON.parse(res.body);
    } catch (e) {
      console.log(`JSON parse error: ${e}`);
      break;
    }

    // Check if we already won
    if (state.winner) {
      console.log(`🎉 WINNER! Total: ${state.total_bank} points in ${totalTurns} turns`);
      break;
    }

    // If we farkled on first roll, move to next turn
    if (state.turn && state.turn.is_game_over) {
      sleep(0.3);
      continue;
    }

    // Play this turn aggressively
    let rollsThisTurn = 0;
    const maxRollsPerTurn = 10;

    while (rollsThisTurn < maxRollsPerTurn) {
      rollsThisTurn++;

      const dice = state.last_roll || [];
      const scoringDice = getScoringDice(dice);

      // If no scoring dice, we farkled
      if (scoringDice.length === 0) {
        break;
      }

      // Keep all scoring dice
      res = http.post(
        `${BASE_URL}/api/roll`,
        JSON.stringify({ dice_to_keep: scoringDice }),
        { headers: JSON_HEADERS }
      );

      if (res.status !== 200) {
        break;
      }

      try {
        state = JSON.parse(res.body);
      } catch (e) {
        break;
      }

      // If we farkled, end this turn
      if (state.turn && state.turn.is_game_over) {
        break;
      }

      // Bank if we have enough points
      const accumulatedScore = state.turn ? state.turn.accumulated_score : 0;
      const totalBank = state.total_bank || 0;
      const minBankScore = totalBank === 0 ? 500 : 300; // House rule: first bank needs 500

      if (accumulatedScore >= minBankScore) {
        res = http.get(`${BASE_URL}/api/bank`);
        check(res, { 'bank ok': (r) => r.status === 200 });

        try {
          state = JSON.parse(res.body);
          if (state.winner) {
            console.log(`🎉 WINNER! Total: ${state.total_bank} points in ${totalTurns} turns`);
            totalTurns = maxTurns; // Exit outer loop
            break;
          }
        } catch (e) {
          // Ignore
        }

        break; // End this turn after banking
      }

      sleep(0.1);
    }

    sleep(0.5);
  }

  // Cooldown between games
  sleep(2);
}
