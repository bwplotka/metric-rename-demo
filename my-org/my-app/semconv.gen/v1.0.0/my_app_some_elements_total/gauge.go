


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

package my_app_some_elements_total

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MustNew returns my_app_some_elements_total~gauge.
func MustNewGauge(reg prometheus.Registerer) prometheus.Gauge {
	reg = prometheus.WrapRegistererWith(prometheus.Labels{"__schema_url__": "https://github.com/bwplotka/metric-rename-demo/tree/main/my-org/semconv/v1.0.0"}, reg)

	return promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "my_app_some_elements_total",
		Help: "some metric",
		// Unit: "{unknown}" // TODO(bwplotka): Add Unit as one of the supported options.
	})
}


