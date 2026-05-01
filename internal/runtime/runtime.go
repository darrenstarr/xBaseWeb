package runtime

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/anomalyco/db4web/internal/compiler"
)

// Runtime is the execution engine for compiled xBase programs.
type Runtime struct {
	WorkAreas *WorkAreaManager
	Variables map[string]Value
	CallStack []string
	Screen    *Screen
	Paused    bool
	prog      *compiler.Program
	DB        *sql.DB // real database connection for USE/REPLACE/APPEND etc.
}

func New() *Runtime {
	return &Runtime{
		WorkAreas: NewWorkAreaManager(),
		Variables: make(map[string]Value),
		CallStack: make([]string, 0),
		Screen:    &Screen{},
	}
}

func (rt *Runtime) GetVar(name string) Value {
	if v, ok := rt.Variables[name]; ok {
		return v
	}
	return Nil()
}

func (rt *Runtime) SetVar(name string, val Value) {
	rt.Variables[name] = val
}

func (rt *Runtime) PushCall(name string) {
	rt.CallStack = append(rt.CallStack, name)
}

func (rt *Runtime) PopCall() {
	if len(rt.CallStack) > 0 {
		rt.CallStack = rt.CallStack[:len(rt.CallStack)-1]
	}
}

// ExecuteInteractive runs a program and returns the screen state at the
// first READ or WAIT pause point. If done is true, the program finished.
// If input is provided, it continues from a paused state by setting
// the READ field's variable, then resuming.
func (rt *Runtime) ExecuteInteractive(prog *compiler.Program, entryPoint string, input map[string]string) (*Screen, bool) {
	rt.Screen = &Screen{}
	rt.Paused = false
	rt.prog = prog

	for k, v := range input {
		rt.SetVar(k, NewString(v))
	}

	atGet := false
	found := false

	// Catch the initial readSentinel from the entry procedure
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(readSentinel); ok {
				rt.Paused = true
				return
			}
		}
	}()

	for _, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *compiler.ProcedureDef:
			if strings.EqualFold(s.Name, entryPoint) {
				found = true
				rt.execBody(s.Body, &atGet)
			}
		case *compiler.FunctionDef:
			if strings.EqualFold(s.Name, entryPoint) {
				found = true
				rt.execBody(s.Body, &atGet)
			}
		}
	}
	if !found {
		rt.Screen.Result = fmt.Sprintf("procedure %s not found", entryPoint)
		rt.Screen.Done = true
	}

	if rt.Paused {
		return rt.Screen, false
	}
	rt.Screen.Done = true
	return rt.Screen, true
}

// interactiveBlock runs a list of statements, stopping at READ/WAIT.
func (rt *Runtime) interactiveBlock(stmts []compiler.Stmt, atGet *bool) {
	// Wrap body execution so READ panics are caught and paused
	rt.execBody(stmts, atGet)
}

func (rt *Runtime) execBody(stmts []compiler.Stmt, atGet *bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(readSentinel); ok {
				rt.Paused = true
				// Don't re-panic for read — hand back to the main loop
				return
			}
			if _, ok := r.(returnSentinel); ok {
				// RETURN exits the block normally
				return
			}
			panic(r)
		}
	}()

	for _, stmt := range stmts {
		if rt.Paused {
			return
		}
		rt.interactiveStmt(stmt, atGet)
	}
}

// interactiveStmt dispatches a single statement in interactive mode.
func (rt *Runtime) interactiveStmt(stmt compiler.Stmt, atGet *bool) {
	switch s := stmt.(type) {
	case *compiler.ProcedureDef:
		rt.interactiveBlock(s.Body, atGet)

	case *compiler.FunctionDef:
		rt.interactiveBlock(s.Body, atGet)

	case *compiler.ReturnStmt:
		if s.Expr != nil {
			rt.Screen.Result = rt.evalExpr(s.Expr).String()
		}
		panic(returnSentinel{})

	case *compiler.AssignmentStmt:
		val := rt.evalExpr(s.Expr)
		rt.SetVar(s.Target, val)

	case *compiler.IfStmt:
		cond := rt.evalExpr(s.Condition)
		if cond.IsTruthy() {
			rt.execBody(s.ThenBody, atGet)
			if rt.Paused {
				return
			}
		} else if s.ElseBody != nil {
			rt.execBody(s.ElseBody, atGet)
			if rt.Paused {
				return
			}
		}

	case *compiler.WhileStmt:
		for rt.evalExpr(s.Condition).IsTruthy() {
			if rt.Paused {
				return
			}
			rt.interactiveBlock(s.Body, atGet)
		}

	case *compiler.ForStmt:
		start := rt.evalExpr(s.Start).AsInt()
		end := rt.evalExpr(s.End).AsInt()
		rt.SetVar(s.Var, NewNumber(float64(start)))
		for i := start; i <= end; i++ {
			if rt.Paused {
				return
			}
			rt.SetVar(s.Var, NewNumber(float64(i)))
			rt.interactiveBlock(s.Body, atGet)
		}

	case *compiler.ClearStmt:
		title, tagline, nav := rt.Screen.Title, rt.Screen.Tagline, rt.Screen.Nav
		rt.Screen = &Screen{}
		rt.Screen.Title, rt.Screen.Tagline = title, tagline
		rt.Screen.Nav = nav

	case *compiler.SayGetStmt:
		rt.interactiveSayGet(s, atGet)

	case *compiler.ReadStmt:
		// A READ signals we need user input for the last GET field.
		// If we're already tracking a GET, pause here.
		if *atGet {
			*atGet = false
			if rt.Screen.Prompt == "" && len(rt.Screen.Fields) > 0 {
				last := rt.Screen.Fields[len(rt.Screen.Fields)-1]
				rt.Screen.Prompt = last.Var
			}
			panic(readSentinel{})
		}

	case *compiler.WaitStmt:
		rt.Screen.Wait = true
		rt.Screen.Result = fmt.Sprintf("WAIT %q", s.Prompt)
		panic(readSentinel{})

	case *compiler.StoreStmt:
		val := rt.evalExpr(s.Expr)
		rt.SetVar(s.Var, val)

	case *compiler.CallStmt:
		rt.executeProcedure(s.Name, atGet)

	case *compiler.SetStmt:
		if len(s.Parts) >= 2 && strings.ToUpper(s.Parts[0]) == "TITLE" && strings.ToUpper(s.Parts[1]) == "TO" && len(s.Parts) >= 3 {
			rt.Screen.Title = s.Parts[2]
		} else if len(s.Parts) >= 2 && strings.ToUpper(s.Parts[0]) == "TAGLINE" && strings.ToUpper(s.Parts[1]) == "TO" && len(s.Parts) >= 3 {
			rt.Screen.Tagline = s.Parts[2]
		}

	case *compiler.UseStmt:
		rt.execUse(s)

	case *compiler.SelectStmt:
		rt.execSelect(s)

	case *compiler.ReplaceStmt:
		rt.execReplace(s)

	case *compiler.AppendStmt:
		rt.execAppend(s)

	case *compiler.SkipStmt:
		rt.execSkip(s)

	case *compiler.CloseStmt:
		rt.WorkAreas.CloseAll()

	case *compiler.GoStmt:
		rt.execGo(s)

	case *compiler.DeleteStmt:
		rt.execDelete(s)

	case *compiler.PackStmt:
		rt.execPack(s)

	case *compiler.LocateStmt:
		rt.execLocate(s)

	case *compiler.SeekStmt:
		if s.Expr != nil {
			rt.addLine(1, 1, fmt.Sprintf("SEEK %s", rt.evalExpr(s.Expr).String()))
		}

	case *compiler.InputStmt:
		rt.Screen.Fields = append(rt.Screen.Fields, ScreenField{
			Var:   s.Var,
			Type:  "get",
			Value: "",
		})
		rt.Screen.Prompt = s.Var
		panic(readSentinel{})

	case *compiler.AcceptStmt:
		rt.Screen.Fields = append(rt.Screen.Fields, ScreenField{
			Var:   s.Var,
			Type:  "get",
			Value: "",
		})
		rt.Screen.Prompt = s.Var
		panic(readSentinel{})

	case *compiler.NavStmt:
		rt.Screen.Nav = s.Entries

	case *compiler.ExecSQLStmt:
		rt.Screen.SQL = s.Query
		rt.Screen.Cols = s.Cols
		rt.Screen.Result = "ACTIONS:" // marker for server to add actions
		for _, a := range s.Actions {
			rt.Screen.Result += a.Label + ":" + a.Procedure + ","
		}
		panic(readSentinel{})

	default:
		if es, ok := stmt.(*compiler.ExprStmt); ok {
			rt.Screen.Result = rt.evalExpr(es.Expr).String()
		} else {
			// Unknown statement type — try EvalStmt for safety
			rt.EvalStmt(stmt)
		}
	}
}

// interactiveSayGet processes @ SAY / GET for screen rendering.
func (rt *Runtime) interactiveSayGet(s *compiler.SayGetStmt, atGet *bool) {
	if s.Row != nil {
		rowVal := rt.evalExpr(s.Row)
		colVal := rt.evalExpr(s.Col)
		row := int(rowVal.AsInt())
		col := int(colVal.AsInt())

		if s.SayExpr != nil {
			text := rt.evalExpr(s.SayExpr).String()
			rt.addLine(row, col, text)
		}

		if s.GetVar != "" {
			*atGet = true
			picture := s.Picture
			rt.Screen.Fields = append(rt.Screen.Fields, ScreenField{
				Row:     row,
				Col:     col,
				Var:     s.GetVar,
				Type:    "get",
				Picture: picture,
				Value:   rt.getVarDisplay(s.GetVar),
			})
		}
	} else {
		if s.SayExpr != nil {
			rt.addLine(0, 0, rt.evalExpr(s.SayExpr).String())
		}
		if s.GetVar != "" {
			*atGet = true
			picture := s.Picture
			rt.Screen.Fields = append(rt.Screen.Fields, ScreenField{
				Var:     s.GetVar,
				Type:    "get",
				Picture: picture,
				Value:   rt.getVarDisplay(s.GetVar),
			})
		}
	}
}

func (rt *Runtime) executeProcedure(name string, atGet *bool) {
	if rt.prog == nil {
		rt.addLine(1, 1, fmt.Sprintf("call %s (no program loaded)", name))
		return
	}
	for _, stmt := range rt.prog.Stmts {
		switch s := stmt.(type) {
		case *compiler.ProcedureDef:
			if strings.EqualFold(s.Name, name) {
				rt.PushCall(name)
				rt.execBody(s.Body, atGet)
				rt.PopCall()
				return
			}
		case *compiler.FunctionDef:
			if strings.EqualFold(s.Name, name) {
				rt.PushCall(name)
				rt.execBody(s.Body, atGet)
				rt.PopCall()
				return
			}
		}
	}
	rt.addLine(1, 1, fmt.Sprintf("procedure %s not found", name))
}

func (rt *Runtime) getVarDisplay(name string) string {
	v := rt.GetVar(name)
	if v.Type == TypeNil {
		return ""
	}
	return v.String()
}

// ─── Real database operations ──────────────────────────────

func (rt *Runtime) execUse(s *compiler.UseStmt) {
	if rt.DB == nil {
		rt.addLine(1, 1, "USE (no database)")
		return
	}
	alias := s.Alias
	if alias == "" {
		alias = s.Table
	}
	wa := rt.WorkAreas.Use(alias, s.Table)
	// Load all fields for the current work area
	rows, err := rt.DB.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 1", s.Table))
	if err != nil {
		rt.addLine(1, 1, fmt.Sprintf("USE error: %v", err))
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	for _, c := range cols {
		wa.Fields[c] = Nil()
	}
	if rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		for i, col := range cols {
			switch x := vals[i].(type) {
			case []byte:
				wa.Fields[col] = NewString(string(x))
			case int64:
				wa.Fields[col] = NewNumber(float64(x))
			case float64:
				wa.Fields[col] = NewNumber(x)
			default:
				wa.Fields[col] = Nil()
			}
		}
		wa.RecNo = 1
		wa.LastRec = rt.countTable(s.Table)
		wa.BOF = true
		wa.EOF = wa.LastRec == 0
	}
}

func (rt *Runtime) countTable(table string) int64 {
	var n int64
	rt.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
	return n
}

func (rt *Runtime) execSelect(s *compiler.SelectStmt) {
	if s.Expr == nil {
		return
	}
	val := rt.evalExpr(s.Expr)
	if val.Type == TypeNumber {
		rt.WorkAreas.Select(int(val.Number))
	} else if val.Type == TypeString {
		rt.WorkAreas.SelectAlias(val.Str)
	}
}

func (rt *Runtime) execReplace(s *compiler.ReplaceStmt) {
	if rt.DB == nil {
		return
	}
	val := rt.evalExpr(s.Expr)
	field := s.Field
	// Find the work area by alias (or use current)
	wa := rt.WorkAreas.Current()
	if s.Alias != "" {
		if a := rt.WorkAreas.Get(s.Alias); a != nil {
			wa = a
		}
	}
	if wa == nil || wa.TableName == "" {
		return
	}
	// Update in-memory field
	if wa.RecNo > 0 {
		wa.Fields[field] = val
	}
	// Persist to database
	rt.DB.Exec(fmt.Sprintf("UPDATE %s SET %s = ? WHERE rowid = ?", wa.TableName, field), val.String(), wa.RecNo)
}

func (rt *Runtime) execAppend(s *compiler.AppendStmt) {
	if rt.DB == nil || !s.Blank {
		return
	}
	wa := rt.WorkAreas.Current()
	if wa == nil || wa.TableName == "" {
		return
	}
	// Get column list for the table
	colRows, err := rt.DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", wa.TableName))
	if err != nil {
		rt.addLine(1, 1, fmt.Sprintf("APPEND error: %v", err))
		return
	}
	var cols []string
	var placeholders []string
	var vals []interface{}
	for colRows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		colRows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk)
		if name == "id" {
			continue // skip auto-increment PK
		}
		cols = append(cols, name)
		placeholders = append(placeholders, "?")
		if colType == "INTEGER" {
			vals = append(vals, 0)
		} else {
			vals = append(vals, "")
		}
	}
	colRows.Close()

	if len(cols) == 0 {
		return
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", wa.TableName,
		strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	result, err := rt.DB.Exec(query, vals...)
	if err != nil {
		rt.addLine(1, 1, fmt.Sprintf("APPEND error: %v", err))
		return
	}
	id, _ := result.LastInsertId()
	wa.LastRec = rt.countTable(wa.TableName)
	wa.RecNo = id
	wa.BOF = false
	wa.EOF = false
	// Reload fields
	for k := range wa.Fields {
		wa.Fields[k] = Nil()
	}
}

func (rt *Runtime) execSkip(s *compiler.SkipStmt) {
	n := int64(1)
	if s.Count != nil {
		n = rt.evalExpr(s.Count).AsInt()
	}
	wa := rt.WorkAreas.Current()
	if wa == nil || wa.TableName == "" {
		return
	}
	// Move to next/prev record
	target := wa.RecNo + n
	if target < 1 {
		target = 1
	}
	if target > wa.LastRec {
		target = wa.LastRec + 1
	}
	wa.RecNo = target
	wa.BOF = target <= 1
	wa.EOF = target > wa.LastRec
	// Reload fields for the new record
	if target > 0 && target <= wa.LastRec && rt.DB != nil {
		rows, err := rt.DB.Query(fmt.Sprintf("SELECT * FROM %s WHERE rowid = ?", wa.TableName), target)
		if err == nil {
			defer rows.Close()
			cols, _ := rows.Columns()
			if rows.Next() {
				vals := make([]interface{}, len(cols))
				ptrs := make([]interface{}, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)
				for i, col := range cols {
					switch x := vals[i].(type) {
					case []byte:
						wa.Fields[col] = NewString(string(x))
					case int64:
						wa.Fields[col] = NewNumber(float64(x))
					case float64:
						wa.Fields[col] = NewNumber(x)
					default:
						wa.Fields[col] = Nil()
					}
				}
			}
		}
	}
}

func (rt *Runtime) execGo(s *compiler.GoStmt) {
	wa := rt.WorkAreas.Current()
	if wa == nil || wa.TableName == "" {
		return
	}
	upper := strings.ToUpper(s.Pos)
	if upper == "TOP" {
		wa.GoTop()
	} else if upper == "BOTTOM" {
		wa.GoBottom()
	} else {
		n := parseInt64(s.Pos)
		if n > 0 {
			wa.GoTo(n)
		}
	}
}

func parseInt64(s string) int64 {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}

func (rt *Runtime) execDelete(s *compiler.DeleteStmt) {
	wa := rt.WorkAreas.Current()
	if wa == nil || wa.TableName == "" || wa.RecNo <= 0 || rt.DB == nil {
		return
	}
	rt.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE rowid = ?", wa.TableName), wa.RecNo)
	wa.LastRec = rt.countTable(wa.TableName)
}

func (rt *Runtime) execPack(s *compiler.PackStmt) {
	// SQLite handles this automatically — VACUUM if needed
}

func (rt *Runtime) execLocate(s *compiler.LocateStmt) {
	wa := rt.WorkAreas.Current()
	if wa == nil || wa.TableName == "" || rt.DB == nil {
		return
	}
	// Simple locate: scan sequentially until condition matches
	for wa.RecNo = 1; wa.RecNo <= wa.LastRec; wa.RecNo++ {
		wa.BOF = wa.RecNo == 1
		wa.EOF = false
		// Reload fields
		rows, err := rt.DB.Query(fmt.Sprintf("SELECT * FROM %s WHERE rowid = ?", wa.TableName), wa.RecNo)
		if err == nil {
			cols, _ := rows.Columns()
			if rows.Next() {
				vals := make([]interface{}, len(cols))
				ptrs := make([]interface{}, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)
				for i, col := range cols {
					switch x := vals[i].(type) {
					case []byte:
						wa.Fields[col] = NewString(string(x))
					case int64:
						wa.Fields[col] = NewNumber(float64(x))
					case float64:
						wa.Fields[col] = NewNumber(x)
					default:
						wa.Fields[col] = Nil()
					}
				}
			}
			rows.Close()
		}
		// Check FOR condition (or true if no condition)
		match := true
		if s.For != nil {
			match = rt.evalExpr(s.For).IsTruthy()
		}
		if match {
			wa.Found = true
			return
		}
	}
	wa.Found = false
	wa.EOF = true
}

// ─── Screen helpers ─────────────────────────────────────────

func (rt *Runtime) addLine(row, col int, text string) {
	rt.Screen.Lines = append(rt.Screen.Lines, ScreenLine{
		Row: row, Col: col, Text: text,
	})
}

// EvalStmt evaluates a statement and returns the result value.
func (rt *Runtime) EvalStmt(stmt compiler.Stmt) Value {
	switch s := stmt.(type) {
	case *compiler.ReturnStmt:
		if s.Expr != nil {
			return rt.evalExpr(s.Expr)
		}
	case *compiler.AssignmentStmt:
		val := rt.evalExpr(s.Expr)
		rt.SetVar(s.Target, val)
		return val
	case *compiler.CallStmt:
		return NewString(fmt.Sprintf("call %s(%d args)", s.Name, len(s.Args)))
	case *compiler.ClearStmt:
		return NewString("CLEAR")
	case *compiler.WaitStmt:
		return NewString(fmt.Sprintf("WAIT %q", s.Prompt))
	}
	return Nil()
}

// evalExpr evaluates an expression and returns its Value.
func (rt *Runtime) evalExpr(expr compiler.Expr) Value {
	switch e := expr.(type) {
	case *compiler.NumberExpr:
		if e.IsInt {
			return NewNumber(float64(e.IntValue))
		}
		return NewNumber(e.Value)

	case *compiler.StringExpr:
		return NewString(e.Value)

	case *compiler.BoolExpr:
		return NewLogical(e.Value)

	case *compiler.IdentExpr:
		if v := rt.GetVar(e.Name); v.Type != TypeNil {
			return v
		}
		wa := rt.WorkAreas.Current()
		if v := wa.GetField(e.Name); v.Type != TypeNil {
			return v
		}
		return Nil()

	case *compiler.FieldRefExpr:
		if wa := rt.WorkAreas.Get(e.Alias); wa != nil {
			return wa.GetField(e.Field)
		}
		return Nil()

	case *compiler.UnaryExpr:
		inner := rt.evalExpr(e.Inner)
		switch e.Op {
		case compiler.T_MINUS:
			if inner.Type == TypeNumber {
				return NewNumber(-inner.Number)
			}
		case compiler.T_NOT:
			return inner.Not()
		}
		return Nil()

	case *compiler.BinaryExpr:
		left := rt.evalExpr(e.Left)
		right := rt.evalExpr(e.Right)
		switch e.Op {
		case compiler.T_PLUS:
			return left.Add(right)
		case compiler.T_MINUS:
			return left.Sub(right)
		case compiler.T_STAR:
			return left.Mul(right)
		case compiler.T_SLASH:
			return left.Div(right)
		case compiler.T_CARET:
			return left.Pow(right)
		case compiler.T_CONCAT:
			return left.Concat(right)
		case compiler.T_EQ:
			return left.Eq(right)
		case compiler.T_NE:
			return left.Neq(right)
		case compiler.T_LT:
			return left.Lt(right)
		case compiler.T_LE:
			return left.Le(right)
		case compiler.T_GT:
			return left.Gt(right)
		case compiler.T_GE:
			return left.Ge(right)
		case compiler.T_AND:
			return left.And(right)
		case compiler.T_OR:
			return left.Or(right)
		case compiler.T_DOLLAR:
			return left.Contains(right)
		}
		return Nil()

	case *compiler.FuncCallExpr:
		rt.Screen.Result = fmt.Sprintf("%s(%d args)", e.Name, len(e.Args))
		return NewString(rt.Screen.Result)
	}

	return NewString(fmt.Sprintf("expr %T", expr))
}

type returnSentinel struct{}
type readSentinel struct{}
