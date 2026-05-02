package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anomalyco/db4web/internal/compiler"
	"github.com/anomalyco/db4web/internal/runtime"
)

// ─── Interpreter (execute .prg in real time) ────────────────

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Program   string                 `json:"program"`
		Procedure string                 `json:"procedure"`
		State     map[string]interface{} `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorJSON(w, "invalid request body", 400)
		return
	}
	if req.Program == "" {
		s.errorJSON(w, "program is required", 400)
		return
	}
	s.executeProgram(w, req.Program, req.Procedure, req.State)
}

func (s *Server) handleExecuteProgram(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	s.executeProgram(w, "examples/"+name+".prg", "", nil)
}

func (s *Server) executeProgram(w http.ResponseWriter, programPath, entryPoint string, state map[string]interface{}) {
	prg, err := os.ReadFile(programPath)
	if err != nil {
		s.errorJSON(w, fmt.Sprintf("cannot read program %s: %v", programPath, err), 404)
		return
	}
	combinedSource := string(prg)

	lex := compiler.NewLexer(combinedSource)
	tokens, lexErrs := lex.Lex()
	if len(lexErrs) > 0 {
		s.errorJSON(w, fmt.Sprintf("lex errors: %v", lexErrs), 400)
		return
	}

	parser := compiler.NewParser(tokens)
	program, parseErrs := parser.Parse()
	if len(parseErrs) > 0 {
		s.errorJSON(w, fmt.Sprintf("parse errors: %v", parseErrs), 400)
		return
	}

	input := make(map[string]string)
	if state != nil {
		if m, ok := state["vars"].(map[string]interface{}); ok {
			for k, v := range m {
				input[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	if state != nil {
		if m, ok := state["input"].(map[string]interface{}); ok {
			for k, v := range m {
				input[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	rt := runtime.New()
	rt.DB = s.db.Conn()
	screen, done := rt.ExecuteInteractive(program, entryPoint, input)

	// If the screen requested an SQL query, execute it and populate Table
	if screen != nil && screen.SQL != "" {
		limitedSQL := screen.SQL + " LIMIT 50"
		rows, err := s.db.Query(limitedSQL)
		if err == nil {
			dbCols, _ := rows.Columns()
			allData := make([][]string, 0)
			for rows.Next() {
				vals := make([]interface{}, len(dbCols))
				ptrs := make([]interface{}, len(dbCols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)
				row := make([]string, len(dbCols))
				for i, v := range vals {
					switch x := v.(type) {
					case []byte:
						row[i] = string(x)
					case int64:
						row[i] = fmt.Sprintf("%d", x)
					case float64:
						row[i] = fmt.Sprintf("%.2f", x)
					default:
						row[i] = fmt.Sprintf("%v", x)
					}
				}
				allData = append(allData, row)
			}
			rows.Close()
			table := &runtime.TableData{}
			if len(screen.Cols) > 0 {
				for _, c := range screen.Cols {
					table.Columns = append(table.Columns, runtime.TableColumn{Name: c})
				}
			} else {
				for _, c := range dbCols {
					table.Columns = append(table.Columns, runtime.TableColumn{Name: c})
				}
			}
			table.Rows = allData
			table.Query = screen.SQL
			table.Offset = 0
			table.Limit = 50
			// Get total count
			var total int
			countSQL := "SELECT COUNT(*) FROM (" + screen.SQL + ")"
			s.db.QueryRow(countSQL).Scan(&total)
			table.Total = total
			// Extract row actions from Result marker
			if strings.HasPrefix(screen.Result, "ACTIONS:") {
				parts := strings.Split(screen.Result[8:], ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					seg := strings.SplitN(p, ":", 2)
					if len(seg) == 2 {
						table.Actions = append(table.Actions, runtime.RowAction{Label: seg[0], Procedure: seg[1]})
					}
				}
				table.KeyCol = 0
			}
			screen.Table = table
			if len(allData) == 0 {
				screen.Lines = append(screen.Lines, runtime.ScreenLine{Row: 0, Col: 0, Text: "No records found."})
			}
		} else {
			screen.Lines = append(screen.Lines, runtime.ScreenLine{Row: 0, Col: 0, Text: fmt.Sprintf("SQL error: %v", err)})
		}
		screen.SQL = ""
		screen.Done = true
	}

	finalDone := done
	if screen != nil && screen.Done {
		finalDone = true
	}
	s.json(w, map[string]interface{}{
		"screen": screen,
		"done":   finalDone,
		"vars":   rt.Variables,
	})
}

// GET /api/page — load next page for infinite scroll
func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query  string `json:"query"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Sort   string `json:"sort,omitempty"`
		Dir    string `json:"dir,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorJSON(w, "invalid request", 400)
		return
	}
	if req.Query == "" {
		s.errorJSON(w, "query required", 400)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	sql := req.Query
	// Strip any existing ORDER BY when a sort is requested
	if req.Sort != "" {
		idx := strings.Index(strings.ToUpper(sql), "ORDER BY")
		if idx >= 0 {
			sql = sql[:idx]
		}
		dir := "ASC"
		if req.Dir == "desc" {
			dir = "DESC"
		}
		sql += fmt.Sprintf(" ORDER BY %s %s", req.Sort, dir)
	}
	sql += fmt.Sprintf(" LIMIT %d OFFSET %d", req.Limit, req.Offset)

	rows, err := s.db.Query(sql)
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
	s.json(w, map[string]interface{}{"rows": results, "offset": req.Offset, "limit": req.Limit})
}
