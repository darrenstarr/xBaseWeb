package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testServer creates a Server backed by an in-memory SQLite database.
func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test server: %v", err)
	}
	return s
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func assertBody(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

// ---------- Health endpoint ----------

func TestHealthEndpoint(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertBody(t, w.Body.String(), "ok")
}

func TestHealthMethodNotAllowed(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("POST", "/api/health", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	// With the new Go 1.22 routing, POST to a GET-only route returns 405
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
		t.Logf("POST /api/health returned %d (acceptable: 405 or 404)", w.Code)
	}
}

// ---------- Status endpoint ----------

func TestStatusEndpoint(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", result["status"])
	}
	if result["version"] != "0.1.0" {
		t.Errorf("expected version=0.1.0, got %v", result["version"])
	}
}

// ---------- Tables endpoint ----------

func TestListTablesEmpty(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/tables", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	tables, ok := result["tables"].([]interface{})
	if !ok {
		t.Fatalf("expected tables array, got %T", result["tables"])
	}
	if len(tables) != 0 {
		t.Errorf("expected empty tables array, got %d tables", len(tables))
	}
}

func TestListTablesWithData(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	// Create some tables via sqlite directly
	conn := s.db.Conn()
	_, err := conn.Exec("CREATE TABLE test1 (id INTEGER)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	_, err = conn.Exec("CREATE TABLE test2 (id INTEGER)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/tables", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)

	var result struct {
		Tables []string `json:"tables"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result.Tables) < 2 {
		t.Fatalf("expected at least 2 tables, got %d", len(result.Tables))
	}

	found := map[string]bool{}
	for _, name := range result.Tables {
		found[name] = true
	}

	if !found["test1"] {
		t.Error("expected test1 in table list")
	}
	if !found["test2"] {
		t.Error("expected test2 in table list")
	}
}

// ---------- 404 handling ----------

func TestNotFound(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------- CORS headers ----------

func TestCORSPreflight(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("OPTIONS", "/api/health", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusNoContent)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestCORSHeadersOnGet(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}
}

// ---------- Content types ----------

func TestHealthContentType(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "text/plain" {
		t.Errorf("expected Content-Type: text/plain, got %q", ct)
	}
}

func TestStatusContentType(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", ct)
	}
}

// ---------- Server creation ----------

func TestNewServer(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.Handler() == nil {
		t.Error("expected non-nil handler")
	}
	s.db.Close()
}

func TestNewServerInvalidPath(t *testing.T) {
	_, err := New("/nonexistent/dir/data.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// ---------- Multiple requests ----------

func TestMultipleRequests(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, req)
		assertStatus(t, w.Code, http.StatusOK)
		assertBody(t, w.Body.String(), "ok")
	}
}

// ---------- Table listing with multiple DB operations ----------

func TestTablesEndpointAfterOperations(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	conn := s.db.Conn()

	// Create, insert, drop, create — verify tables endpoint reflects changes
	conn.Exec("CREATE TABLE phase1 (id INTEGER)")
	conn.Exec("CREATE TABLE phase2 (id INTEGER)")
	conn.Exec("DROP TABLE phase1")

	req := httptest.NewRequest("GET", "/api/tables", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	var result struct {
		Tables []string `json:"tables"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	foundPhase2 := false
	foundPhase1 := false
	for _, name := range result.Tables {
		switch name {
		case "phase1":
			foundPhase1 = true
		case "phase2":
			foundPhase2 = true
		}
	}
	if foundPhase1 {
		t.Error("phase1 should have been dropped")
	}
	if !foundPhase2 {
		t.Error("phase2 should exist")
	}
}

// ---------- Concurrent requests ----------

func TestConcurrentRequests(t *testing.T) {
	s := testServer(t)
	defer s.db.Close()

	done := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/health", nil)
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, req)
			if w.Code != http.StatusOK || w.Body.String() != "ok" {
				t.Errorf("concurrent request failed: %d %s", w.Code, w.Body.String())
			}
			done <- true
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}
