package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/blkst8/client-service/internal/models"
)

type fakeRepo struct {
	createErr error
	created   models.Client
	nextID    uint32
}

func (f *fakeRepo) Create(ctx context.Context, client models.Client) (uint32, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.created = client
	f.nextID++
	return f.nextID, nil
}

func (f *fakeRepo) Get(ctx context.Context, id uint32) (*models.Client, error) {
	return nil, nil
}

type fakeNotifier struct {
	sendErr error
	chatID  string
	msg     string
	calls   int
}

func (f *fakeNotifier) Send(ctx context.Context, chatID string, msg string) error {
	f.calls++
	f.chatID = chatID
	f.msg = msg
	return f.sendErr
}

func TestClientUsecaseCreate(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		notifierErr   error
		telegramID    string
		wantErr       bool
		wantNotifCall bool
	}{
		{
			name:          "sends welcome message on success",
			telegramID:    "12345",
			wantNotifCall: true,
		},
		{
			name:          "succeeds even when telegram is down",
			telegramID:    "12345",
			notifierErr:   errors.New("unexpected status from provider: 500 Internal Server Error"),
			wantNotifCall: true,
		},
		{
			name:       "skips notification without telegram id",
			telegramID: "",
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("duplicate key"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{createErr: tt.repoErr}
			notifier := &fakeNotifier{sendErr: tt.notifierErr}
			uc := NewClientUsecase(repo, notifier)

			client, err := uc.Create(context.Background(), CreateClientRequest{
				Name:       "Alice",
				Email:      "alice@example.com",
				TelegramID: tt.telegramID,
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if client == nil {
				t.Fatalf("expected client, got nil")
			}
			if client.ID == 0 {
				t.Fatalf("expected generated id to be set")
			}

			if got := notifier.calls; got != boolToInt(tt.wantNotifCall) {
				t.Fatalf("notifier calls = %d, want %d", got, boolToInt(tt.wantNotifCall))
			}
			if tt.wantNotifCall {
				if notifier.chatID != tt.telegramID {
					t.Fatalf("notifier chatID = %q, want %q", notifier.chatID, tt.telegramID)
				}
				if notifier.msg == "" {
					t.Fatalf("expected non-empty welcome message")
				}
			}
		})
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
