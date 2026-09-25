// Package notifier is a client for an external notification webhook
// (e.g. an ops Slack/Telegram bridge). It performs external I/O only — it
// never imports repositories, opens DB connections, or runs SQL.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Interface definition
type Notifier interface {
	Send(ctx context.Context, subject string, message string) error
}

// Implementation struct (private)
type webhookNotifier struct {
	webhookURL string
	http       *http.Client
}

// Constructor returns the interface. With an empty webhook URL the notifier
// is disabled and Send is a no-op.
func New(webhookURL string) Notifier {
	return &webhookNotifier{
		webhookURL: webhookURL,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts the notification payload to the configured webhook.
func (n *webhookNotifier) Send(ctx context.Context, subject string, message string) error {
	if n.webhookURL == "" {
		return nil
	}

	body, err := json.Marshal(map[string]string{"subject": subject, "message": message})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("unexpected status from provider: %s", resp.Status)
	}

	return nil
}
