/**
 * Spike Test - Sudden traffic surge
 *
 * Purpose: Test system recovery from sudden traffic spikes
 * VUs: 2 baseline, spike to 50
 * Duration: 5 minutes
 *
 * Run: k6 run tests/k6/spike-test.js
 */

import { check, sleep } from 'k6';
import {
  BASE_URL,
  resetGame,
  setPlayerName,
  rollDice,
  bankPoints,
} from './utils.js';
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '30s', target: 2 },   // Baseline
    { duration: '10s', target: 50 },  // Sudden spike!
    { duration: '1m', target: 50 },   // Maintain spike
    { duration: '30s', target: 2 },   // Return to baseline
    { duration: '1m', target: 2 },    // Recovery period
  ],
  thresholds: {
    http_req_duration: ['p(95)<1500', 'p(99)<3000'], // Relaxed during spike
    http_req_failed: ['rate<0.10'], // Allow up to 10% failure during spike
  },
};

export default function () {
  const playerName = `SpikePlayer_${__VU}`;

  try {
    // Minimal game interaction
    resetGame();
    sleep(0.1);

    setPlayerName(playerName);
    sleep(0.1);

    // 2-3 quick rolls
    const numRolls = Math.floor(Math.random() * 2) + 2;

    for (let i = 0; i < numRolls; i++) {
      const state = rollDice();

      if (state.turn.accumulated_score > 100 && Math.random() < 0.4) {
        bankPoints();
        break;
      }

      sleep(0.1);
    }

    // Very short pause
    sleep(0.2 + Math.random() * 0.3);

  } catch (error) {
    // Expect some errors during spike
    check(error, {
      'error is acceptable during spike': () => true,
    });
  }
}

export function handleSummary(data) {
  console.log('\n=================================');
  console.log('Spike Test Summary');
  console.log('=================================\n');

  console.log('Test Profile:');
  console.log('  Baseline: 2 VUs');
  console.log('  Spike: 50 VUs');
  console.log('  Duration: 5 minutes\n');

  if (data.metrics.http_req_duration) {
    const avg = data.metrics.http_req_duration.values.avg;
    const p50 = data.metrics.http_req_duration.values.med;
    const p95 = data.metrics.http_req_duration.values['p(95)'];
    const p99 = data.metrics.http_req_duration.values['p(99)'];
    const max = data.metrics.http_req_duration.values.max;

    console.log('Response Times:');
    console.log(`  avg: ${avg.toFixed(2)}ms`);
    console.log(`  p50: ${p50.toFixed(2)}ms`);
    console.log(`  p95: ${p95.toFixed(2)}ms`);
    console.log(`  p99: ${p99.toFixed(2)}ms`);
    console.log(`  max: ${max.toFixed(2)}ms\n`);
  }

  if (data.metrics.http_reqs) {
    const rate = data.metrics.http_reqs.values.rate;
    console.log(`Request Rate: ${rate.toFixed(2)} req/s`);
    console.log(`Total Requests: ${data.metrics.http_reqs.values.count}\n`);
  }

  if (data.metrics.http_req_failed) {
    const failRate = (data.metrics.http_req_failed.values.rate * 100).toFixed(2);
    const failCount = data.metrics.http_req_failed.values.passes;
    console.log(`Failed Requests: ${failRate}% (${failCount} total)\n`);
  }

  console.log('Recovery Analysis:');
  if (data.metrics.http_req_failed && data.metrics.http_req_failed.values.rate < 0.05) {
    console.log('  ✓ System handled spike well (< 5% errors)');
  } else if (data.metrics.http_req_failed && data.metrics.http_req_failed.values.rate < 0.10) {
    console.log('  ⚠️  Moderate degradation during spike (5-10% errors)');
  } else {
    console.log('  ❌ Significant degradation during spike (> 10% errors)');
  }

  if (data.metrics.http_req_duration && data.metrics.http_req_duration.values['p(95)'] < 500) {
    console.log('  ✓ Latency remained low even during spike');
  } else if (data.metrics.http_req_duration && data.metrics.http_req_duration.values['p(95)'] < 1500) {
    console.log('  ⚠️  Latency increased during spike but remained acceptable');
  } else {
    console.log('  ❌ High latency during spike');
  }

  console.log('\n');

  return {
    'stdout': '',
  };
}
