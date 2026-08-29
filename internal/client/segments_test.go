// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

func TestCreateSegment(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody model.CreateSegmentRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Segment{
			ID:       "seg-1",
			Name:     gotBody.Name,
			Selector: gotBody.Selector,
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	segment, err := c.CreateSegment(model.CreateSegmentRequest{Name: "sre", Selector: `{team="sre"}`})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/adaptive-logs/segment", gotPath)
	assert.Equal(t, "sre", gotBody.Name)
	assert.Equal(t, "seg-1", segment.ID)
	assert.Equal(t, "sre", segment.Name)
}

func TestReadSegmentNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "missing", r.URL.Query().Get("segment"))
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "segment not found: missing"})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	_, err = c.ReadSegment("missing")
	require.Error(t, err)
	assert.True(t, client.IsErrNotFound(err))
}

func TestListSegments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/adaptive-logs/segments", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]model.Segment{
			{ID: "seg-1", Name: "sre", Selector: `{team="sre"}`},
			{ID: "seg-2", Name: "billing", Selector: `{team="billing"}`},
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	segments, err := c.ListSegments()
	require.NoError(t, err)
	assert.Len(t, segments, 2)
}

func TestUpdateSegment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "seg-1", r.URL.Query().Get("segment"))

		var body model.CreateSegmentRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Segment{ID: "seg-1", Name: body.Name, Selector: body.Selector})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	segment, err := c.UpdateSegment("seg-1", model.CreateSegmentRequest{Name: "renamed", Selector: `{team="sre"}`})
	require.NoError(t, err)
	assert.Equal(t, "renamed", segment.Name)
}

func TestDeleteSegment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "seg-1", r.URL.Query().Get("segment"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	require.NoError(t, c.DeleteSegment("seg-1"))
}
