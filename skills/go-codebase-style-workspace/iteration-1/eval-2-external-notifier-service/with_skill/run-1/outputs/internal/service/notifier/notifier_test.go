package notifier

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelegramNotifierSend(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantErr    bool
		wantChatID string
		wantText   string
	}{
		{
			name:       "sends message on 200",
			status:     http.StatusOK,
			wantChatID: "12345",
			wantText:   "Welcome!",
		},
		{
			name:    "fails on provider error",
			status:  http.StatusServiceUnavailable,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotChatID, gotText string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("content-type = %s, want application/json", got)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}

				var payload map[string]string
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatalf("failed to unmarshal payload: %v", err)
				}
				gotChatID = payload["chat_id"]
				gotText = payload["text"]

				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			n := &telegramNotifier{
				token:   "test-token",
				baseURL: srv.URL,
				http:    srv.Client(),
			}

			err := n.Send(context.Background(), "12345", "Welcome!")

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotChatID != tt.wantChatID {
				t.Fatalf("chat_id = %q, want %q", gotChatID, tt.wantChatID)
			}
			if gotText != tt.wantText {
				t.Fatalf("text = %q, want %q", gotText, tt.wantText)
			}
		})
	}
}
