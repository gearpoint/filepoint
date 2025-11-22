# Testing Observability Locally

This guide shows you how to test Prometheus metrics and health checks on your local machine.

## Quick Start

### Option 1: Test with Docker Compose (Full Stack)

```bash
# 1. Build and start all services (including Prometheus + Grafana)
docker compose up --build

# 2. Wait for services to start (about 30 seconds)

# 3. Access the services:
# - Filepoint API: http://localhost:9001/v1/health
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (user: admin, pass: admin)
# - Metrics endpoint: http://localhost:9001/v1/metrics
```

### Option 2: Test Locally (Without Docker)

```bash
# 1. Start required services only
make run-services

# 2. Run Filepoint locally
make run

# 3. In another terminal, run webhook sender
make run-webhooks-sender
```

---

## Testing Endpoints

### 1. Health Checks

```bash
# Basic health check
curl http://localhost:9001/v1/health

# Expected response:
# {
#   "status": "up",
#   "timestamp": "2025-01-20T10:30:00Z"
# }

# Liveness probe
curl http://localhost:9001/v1/health/live

# Expected: "OK"

# Readiness probe (checks all dependencies)
curl http://localhost:9001/v1/health/ready | jq

# Expected when healthy:
# {
#   "status": "up",
#   "timestamp": "2025-01-20T10:30:00Z",
#   "dependencies": {
#     "dynamodb": {"status": "up", "latency": "12ms"},
#     "redis": {"status": "up", "latency": "2.5ms"},
#     "s3": {"status": "up", "latency": "15ms"}
#   }
# }
```

### 2. Metrics Endpoint

```bash
# View all metrics
curl http://localhost:9001/v1/metrics

# Grep for specific metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_uploads_total
curl -s http://localhost:9001/v1/metrics | grep filepoint_dependency_up
curl -s http://localhost:9001/v1/metrics | grep filepoint_http_requests_total
```

### 3. Generate Some Traffic

```bash
# Upload a test file to generate metrics
curl -X POST http://localhost:9001/v1/upload \
  -F "userId=test-user-123" \
  -F "title=Test Upload" \
  -F "content=@/path/to/image.jpg"

# Check upload metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_uploads_total
curl -s http://localhost:9001/v1/metrics | grep filepoint_upload_processing
```

---

## Using Prometheus

### Access Prometheus UI

1. Open http://localhost:9090
2. Go to **Status > Targets** to verify Filepoint is being scraped
3. Go to **Graph** to query metrics

### Example Queries

Try these in the Prometheus query box:

**Request Rate:**
```promql
rate(filepoint_http_requests_total[5m])
```

**Upload Success Rate:**
```promql
rate(filepoint_uploads_total{status="success"}[5m])
/
rate(filepoint_uploads_total[5m])
* 100
```

**P95 Latency:**
```promql
histogram_quantile(0.95,
  rate(filepoint_http_request_duration_seconds_bucket[5m])
)
```

**Dependency Health:**
```promql
filepoint_dependency_up
```

**Cache Hit Rate:**
```promql
sum(rate(filepoint_cache_hits_total[5m]))
/
(sum(rate(filepoint_cache_hits_total[5m])) + sum(rate(filepoint_cache_misses_total[5m])))
* 100
```

---

## Using Grafana

### Initial Setup

1. Open http://localhost:3000
2. Login with:
   - **Username:** `admin`
   - **Password:** `admin`
3. Skip password change (or change it)

### Create Your First Dashboard

1. Click **+** → **Create Dashboard**
2. Click **Add visualization**
3. Select **Prometheus** as data source
4. Enter a query (see examples above)
5. Click **Apply**

### Import a Dashboard

Instead of building from scratch, create a dashboard with these panels:

#### Panel 1: Request Rate
- **Query:** `sum(rate(filepoint_http_requests_total[5m])) by (endpoint)`
- **Legend:** `{{endpoint}}`
- **Type:** Graph

#### Panel 2: Upload Success Rate
- **Query:**
  ```promql
  rate(filepoint_uploads_total{status="success"}[5m])
  /
  rate(filepoint_uploads_total[5m])
  * 100
  ```
- **Unit:** Percent (0-100)
- **Type:** Stat

#### Panel 3: P95 Latency
- **Query:**
  ```promql
  histogram_quantile(0.95,
    rate(filepoint_http_request_duration_seconds_bucket[5m])
  )
  ```
- **Unit:** Seconds
- **Type:** Graph

#### Panel 4: Dependency Health
- **Query:** `filepoint_dependency_up`
- **Legend:** `{{service}}`
- **Type:** Stat
- **Thresholds:**
  - Red: 0
  - Green: 1

#### Panel 5: Error Rate
- **Query:**
  ```promql
  sum(rate(filepoint_http_requests_total{status=~"5.."}[5m]))
  /
  sum(rate(filepoint_http_requests_total[5m]))
  * 100
  ```
- **Unit:** Percent
- **Type:** Graph

#### Panel 6: Cache Hit Rate
- **Query:**
  ```promql
  sum(rate(filepoint_cache_hits_total[5m]))
  /
  (sum(rate(filepoint_cache_hits_total[5m])) + sum(rate(filepoint_cache_misses_total[5m])))
  * 100
  ```
- **Unit:** Percent
- **Type:** Graph

---

## Testing Scenarios

### Scenario 1: Normal Operation

1. Start all services
2. Upload a few files
3. Check Prometheus:
   - `filepoint_dependency_up` should be `1` for all services
   - `filepoint_uploads_total` should increment
   - `filepoint_http_requests_total` should show traffic

### Scenario 2: Dependency Down

```bash
# Stop Redis
docker stop redis

# Check readiness endpoint
curl http://localhost:9001/v1/health/ready | jq

# Expected: 503 status, redis dependency shows "down"

# Check Prometheus
# Query: filepoint_dependency_up{service="redis"}
# Expected: 0

# Restart Redis
docker start redis
```

### Scenario 3: High Load

```bash
# Generate load with apache bench (install if needed: apt-get install apache2-utils)
ab -n 1000 -c 10 http://localhost:9001/v1/health

# Watch metrics in Prometheus:
# - rate(filepoint_http_requests_total[1m])
# - histogram_quantile(0.95, rate(filepoint_http_request_duration_seconds_bucket[1m]))
```

### Scenario 4: Upload Validation

```bash
# Try to upload a fake image (text file with .jpg extension)
echo "fake image" > fake.jpg
curl -X POST http://localhost:9001/v1/upload \
  -F "userId=test" \
  -F "content=@fake.jpg"

# Expected: 400 Bad Request (file validation failed)

# Check metrics:
curl -s http://localhost:9001/v1/metrics | grep filepoint_upload_errors_total
# Should show validation errors
```

---

## Verifying Metrics Collection

### Check All Metrics are Being Collected

```bash
# HTTP metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_http_requests_total
curl -s http://localhost:9001/v1/metrics | grep filepoint_http_request_duration

# Upload metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_uploads_total
curl -s http://localhost:9001/v1/metrics | grep filepoint_upload_size

# Cache metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_cache_hits
curl -s http://localhost:9001/v1/metrics | grep filepoint_cache_misses

# Dependency metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_dependency_up

# AWS metrics
curl -s http://localhost:9001/v1/metrics | grep filepoint_s3_operations
curl -s http://localhost:9001/v1/metrics | grep filepoint_dynamodb_operations
```

---

## Troubleshooting

### Prometheus Can't Scrape Filepoint

**Check Prometheus targets:**
1. Go to http://localhost:9090/targets
2. Look for `filepoint` job
3. If "DOWN":
   - Check Filepoint is running: `curl http://localhost:9001/v1/health`
   - Check network: containers should be on same network
   - Check logs: `docker logs prometheus`

### No Metrics Appearing

**Verify metrics endpoint:**
```bash
curl http://localhost:9001/v1/metrics
```

If empty or error:
- Check Filepoint logs: `docker logs filepoint`
- Verify metrics middleware is active
- Try restarting Filepoint

### Grafana Shows "No Data"

**Check:**
1. Prometheus datasource is configured correctly
2. Prometheus is scraping: http://localhost:9090/targets
3. Query is correct: test in Prometheus first
4. Time range is appropriate (try "Last 5 minutes")

### Health Check Shows Dependencies Down

```bash
# Check each service individually:

# Redis
docker ps | grep redis
redis-cli ping

# S3 (LocalStack)
curl http://localhost:4566/_localstack/health

# DynamoDB (LocalStack)
aws dynamodb list-tables --endpoint-url http://localhost:4566
```

---

## Performance Testing

### Generate Realistic Load

```bash
# Install vegeta (load testing tool)
go install github.com/tsenart/vegeta@latest

# Create attack plan
echo "GET http://localhost:9001/v1/health" | \
  vegeta attack -duration=30s -rate=50 | \
  vegeta report

# Watch metrics during load test in Grafana
```

---

## What to Look For

### Good Indicators ✅
- `filepoint_dependency_up` = 1 for all services
- Upload success rate > 95%
- P95 latency < 1s for most endpoints
- Cache hit rate > 70%
- Error rate < 1%

### Warning Signs ⚠️
- Any dependency showing 0 (down)
- Upload failure rate > 5%
- P95 latency > 2s
- Cache hit rate < 50%
- Increasing poison queue messages

### Critical Issues 🔴
- All dependencies down
- Upload failure rate > 20%
- P95 latency > 5s
- Error rate > 10%

---

## Next Steps

After verifying everything works locally:

1. **Set up alerts** - Copy alert rules from `docs/OBSERVABILITY.md`
2. **Create dashboards** - Save your Grafana dashboards
3. **Deploy to production** - Use the same Docker Compose or Kubernetes configs
4. **Monitor SLAs** - Track success rates and latencies over time

---

## Cleanup

```bash
# Stop all services
docker compose down

# Remove volumes (clean slate)
docker compose down -v

# Remove only observability volumes
docker volume rm filepoint_prometheus-data filepoint_grafana-data
```
