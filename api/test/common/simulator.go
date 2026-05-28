package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"atom-maintenance/internal/adapters/simulator"
)

func FakeSimulator(t *testing.T, items []simulator.TelemetryItem) *simulator.Client {
	t.Helper()

	byID := make(map[int64]simulator.TelemetryItem, len(items))
	for _, item := range items {
		byID[item.EquipmentID] = item
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/telemetry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	})

	mux.HandleFunc("/api/v1/telemetry/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/telemetry/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		item, ok := byID[id]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return simulator.NewClient(srv.URL, 5*time.Second)
}
