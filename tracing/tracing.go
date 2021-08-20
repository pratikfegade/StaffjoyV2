// healthcheck is a library that provides a basic health check handler for Staffjoy applications.
// We generally host this endpoint at "/health" on port 80
//
// Usage:
// r.HandleFunc(healthcheck.HEALTHPATH healthcheck.Handler)

package tracing

import (
	"io"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics"


	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func InitTracer(serviceName string) (opentracing.Tracer, io.Closer) {
	if !opentracing.IsGlobalTracerRegistered() {
		// Sample configuration for testing. Use constant sampling to sample every trace
		// and enable LogSpan to log every span via configured Logger.
		cfg := jaegercfg.Configuration{
			ServiceName: serviceName,
			Sampler:     &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
			Reporter:    &jaegercfg.ReporterConfig{
				LogSpans: true,
			},
		}

		// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
		// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
		// frameworks.
		jLogger := jaegerlog.StdLogger
		jMetricsFactory := metrics.NullFactory

		// Initialize tracer with a logger and a metrics factory
		tracer, closer, _ := cfg.NewTracer(
			jaegercfg.Logger(jLogger),
			jaegercfg.Metrics(jMetricsFactory),
		)
		// Set the singleton opentracing.Tracer with the Jaeger tracer.
		opentracing.SetGlobalTracer(tracer)
		return tracer, closer
	} else {
		return opentracing.GlobalTracer(), nil
	}
}

func GetTracer() (opentracing.Tracer) {
	if opentracing.IsGlobalTracerRegistered() {
		return opentracing.GlobalTracer()
	} else {
		return nil
	}
}
