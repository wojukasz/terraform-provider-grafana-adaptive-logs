// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

// CheckAccTestsEnabled skips the test unless TF_ACC is set. Unlike a typical
// Terraform provider acceptance test, these tests never call a real Grafana
// Cloud tenant: they run the full create/read/update/delete lifecycle
// against an in-memory fakeAdaptiveLogsAPI so no live writes ever happen.
func CheckAccTestsEnabled(t *testing.T) {
	t.Helper()

	if enabled, _ := strconv.ParseBool(os.Getenv(resource.EnvTfAcc)); enabled {
		return
	}

	t.Skip("Set TF_ACC=1 to enable acceptance tests.")
}

// fakeAdaptiveLogsAPI is a minimal in-memory stand-in for the Adaptive Logs
// segment endpoints, so resource lifecycle tests can exercise real
// create/read/update/delete/import flows without touching a live tenant.
type fakeAdaptiveLogsAPI struct {
	mu       sync.Mutex
	segments map[string]model.Segment
	nextID   int
}

func newFakeAdaptiveLogsAPI() *fakeAdaptiveLogsAPI {
	return &fakeAdaptiveLogsAPI{
		segments: map[string]model.Segment{},
	}
}

func (f *fakeAdaptiveLogsAPI) Server() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/adaptive-logs/segments", f.handleSegmentList)
	mux.HandleFunc("/adaptive-logs/segment", f.handleSegmentSingle)
	return httptest.NewServer(mux)
}

func (f *fakeAdaptiveLogsAPI) id(prefix string) string {
	f.nextID++
	return fmt.Sprintf("fake-%s-%d", prefix, f.nextID)
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// --- segments ---

func (f *fakeAdaptiveLogsAPI) handleSegmentList(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	segments := make([]model.Segment, 0, len(f.segments))
	for _, s := range f.segments {
		segments = append(segments, s)
	}
	writeJSON(w, http.StatusOK, segments)
}

func (f *fakeAdaptiveLogsAPI) handleSegmentSingle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req model.CreateSegmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f.mu.Lock()
		defer f.mu.Unlock()

		segment := model.Segment{ID: f.id("segment"), Name: req.Name, Selector: req.Selector, CreatedAt: now(), UpdatedAt: now()}
		f.segments[segment.ID] = segment
		writeJSON(w, http.StatusCreated, segment)
	case http.MethodGet:
		id := r.URL.Query().Get("segment")
		f.mu.Lock()
		defer f.mu.Unlock()
		segment, ok := f.segments[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("segment not found: %s", id)})
			return
		}
		writeJSON(w, http.StatusOK, segment)
	case http.MethodPut:
		id := r.URL.Query().Get("segment")
		var req model.CreateSegmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		segment, ok := f.segments[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("segment not found: %s", id)})
			return
		}
		segment.Name = req.Name
		segment.Selector = req.Selector
		segment.UpdatedAt = now()
		f.segments[id] = segment
		writeJSON(w, http.StatusOK, segment)
	case http.MethodDelete:
		id := r.URL.Query().Get("segment")
		f.mu.Lock()
		defer f.mu.Unlock()
		if _, ok := f.segments[id]; !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("segment not found: %s", id)})
			return
		}
		delete(f.segments, id)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
