// Package metrics provides business and monitoring metrics exposed on
// /metrics in Prometheus format.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ReconcileRuns = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "invoice_service",
		Name:      "reconcile_runs_total",
		Help:      "Total number of invoice reconciliation runs.",
	})
	ReconcileChecked = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "invoice_service",
		Name:      "reconcile_invoices_checked_total",
		Help:      "Total number of pending invoices examined by reconciliation.",
	})
	ReconcileOverdue = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "invoice_service",
		Name:      "reconcile_invoices_overdue_total",
		Help:      "Total number of invoices marked overdue by reconciliation.",
	})
)

func init() {
	prometheus.MustRegister(ReconcileRuns, ReconcileChecked, ReconcileOverdue)
}
