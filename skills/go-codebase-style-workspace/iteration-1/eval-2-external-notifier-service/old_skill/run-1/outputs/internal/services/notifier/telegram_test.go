package notifier

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/blkst8/client-service/internal/config"
)

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		resp    sendMessageResponse
		wantErr bool
	}{
		{
			name:   "success",
			status: http.StatusOK,
			resp:   sendMessageResponse{OK: true},
		},
		{
			name:    "telegram api returns ok false",
			status:  http.StatusBadRequest,
			resp:    sendMessageResponse{OK: false, Description: "Bad Request: chat not found"},
			wantErr: true,
		},
		{
			name:    "telegram api is down",
			status:  http.StatusServiceUnavailable,
			resp:    sendMessageResponse{OK: false, Description: "service unavailable"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotRequest sendMessageRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("failed to read request body: %v", err)
				}
				_ = json.Unmarshal(body, &gotRequest)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_ = json.NewEncoder(w).Encode(tt.resp)
			}))
			defer server.Close()

			n := New(config.Telegram{
				Enabled:    true,
				BotToken:   "test-token",
				APIBaseURL: server.URL,
				Timeout:    time.Second,
			})

			err := n.SendMessage(context.Background(), "42", "hello")
			if (err != nil) != tt.wantErr {
				t.Fatalf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if want := "/bottest-token/sendMessage"; gotPath != want {
				t.Errorf("request path = %q, want %q", gotPath, want)
			}
			if gotRequest.ChatID != "42" || gotRequest.Text != "hello" {
				t.Errorf("request payload = %+v, want chat_id %q text %q", gotRequest, "42", "hello")
			}
		})
	}
}

func TestNewDisabledReturnsNoop(t *testing.T) {
	n := New(config.Telegram{Enabled: false})

	if err := n.SendMessage(context.Background(), "42", "hello"); err != nil {
		t.Errorf("SendMessage() error = %v, want nil", err)
	}
}

func TestNewAppliesDefaultTimeout(t *testing.T) {
	n := newTelegram(config.Telegram{Enabled: true, APIBaseURL: "http://localhost"})

	if n.httpClient.Timeout != DefaultTelegramTimeout {
		t.Errorf("timeout = %v, want %v", n.httpClient.Timeout, DefaultTelegramTimeout)
	}
}

func TestSendMessageTransportFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	server.Close()

	n := New(config.Telegram{
		Enabled:    true,
		BotToken:   "test-token",
		APIBaseURL: server.URL,
		Timeout:    time.Second,
	})

	err := n.SendMessage(context.Background(), "42", "hello")
	if err == nil {
		t.Fatal("SendMessage() error = nil, want transport error")
	}
}
