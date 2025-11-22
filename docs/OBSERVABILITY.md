# Filepoint Observability Guide

This document describes the observability features available in Filepoint, including metrics, health checks, and monitoring.

## Table of Contents
- [Metrics](#metrics)
- [Health Checks](#health-checks)
- [Prometheus Setup](#prometheus-setup)
- [Grafana Dashboards](#grafana-dashboards)
- [Kubernetes Integration](#kubernetes-integration)
- [Alerts](#alerts)

---

## Metrics

Filepoint exposes Prometheus-compatible metrics at `/v1/metrics`.

### Available Metrics

#### HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `filepoint_http_requests_total` | Counter | method, endpoint, status | Total HTTP requests |
| `filepoint_http_request_duration_seconds` | Histogram | method, endpoint, status | HTTP request latency |
| `filepoint_http_request_size_bytes` | Histogram | method, endpoint | HTTP request size |
| `filepoint_http_response_size_bytes` | Histogram | method, endpoint | HTTP response size |

#### Upload Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `filepoint_uploads_total` | Counter | file_type, status | Total file uploads (status: success/failed/rejected) |
| `filepoint_upload_size_bytes` | Histogram | file_type | Upload file size distribution |
| `filepoint_upload_processing_duration_seconds` | Histogram | file_type, definition | Time to process files |
| `filepoint_upload_errors_total` | Counter | error_type | Upload errors by type (validation/storage/processing/webhook) |
| `filepoint_temp_file_uploads_total` | Counter | - | Temporary file uploads to S3 |
| `filepoint_final_file_uploads_total` | Counter | - | Final processed file uploads |
| `filepoint_file_type_total` | Counter | content_type | Files processed by content type |

#### Queue/PubSub Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `filepoint_messages_published_total` | Counter | topic, event_type | Messages published to queue |
| `filepoint_messages_consumed_total` | Counter | topic, status | Messages consumed (status: success/error/retry) |
| `filepoint_message_processing_duration_seconds` | Histogram | topic, event_type | Message processing time |
| `filepoint_poison_queue_messages_total` | Counter | - | Messages sent to poison queue |

#### Cache Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `filepoint_cache_hits_total` | Counter | cache_type | Cache hits (cache_type: signed_url/prefixes/upload) |
| `filepoint_cache_misses_total` | Counter | cache_type | Cache misses |
| `filepoint_cache_errors_total` | Counter | cache_type, operation | Cache errors (operation: get/set/delete) |

#### Dependency Health

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `filepoint_dependency_up` | Gauge | service | Dependency status (service: s3/dynamodb/redis/kafka/sqs) - 1=up, 0=down |

#### AWS Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `filepoint_s3_operations_total` | Counter | operation, status | S3 operations (operation: get/put/delete/list) |
| `filepoint_s3_operation_duration_seconds` | Histogram | operation | S3 operation duration |
| `filepoint_dynamodb_operations_total` | Counter | operation, status | DynamoDB operations (operation: get/put/update/delete/query) |
| `filepoint_dynamodb_operation_duration_seconds` | Histogram | operation | DynamoDB operation duration |

---

## Health Checks

Filepoint provides multiple health check endpoints for different purposes.

### Endpoints

#### `GET /v1/health`
**Basic health check** - Returns 200 if the service is running.

**Response:**
```json
{
  "status": "up",
  "timestamp": "2025-01-20T10:30:00Z"
}
```

**Use case:** Simple availability check

---

#### `GET /v1/health/live`
**Liveness probe** - Kubernetes liveness check.

**Response:** `200 OK` (plain text)

**Use case:** Kubernetes liveness probe - restarts pod if it fails

**Kubernetes config:**
```yaml
livenessProbe:
  httpGet:
    path: /v1/health/live
    port: 9001
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

---

#### `GET /v1/health/ready`
**Readiness probe** - Checks all dependencies before serving traffic.

**Response (healthy):**
```json
{
  "status": "up",
  "timestamp": "2025-01-20T10:30:00Z",
  "dependencies": {
    "redis": {
      "status": "up",
      "latency": "2.5ms"
    },
    "s3": {
      "status": "up",
      "latency": "15ms"
    },
    "dynamodb": {
      "status": "up",
      "latency": "12ms"
    }
  }
}
```

**Response (degraded) - 503 Service Unavailable:**
```json
{
  "status": "degraded",
  "timestamp": "2025-01-20T10:30:00Z",
  "dependencies": {
    "redis": {
      "status": "down",
      "message": "connection refused",
      "latency": "100ms"
    },
    "s3": {
      "status": "up",
      "latency": "15ms"
    },
    "dynamodb": {
      "status": "up",
      "latency": "12ms"
    }
  }
}
```

**Use case:** Kubernetes readiness probe - prevents traffic to unhealthy pods

**Kubernetes config:**
```yaml
readinessProbe:
  httpGet:
    path: /v1/health/ready
    port: 9001
  initialDelaySeconds: 15
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

---

## Prometheus Setup

### Prometheus Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'filepoint'
    scrape_interval: 15s
    static_configs:
      - targets: ['filepoint:9001']
    metrics_path: '/v1/metrics'
    scheme: http
```

### Docker Compose Example

```yaml
services:
  filepoint:
    image: filepoint:latest
    ports:
      - "9001:9001"

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
```

### Kubernetes ServiceMonitor

For Prometheus Operator:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: filepoint
  labels:
    app: filepoint
spec:
  selector:
    matchLabels:
      app: filepoint
  endpoints:
    - port: http
      path: /v1/metrics
      interval: 15s
```

---

## Grafana Dashboards

### Key Metrics to Monitor

#### 1. Upload Success Rate

```promql
rate(filepoint_uploads_total{status="success"}[5m])
/
rate(filepoint_uploads_total[5m])
* 100
```

#### 2. Average Upload Processing Time

```promql
rate(filepoint_upload_processing_duration_seconds_sum[5m])
/
rate(filepoint_upload_processing_duration_seconds_count[5m])
```

#### 3. P95 Request Latency

```promql
histogram_quantile(0.95,
  rate(filepoint_http_request_duration_seconds_bucket[5m])
)
```

#### 4. Error Rate

```promql
sum(rate(filepoint_http_requests_total{status=~"5.."}[5m]))
/
sum(rate(filepoint_http_requests_total[5m]))
* 100
```

#### 5. Cache Hit Rate

```promql
sum(rate(filepoint_cache_hits_total[5m]))
/
(sum(rate(filepoint_cache_hits_total[5m])) + sum(rate(filepoint_cache_misses_total[5m])))
* 100
```

#### 6. Dependency Health

```promql
filepoint_dependency_up
```

#### 7. Queue Depth (if using Kafka)

```promql
sum(kafka_consumergroup_lag) by (consumergroup, topic)
```

#### 8. Poison Queue Messages

```promql
rate(filepoint_poison_queue_messages_total[5m])
```

### Sample Grafana Dashboard JSON

Create a dashboard with panels for:
- Request rate and latency
- Upload success/failure rates
- File processing time by type
- Cache hit rates
- Dependency health status
- AWS operation latencies
- Queue metrics

---

## Kubernetes Integration

### Complete Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: filepoint
spec:
  replicas: 3
  selector:
    matchLabels:
      app: filepoint
  template:
    metadata:
      labels:
        app: filepoint
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9001"
        prometheus.io/path: "/v1/metrics"
    spec:
      containers:
        - name: filepoint
          image: filepoint:latest
          ports:
            - containerPort: 9001
              name: http
          env:
            - name: ENVIRONMENT
              value: "production"
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /v1/health/live
              port: 9001
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /v1/health/ready
              port: 9001
            initialDelaySeconds: 15
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
---
apiVersion: v1
kind: Service
metadata:
  name: filepoint
  labels:
    app: filepoint
spec:
  selector:
    app: filepoint
  ports:
    - name: http
      port: 9001
      targetPort: 9001
```

---

## Alerts

### Prometheus Alert Rules

Create `filepoint-alerts.yml`:

```yaml
groups:
  - name: filepoint
    interval: 30s
    rules:
      # High error rate
      - alert: FilePointHighErrorRate
        expr: |
          sum(rate(filepoint_http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(filepoint_http_requests_total[5m]))
          > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate on Filepoint"
          description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"

      # High latency
      - alert: FilePointHighLatency
        expr: |
          histogram_quantile(0.95,
            rate(filepoint_http_request_duration_seconds_bucket[5m])
          ) > 2
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High latency on Filepoint"
          description: "P95 latency is {{ $value }}s (threshold: 2s)"

      # Dependency down
      - alert: FilePointDependencyDown
        expr: filepoint_dependency_up == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Filepoint dependency {{ $labels.service }} is down"
          description: "Service {{ $labels.service }} has been down for 2 minutes"

      # High upload failure rate
      - alert: FilePointUploadFailureRate
        expr: |
          sum(rate(filepoint_uploads_total{status="failed"}[5m]))
          /
          sum(rate(filepoint_uploads_total[5m]))
          > 0.10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High upload failure rate"
          description: "Upload failure rate is {{ $value | humanizePercentage }} (threshold: 10%)"

      # Poison queue growing
      - alert: FilePointPoisonQueueGrowing
        expr: rate(filepoint_poison_queue_messages_total[5m]) > 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Messages being sent to poison queue"
          description: "{{ $value }} messages/sec are failing after max retries"

      # Low cache hit rate
      - alert: FilePointLowCacheHitRate
        expr: |
          sum(rate(filepoint_cache_hits_total[5m]))
          /
          (sum(rate(filepoint_cache_hits_total[5m])) + sum(rate(filepoint_cache_misses_total[5m])))
          < 0.50
        for: 15m
        labels:
          severity: info
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }} (threshold: 50%)"

      # S3 high error rate
      - alert: FilePointS3ErrorRate
        expr: |
          sum(rate(filepoint_s3_operations_total{status="error"}[5m]))
          /
          sum(rate(filepoint_s3_operations_total[5m]))
          > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High S3 error rate"
          description: "S3 error rate is {{ $value | humanizePercentage }}"

      # DynamoDB throttling
      - alert: FilePointDynamoDBThrottling
        expr: |
          sum(rate(filepoint_dynamodb_operations_total{status="throttled"}[5m])) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "DynamoDB throttling detected"
          description: "DynamoDB operations are being throttled"
```

---

## Best Practices

### 1. Metric Cardinality
- Avoid high-cardinality labels (e.g., user IDs, request IDs)
- Use predefined label values
- Aggregate where possible

### 2. Health Check Timeouts
- Keep health check timeouts short (< 5s)
- Use separate liveness/readiness probes
- Don't check deep dependencies in liveness

### 3. Alert Fatigue
- Set appropriate thresholds
- Use `for:` clause to avoid flapping
- Group related alerts
- Set severity levels correctly

### 4. Dashboard Organization
- Create service-level overview dashboard
- Separate operational metrics from business metrics
- Use consistent naming conventions

### 5. Retention
- Configure appropriate retention for your metrics
- Use downsampling for long-term storage
- Archive important metrics to cold storage

---

## Troubleshooting

### Metrics not appearing

1. Check `/v1/metrics` endpoint is accessible:
   ```bash
   curl http://localhost:9001/v1/metrics
   ```

2. Verify Prometheus is scraping:
   ```bash
   # Check Prometheus targets
   http://localhost:9090/targets
   ```

3. Check for errors in Prometheus logs

### Health checks failing

1. Check individual dependency:
   ```bash
   curl http://localhost:9001/v1/health/ready | jq
   ```

2. Review application logs for connection errors

3. Verify network connectivity to dependencies

### High cardinality warnings

If you see warnings about high cardinality:
1. Review metric labels
2. Reduce dynamic label values
3. Consider using aggregation

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [Kubernetes Monitoring](https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/)
- [OpenTelemetry](https://opentelemetry.io/) (for future distributed tracing)

