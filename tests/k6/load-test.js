/**
 * Load Test - Expected production traffic
 *
 * Purpose: Test system under expected load
 * VUs: 10
 * Duration: 5 minutes (1min ramp-up, 3min sustained, 1min ramp-down)
 *
 * Run: k6 run tests/k6/load-test.js
 */

import { check, sleep } from 'k6';
import {
  BASE_URL,
  resetGame,
  setPlayerName,
  playRealisticTurn,
  getCommonThresholds,
} from './utils.js';
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '1m', target: 10 }, // Ramp-up to 10 users
    { duration: '3m', target: 10 }, // Stay at 10 users
    { duration: '1m', target: 0 },  // Ramp-down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<250', 'p(99)<750'], // Slightly relaxed for sustained load
    http_req_failed: ['rate<0.01'],
    checks: ['rate>0.95'], // 95% of checks should pass
  },
};

export default function () {
  // Each VU represents a player
  const playerName = `Player_${__VU}`;

  // Reset game
  resetGame();
  sleep(0.5);

  // Set player name
  setPlayerName(playerName);
  sleep(0.5);

  // Play multiple turns (5-10 turns per iteration)
  const numTurns = Math.floor(Math.random() * 6) + 5;

  for (let turn = 0; turn < numTurns; turn++) {
    playRealisticTurn();

    // Random think time between turns (1-3 seconds)
    sleep(1 + Math.random() * 2);
  }

  // Check metrics endpoint occasionally
  if (Math.random() < 0.1) {
    const metricsResponse = http.get(`${BASE_URL}/metrics`);
    check(metricsResponse, {
      'metrics endpoint accessible': (r) => r.status === 200,
    });
  }

  // Longer pause between game sessions
  sleep(2 + Math.random() * 3);
}

export function handleSummary(data) {
  console.log('\n=================================');
  console.log('Load Test Summary');
  console.log('=================================\n');

  if (data.metrics.http_req_duration) {
    const avg = data.metrics.http_req_duration.values.avg;
    const p95 = data.metrics.http_req_duration.values['p(95)'];
    const p99 = data.metrics.http_req_duration.values['p(99)'];
    console.log('Response Times:');
    console.log(`  avg: ${avg.toFixed(2)}ms`);
    console.log(`  p95: ${p95.toFixed(2)}ms`);
    console.log(`  p99: ${p99.toFixed(2)}ms\n`);
  }

  if (data.metrics.http_reqs) {
    const rate = data.metrics.http_reqs.values.rate;
    console.log(`Request Rate: ${rate.toFixed(2)} req/s`);
    console.log(`Total Requests: ${data.metrics.http_reqs.values.count}\n`);
  }

  if (data.metrics.checks) {
    const passed = data.metrics.checks.values.passes;
    const failed = data.metrics.checks.values.fails;
    const total = passed + failed;
    const passRate = ((passed / total) * 100).toFixed(2);
    console.log(`Checks: ${passed}/${total} passed (${passRate}%)\n`);
  }

  if (data.metrics.http_req_failed) {
    const failRate = (data.metrics.http_req_failed.values.rate * 100).toFixed(2);
    console.log(`Failed Requests: ${failRate}%\n`);
  }

  return {
    'stdout': '',
  };
}
