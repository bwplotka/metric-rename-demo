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

type Category string

const (
		FirstCategory Category = "first"
		SecondCategory Category = "second"
		OtherCategory Category = "other"
)

// MustNew returns my_app_custom_elements.
func MustNewMyAppCustomElementsTotal(reg prometheus.Registerer) *MyAppCustomElementsTotal {
	return &MyAppCustomElementsTotal{promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "my_app_custom_elements_total",
		Help: "Custom counter metric (1.0.0) for my app counting important elements. It serves as an example of a very important metric that everyone is using.",
		Unit: "elements", // Yolo parsing of UCUM.
		ConstLabels: map[string]string{
			"__schema_url__": "https://bwplotka.dev/semconv/1.0.0",
		},
		
	}, []string{
		// Important label that specifies the integer for this count.
		"integer",
		// Important label that specifies the category for this count.
		"category",
		// This is an important label that specifies the fraction for this count.
		"fraction",
	})}
}

type MyAppCustomElementsTotal struct {
	*prometheus.CounterVec
}

func (x *MyAppCustomElementsTotal) WithLabelValues(
	integer int,
	category Category,
	fraction float64,
) prometheus.Counter {
	// TODO(bwplotka): This is actually not ideal for efficiency reasons (type conversions to string).
  // Fix might require internals to completely differ in the client_golang for the efficient solution.
	return x.CounterVec.WithLabelValues(
		fmt.Sprintf("%v", integer),
		fmt.Sprintf("%v", category),
		fmt.Sprintf("%v", fraction),
	)
}


