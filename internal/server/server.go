package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anomalyco/db4web/internal/sqlite"
)

type Server struct {
	db  *sqlite.DB
	mux *http.ServeMux
}

func New(dbPath string) (*Server, error) {
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}
	s := &Server{db: db, mux: http.NewServeMux()}
	s.routes()
	return s, nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/tables", s.handleListTables)
	s.mux.HandleFunc("GET /api/workspace", s.handleWorkspace)
	s.mux.HandleFunc("GET /api/forms/{name}", s.handleGetForm)
	s.mux.HandleFunc("GET /api/query", s.handleQuery)
	s.mux.HandleFunc("POST /api/execute", s.handleExecute)
	s.mux.HandleFunc("GET /api/execute/{name}", s.handleExecuteProgram)
	s.mux.HandleFunc("POST /api/page", s.handlePage)
	s.mux.HandleFunc("GET /api/data/{table}", s.handleDataList)
	s.mux.HandleFunc("GET /api/data/{table}/{id}", s.handleDataGet)
	s.mux.HandleFunc("POST /api/data/{table}", s.handleDataCreate)
	s.mux.HandleFunc("PUT /api/data/{table}/{id}", s.handleDataUpdate)
	s.mux.HandleFunc("DELETE /api/data/{table}/{id}", s.handleDataDelete)
}

func (s *Server) Handler() http.Handler {
	return withMiddleware(s.mux)
}

func (s *Server) json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) errorJSON(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func sanitizeName(s string) string {
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' {
			out = append(out, c)
		}
	}
	return string(out)
}

// ─── Infrastructure endpoints ──────────────────────────────

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.json(w, map[string]string{"status": "ok", "version": "0.1.0"})
}

func (s *Server) handleListTables(w http.ResponseWriter, r *http.Request) {
	tables, err := s.db.Tables()
	if err != nil {
		s.errorJSON(w, err.Error(), 500)
		return
	}
	s.json(w, map[string][]string{"tables": tables})
}

func (s *Server) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "examples/cureforwoke/app.json")
}

func (s *Server) handleGetForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, fmt.Sprintf("examples/cureforwoke/forms/%s.json", r.PathValue("name")))
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("sql")
	if query == "" {
		s.errorJSON(w, "sql parameter required", 400)
		return
	}
	if strings.Contains(query, ";") {
		s.errorJSON(w, "multi-statement queries not allowed", 400)
		return
	}
	trimmed := strings.TrimSpace(query)
	if len(trimmed) < 6 || strings.ToUpper(trimmed[:6]) != "SELECT" {
		s.errorJSON(w, "only SELECT queries allowed", 400)
		return
	}
	rows, err := s.db.Query(query)
	if err != nil {
		s.errorJSON(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		row := make(map[string]interface{})
		for i, col := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		results = append(results, row)
	}
	s.json(w, map[string]interface{}{"columns": cols, "rows": results})
}

// ─── Generic Data CRUD (sanitized) ──────────────────────────

func (s *Server) handleDataList(w http.ResponseWriter, r *http.Request) {
	table := sanitizeName(r.PathValue("table"))
	if table == "" {
		s.errorJSON(w, "invalid table name", 400)
		return
	}
	rows, err := s.db.Query(fmt.Sprintf("SELECT * FROM %s ORDER BY id", table))
	if err != nil {
		s.errorJSON(w, err.Error(), 400)
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		row := make(map[string]interface{})
		for i, col := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		results = append(results, row)
	}
	s.json(w, map[string]interface{}{"rows": results})
}

func (s *Server) handleDataGet(w http.ResponseWriter, r *http.Request) {
	table := sanitizeName(r.PathValue("table"))
	id := r.PathValue("id")
	if table == "" {
		s.errorJSON(w, "invalid table name", 400)
		return
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
	rows, err := s.db.Query(query, id)
	if err != nil {
		s.errorJSON(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	if rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		row := make(map[string]interface{})
		for i, col := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		s.json(w, row)
		return
	}
	s.errorJSON(w, "not found", 404)
}

func (s *Server) handleDataCreate(w http.ResponseWriter, r *http.Request) {
	table := sanitizeName(r.PathValue("table"))
	if table == "" {
		s.errorJSON(w, "invalid table name", 400)
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.errorJSON(w, "invalid JSON body", 400)
		return
	}
	var cols []string
	var vals []interface{}
	ph := []string{}
	for k, v := range body {
		col := sanitizeName(k)
		if col == "" || k == "id" {
			continue
		}
		cols = append(cols, col)
		vals = append(vals, v)
		ph = append(ph, "?")
	}
	if len(cols) == 0 {
		s.errorJSON(w, "no fields", 400)
		return
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(ph, ", "))
	result, err := s.db.Exec(query, vals...)
	if err != nil {
		s.errorJSON(w, err.Error(), 500)
		return
	}
	newID, _ := result.LastInsertId()
	s.json(w, map[string]interface{}{"id": newID, "status": "created"})
}

func (s *Server) handleDataUpdate(w http.ResponseWriter, r *http.Request) {
	table := sanitizeName(r.PathValue("table"))
	if table == "" {
		s.errorJSON(w, "invalid table name", 400)
		return
	}
	id := r.PathValue("id")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.errorJSON(w, "invalid JSON body", 400)
		return
	}
	var setClauses []string
	var vals []interface{}
	for k, v := range body {
		col := sanitizeName(k)
		if col == "" || k == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		vals = append(vals, v)
	}
	if len(setClauses) == 0 {
		s.errorJSON(w, "no fields", 400)
		return
	}
	vals = append(vals, id)
	_, err := s.db.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(setClauses, ", ")), vals...)
	if err != nil {
		s.errorJSON(w, err.Error(), 500)
		return
	}
	s.json(w, map[string]interface{}{"status": "updated"})
}

func (s *Server) handleDataDelete(w http.ResponseWriter, r *http.Request) {
	table := sanitizeName(r.PathValue("table"))
	if table == "" {
		s.errorJSON(w, "invalid table name", 400)
		return
	}
	id := r.PathValue("id")
	intID, err := strconv.Atoi(id)
	if err == nil {
		_, err = s.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", table), intID)
	} else {
		_, err = s.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", table), id)
	}
	if err != nil {
		s.errorJSON(w, err.Error(), 500)
		return
	}
	s.json(w, map[string]interface{}{"status": "deleted"})
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lrw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.status, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}
