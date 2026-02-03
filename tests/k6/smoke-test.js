/**
 * Smoke Test - Minimal load to verify basic functionality
 *
 * Purpose: Verify all endpoints work correctly with minimal load
 * VUs: 1
 * Duration: 30 seconds
 *
 * Run: k6 run tests/k6/smoke-test.js
 */

import { check, sleep } from 'k6';
import http from 'k6/http';
import {
  BASE_URL,
  JSON_HEADERS,
  checkSuccess,
  resetGame,
  setPlayerName,
  rollDice,
  bankPoints,
  getCommonThresholds,
} from './utils.js';

export const options = {
  vus: 1,
  duration: '30s',
  thresholds: getCommonThresholds(),
};

export default function () {
  // Test 1: Health Check
  let response = http.get(`${BASE_URL}/health`);
  check(response, {
    'health check: status is 200': (r) => r.status === 200,
  });

  // Test 2: Reset Game
  let state = resetGame();
  check(state, {
    'reset: total_bank is 0': (s) => s.total_bank === 0,
    'reset: winner is false': (s) => s.winner === false,
  });

  sleep(0.5);

  // Test 3: Set Player Name
  state = setPlayerName('K6TestPlayer');
  check(state, {
    'set name: player_name is set': (s) => s.player_name === 'K6TestPlayer',
  });

  sleep(0.5);

  // Test 4: Roll Dice
  state = rollDice();
  check(state, {
    'roll: has last_roll': (s) => s.last_roll && s.last_roll.length > 0,
    'roll: dice_remaining is valid': (s) => s.turn.dice_remaining >= 0 && s.turn.dice_remaining <= 6,
  });

  sleep(0.5);

  // Test 5: Bank Points (if we have any)
  if (state.turn.accumulated_score > 0 && !state.turn.is_game_over) {
    state = bankPoints();
    check(state, {
      'bank: accumulated_score is reset': (s) => s.turn.accumulated_score === 0,
      'bank: dice_remaining is 6': (s) => s.turn.dice_remaining === 6,
    });
  }

  sleep(0.5);

  // Test 6: Metrics Endpoint
  response = http.get(`${BASE_URL}/metrics`);
  check(response, {
    'metrics: status is 200': (r) => r.status === 200,
    'metrics: contains prometheus metrics': (r) => r.body.includes('farkle_'),
  });

  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options = {}) {
  const indent = options.indent || '';
  const enableColors = options.enableColors !== false;

  let summary = `\n${indent}Smoke Test Results:\n`;
  summary += `${indent}==================\n\n`;

  if (data.metrics.http_req_duration) {
    const p95 = data.metrics.http_req_duration.values['p(95)'];
    const p99 = data.metrics.http_req_duration.values['p(99)'];
    summary += `${indent}Response Times:\n`;
    summary += `${indent}  p95: ${p95.toFixed(2)}ms\n`;
    summary += `${indent}  p99: ${p99.toFixed(2)}ms\n\n`;
  }

  if (data.metrics.http_reqs) {
    summary += `${indent}Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  }

  if (data.metrics.checks) {
    const passed = data.metrics.checks.values.passes;
    const failed = data.metrics.checks.values.fails;
    const total = passed + failed;
    const passRate = ((passed / total) * 100).toFixed(2);
    summary += `${indent}Checks: ${passed}/${total} passed (${passRate}%)\n\n`;
  }

  return summary;
}
