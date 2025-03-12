
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

package semconv // TODO(bwplotka): Use id prefix or something more unique?

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type CustomElementsClass string

const (
		MetricMyOrgCustomElementsCategoryVal1CustomElementsClass CustomElementsClass = "FIRST"
		MetricMyOrgCustomElementsCategoryVal2CustomElementsClass CustomElementsClass = "SECOND"
		MetricMyOrgCustomElementsCategoryOtherCustomElementsClass CustomElementsClass = "OTHER"
)

func MustNewCustomElementsCounterVec(reg prometheus.Registerer) *prometheus.CounterVec {
	return promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "my_app_custom_elements_changed_total",
		Help: "Changed custom counter metric (v0.2.0) for my app counting important elements. It serves as an example of a very important metric that everyone is using.",
		// Unit: "{elements}" // TODO(bwplotka): Add Unit as one of the supported options.
	}, []string{
		"number",
		"class",
		"fraction",
	})
}

/*
TODO(bwplotka): Add more type safety e.g. for CustomElementsCounterVec:

type CustomElementsCounterVec struct {
	prometheus.CounterVec
}

func (v *CustomElementsCounterVec) WithLabelValues(integer int, category CustomElementsCategory, fraction float64) prometheus.Counter {
	// This is not ideal as we do, potentially expensive stringifying on the hot path.
  // Fix might require internals to completely differ in the client_golang for the efficient solution.
	return v.CounterVec.WithLabelValues(fmt.Sprintf("%v", integer), string(category), fmt.Sprintf("%v", fraction))
}
*/



