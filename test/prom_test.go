package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/efficientgo/core/testutil"
	"github.com/efficientgo/e2e"
	e2emon "github.com/efficientgo/e2e/monitoring"
)

const (
	myAppImage = "quay.io/bwplotka/my-app:latest"
)

// Requires make docker DOCKER_TAG=latest before starting.
func TestMyApp_PrometheusWriting(t *testing.T) {
	e, err := e2e.New()
	t.Cleanup(e.Close)
	testutil.Ok(t, err)

	// TODO Create my-app, create my-app-new containers (with new metric).
	myApp := newMyApp(e, "my-app", myAppImage, nil)
	myApp2 := newMyApp(e, "my-app", myAppImage, nil)
	// Create self-scraping Prometheus writing two streams of PRW writes to sink-1: v1 and v2.
	prom := newPrometheus(e, "prom-1", "quay.io/prometheus/prometheus:main", []string{myApp.InternalEndpoint("http"), myApp2.InternalEndpoint("http")}, nil)
	testutil.Ok(t, e2e.StartAndWaitReady(myApp, prom))

	//const expectSamples float64 = 2e3
	//
	//testutil.Ok(t, prom.WaitSumMetricsWithOptions(
	//	e2emon.Greater(expectSamples), []string{"prometheus_remote_storage_samples_total"},
	//	e2emon.WithLabelMatchers(&matchers.Matcher{Name: "remote_name", Value: "v2-to-sink", Type: matchers.MatchEqual}),
	//	e2emon.WithWaitBackoff(&backoff.Config{Min: 1 * time.Second, Max: 1 * time.Second, MaxRetries: 300}), // Wait 5m max.
	//))
	//testutil.Ok(t, prom.WaitSumMetricsWithOptions(
	//	e2emon.Greater(expectSamples), []string{"prometheus_remote_storage_samples_total"},
	//	e2emon.WithLabelMatchers(&matchers.Matcher{Name: "remote_name", Value: "v1-to-sink", Type: matchers.MatchEqual}),
	//))

	// Uncomment for the interactive run (test will run until you kill the test or hit endpoint that was logged)
	// so you can explore Prometheus UI and sink metrics.
	//testutil.Ok(t, e2einteractive.OpenInBrowser("http://"+prom.Endpoint("http"))) // Open Prometheus UI
	//testutil.Ok(t, e2einteractive.OpenInBrowser("http://"+sink.Endpoint("http")+"/metrics"))
	//testutil.Ok(t, e2einteractive.RunUntilEndpointHit())

}

func newMyApp(e e2e.Environment, name, image string, flagOverride map[string]string) *e2emon.InstrumentedRunnable {
	ports := map[string]int{"http": 9011}

	f := e.Runnable(name).WithPorts(ports).Future()
	args := map[string]string{
		"-listen-address": fmt.Sprintf(":%d", ports["http"]),
	}
	if flagOverride != nil {
		args = e2e.MergeFlagsWithoutRemovingEmpty(args, flagOverride)
	}

	return e2emon.AsInstrumented(f.Init(e2e.StartOptions{
		Image:     image,
		Command:   e2e.NewCommandWithoutEntrypoint("my-app", e2e.BuildArgs(args)...),
		Readiness: e2e.NewHTTPReadinessProbe("http", "/-/ready", 200, 200),
		User:      strconv.Itoa(os.Getuid()),
	}), "http")
}

func newPrometheus(env e2e.Environment, name, image string, scrapeAddrs []string, flagOverride map[string]string) *e2emon.Prometheus {
	ports := map[string]int{"http": 9090}

	f := env.Runnable(name).WithPorts(ports).Future()
	config := fmt.Sprintf(`
global:
  external_labels:
    prometheus: %v
scrape_configs:
- job_name: 'self'
  scrape_interval: 5s
  scrape_timeout: 5s
  static_configs:
  - targets: [%v]
`, name, ports["http"], strings.Join(scrapeAddrs, ","))
	if err := os.WriteFile(filepath.Join(f.Dir(), "prometheus.yml"), []byte(config), 0o600); err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("create prometheus config failed: %w", err))}
	}

	args := map[string]string{
		"--web.listen-address":                  fmt.Sprintf(":%d", ports["http"]),
		"--config.file":                         filepath.Join(f.Dir(), "prometheus.yml"),
		"--storage.tsdb.path":                   f.Dir(),
		"--enable-feature=exemplar-storage":     "",
		"--enable-feature=native-histograms":    "",
		"--enable-feature=metadata-wal-records": "",
		"--storage.tsdb.no-lockfile":            "",
		"--storage.tsdb.retention.time":         "1d",
		"--storage.tsdb.wal-compression":        "",
		"--storage.tsdb.min-block-duration":     "2h",
		"--storage.tsdb.max-block-duration":     "2h",
		"--web.enable-lifecycle":                "",
		"--log.format":                          "json",
		"--log.level":                           "info",
	}
	if flagOverride != nil {
		args = e2e.MergeFlagsWithoutRemovingEmpty(args, flagOverride)
	}

	p := e2emon.AsInstrumented(f.Init(e2e.StartOptions{
		Image:     image,
		Command:   e2e.NewCommandWithoutEntrypoint("prometheus", e2e.BuildArgs(args)...),
		Readiness: e2e.NewHTTPReadinessProbe("http", "/-/ready", 200, 200),
		User:      strconv.Itoa(os.Getuid()),
	}), "http")

	return &e2emon.Prometheus{
		Runnable:     p,
		Instrumented: p,
	}
}
