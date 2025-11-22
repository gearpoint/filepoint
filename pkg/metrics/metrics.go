// Package metrics provides Prometheus metrics collection for the Filepoint service.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics collectors
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Upload metrics
	UploadsTotal          *prometheus.CounterVec
	UploadSize            *prometheus.HistogramVec
	UploadProcessingTime  *prometheus.HistogramVec
	UploadErrors          *prometheus.CounterVec
	TempFileUploadsTotal  prometheus.Counter
	FinalFileUploadsTotal prometheus.Counter

	// File type metrics
	FileTypeCounter *prometheus.CounterVec

	// Queue/PubSub metrics
	MessagesPublished *prometheus.CounterVec
	MessagesConsumed  *prometheus.CounterVec
	MessageProcessingTime *prometheus.HistogramVec
	PoisonQueueTotal  prometheus.Counter

	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec
	CacheErrors *prometheus.CounterVec

	// Dependency health
	DependencyUp *prometheus.GaugeVec

	// AWS metrics
	S3OperationsTotal      *prometheus.CounterVec
	S3OperationDuration    *prometheus.HistogramVec
	DynamoDBOperationsTotal *prometheus.CounterVec
	DynamoDBOperationDuration *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency in seconds",
				Buckets:   prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to 10GB
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),

		// Upload metrics
		UploadsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "uploads_total",
				Help:      "Total number of file uploads",
			},
			[]string{"file_type", "status"}, // status: success, failed, rejected
		),
		UploadSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "upload_size_bytes",
				Help:      "Upload file size in bytes",
				Buckets:   []float64{1024, 10240, 102400, 1048576, 10485760, 104857600, 1073741824}, // 1KB to 1GB
			},
			[]string{"file_type"},
		),
		UploadProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "upload_processing_duration_seconds",
				Help:      "Time taken to process uploaded files",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120}, // Up to 2 minutes
			},
			[]string{"file_type", "definition"}, // definition: low-def, medium-def, high-def
		),
		UploadErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "upload_errors_total",
				Help:      "Total number of upload errors",
			},
			[]string{"error_type"}, // validation, storage, processing, webhook
		),
		TempFileUploadsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "temp_file_uploads_total",
				Help:      "Total number of temporary file uploads to S3",
			},
		),
		FinalFileUploadsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "final_file_uploads_total",
				Help:      "Total number of final processed file uploads to S3",
			},
		),

		// File type metrics
		FileTypeCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "file_type_total",
				Help:      "Counter for each file type processed",
			},
			[]string{"content_type"},
		),

		// Queue/PubSub metrics
		MessagesPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_published_total",
				Help:      "Total messages published to queue",
			},
			[]string{"topic", "event_type"},
		),
		MessagesConsumed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_consumed_total",
				Help:      "Total messages consumed from queue",
			},
			[]string{"topic", "status"}, // status: success, error, retry
		),
		MessageProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "message_processing_duration_seconds",
				Help:      "Time taken to process messages from queue",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300}, // Up to 5 minutes
			},
			[]string{"topic", "event_type"},
		),
		PoisonQueueTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "poison_queue_messages_total",
				Help:      "Total messages sent to poison queue after max retries",
			},
		),

		// Cache metrics
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total cache hits",
			},
			[]string{"cache_type"}, // signed_url, prefixes, upload
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total cache misses",
			},
			[]string{"cache_type"},
		),
		CacheErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_errors_total",
				Help:      "Total cache errors",
			},
			[]string{"cache_type", "operation"}, // operation: get, set, delete
		),

		// Dependency health (1 = up, 0 = down)
		DependencyUp: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "dependency_up",
				Help:      "Dependency health status (1 = up, 0 = down)",
			},
			[]string{"service"}, // s3, dynamodb, redis, kafka, sqs
		),

		// AWS metrics
		S3OperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "s3_operations_total",
				Help:      "Total S3 operations",
			},
			[]string{"operation", "status"}, // operation: get, put, delete, list
		),
		S3OperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "s3_operation_duration_seconds",
				Help:      "S3 operation duration in seconds",
				Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"operation"},
		),
		DynamoDBOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "dynamodb_operations_total",
				Help:      "Total DynamoDB operations",
			},
			[]string{"operation", "status"}, // operation: get, put, update, delete, query
		),
		DynamoDBOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "dynamodb_operation_duration_seconds",
				Help:      "DynamoDB operation duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
	}
}

// Global metrics instance
var globalMetrics *Metrics

// Init initializes the global metrics instance
func Init(namespace string) {
	globalMetrics = NewMetrics(namespace)
}

// Get returns the global metrics instance
func Get() *Metrics {
	if globalMetrics == nil {
		Init("filepoint")
	}
	return globalMetrics
}
