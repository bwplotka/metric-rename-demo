


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

package my_app_latency_seconds_total

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MustNew returns my_app_latency_seconds_total~seconds.histogram.
func MustNewHistogramVec(reg prometheus.Registerer) *HistogramVec {
	reg = prometheus.WrapRegistererWith(prometheus.Labels{"__schema_url__": "https://github.com/bwplotka/metric-rename-demo/tree/main/my-org/semconv/v1.1.0"}, reg)

	return &HistogramVec{promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
		Name: "my_app_latency_seconds_total",
		Help: "Histogram with my-app latency seconds (v1.1.0)",
		// Unit: "seconds" // TODO(bwplotka): Add Unit as one of the supported options.
	}, []string{
		// HTTP status code.
		"code",
	})}
}

type HistogramVec struct {
	*prometheus.HistogramVec
}

func (x *HistogramVec) WithLabelValues(
	code int,
) prometheus.Observer {
	// TODO(bwplotka): This is actually not ideal for efficiency reasons (type conversions to string).
  // Fix might require internals to completely differ in the client_golang for the efficient solution.
	return x.HistogramVec.WithLabelValues(
		fmt.Sprintf("%v", code),
	)
}


