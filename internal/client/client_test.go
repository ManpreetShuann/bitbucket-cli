package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/1.0/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth: %s", r.Header.Get("Authorization"))
		}
		if r.Method != "GET" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"key": "PROJ"})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Get(context.Background(), "/projects", nil, &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result["key"] != "PROJ" {
		t.Errorf("expected key=PROJ, got %v", result["key"])
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "my-repo" {
			t.Errorf("unexpected body: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{"slug": "my-repo"})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	body := map[string]any{"name": "my-repo"}
	var result map[string]any
	err := c.Post(context.Background(), "/projects/PROJ/repos", body, nil, &result)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if result["slug"] != "my-repo" {
		t.Errorf("expected slug=my-repo, got %v", result["slug"])
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{{"message": "Project not found"}},
		})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Get(context.Background(), "/projects/NOPE", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestClient_Retry429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_204NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Post(context.Background(), "/test", nil, nil, &result)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
}

func TestGetPaged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		w.Header().Set("Content-Type", "application/json")
		if start == "" || start == "0" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"values":        []map[string]any{{"key": "PROJ1"}, {"key": "PROJ2"}},
				"size":          2,
				"start":         0,
				"limit":         2,
				"isLastPage":    false,
				"nextPageStart": 2,
			})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"values":     []map[string]any{{"key": "PROJ3"}},
				"size":       1,
				"start":      2,
				"limit":      2,
				"isLastPage": true,
			})
		}
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	ctx := context.Background()

	page, err := GetPaged[Project](ctx, c, "/projects", nil, 0, 2)
	if err != nil {
		t.Fatalf("GetPaged error: %v", err)
	}
	if len(page.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(page.Values))
	}
	if page.IsLastPage {
		t.Error("expected IsLastPage=false")
	}

	all, err := GetAll[Project](ctx, c, "/projects", nil, 2)
	if err != nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 values, got %d", len(all))
	}
}
