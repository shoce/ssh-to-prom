package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var acceptedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ssh_accepted_total",
		Help: "Number of accepted ssh connection attempts",
	},
	[]string{"authmethod", "username"},
)

var failedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ssh_failed_total",
		Help: "Number of failed ssh connection attempts",
	},
	[]string{"authmethod", "username"},
)

// Reporter defines the behaviour for failed connection event reporters
type Reporter interface {
	Report(e ConnEvent) error
}

type prometheusReporter struct{}

func (pr prometheusReporter) Report(e ConnEvent) error {
	if e.Accepted {
		acceptedCounter.WithLabelValues(e.AuthMethod, e.User).Inc()
	} else {
		failedCounter.WithLabelValues(e.AuthMethod, e.User).Inc()
	}
	return nil
}

func init() {
	prometheus.MustRegister(acceptedCounter)
	prometheus.MustRegister(failedCounter)
}
