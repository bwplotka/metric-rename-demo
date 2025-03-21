package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func mustNewCustomElements(reg prometheus.Registerer) *prometheus.CounterVec {
	return promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "my_app_custom_elements_total",
		Help: "Custom counter metric for my app counting important elements. It serves as an example " +
				"of a very important metric that everyone is using.",
	}, []string{"integer", "category", "fraction"})
}

// MustNew returns my_app_latency_milliseconds_total~milliseconds.histogram.
func mustNewLatency(reg prometheus.Registerer) *prometheus.HistogramVec {
	return promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
		Name: "my_app_latency_milliseconds_total",
		Help: "Histogram with my-app latency milliseconds (v1.0.0)",
		// Unit: "milliseconds" // TODO(bwplotka): Add Unit as one of the supported options.
	}, []string{
		// HTTP status code.
		"code",
	})
}
