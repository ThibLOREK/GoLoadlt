package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

// Prometheus metrics
var (
	RunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "etl_runs_total",
		Help: "Total ETL runs by pipeline and status",
	}, []string{"pipeline_id", "status"})

	RunDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "etl_run_duration_seconds",
		Help:    "ETL run duration in seconds",
		Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 120, 300},
	}, []string{"pipeline_id"})

	RecordsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "etl_records_processed_total",
		Help: "Total records processed by pipeline",
	}, []string{"pipeline_id", "stage"})

	ActiveRuns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "etl_active_runs",
		Help: "Number of currently running pipelines",
	})
)

type Provider struct {
	tracerProvider *sdktrace.TracerProvider
}

func Init(ctx context.Context, serviceName string) (*Provider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("telemetry: init exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: init resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	Tracer = otel.Tracer(serviceName)

	return &Provider{tracerProvider: tp}, nil
}

func (p *Provider) Shutdown(ctx context.Context) error {
	return p.tracerProvider.Shutdown(ctx)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func RecordRun(pipelineID, status string, duration time.Duration, read, loaded int64) {
	RunsTotal.WithLabelValues(pipelineID, status).Inc()
	RunDuration.WithLabelValues(pipelineID).Observe(duration.Seconds())
	RecordsProcessed.WithLabelValues(pipelineID, "read").Add(float64(read))
	RecordsProcessed.WithLabelValues(pipelineID, "loaded").Add(float64(loaded))
}
