import http from 'k6/http';
import { sleep, check } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { randomIntBetween } from './utils.js';

// Custom metrics
const gameRolls = new Counter('game_rolls');
const gameBanks = new Counter('game_banks');
const gameFarkles = new Counter('game_farkles');
const turnDuration = new Trend('turn_duration');

// Continuous background traffic configuration
export const options = {
  scenarios: {
    // Low, steady background traffic
    background: {
      executor: 'constant-vus',
      vus: 3,
      duration: '24h', // Run for 24 hours
      gracefulStop: '30s',
    },
    // Periodic burst traffic (simulates usage patterns)
    periodic_bursts: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5m', target: 0 },   // Quiet
        { duration: '2m', target: 10 },  // Burst up
        { duration: '3m', target: 10 },  // Stay elevated
        { duration: '2m', target: 0 },   // Drop back down
      ],
      gracefulStop: '30s',
      startTime: '5m', // Start 5 minutes in
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Less than 1% of requests should fail
  },
};

const BASE_URL = 'http://localhost:8080';

// Simulate realistic game behavior
export default function () {
  const turnStart = Date.now();

  // 1. Start a new game with a random player name
  const playerNames = ['Alice', 'Bob', 'Charlie', 'Dana', 'Eve', 'Frank', 'Grace'];
  const playerName = playerNames[Math.floor(Math.random() * playerNames.length)];

  http.post(`${BASE_URL}/api/set-player-name`, JSON.stringify({
    player_name: playerName,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  sleep(randomIntBetween(1, 3));

  // 2. Play multiple turns (realistic gameplay)
  const numTurns = randomIntBetween(3, 8); // Play 3-8 turns

  for (let turn = 0; turn < numTurns; turn++) {
    let turnScore = 0;
    let rollsThisTurn = 0;
    const maxRolls = randomIntBetween(2, 6); // Each turn has 2-6 rolls

    // Keep rolling until we decide to bank or farkle
    for (let roll = 0; roll < maxRolls; roll++) {
      // Roll dice
      const rollRes = http.post(`${BASE_URL}/api/roll`, JSON.stringify({
        dice_to_keep: [],
      }), {
        headers: { 'Content-Type': 'application/json' },
      });

      gameRolls.add(1);
      rollsThisTurn++;

      check(rollRes, {
        'roll successful': (r) => r.status === 200,
      });

      if (rollRes.status === 200) {
        const gameState = rollRes.json();

        // Check if we farkled
        if (gameState.turn.is_game_over) {
          gameFarkles.add(1);
          break; // Farkled, end turn
        }

        turnScore = gameState.turn.accumulated_score;

        // Decide whether to bank (more likely as score increases)
        const shouldBank = (
          turnScore >= 500 && // Must have at least 500 for first bank
          (turnScore > 1000 || Math.random() < 0.3) // Bank if > 1000 or 30% chance
        );

        if (shouldBank || roll === maxRolls - 1) {
          // Bank points
          const bankRes = http.get(`${BASE_URL}/api/bank`);

          check(bankRes, {
            'bank successful': (r) => r.status === 200,
          });

          if (bankRes.status === 200) {
            gameBanks.add(1);
            turnDuration.add(Date.now() - turnStart);
          }

          break;
        }
      }

      sleep(randomIntBetween(1, 3)); // Think time between rolls
    }

    // Check win condition
    const stateRes = http.get(`${BASE_URL}/api/state`);
    if (stateRes.status === 200) {
      const state = stateRes.json();
      if (state.winner || state.total_bank >= 10000) {
        // Game won, reset for next iteration
        http.get(`${BASE_URL}/api/reset`);
        sleep(randomIntBetween(2, 5));
        break;
      }
    }

    sleep(randomIntBetween(2, 4)); // Think time between turns
  }

  // Reset game for next VU iteration
  http.get(`${BASE_URL}/api/reset`);

  // Longer sleep between complete game sessions
  sleep(randomIntBetween(10, 30));
}
