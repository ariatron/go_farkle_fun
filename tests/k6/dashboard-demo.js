/**
 * Dashboard Demo Test - Generate impressive visualizations
 *
 * Purpose: Create varied traffic patterns to showcase the dashboard
 * VUs: Varies by stage
 * Duration: 10 minutes
 *
 * This test creates:
 * - Variable request rates for interesting graphs
 * - Mix of successful and failed requests
 * - Different game patterns (quick games, long games, farkles)
 * - Logs with different levels
 * - Traces with various durations
 *
 * Run: k6 run tests/k6/dashboard-demo.js
 */

import { check, sleep, group } from 'k6';
import { Counter, Trend, Rate, Gauge } from 'k6/metrics';
import http from 'k6/http';
import {
  BASE_URL,
  JSON_HEADERS,
  resetGame,
  setPlayerName,
  rollDice,
  bankPoints,
} from './utils.js';

// Custom metrics for dashboard
const gameSessionDuration = new Trend('game_session_duration');
const pointsPerGame = new Trend('points_per_game');
const turnsPerGame = new Trend('turns_per_game');
const activePlayers = new Gauge('active_players');

export const options = {
  stages: [
    // Warm-up: gradual increase
    { duration: '1m', target: 5 },

    // Peak traffic: showcase high load
    { duration: '2m', target: 15 },

    // Variable load: create interesting patterns
    { duration: '1m', target: 8 },
    { duration: '1m', target: 20 },
    { duration: '1m', target: 5 },

    // Sustained moderate load
    { duration: '3m', target: 10 },

    // Cool down
    { duration: '1m', target: 2 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
    game_session_duration: ['avg<120000'], // 2 minutes
  },
};

export default function () {
  const playerName = `Player_${__VU}_${__ITER}`;
  activePlayers.add(1);

  const sessionStart = Date.now();
  let totalPoints = 0;
  let turns = 0;

  // Randomly choose game pattern
  const gamePattern = Math.random();

  if (gamePattern < 0.3) {
    // Quick game: aggressive banking (30%)
    quickAggressiveGame(playerName);
  } else if (gamePattern < 0.6) {
    // Normal game: balanced strategy (30%)
    normalBalancedGame(playerName);
  } else if (gamePattern < 0.85) {
    // Risky game: many farkles (25%)
    riskyGame(playerName);
  } else {
    // Full game: play to 10,000 (15%)
    fullGame(playerName);
  }

  activePlayers.add(-1);

  const sessionDuration = Date.now() - sessionStart;
  gameSessionDuration.add(sessionDuration);

  // Random pause between games
  sleep(1 + Math.random() * 3);
}

function quickAggressiveGame(playerName) {
  group('Quick Aggressive Game', function () {
    resetGame();
    setPlayerName(playerName);

    let turns = 0;
    const maxTurns = 5;

    while (turns < maxTurns) {
      const state = rollDice();

      // Bank aggressively (even small amounts)
      if (state.turn.accumulated_score > 50 && !state.turn.is_game_over) {
        bankPoints();
        turnsPerGame.add(turns);
        break;
      }

      turns++;
      sleep(0.2 + Math.random() * 0.3);
    }
  });
}

function normalBalancedGame(playerName) {
  group('Normal Balanced Game', function () {
    resetGame();
    setPlayerName(playerName);

    let turns = 0;
    const maxTurns = 15;

    while (turns < maxTurns) {
      const state = rollDice();

      // Bank when we have decent points (200+)
      if (state.turn.accumulated_score >= 200 && !state.turn.is_game_over) {
        bankPoints();
      }

      // Sometimes risk another roll
      if (state.turn.accumulated_score < 500 && !state.turn.is_game_over && Math.random() < 0.6) {
        rollDice([1]); // Keep simple scoring die
      }

      turns++;
      sleep(0.5 + Math.random() * 0.5);

      if (state.turn.is_game_over) {
        break;
      }
    }

    turnsPerGame.add(turns);
  });
}

function riskyGame(playerName) {
  group('Risky Game - Many Farkles', function () {
    resetGame();
    setPlayerName(playerName);

    let turns = 0;
    const maxTurns = 10;

    while (turns < maxTurns) {
      // Roll multiple times per turn (risky!)
      for (let i = 0; i < 3; i++) {
        const state = rollDice();

        if (state.turn.is_game_over) {
          // Farkled!
          break;
        }

        sleep(0.3);
      }

      turns++;
      sleep(0.5);
    }

    turnsPerGame.add(turns);
  });
}

function fullGame(playerName) {
  group('Full Game to 10,000', function () {
    resetGame();
    setPlayerName(playerName);

    let turns = 0;
    const maxTurns = 50;

    while (turns < maxTurns) {
      const state = rollDice();

      // Bank when we have good points
      if (state.turn.accumulated_score >= 300 && !state.turn.is_game_over) {
        const bankState = bankPoints();

        if (bankState.winner) {
          pointsPerGame.add(bankState.total_bank);
          turnsPerGame.add(turns);
          return; // Won!
        }
      }

      turns++;
      sleep(0.5 + Math.random() * 0.5);

      if (state.turn.is_game_over) {
        sleep(0.5);
      }
    }

    turnsPerGame.add(turns);
  });
}

export function handleSummary(data) {
  console.log('\n====================================');
  console.log('Dashboard Demo Test - Summary');
  console.log('====================================\n');

  if (data.metrics.http_req_duration) {
    const avg = data.metrics.http_req_duration.values.avg;
    const p95 = data.metrics.http_req_duration.values['p(95)'];
    const p99 = data.metrics.http_req_duration.values['p(99)'];
    console.log('HTTP Performance:');
    console.log(`  avg: ${avg.toFixed(2)}ms`);
    console.log(`  p95: ${p95.toFixed(2)}ms`);
    console.log(`  p99: ${p99.toFixed(2)}ms\n`);
  }

  if (data.metrics.http_reqs) {
    const rate = data.metrics.http_reqs.values.rate;
    const total = data.metrics.http_reqs.values.count;
    console.log('Request Stats:');
    console.log(`  total: ${total}`);
    console.log(`  rate: ${rate.toFixed(2)} req/s\n`);
  }

  if (data.metrics.game_session_duration) {
    const avg = data.metrics.game_session_duration.values.avg / 1000;
    console.log('Game Sessions:');
    console.log(`  avg duration: ${avg.toFixed(1)}s\n`);
  }

  if (data.metrics.turns_per_game) {
    const avg = data.metrics.turns_per_game.values.avg;
    console.log('Game Turns:');
    console.log(`  avg turns: ${avg.toFixed(1)}\n`);
  }

  console.log('✨ Dashboard should now have rich data!');
  console.log('View at: https://your-stack.grafana.net\n');

  return {
    'stdout': '',
  };
}
