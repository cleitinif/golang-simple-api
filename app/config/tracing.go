package config

import (
	"context"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CustomMetrics struct {
	TotalRequests prometheus.Counter
	ErrorRequests prometheus.Counter
}

var customMetrics *CustomMetrics

// Helper function to define sampling.
// When in development mode, AlwaysSample is defined,
// otherwise, sample based on Parent and IDRatio will be used.
func getSampler() trace.Sampler {
	ENV := os.Getenv("GO_ENV")
	switch ENV {
	case "development":
		return trace.AlwaysSample()
	case "production":
		return trace.ParentBased(trace.TraceIDRatioBased(0.5))
	default:
		return trace.AlwaysSample()
	}
}

func newResource(ctx context.Context) *resource.Resource {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(os.Getenv("SERVICE_NAME")),
			attribute.String("environment", os.Getenv("GO_ENV")),
		),
	)
	if err != nil {
		log.Fatalf("%s: %v", "Failed to create resource", err)
	}
	return res
}

func exporterToJaeger(ctx context.Context) (*otlptrace.Exporter, error) {
	return otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
}

func InitProviderWithJaegerExporter(ctx context.Context) func(context.Context) error {
	if os.Getenv("OPEN_TELEMETRY_ENABLED") != "true" {
		return func(context.Context) error { return nil }
	}

	exp, err := exporterToJaeger(ctx)
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(getSampler()),
		trace.WithBatcher(exp),
		trace.WithResource(newResource(ctx)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

func InitMetricsExporter(ctx context.Context) func(context.Context) error {
	if os.Getenv("OPEN_TELEMETRY_ENABLED") != "true" {
		return func(context.Context) error { return nil }
	}

	reg := prometheus.NewRegistry()

	customMetrics := GetCustomMetrics()

	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
	))
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(customMetrics.TotalRequests)
	reg.MustRegister(customMetrics.ErrorRequests)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
	return func(context.Context) error { return nil }
}

func GetCustomMetrics() *CustomMetrics {
	if customMetrics == nil {
		customMetrics = &CustomMetrics{
			TotalRequests: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "The total number of HTTP requests",
			}),
			ErrorRequests: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "http_requests_error",
				Help: "The total number of HTTP requests",
			}),
		}
	}

	return customMetrics
}
