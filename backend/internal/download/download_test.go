package download

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
	"netfusion/backend/internal/monitor"
	"netfusion/backend/internal/websocket"
)

func TestDownloaderLifecycle(t *testing.T) {
	// 1. Temporary test server simulating chunked range HTTP endpoint
	dummyData := make([]byte, 1024*64) // 64 KB
	for i := range dummyData {
		dummyData[i] = byte(i % 256)
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(dummyData)))

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			var start, end int
			_, _ = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
			if end >= len(dummyData) {
				end = len(dummyData) - 1
			}
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(dummyData)))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(dummyData[start : end+1])
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(dummyData)
	}))
	defer testServer.Close()

	// 2. Setup SQLite DB and Downloader
	tmpFile, err := os.CreateTemp("", "netfusion_dl_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cfg := &config.Config{
		ServerHost:        "127.0.0.1",
		ServerPort:        "8080",
		DBDriver:          "sqlite",
		DBURL:             tmpFile.Name(),
		SimulationMode:    true,
		TelemetryInterval: 100 * time.Millisecond,
		LogLevel:          slog.LevelError,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := database.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mon := monitor.New(cfg, db, logger)
	mon.Start(ctx)
	defer mon.Stop()

	wsHub := websocket.NewHub(logger)
	go wsHub.Run(ctx)

	dl := NewDownloader(db, mon, wsHub, logger)
	defer os.RemoveAll(dl.downloadDir)

	// 3. Test starting single mode download
	sessionSingle, err := dl.StartDownload(testServer.URL+"/testfile.bin", "single", "")
	if err != nil {
		t.Fatalf("failed to start single download: %v", err)
	}
	if sessionSingle == nil || sessionSingle.ID == "" {
		t.Fatalf("expected valid session")
	}

	// 4. Test starting multi mode download
	sessionMulti, err := dl.StartDownload(testServer.URL+"/testfile-multi.bin", "multi", "")
	if err != nil {
		t.Fatalf("failed to start multi download: %v", err)
	}
	if sessionMulti == nil || sessionMulti.ID == "" {
		t.Fatalf("expected valid session")
	}

	// 5. Test stop download
	err = dl.StopDownload(sessionSingle.ID)
	if err != nil {
		t.Errorf("failed to stop download: %v", err)
	}

	// 6. Test list sessions
	sessions := dl.ListSessions()
	if len(sessions) < 2 {
		t.Errorf("expected at least 2 sessions, got %d", len(sessions))
	}
}
