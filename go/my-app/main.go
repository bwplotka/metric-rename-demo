package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"syscall"

	semconv "github.com/bwplotka/metric-rename-demo/go/my-app/my-app/semconv/v0.1.0"
	"github.com/nelkinda/health-go"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addrFlag := flag.String("listen-address", ":9011", "Address to listen on. Available HTTP paths: /metrics")
	metricDefinition := flag.String("metric-source", "manual", "Metric definition source to use ['manual', 'generated@v0.1.0'")
	flag.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	switch *metricDefinition {
	case "manual":
		m := mustNewCustomStableMetric(reg)
		m.WithLabelValues("101", "a", "1.22314").Inc()
	case "generated@v0.1.0":
		m := semconv.MustNewCustomElementsCounterVec(reg)
		m.WithLabelValues("101", string(semconv.ACustomElementsCategory), "1.22314").Inc() // TODO(bwplotka): Make it more type safe.
	default:
		log.Fatalf("unknown -metric-source source, got %v", *metricDefinition)
	}

	var g run.Group
	{
		healthHandler := health.New(health.Health{}).Handler
		httpSrv := &http.Server{Addr: *addrFlag}
		http.HandleFunc("/-/health", healthHandler)
		http.HandleFunc("/-/ready", healthHandler)
		http.HandleFunc("/metrics", instrument(reg, "/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		})))
		g.Add(func() error {
			slog.Info("starting HTTP server", "address", *addrFlag)
			return httpSrv.ListenAndServe()
		}, func(_ error) {
			_ = httpSrv.Shutdown(context.Background())
		})
	}
	g.Add(run.SignalHandler(context.Background(), os.Interrupt, syscall.SIGTERM))

	slog.Info("my-app starting...")
	if err := g.Run(); err != nil {
		slog.Error("running my-app failed", "err", err)
		os.Exit(1)
	}
	slog.Info("my-app finished")
}

func instrument(reg prometheus.Registerer, handlerName string, handler http.Handler) http.HandlerFunc {
	reg = prometheus.WrapRegistererWith(prometheus.Labels{"handler": handlerName}, reg)

	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Tracks the latencies for HTTP requests.",

			NativeHistogramBucketFactor: 1.1,
		},
		[]string{"method", "code"},
	)
	requestSize := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",

			// Custom buckets, so key metric is visible in the text format (for testing and local debugging).
			Buckets: []float64{0, 200, 1024, 2048, 10240},

			NativeHistogramBucketFactor: 1.1,
		},
		[]string{"method", "code"},
	)
	requestsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)
	responseSize := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",

			NativeHistogramBucketFactor: 1.1,
		},
		[]string{"method", "code"},
	)

	base := promhttp.InstrumentHandlerRequestSize(
		requestSize,
		promhttp.InstrumentHandlerCounter(
			requestsTotal,
			promhttp.InstrumentHandlerResponseSize(
				responseSize,
				promhttp.InstrumentHandlerDuration(
					requestDuration,
					http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
						handler.ServeHTTP(writer, r)
					}),
				),
			),
		),
	)
	return base.ServeHTTP
}
