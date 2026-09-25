# Feature: Invoice Reconciliation

Every `worker.jobs_intervals.reconcile_invoices` (default `5m`) the ticker
worker (`internal/worker/handlers/reconcile_invoices.go`) invokes
`InvoiceUsecase.Reconcile`:

1. List all invoices with status `pending` (`repository.Invoice.ListByStatus`).
2. For each with `due_date <= now`: update status to `overdue`. Individual
   update failures are logged and retried on the next run (log-and-continue).
3. Emit business metrics: `invoice_service_reconcile_runs_total`,
   `invoice_service_reconcile_invoices_checked_total`,
   `invoice_service_reconcile_invoices_overdue_total`.
4. If any invoices were marked overdue, notify ops through the
   `service/notifier` webhook (non-fatal on failure).

The same flow is also available as an on-demand task
(`internal/workers.NewReconcileInvoicesTask`) for the worker pool.
