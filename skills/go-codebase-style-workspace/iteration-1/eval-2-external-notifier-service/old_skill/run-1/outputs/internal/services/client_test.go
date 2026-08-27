package services

import (
	"context"
	"errors"
	"testing"

	"github.com/blkst8/client-service/internal/models"
	"github.com/blkst8/client-service/internal/repository"
)

// mockClientRepository is a test double for repository.Client.
type mockClientRepository struct {
	createErr error
	getErr    error
}

func (m *mockClientRepository) Create(_ context.Context, client *models.Client) error {
	if m.createErr != nil {
		return m.createErr
	}
	client.ID = 1
	return nil
}

func (m *mockClientRepository) Get(_ context.Context, id uint32) (*models.Client, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &models.Client{ID: id, Name: "Alice", Email: "alice@example.com", TelegramID: "42"}, nil
}

// mockNotifier is a test double for notifier.Notifier that records calls.
type mockNotifier struct {
	err     error
	calls   int
	chatID  string
	message string
}

func (m *mockNotifier) SendMessage(_ context.Context, chatID string, text string) error {
	m.calls++
	m.chatID = chatID
	m.message = text
	return m.err
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateClientRequest
		repoErr     error
		notifierErr error
		wantErr     bool
		wantCalls   int
	}{
		{
			name:      "creation succeeds and sends welcome message",
			req:       CreateClientRequest{Name: "Alice", Email: "alice@example.com", TelegramID: "42"},
			wantCalls: 1,
		},
		{
			name:        "telegram down does not fail creation",
			req:         CreateClientRequest{Name: "Alice", Email: "alice@example.com", TelegramID: "42"},
			notifierErr: errors.New("failed to call telegram api: connection refused"),
			wantCalls:   1,
		},
		{
			name:      "client without telegram id skips notification",
			req:       CreateClientRequest{Name: "Alice", Email: "alice@example.com"},
			wantCalls: 0,
		},
		{
			name:      "repository failure fails creation without notifying",
			req:       CreateClientRequest{Name: "Alice", Email: "alice@example.com", TelegramID: "42"},
			repoErr:   errors.New("failed to insert client"),
			wantErr:   true,
			wantCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClientRepository{createErr: tt.repoErr}
			notifier := &mockNotifier{err: tt.notifierErr}
			svc := NewClientService(repo, notifier)

			client, err := svc.Create(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if notifier.calls != tt.wantCalls {
				t.Errorf("SendMessage calls = %d, want %d", notifier.calls, tt.wantCalls)
			}
			if tt.wantErr {
				return
			}
			if client == nil || client.ID == 0 {
				t.Errorf("Create() = %+v, want stored client with generated ID", client)
			}
			if notifier.err == nil && tt.req.TelegramID != "" && notifier.chatID != tt.req.TelegramID {
				t.Errorf("SendMessage chatID = %q, want %q", notifier.chatID, tt.req.TelegramID)
			}
			if notifier.err == nil && tt.req.TelegramID != "" && notifier.message == "" {
				t.Error("SendMessage received an empty message")
			}
		})
	}
}

func TestGet(t *testing.T) {
	repo := &mockClientRepository{}
	svc := NewClientService(repo, &mockNotifier{})

	client, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if client.Name != "Alice" {
		t.Errorf("Get() name = %q, want %q", client.Name, "Alice")
	}

	repo.getErr = repository.ErrClientNotFound
	if _, err := svc.Get(context.Background(), 99); !errors.Is(err, repository.ErrClientNotFound) {
		t.Errorf("Get() error = %v, want ErrClientNotFound", err)
	}
}
