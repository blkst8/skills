// Package notifier provides a client for sending messages through the
// Telegram Bot API.
//
// The notifier performs external I/O only: it never touches repositories,
// opens database connections, or runs SQL.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier sends messages to chat rooms of an external messaging provider.
type Notifier interface {
	Send(ctx context.Context, chatID string, msg string) error
}

// telegramNotifier is the Telegram Bot API implementation of Notifier.
type telegramNotifier struct {
	token   string
	baseURL string
	http    *http.Client
}

// New returns a Telegram-backed Notifier using the given bot token.
func New(token string) Notifier {
	return &telegramNotifier{
		token:   token,
		baseURL: "https://api.telegram.org",
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts a text message to the given chat via the sendMessage endpoint.
// Failures are wrapped with context; callers decide the retry/log policy.
func (n *telegramNotifier) Send(ctx context.Context, chatID string, msg string) error {
	endpoint := n.baseURL + "/bot" + n.token + "/sendMessage"

	body, err := json.Marshal(map[string]string{"chat_id": chatID, "text": msg})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status from provider: %s", resp.Status)
	}

	return nil
}
