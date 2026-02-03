/**
 * Stress Test - Find the breaking point
 *
 * Purpose: Push system beyond normal limits to find breaking point
 * VUs: Up to 50
 * Duration: 9 minutes (2min ramp-up, 5min stress, 2min ramp-down)
 *
 * Run: k6 run tests/k6/stress-test.js
 */

import { check, sleep } from 'k6';
import {
  BASE_URL,
  resetGame,
  setPlayerName,
  rollDice,
  bankPoints,
  playRealisticTurn,
} from './utils.js';
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '2m', target: 20 },  // Ramp-up to 20 users
    { duration: '2m', target: 50 },  // Scale to 50 users
    { duration: '5m', target: 50 },  // Maintain stress
    { duration: '2m', target: 0 },   // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'], // More relaxed for stress test
    http_req_failed: ['rate<0.05'], // Allow up to 5% failure
  },
};

export default function () {
  const playerName = `StressPlayer_${__VU}`;

  try {
    // Quick game session
    resetGame();
    sleep(0.2);

    setPlayerName(playerName);
    sleep(0.2);

    // Play 3-5 quick turns
    const numTurns = Math.floor(Math.random() * 3) + 3;

    for (let turn = 0; turn < numTurns; turn++) {
      // Faster gameplay under stress
      const state = rollDice();

      if (!state.turn.is_game_over && state.turn.accumulated_score > 0) {
        // Quick decision to bank
        if (Math.random() < 0.5) {
          bankPoints();
        } else {
          // Or roll again
          rollDice([1]); // Keep simple dice
        }
      }

      sleep(0.1 + Math.random() * 0.2); // Minimal think time
    }

    // Shorter pause between sessions under stress
    sleep(0.5 + Math.random());

  } catch (error) {
    console.error(`Error in VU ${__VU}: ${error}`);
  }
}

export function handleSummary(data) {
  console.log('\n=================================');
  console.log('Stress Test Summary');
  console.log('=================================\n');

  if (data.metrics.http_req_duration) {
    const avg = data.metrics.http_req_duration.values.avg;
    const p95 = data.metrics.http_req_duration.values['p(95)'];
    const p99 = data.metrics.http_req_duration.values['p(99)'];
    const max = data.metrics.http_req_duration.values.max;
    console.log('Response Times:');
    console.log(`  avg: ${avg.toFixed(2)}ms`);
    console.log(`  p95: ${p95.toFixed(2)}ms`);
    console.log(`  p99: ${p99.toFixed(2)}ms`);
    console.log(`  max: ${max.toFixed(2)}ms\n`);
  }

  if (data.metrics.http_reqs) {
    const rate = data.metrics.http_reqs.values.rate;
    console.log(`Request Rate: ${rate.toFixed(2)} req/s`);
    console.log(`Total Requests: ${data.metrics.http_reqs.values.count}\n`);
  }

  if (data.metrics.checks) {
    const passed = data.metrics.checks.values.passes || 0;
    const failed = data.metrics.checks.values.fails || 0;
    const total = passed + failed;
    if (total > 0) {
      const passRate = ((passed / total) * 100).toFixed(2);
      console.log(`Checks: ${passed}/${total} passed (${passRate}%)\n`);
    }
  }

  if (data.metrics.http_req_failed) {
    const failRate = (data.metrics.http_req_failed.values.rate * 100).toFixed(2);
    const failCount = data.metrics.http_req_failed.values.passes;
    console.log(`Failed Requests: ${failRate}% (${failCount} total)\n`);
  }

  if (data.metrics.vus_max) {
    console.log(`Peak VUs: ${data.metrics.vus_max.values.max}\n`);
  }

  console.log('Analysis:');
  if (data.metrics.http_req_failed && data.metrics.http_req_failed.values.rate > 0.01) {
    console.log('  ⚠️  Elevated error rate detected under stress');
  } else {
    console.log('  ✓ System handled stress well');
  }

  if (data.metrics.http_req_duration && data.metrics.http_req_duration.values['p(99)'] > 1000) {
    console.log('  ⚠️  High latency observed at p99');
  } else {
    console.log('  ✓ Latency remained acceptable');
  }

  console.log('\n');

  return {
    'stdout': '',
  };
}
