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

package my_app_some_elements

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MustNew returns my_app_some_elements.
// Deprecated {"obsoleted": {"note": "Not useful anymore"}}
func MustNewMyAppSomeElementsTotals(reg prometheus.Registerer) prometheus.Gauge {
	return promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "my_app_some_elements_totals",
		Help: "some metric",
		// Unit: "{unknown}" // TODO(bwplotka): Add Unit as one of the supported options.
		ConstLabels: map[string]string{
			"__schema_url__": "https://raw.githubusercontent.com/bwplotka/metric-rename-demo/refs/heads/diff/my-org/semconv/v1.1.0",
		},
	})
}


