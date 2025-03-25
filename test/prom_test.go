package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/efficientgo/core/testutil"
	"github.com/efficientgo/e2e"
	e2einteractive "github.com/efficientgo/e2e/interactive"
	e2emon "github.com/efficientgo/e2e/monitoring"
)

const (
	myAppImage = "quay.io/bwplotka/my-app:latest"

	// Prometheus built from "rename-kubecon" branch.
	promImage = "quay.io/bwplotka/prometheus:semconv-v1"
)

// Requires make docker DOCKER_TAG=latest before starting.
func TestMyApp_PrometheusWriting(t *testing.T) {
	e, err := e2e.New()
	t.Cleanup(e.Close)
	testutil.Ok(t, err)

	schemaVersions := [2]string{
		"generated@v1.0.0",
		"generated@v1.1.0",
	}

	// Create my-app-new containers. One creating metrics from , second from
	myApp := newMyApp(e, "my-app-v1.0.0-metrics", myAppImage, map[string]string{"-metric-source": schemaVersions[0]})
	myApp2 := newMyApp(e, "my-app-v1.1.0-metrics", myAppImage, map[string]string{"-metric-source": schemaVersions[1]})
	testutil.Ok(t, e2e.StartAndWaitReady(myApp, myApp2))

	// Create a go routine that switches runnables under a single name for different versions.
	switchInterval := 5 * time.Minute
	activeSchemaVersion := 0
	myAppSwitching := newMyApp(e, "my-app", myAppImage, map[string]string{"-metric-source": schemaVersions[activeSchemaVersion]})
	testutil.Ok(t, e2e.StartAndWaitReady(myAppSwitching))
	{
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.TODO())
		e.AddCloser(func() {
			cancel()
			wg.Wait()
		})

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					// Env will close the app.
					return
				case <-time.After(switchInterval):
					testutil.Ok(t, myAppSwitching.Stop())

					activeSchemaVersion = (activeSchemaVersion + 1) % 2

					// Must be same name for InternalEndpoint to stay consistent!
					myAppSwitching = newMyApp(e, "my-app", myAppImage, map[string]string{"-metric-source": schemaVersions[activeSchemaVersion]})
					testutil.Ok(t, e2e.StartAndWaitReady(myAppSwitching))
				}
			}
		}()
	}

	prom := newPrometheus(e, "prom-1", promImage, []string{
		myApp.InternalEndpoint("http"),
		myApp2.InternalEndpoint("http"),
		myAppSwitching.InternalEndpoint("http"),
	}, nil)
	testutil.Ok(t, e2e.StartAndWaitReady(prom))

	testutil.Ok(t, e2einteractive.OpenInBrowser("http://"+prom.Endpoint("http")))
	testutil.Ok(t, e2einteractive.RunUntilEndpointHit())
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
- job_name: 'my-app'
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
