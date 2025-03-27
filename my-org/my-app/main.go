package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/bwplotka/metric-rename-demo/my-org/my-app/semconv.gen/1.0.0/my_app_custom_elements"
	"github.com/bwplotka/metric-rename-demo/my-org/my-app/semconv.gen/1.1.0/my_app_custom_elements/2"
	"github.com/bwplotka/metric-rename-demo/my-org/my-app/semconv.gen/my_app_custom_elements/3"

	"github.com/bwplotka/metric-rename-demo/my-org/my-app/semconv.gen/1.0.0/my_app_latency"
	"github.com/bwplotka/metric-rename-demo/my-org/my-app/semconv.gen/my_app_latency/2"

	"github.com/nelkinda/health-go"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	// Hack for demo purposes. Long term schema and unit should be part of the API.
	prometheus.AllowReservedLabels = true
}

func main() {
	addrFlag := flag.String("listen-address", ":9011", "Address to listen on. Available HTTP paths: /metrics")
	metricDefinition := flag.String("metric-source", "manual", "Metric definition source to use ['manual', 'generated@v1.0.0', 'generated@v1.1.0', 'generated@v1.2.0'")
	flag.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	var (
		elementsCount prometheus.Counter
		latency       prometheus.Observer
	)

	switch *metricDefinition {
	case "manual":
		elementsCount = mustNewCustomElements(reg).
			WithLabelValues("100", "first", "1.2414")
		latency = mustNewLatency(reg).
			WithLabelValues("200")

		// Notice the type safety of generated code below vs manual above.
	case "generated@v1.0.0":
		elementsCount = my_app_custom_elements.MustNewMyAppCustomElementsTotal(reg).
			WithLabelValues(100, my_app_custom_elements.FirstCategory, 1.2414)
		latency = my_app_latency.MustNewMyAppLatencyMilliseconds(reg, []float64{1000, 10000, 100000}).
			WithLabelValues(200)
	case "generated@v1.1.0":
		elementsCount = my_app_custom_elements_2.MustNewMyAppCustomChangedElementsTotal(reg).
			WithLabelValues(100, my_app_custom_elements_2.FirstClass, 1.2414)
		latency = my_app_latency_2.MustNewMyAppLatencySeconds(reg, []float64{1, 10, 100}). // Buckets has to scaled too.
			WithLabelValues(200)
	case "generated@v1.2.0":
		elementsCount = my_app_custom_elements_3.MustNewMyAppCustomChangedElementsTotal(reg).
			WithLabelValues(100, my_app_custom_elements_3.FirstClass, 1.2414)
		latency = my_app_latency_2.MustNewMyAppLatencySeconds(reg, []float64{1, 10, 100}). // Buckets has to scaled too.
			WithLabelValues(200)
	default:
		log.Fatalf("unknown -metric-source source, got %v", *metricDefinition)
	}

	var g run.Group
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {

			for {
				const interval = 10 * time.Second
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(interval):
					elementsCount.Inc()
					if *metricDefinition == "manual" || *metricDefinition == "generated@v1.0.0" {
						latency.Observe(float64(interval.Milliseconds()))
					} else {
						// From v1.1.0 metric reports base units
						latency.Observe(float64(interval.Seconds()))
					}
				}
			}
		}, func(err error) {
			cancel()
		})
	}
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
