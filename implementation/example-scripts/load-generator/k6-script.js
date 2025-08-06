import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';

const VUS = 1000;
const DURATION = '10s';
const RPS = 1000;

// Define custom trends for each endpoint
const browseTrend = new Trend('http_req_duration_browse');
const submitTrend = new Trend('http_req_duration_submit');

export const options = {
  scenarios: {
    main_scenario: {
      executor: 'constant-arrival-rate',
      rate: RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: VUS,
      maxVUs: VUS,
      exec: 'main',
    },
  },
  thresholds: {
    // Define thresholds if needed, e.g., http_req_failed < 0.01
  },
};

// --- Test Logic ---

// 1. Login once at the beginning of the test to get a token.
// The data returned from setup() is passed to the main execution function.
export function setup() {
    const loginRes = http.post('http://localhost:8080/login', JSON.stringify({ username: 'testuser' }));
    check(loginRes, { 'login status was 200': (r) => r.status === 200 });
    const token = loginRes.json('token');
    console.log(`Setup: Logged in and got token.`);
    return { authToken: token };
}


// 2. Main execution function, called by each VU.
// The 'data' parameter is the object returned from setup().
export function main(data) {
    const token = data.authToken;
    if (!token) {
        console.error("VU has no auth token, skipping iteration.");
        return;
    }

    // Simple weighted random selection to mix API calls
    const random = Math.random();

    if (random < 0.8) { // 80% chance
        const res = http.get('http://localhost:8080/browse', {
            headers: { Authorization: `Bearer ${token}` },
        });
        check(res, { 'browse status was 200': (r) => r.status === 200 });
        browseTrend.add(res.timings.duration);

    } else { // 20% chance
        const res = http.post('http://localhost:8080/submit', JSON.stringify({ data: 'sample' }), {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
        });
        check(res, { 'submit status was 200': (r) => r.status === 200 });
        submitTrend.add(res.timings.duration);
    }

    // sleep(0.1); // Small sleep to prevent overwhelming the system in a tight loop
}
