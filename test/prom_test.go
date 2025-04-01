package test

import (
	"context"
	"fmt"
	"net/url"
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
	promImage = "quay.io/bwplotka/prometheus:semconv-v1.7"
)

// Requires make docker DOCKER_TAG=latest before starting.
func TestMyApp_PrometheusWriting(t *testing.T) {
	e, err := e2e.New()
	t.Cleanup(e.Close)
	testutil.Ok(t, err)

	schemaVersions := [2]string{
		"generated@v1.0.0",
		"generated@v1.2.0",
	}

	// Create my-app-new containers. One creating metrics from , second from
	myApp := newMyApp(e, "my-app-v1.0.0-pinned", myAppImage, map[string]string{"-metric-source": schemaVersions[0]})
	myApp2 := newMyApp(e, "my-app-v1.2.0-pinned", myAppImage, map[string]string{"-metric-source": schemaVersions[1]})
	testutil.Ok(t, e2e.StartAndWaitReady(myApp, myApp2))

	// Create a go routine that switches runnables under a single name for different versions.
	switchInterval := 10 * time.Minute
	activeSchemaVersion := 0
	myAppSwitchingFuture := newMyAppFuture(e, "my-app")
	myAppSwitching := newMyAppFromFuture(myAppSwitchingFuture, myAppImage, map[string]string{"-metric-source": schemaVersions[activeSchemaVersion]})

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
					myAppSwitching = newMyAppFromFuture(myAppSwitchingFuture, myAppImage, map[string]string{"-metric-source": schemaVersions[activeSchemaVersion]})
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

	testutil.Ok(t, e2einteractive.OpenInBrowser("http://"+prom.Endpoint("http")+promURL1))
	testutil.Ok(t, e2einteractive.OpenInBrowser("http://"+prom.Endpoint("http")+promURL2))
	testutil.Ok(t, e2einteractive.OpenInBrowser("http://"+prom.Endpoint("http")+promURL3))
	testutil.Ok(t, e2einteractive.RunUntilEndpointHit())
}

/*
https://www.bwplotka.dev/semconv/v1.0.0/my-app.yaml
*/

var promURL1 = func() string { ret, _ := url.QueryUnescape(`/query?g0.expr=my_app_latency_seconds_sum%7B__schema_url__%3D"https%3A%2F%2Fbwplotka.dev%2Fsemconv%2F1.1.0"%7D&g0.show_tree=0&g0.tab=table&g0.range_input=1h&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0&g1.expr=my_app_latency_seconds_sum%7B%7D&g1.show_tree=0&g1.tab=table&g1.range_input=1h&g1.res_type=auto&g1.res_density=medium&g1.display_mode=lines&g1.show_exemplars=0&g2.expr=my_app_latency_milliseconds_sum%7B%7D&g2.show_tree=0&g2.tab=table&g2.range_input=1h&g2.res_type=auto&g2.res_density=medium&g2.display_mode=lines&g2.show_exemplars=0&g3.expr=histogram_quantile%28%0A++0.9%2C%0A++sum+by+%28le%2C+instance%2C+code%29+%28%0A++++rate%28%0A++++++my_app_latency_seconds_bucket%7B__schema_url__%3D"https%3A%2F%2Fbwplotka.dev%2Fsemconv%2F1.1.0"%7D%5B1m%5D%0A++++%29%0A++%29%0A%29&g3.show_tree=0&g3.tab=table&g3.range_input=1h&g3.res_type=auto&g3.res_density=medium&g3.display_mode=lines&g3.show_exemplars=0`); return ret }()

var promURL2 = func() string { ret, _ := url.QueryUnescape(`/query?g0.expr=my_app_custom_changed_elements_total%7B__schema_url__%3D"https%3A%2F%2Fbwplotka.dev%2Fsemconv%2F1.1.0"%2C+number%3D"100"%7D&g0.show_tree=0&g0.tab=table&g0.range_input=1h&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0&g1.expr=my_app_custom_changed_elements_total%7B%7D&g1.show_tree=0&g1.tab=table&g1.range_input=1h&g1.res_type=auto&g1.res_density=medium&g1.display_mode=lines&g1.show_exemplars=0&g2.expr=my_app_custom_elements_total%7B%7D&g2.show_tree=0&g2.tab=table&g2.range_input=1h&g2.res_type=auto&g2.res_density=medium&g2.display_mode=lines&g2.show_exemplars=0&g3.expr=my_app_custom_changed_elements_total%7B__schema_url__%3D"https%3A%2F%2Fbwplotka.dev%2Fsemconv%2F1.1.0"%2C+number%3D"100"%7D&g3.show_tree=0&g3.tab=table&g3.range_input=1h&g3.res_type=auto&g3.res_density=medium&g3.display_mode=lines&g3.show_exemplars=0&g4.expr=my_app_custom_elements_total%7B__schema_url__%3D"https%3A%2F%2Fbwplotka.dev%2Fsemconv%2F1.0.0"%2C+integer%3D"100"%7D&g4.show_tree=0&g4.tab=table&g4.range_input=1h&g4.res_type=auto&g4.res_density=medium&g4.display_mode=lines&g4.show_exemplars=0`); return ret }()

var promURL3 = func() string { ret, _ := url.QueryUnescape(`/query?g0.expr=rate%28my_app_custom_changed_elements_total%7B__schema_url__%3D"https%3A%2F%2Fbwplotka.dev%2Fsemconv%2F1.2.0"%2C+my_number%3D"100"%2C+instance%21~".*-pinned%3A9011"%7D%5B1m%5D%29&g0.show_tree=0&g0.tab=graph&g0.range_input=5m&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0&g1.expr=rate%28my_app_custom_changed_elements_total%7Bmy_number%3D"100"%2C+instance%21~".*-pinned%3A9011"%7D%5B1m%5D%29&g1.show_tree=0&g1.tab=graph&g1.range_input=5m&g1.res_type=auto&g1.res_density=medium&g1.display_mode=lines&g1.show_exemplars=0&g2.expr=rate%28my_app_custom_elements_total%7Binteger%3D"100"%2C+instance%21~".*-pinned%3A9011"%7D%5B1m%5D%29&g2.show_tree=0&g2.tab=graph&g2.range_input=5m&g2.res_type=auto&g2.res_density=medium&g2.display_mode=lines&g2.show_exemplars=0`); return ret }()

func newMyApp(e e2e.Environment, name, image string, flagOverride map[string]string) *e2emon.InstrumentedRunnable {
	return newMyAppFromFuture(newMyAppFuture(e, name), image, flagOverride)
}

func newMyAppFuture(e e2e.Environment, name string) e2e.FutureRunnable {
	return e.Runnable(name).WithPorts(map[string]int{"http": 9011}).Future()
}

func newMyAppFromFuture(f e2e.FutureRunnable, image string, flagOverride map[string]string) *e2emon.InstrumentedRunnable {
	args := map[string]string{
		"-listen-address": f.InternalEndpoint("http"),
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

	if err := os.MkdirAll(filepath.Join(f.Dir(), "semconv"), os.ModePerm); err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("create local semconv dir: %w", err))}
	}
	chBytes, err := os.ReadFile("../my-org/semconv/changelog.yaml")
	if err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("read changelog file: %w", err))}
	}
	if err := os.WriteFile(filepath.Join(f.Dir(), "semconv/changlog.yml"), chBytes, 0o600); err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("create local changelog file failed: %w", err))}
	}
	idsBytes, err := os.ReadFile("../my-org/semconv/ids.yaml")
	if err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("read ids file: %w", err))}
	}
	if err := os.WriteFile(filepath.Join(f.Dir(), "semconv/ids.yml"), idsBytes, 0o600); err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("create local ids file failed: %w", err))}
	}

	config := fmt.Sprintf(`
global:
  external_labels:
    prometheus: %v
semconv:
  schema_overrides:
    https://bwplotka.dev/semconv: %v/semconv
  
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
`, name, f.Dir(), f.InternalEndpoint("http"), strings.Join(scrapeAddrs, ","))
	if err := os.WriteFile(filepath.Join(f.Dir(), "prometheus.yml"), []byte(config), 0o600); err != nil {
		return &e2emon.Prometheus{Runnable: e2e.NewFailedRunnable(name, fmt.Errorf("create prometheus config failed: %w", err))}
	}

	args := map[string]string{
		"--web.listen-address":                  fmt.Sprintf(":%d", ports["http"]),
		"--config.file":                         filepath.Join(f.Dir(), "prometheus.yml"),
		"--storage.tsdb.path":                   f.Dir(),
		"--enable-feature=native-histograms":    "",
		"--enable-feature=type-and-unit-labels": "",
		//"--enable-feature=promql-delayed-name-removal": "",
		"--enable-feature=semconv-versioned-read": "",
		"--storage.tsdb.no-lockfile":              "",
		"--storage.tsdb.retention.time":           "1d",
		"--storage.tsdb.wal-compression":          "",
		"--storage.tsdb.min-block-duration":       "2h",
		"--storage.tsdb.max-block-duration":       "2h",
		"--web.enable-lifecycle":                  "",
		"--log.format":                            "json",
		"--log.level":                             "info",
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
