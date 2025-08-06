# k6 Load Generator Script

This directory contains the `k6` script used to generate a high-traffic load against the `app-server`.

## Script: `k6-script.js`

This script is designed to simulate 1,000 concurrent users generating a constant 1,000 requests per second (RPS) for 60 seconds.

### Workflow

1.  **`setup()` function:** Before the main test starts, a single user logs in to get an authentication token.
2.  **`main()` function:** Each virtual user then repeatedly hits the `/browse` and `/submit` endpoints based on a weighted probability (80% browse, 20% submit).

This script is executed automatically by the main `run_test.sh` script.
