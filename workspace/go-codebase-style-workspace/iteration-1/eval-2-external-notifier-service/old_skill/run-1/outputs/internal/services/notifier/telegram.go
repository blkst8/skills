package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/blkst8/client-service/internal/config"
)

// DefaultTelegramTimeout bounds every call to the Telegram Bot API when no
// timeout is configured. It keeps a Telegram outage from hanging requests.
const DefaultTelegramTimeout = 5 * time.Second

// maxResponseBody caps how much of a Telegram API response is read into
// memory.
const maxResponseBody = 4096

// sendMessagePath is the Bot API method used to deliver messages. The bot
// token is part of the path: /bot<token>/sendMessage.
const sendMessagePath = "/bot%s/sendMessage"

// telegram is a Notifier backed by the Telegram Bot API.
type telegram struct {
	apiBaseURL string
	botToken   string
	httpClient *http.Client
}

// newTelegram builds the Telegram Bot API client.
func newTelegram(cfg config.Telegram) *telegram {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTelegramTimeout
	}

	return &telegram{
		apiBaseURL: strings.TrimRight(cfg.APIBaseURL, "/"),
		botToken:   cfg.BotToken,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// sendMessageRequest is the payload for the sendMessage Bot API method.
type sendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// sendMessageResponse is the reply from the sendMessage Bot API method.
type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// SendMessage delivers text to the Telegram chat identified by chatID using
// the sendMessage Bot API method.
//
// A non-2xx status, a transport failure, or an API-level "ok": false reply is
// reported as an error. Callers decide how to handle it — in this codebase the
// welcome message is best effort, so the error is only logged.
func (t *telegram) SendMessage(ctx context.Context, chatID string, text string) error {
	payload, err := json.Marshal(sendMessageRequest{ChatID: chatID, Text: text})
	if err != nil {
		return fmt.Errorf("failed to marshal telegram request: %w", err)
	}

	apiURL := t.apiBaseURL + fmt.Sprintf(sendMessagePath, t.botToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to build telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call telegram api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return fmt.Errorf("failed to read telegram response: %w", err)
	}

	var apiResponse sendMessageResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to decode telegram response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !apiResponse.OK {
		return fmt.Errorf("telegram api rejected message: status=%d description=%q", resp.StatusCode, apiResponse.Description)
	}

	return nil
}
