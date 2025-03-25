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

package my_app_custom_elements

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TODO(bwplotka): Detect changes and not add V2 prefix if no change was made?
type ClassV2 string

const (
		FirstClassV2 ClassV2 = "FIRST"
		SecondClassV2 ClassV2 = "SECOND"
		OtherClassV2 ClassV2 = "OTHER"
)

// MustNew returns my_app_custom_elements.2.
func MustNewMyAppCustomElementsChangedTotalV2(reg prometheus.Registerer) *my_app_custom_elements_changed_totalCounterVecV2 {
	reg = prometheus.WrapRegistererWith(prometheus.Labels{"__schema_url__": "https://github.com/bwplotka/metric-rename-demo/tree/main/my-org/semconv/v1.1.0"}, reg)

	return &my_app_custom_elements_changed_totalCounterVecV2{promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "my_app_custom_elements_changed_total",
		Help: "Custom counter metric (v1.1.0) for my app counting important elements. It serves as an example of a very important metric that everyone is using. Replacement to my_app_custom_elements_total~counter",
		// Unit: "{unknown}" // TODO(bwplotka): Add Unit as one of the supported options.
	}, []string{
		// Important label that specifies the integer for this count.
		"number",
		// Important label that specifies the category for this count.
		"class",
		// This is an important label that specifies the fraction for this count.
		"fraction",
	})}
}

type my_app_custom_elements_changed_totalCounterVecV2 struct {
	*prometheus.CounterVec
}

func (x *my_app_custom_elements_changed_totalCounterVecV2) WithLabelValues(
	number int,
	class Class,
	fraction float64,
) prometheus.Counter {
	// TODO(bwplotka): This is actually not ideal for efficiency reasons (type conversions to string).
  // Fix might require internals to completely differ in the client_golang for the efficient solution.
	return x.CounterVec.WithLabelValues(
		fmt.Sprintf("%v", number),
		fmt.Sprintf("%v", class),
		fmt.Sprintf("%v", fraction),
	)
}


