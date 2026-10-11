package websocket

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubRegistrationAndBroadcast(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create test HTTP server with WS endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect client 1
	dialer := websocket.DefaultDialer
	conn1, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn1.Close()

	// Connect client 2
	conn2, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial second websocket: %v", err)
	}
	defer conn2.Close()

	// Wait briefly for registration
	time.Sleep(50 * time.Millisecond)

	// Broadcast test payload
	testPayload := map[string]string{"greeting": "hello multi-network"}
	hub.Broadcast("test_event", testPayload)

	// Verify client 1 receives the message
	_ = conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
	var msg1 WSMessage
	if err := conn1.ReadJSON(&msg1); err != nil {
		t.Fatalf("client 1 failed to read message: %v", err)
	}

	if msg1.Type != "test_event" {
		t.Errorf("expected type 'test_event', got '%s'", msg1.Type)
	}

	// Verify client 2 receives the message
	_ = conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	var msg2 WSMessage
	if err := conn2.ReadJSON(&msg2); err != nil {
		t.Fatalf("client 2 failed to read message: %v", err)
	}

	if msg2.Type != "test_event" {
		t.Errorf("expected type 'test_event', got '%s'", msg2.Type)
	}
}
