/**
 * Game Scenario Test - Realistic player behavior
 *
 * Purpose: Simulate complete games with realistic player behavior
 * VUs: 5
 * Duration: 10 minutes
 *
 * Run: k6 run tests/k6/game-scenario.js
 */

import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';
import {
  BASE_URL,
  resetGame,
  setPlayerName,
  playFullGame,
  playRealisticTurn,
} from './utils.js';

// Custom metrics
const gamesCompleted = new Counter('games_completed');
const gamesWon = new Counter('games_won');
const averageTurnsPerGame = new Trend('avg_turns_per_game');
const gameCompletionRate = new Rate('game_completion_rate');

export const options = {
  vus: 5,
  duration: '10m',
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<800'],
    http_req_failed: ['rate<0.01'],
    games_completed: ['count>10'], // Expect at least 10 games completed
    game_completion_rate: ['rate>0.8'], // 80% of attempts should complete
  },
};

export default function () {
  const playerName = `GamePlayer_${__VU}_${Date.now()}`;

  // Reset and start new game
  resetGame();
  sleep(1);

  setPlayerName(playerName);
  sleep(0.5);

  console.log(`${playerName} starting a new game...`);

  // Play a full game (up to 100 turns)
  const result = playFullGame(100);

  gamesCompleted.add(1);
  averageTurnsPerGame.add(result.turns);

  if (result.winner) {
    gamesWon.add(1);
    gameCompletionRate.add(1);
    console.log(`${playerName} WON after ${result.turns} turns! Final score: ${result.finalScore}`);
  } else {
    gameCompletionRate.add(0);
    console.log(`${playerName} reached turn limit. Score: ${result.finalScore}`);
  }

  // Pause between games
  sleep(2 + Math.random() * 3);
}

export function handleSummary(data) {
  console.log('\n=================================');
  console.log('Game Scenario Test Summary');
  console.log('=================================\n');

  // HTTP Performance
  if (data.metrics.http_req_duration) {
    const avg = data.metrics.http_req_duration.values.avg;
    const p95 = data.metrics.http_req_duration.values['p(95)'];
    const p99 = data.metrics.http_req_duration.values['p(99)'];
    console.log('HTTP Performance:');
    console.log(`  avg response time: ${avg.toFixed(2)}ms`);
    console.log(`  p95 response time: ${p95.toFixed(2)}ms`);
    console.log(`  p99 response time: ${p99.toFixed(2)}ms\n`);
  }

  // Request Stats
  if (data.metrics.http_reqs) {
    const rate = data.metrics.http_reqs.values.rate;
    const total = data.metrics.http_reqs.values.count;
    console.log('Request Stats:');
    console.log(`  total requests: ${total}`);
    console.log(`  request rate: ${rate.toFixed(2)} req/s\n`);
  }

  // Game Stats
  if (data.metrics.games_completed) {
    const completed = data.metrics.games_completed.values.count;
    console.log('Game Stats:');
    console.log(`  games completed: ${completed}`);
  }

  if (data.metrics.games_won) {
    const won = data.metrics.games_won.values.count;
    const completed = data.metrics.games_completed?.values.count || 1;
    const winRate = ((won / completed) * 100).toFixed(2);
    console.log(`  games won: ${won}`);
    console.log(`  win rate: ${winRate}%`);
  }

  if (data.metrics.avg_turns_per_game) {
    const avg = data.metrics.avg_turns_per_game.values.avg;
    const min = data.metrics.avg_turns_per_game.values.min;
    const max = data.metrics.avg_turns_per_game.values.max;
    console.log(`  avg turns per game: ${avg.toFixed(1)}`);
    console.log(`  min turns: ${min}`);
    console.log(`  max turns: ${max}\n`);
  }

  // Reliability
  if (data.metrics.http_req_failed) {
    const failRate = (data.metrics.http_req_failed.values.rate * 100).toFixed(2);
    console.log('Reliability:');
    console.log(`  failed requests: ${failRate}%`);
  }

  if (data.metrics.game_completion_rate) {
    const completionRate = (data.metrics.game_completion_rate.values.rate * 100).toFixed(2);
    console.log(`  game completion rate: ${completionRate}%\n`);
  }

  // Analysis
  console.log('Analysis:');
  const p95 = data.metrics.http_req_duration?.values['p(95)'] || 0;
  const failRate = data.metrics.http_req_failed?.values.rate || 0;

  if (p95 < 200 && failRate < 0.01) {
    console.log('  ✓ Excellent: Low latency and high reliability');
  } else if (p95 < 300 && failRate < 0.02) {
    console.log('  ✓ Good: Acceptable performance under realistic load');
  } else if (p95 < 500 && failRate < 0.05) {
    console.log('  ⚠️  Fair: Performance degradation observed');
  } else {
    console.log('  ❌ Poor: Significant performance issues detected');
  }

  console.log('\n');

  return {
    'stdout': '',
  };
}
