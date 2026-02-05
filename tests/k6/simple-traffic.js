/**
 * Simple Traffic Generator - Robust continuous load
 *
 * Purpose: Generate steady traffic to single-player endpoints
 * VUs: 3
 * Duration: 30 minutes
 */

import http from 'k6/http';
import { sleep, check } from 'k6';

const BASE_URL = 'http://localhost:8080';
const JSON_HEADERS = { 'Content-Type': 'application/json' };

export const options = {
  vus: 1, // Single VU to avoid concurrency issues with shared game state
  duration: '30m',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
  },
};

export default function () {
  // Reset game
  let res = http.get(`${BASE_URL}/api/reset`);
  check(res, { 'reset ok': (r) => r.status === 200 });

  // Set player name
  const playerName = `Player_${__VU}`;
  res = http.post(
    `${BASE_URL}/api/set-player-name`,
    JSON.stringify({ player_name: playerName }),
    { headers: JSON_HEADERS }
  );
  check(res, { 'set name ok': (r) => r.status === 200 });

  sleep(1);

  // Play more turns to increase chance of reaching 10k
  for (let turn = 0; turn < 15; turn++) {
    // Roll dice
    res = http.post(
      `${BASE_URL}/api/roll`,
      JSON.stringify({ dice_to_keep: [] }),
      { headers: JSON_HEADERS }
    );

    if (res.status !== 200) {
      continue; // Skip if error
    }

    let state;
    try {
      state = JSON.parse(res.body);
    } catch (e) {
      console.log(`JSON parse error: ${e}, body: ${res.body.substring(0, 100)}`);
      continue;
    }

    // If game over (farkled), move to next turn
    if (state.turn && state.turn.is_game_over) {
      sleep(0.5);
      continue;
    }

    // Keep some random dice (1s or 5s)
    const dice = state.last_roll || [];
    const diceToKeep = dice.filter(d => d === 1 || d === 5).slice(0, 2);

    if (diceToKeep.length > 0) {
      res = http.post(
        `${BASE_URL}/api/roll`,
        JSON.stringify({ dice_to_keep: diceToKeep }),
        { headers: JSON_HEADERS }
      );

      if (res.status === 200) {
        try {
          state = JSON.parse(res.body);

          // Try to bank if we have enough points (more aggressive banking)
          if (state.turn && state.turn.accumulated_score >= 300) {
            res = http.get(`${BASE_URL}/api/bank`);
            check(res, { 'bank ok': (r) => r.status === 200 });

            // Check if we won
            try {
              const bankState = JSON.parse(res.body);
              if (bankState.winner) {
                console.log(`🎉 WINNER! Total bank: ${bankState.total_bank}`);
              }
            } catch (e) {
              // Ignore
            }
          }
        } catch (e) {
          // Ignore JSON errors
        }
      }
    }

    sleep(0.5);
  }

  sleep(2);
}
