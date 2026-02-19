# Performance Testing with k6

## Audience

This document targets backend engineers who want to measure, understand, and improve the performance of server-side APIs and services. It assumes familiarity with HTTP, basic Go, and command-line tooling, but no prior load-testing experience.

---

## What Is Performance Testing?

**Performance testing** is a category of non-functional testing that validates how a system behaves under a given load. Unlike functional tests, which verify _correctness_, performance tests verify _efficiency_ — how fast, how stable, and how scalable a system is.

The field contains several distinct test types. Understanding which type answers which question prevents wasted effort.

| Test Type       | Question Answered                            | Typical Load Pattern                  |
| --------------- | -------------------------------------------- | ------------------------------------- |
| **Load test**   | Does the system meet SLOs at expected load?  | Steady state at realistic concurrency |
| **Stress test** | Where is the breaking point?                 | Gradually increasing until failure    |
| **Spike test**  | Can the system absorb sudden traffic bursts? | Step-change to peak, then back down   |
| **Soak test**   | Does the system degrade over time?           | Sustained moderate load for hours     |
| **Smoke test**  | Does the script itself work correctly?       | 1–2 virtual users, short duration     |

k6 supports all of these patterns through its **executor** and **scenario** configuration.

---

## Why Do Performance Testing?

Backend systems fail in production for reasons that unit and integration tests never catch:

- A database query is correct but takes 4 seconds under concurrency due to lock contention.
- A memory leak causes a process to exhaust heap after 6 hours of continuous traffic.
- A third-party API rate-limits at 50 req/s, but no backoff is implemented, causing cascading 429 failures.
- Auto-scaling kicks in too slowly, causing a 30-second latency spike on traffic bursts.

Performance testing surfaces these issues **before they affect users**. More specifically, it enables you to:

1. **Establish a performance baseline** — know what "normal" looks like.
2. **Set and enforce SLOs** — define acceptable p95 latency and error rates as code.
3. **Catch regressions** — run tests in CI to detect slowdowns introduced by new code.
4. **Inform capacity planning** — understand how many instances you need to handle projected load.
5. **Validate infrastructure changes** — confirm that a new caching layer, index, or CDN actually helps.

---

## When to Do Performance Testing

Run performance tests at multiple stages of the development lifecycle, not only before a major release.

### During Development

- Run a **smoke test** (`1–2 VUs, 30 seconds`) after writing a new endpoint.
  - Goal: confirm the k6 script works and the endpoint doesn't immediately error.
- Run a **quick load test** when you change a query, add middleware, or modify serialization logic.
  - Goal: detect obvious regressions before committing.

### In CI/CD

- Run a smoke test on every pull request against a staging environment.
- Run a full load test on every merge to `main` or `develop`.
- Fail the pipeline if thresholds (p95 latency, error rate) are exceeded.

### Before Production Releases

- Run a **load test** at expected peak traffic (e.g., 2× normal daily traffic).
- Run a **soak test** for 30–60 minutes to surface memory or connection leaks.
- Run a **spike test** if the release includes a feature that may cause irregular traffic.

### Periodically

- Re-run soak tests monthly or after infrastructure changes.
- Re-baseline after major refactors (e.g., migrating from REST to gRPC).

---

## How to Do Performance Testing Effectively with k6

### 1. Install k6

```bash
# macOS
brew install k6

# Docker (no install required)
docker run --rm -i grafana/k6 run - <script.js
```

### 2. Write a Minimal Smoke Test First

Always start with a smoke test. It validates your script logic before you scale up.

```javascript
// smoke.js
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 1,
  duration: "30s",
};

export default function () {
  const res = http.get("http://localhost:8080/api/v1/health");

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time < 200ms": (r) => r.timings.duration < 200,
  });

  sleep(1);
}
```

Run it:

```bash
k6 run smoke.js
```

### 3. Define Thresholds as Code

Encode your SLOs directly in the script. k6 will exit with a non-zero code if thresholds are violated — which integrates cleanly with CI.

```javascript
export const options = {
  thresholds: {
    // 95th-percentile latency must stay below 500ms
    http_req_duration: ["p(95)<500"],
    // Error rate must stay below 1%
    http_req_failed: ["rate<0.01"],
  },
};
```

### 4. Model Realistic Scenarios

Avoid testing a single isolated endpoint. Real traffic hits multiple endpoints with different weights.

```javascript
// scenarios.js
import http from "k6/http";
import { sleep } from "k6";

export const options = {
  scenarios: {
    browse_products: {
      executor: "constant-arrival-rate",
      rate: 100, // 100 iterations/second
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 50,
    },
    place_order: {
      executor: "constant-arrival-rate",
      rate: 10, // 10 iterations/second
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 20,
    },
  },
};

export function browse_products() {
  http.get("http://localhost:8080/api/v1/products");
  sleep(1);
}

export function place_order() {
  http.post(
    "http://localhost:8080/api/v1/orders",
    JSON.stringify({ product_id: 1, qty: 2 }),
    {
      headers: { "Content-Type": "application/json" },
    },
  );
  sleep(1);
}
```

### 5. Use `constant-arrival-rate` for Throughput Tests

k6 offers two executor families:

- **VU-based** (`constant-vus`, `ramping-vus`): fixes the number of concurrent users. Throughput varies with response time — slower responses mean fewer requests per second.
- **Arrival-rate** (`constant-arrival-rate`, `ramping-arrival-rate`): fixes the request rate. k6 spawns as many VUs as needed to maintain the rate.

For backend API load tests, **prefer arrival-rate executors**. They more accurately model real traffic and isolate the effect of latency changes on the system rather than on the test itself.

### 6. Parameterize Test Data

Hardcoded IDs and tokens cause artificial cache hits and skew results. Use k6's `SharedArray` to load realistic test data.

```javascript
import { SharedArray } from "k6/data";

const users = new SharedArray("users", function () {
  return JSON.parse(open("./users.json"));
});

export default function () {
  const user = users[Math.floor(Math.random() * users.length)];
  http.get(`http://localhost:8080/api/v1/profile/${user.id}`, {
    headers: { Authorization: `Bearer ${user.token}` },
  });
}
```

### 7. Collect and Analyze Metrics

k6 emits rich built-in metrics. The most important for backend services:

| Metric              | Meaning                                                      |
| ------------------- | ------------------------------------------------------------ |
| `http_req_duration` | End-to-end request latency (connect + send + wait + receive) |
| `http_req_waiting`  | Time-to-first-byte (TTFB): proxy for server processing time  |
| `http_req_failed`   | Fraction of requests that failed (non-2xx or network error)  |
| `http_reqs`         | Total request count and rate                                 |
| `vus`               | Active virtual users at each point in time                   |

Always look at percentiles (`p(50)`, `p(95)`, `p(99)`), not averages. A low average latency can hide a degraded tail that affects a significant portion of users.

### 8. Send Results to a Dashboard

Plain terminal output is fine for smoke tests. For load and soak tests, stream results to Grafana + InfluxDB or Grafana Cloud k6 for time-series visualization.

```bash
# Stream to InfluxDB
k6 run --out influxdb=http://localhost:8086/k6 script.js

# Stream to Grafana Cloud k6
K6_CLOUD_TOKEN=<token> k6 cloud script.js
```

---

## Common Mistakes and Pitfalls

### Testing Against Localhost

Running k6 on the same host as the server under test shares CPU, memory, and network stack. k6 itself uses resources, which distorts the results. **Always run k6 from a separate machine or container**.

### Ignoring the Coordinated Omission Problem

VU-based executors with `sleep()` hide queuing delays. When the server is slow, VUs sleep at end of an iteration, so the test sends fewer requests — masking the overload. Use `constant-arrival-rate` to send requests regardless of iteration latency.

### Not Warming Up the System

JIT compilation (Go's runtime, JVM, etc.), cold caches, and connection pool initialization all cause artificially high latency in the first 10–30 seconds. Add a ramp-up stage to your load profile so that baseline metrics reflect a warm system.

```javascript
export const options = {
  stages: [
    { duration: "30s", target: 50 }, // ramp up
    { duration: "3m", target: 50 }, // steady state — measure here
    { duration: "15s", target: 0 }, // ramp down
  ],
};
```

### Using a Single Endpoint or Fixed Data

Testing one endpoint with one data point exercises only one code path, one database row, and one cache key. Results will not generalize to production traffic. Diversify endpoints and use randomized test data (see §6 above).

### Neglecting the Database

The application server often scales easily; the database rarely does. Always monitor database metrics (query latency, active connections, lock waits, table sizes) during load tests. k6 outputs tell you only what the client sees — combine them with server-side observability.

### Setting Thresholds Without a Baseline

A threshold of `p(95)<500ms` is arbitrary without knowing current performance. Run a baseline load test first; then tighten thresholds incrementally based on measured behavior and business requirements.

### Running Load Tests Against Production

Even a well-tuned load test can saturate connection pools or degrade cache hit rates in a shared production database. **Always test against a production-like staging environment** with a separate database instance.

### Treating a Passing Test as a Sign-off

A load test answers one specific question under one specific condition. A passing test at 100 RPS says nothing about behavior at 500 RPS, or after 8 hours, or with a slow downstream dependency. Interpret results in context; escalate test scope as risk warrants.

---

## Quick Reference

```bash
# Smoke test (1 VU, 30 s)
k6 run --vus 1 --duration 30s script.js

# Load test (50 VUs, 5 min)
k6 run --vus 50 --duration 5m script.js

# Stress test (ramp to 500 VUs via stages defined in script)
k6 run script.js

# Output summary as JSON for CI parsing
k6 run --summary-export=summary.json script.js
```

---

## Further Reading

- [k6 documentation](https://grafana.com/docs/k6/latest/)
- [k6 executor reference](https://grafana.com/docs/k6/latest/using-k6/scenarios/executors/)
- [k6 thresholds](https://grafana.com/docs/k6/latest/using-k6/thresholds/)
- [The Coordinated Omission Problem – Gil Tene](https://www.youtube.com/watch?v=lJ8ydIuPFeU)
- [How NOT to measure latency – Gil Tene](https://www.infoq.com/presentations/latency-pitfalls/)
