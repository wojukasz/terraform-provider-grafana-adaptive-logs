// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
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
// segment, drop-rule, and exemption endpoints, so resource lifecycle tests
// can exercise real create/read/update/delete/import flows without touching
// a live tenant. It deliberately reproduces the real API's inconsistent
// envelopes: segments and drop-rules list as bare arrays, exemptions list
// wrapped as {"result": [...]}.
type fakeAdaptiveLogsAPI struct {
	mu              sync.Mutex
	segments        map[string]model.Segment
	dropRules       map[string]model.DropRule
	exemptions      map[string]model.Exemption
	recommendations []model.Recommendation
	labelValues     map[string][]string
	nextID          int
}

func newFakeAdaptiveLogsAPI() *fakeAdaptiveLogsAPI {
	return &fakeAdaptiveLogsAPI{
		segments:   map[string]model.Segment{},
		dropRules:  map[string]model.DropRule{},
		exemptions: map[string]model.Exemption{},
	}
}

func (f *fakeAdaptiveLogsAPI) Server() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/adaptive-logs/segments", f.handleSegmentList)
	mux.HandleFunc("/adaptive-logs/segment", f.handleSegmentSingle)
	mux.HandleFunc("/adaptive-logs/drop-rules", f.handleDropRuleListCreate)
	mux.HandleFunc("/adaptive-logs/drop-rules/", f.handleDropRuleSingle)
	mux.HandleFunc("/adaptive-logs/exemptions", f.handleExemptionListCreate)
	mux.HandleFunc("/adaptive-logs/exemptions/", f.handleExemptionSingle)
	mux.HandleFunc("/adaptive-logs/recommendations", f.handleRecommendationsList)
	mux.HandleFunc("/loki/api/v1/label/", f.handleLabelValues)
	return httptest.NewServer(mux)
}

func (f *fakeAdaptiveLogsAPI) handleRecommendationsList(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	writeJSON(w, http.StatusOK, f.recommendations)
}

func (f *fakeAdaptiveLogsAPI) handleLabelValues(w http.ResponseWriter, r *http.Request) {
	label := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/loki/api/v1/label/"), "/values")

	f.mu.Lock()
	defer f.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   f.labelValues[label],
	})
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

// --- drop rules (bare array list, path-based single id, like the real API) ---

func (f *fakeAdaptiveLogsAPI) handleDropRuleListCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		f.mu.Lock()
		defer f.mu.Unlock()
		rules := make([]model.DropRule, 0, len(f.dropRules))
		for _, dr := range f.dropRules {
			rules = append(rules, dr)
		}
		writeJSON(w, http.StatusOK, rules)
	case http.MethodPost:
		var req model.DropRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		rule := model.DropRule{
			ID: f.id("droprule"), SegmentID: req.SegmentID, Name: req.Name, Version: 1,
			Disabled: req.Disabled, ExpiresAt: req.ExpiresAt, Body: req.Body,
			CreatedAt: now(), UpdatedAt: now(),
		}
		f.dropRules[rule.ID] = rule
		writeJSON(w, http.StatusCreated, rule)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (f *fakeAdaptiveLogsAPI) handleDropRuleSingle(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/adaptive-logs/drop-rules/")

	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		rule, ok := f.dropRules[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("drop rule not found: %s", id)})
			return
		}
		writeJSON(w, http.StatusOK, rule)
	case http.MethodPut:
		var req model.DropRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rule, ok := f.dropRules[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("drop rule not found: %s", id)})
			return
		}
		rule.SegmentID = req.SegmentID
		rule.Name = req.Name
		rule.Disabled = req.Disabled
		rule.ExpiresAt = req.ExpiresAt
		rule.Body = req.Body
		rule.Version++
		rule.UpdatedAt = now()
		f.dropRules[id] = rule
		writeJSON(w, http.StatusOK, rule)
	case http.MethodDelete:
		if _, ok := f.dropRules[id]; !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("drop rule not found: %s", id)})
			return
		}
		delete(f.dropRules, id)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- exemptions (list wrapped as {"result": [...]}, like the real API) ---

func (f *fakeAdaptiveLogsAPI) handleExemptionListCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		f.mu.Lock()
		defer f.mu.Unlock()
		exemptions := make([]model.Exemption, 0, len(f.exemptions))
		for _, e := range f.exemptions {
			exemptions = append(exemptions, e)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"result": exemptions})
	case http.MethodPost:
		var req model.ExemptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		exemption := model.Exemption{
			ID: f.id("exemption"), StreamSelector: req.StreamSelector, Reason: req.Reason,
			ExpiresAt: req.ExpiresAt, CreatedAt: now(), UpdatedAt: now(),
		}
		f.exemptions[exemption.ID] = exemption
		writeJSON(w, http.StatusCreated, exemption)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (f *fakeAdaptiveLogsAPI) handleExemptionSingle(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/adaptive-logs/exemptions/")

	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		exemption, ok := f.exemptions[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("exemption not found: %s", id)})
			return
		}
		writeJSON(w, http.StatusOK, exemption)
	case http.MethodPut:
		var req model.ExemptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exemption, ok := f.exemptions[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("exemption not found: %s", id)})
			return
		}
		exemption.StreamSelector = req.StreamSelector
		exemption.Reason = req.Reason
		exemption.ExpiresAt = req.ExpiresAt
		exemption.UpdatedAt = now()
		f.exemptions[id] = exemption
		writeJSON(w, http.StatusOK, exemption)
	case http.MethodDelete:
		if _, ok := f.exemptions[id]; !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("exemption not found: %s", id)})
			return
		}
		delete(f.exemptions, id)
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
