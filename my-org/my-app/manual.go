package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func mustNewCustomStableMetric(reg prometheus.Registerer) *prometheus.CounterVec {
	return promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "my_app_custom_elements_total",
		Help: "Custom counter metric for my app counting important elements. It serves as an example " +
				"of a very important metric that everyone is using.",
	}, []string{"integer", "category", "fraction"})
}
