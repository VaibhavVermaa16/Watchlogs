package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMux(t *testing.T) {
	mux := newMux()

	tests := []struct {
		route       string
		expectedHdl http.HandlerFunc
	}{
		{route: "/ingest", expectedHdl: ingest},
		{route: "/search", expectedHdl: search},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.route, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("expected route %s to be registered, but got 404 Not Found", tt.route)
		}
	}
}

func TestIngestEndpoint(t *testing.T) {
	t.Run("ingest test log with POST request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/ingest", nil)
		response := httptest.NewRecorder()

		ingest(response, request)
		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}
		if response.Body.String() != "ingested" {
			t.Errorf("expected body 'ingested', got '%s'", response.Body.String())
		}
	})
	t.Run("ingest test log with non POST request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ingest", nil)
		response := httptest.NewRecorder()

		ingest(response, request)
		if response.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 Method Not Allowed, got %d", response.Code)
		}
	})

}

func TestSearchEndpoint(t *testing.T) {
	t.Run("Get test logs", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/search", nil)
		response := httptest.NewRecorder()

		search(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}
		if response.Body.String() != "search endpoint" {
			t.Errorf("expected body 'search endpoint', got '%s'", response.Body.String())
		}
	})
}
