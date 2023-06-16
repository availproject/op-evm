package server

import (
	"fmt"
	"os"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

// setupTelemetry initializes the metrics system for the server.
// It configures an in-memory metrics sink and a Prometheus metrics sink.
func (s *Server) setupTelemetry() error {
	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(inm)

	promSink, err := prometheus.NewPrometheusSinkFrom(prometheus.PrometheusOpts{
		Name:       "edge_prometheus_sink",
		Expiration: 0,
	})
	if err != nil {
		return err
	}

	metricsConf := metrics.DefaultConfig("edge")
	metricsConf.EnableHostname = false
	_, err = metrics.NewGlobal(metricsConf, metrics.FanoutSink{
		inm, promSink,
	})

	return err
}

// enableDataDogProfiler enables DataDog profiler. It is enabled by setting
// DD_ENABLE environment variable. Additional parameters can be set
// with environment variables prefixed with "DD_". See https://docs.datadoghq.com/profiler/enabling/go/
func (s *Server) enableDataDogProfiler() error {
	if os.Getenv("DD_PROFILING_ENABLED") == "" {
		s.logger.Debug("DataDog profiler disabled, set DD_PROFILING_ENABLED env var to enable it.")

		return nil
	}
	// For containerized solutions, we want to be able to set the ip and port that the agent will bind to
	// by defining DD_AGENT_HOST and DD_TRACE_AGENT_PORT env vars.
	// If these env vars are not defined, the agent will bind to default ip:port ( localhost:8126 )
	ddIP := "localhost"
	ddPort := "8126"

	if os.Getenv("DD_AGENT_HOST") != "" {
		ddIP = os.Getenv("DD_AGENT_HOST")
	}

	if os.Getenv("DD_TRACE_AGENT_PORT") != "" {
		ddPort = os.Getenv("DD_TRACE_AGENT_PORT")
	}

	if err := profiler.Start(
		// enable all profiles
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
			profiler.BlockProfile,
			profiler.MutexProfile,
			profiler.GoroutineProfile,
			profiler.MetricsProfile,
		),
		profiler.WithAgentAddr(ddIP+":"+ddPort),
	); err != nil {
		return fmt.Errorf("could not start datadog profiler: %w", err)
	}

	// start the tracer
	tracer.Start()
	s.logger.Info("DataDog profiler started")

	return nil
}

// closeDataDogProfiler stops the DataDog profiler and tracer when they are no longer needed.
func (s *Server) closeDataDogProfiler() {
	s.logger.Debug("closing DataDog profiler")
	profiler.Stop()

	s.logger.Debug("closing DataDog tracer")
	tracer.Stop()
}
