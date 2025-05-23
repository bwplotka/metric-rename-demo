// Copyright 2025 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated from semantic convention specification. DO NOT EDIT.

package my_app_latency

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MustNew returns my_app_latency.
// Deprecated: Use my_app_latency_2 instead.
// Note: {"updated": {"backward_promql": "value{} * 1000", "forward_promql": "value{} / 1000", "note": "Ups, we did not use base unit, our bad. This metric should be auto-transformable, see diff for all changes.", "replaced_by_id": "my_app_latency.2"}}
func MustNewMyAppLatencyMilliseconds(reg prometheus.Registerer, buckets []float64) *MyAppLatencyMilliseconds {
	return &MyAppLatencyMilliseconds{promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
		Name: "my_app_latency_milliseconds",
		Help: "Histogram with my-app latency milliseconds (1.0.0)",
		Unit: "milliseconds", // Yolo parsing of UCUM.
		ConstLabels: map[string]string{
			"__schema_url__": "https://bwplotka.dev/semconv/1.1.0",
		},
		Buckets: buckets,
	}, []string{
		// HTTP status code.
		"code",
	})}
}

type MyAppLatencyMilliseconds struct {
	*prometheus.HistogramVec
}

func (x *MyAppLatencyMilliseconds) WithLabelValues(
	code int,
) prometheus.Observer {
	// TODO(bwplotka): This is actually not ideal for efficiency reasons (type conversions to string).
  // Fix might require internals to completely differ in the client_golang for the efficient solution.
	return x.HistogramVec.WithLabelValues(
		fmt.Sprintf("%v", code),
	)
}


