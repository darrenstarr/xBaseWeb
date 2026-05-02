package compiler

type Node interface {
	nodeMarker()
}

// ----- Statements -----

type Stmt interface {
	Node
	stmtNode()
}

type Program struct {
	Stmts []Stmt
}

func (p *Program) nodeMarker() {}

type UseStmt struct {
	Table  string
	Alias  string
	In     int64
	Shared bool
	Exclusive bool
	NoUpdate  bool
}

func (s *UseStmt) nodeMarker()   {}
func (s *UseStmt) stmtNode()     {}

type SelectStmt struct {
	Expr Expr
}

func (s *SelectStmt) nodeMarker() {}
func (s *SelectStmt) stmtNode()   {}

type ReplaceStmt struct {
	Field string
	Alias string
	Expr  Expr
}

func (s *ReplaceStmt) nodeMarker() {}
func (s *ReplaceStmt) stmtNode()   {}

type AppendStmt struct {
	Blank bool
}

func (s *AppendStmt) nodeMarker() {}
func (s *AppendStmt) stmtNode()   {}

type DeleteStmt struct {
	Scope string
	While Expr
	For   Expr
}

func (s *DeleteStmt) nodeMarker() {}
func (s *DeleteStmt) stmtNode()   {}

type PackStmt struct{}

func (s *PackStmt) nodeMarker() {}
func (s *PackStmt) stmtNode()   {}

type ZapStmt struct{}

func (s *ZapStmt) nodeMarker() {}
func (s *ZapStmt) stmtNode()   {}

type GoStmt struct {
	Pos  string
	Expr Expr // optional expression for GO VAL(mId) etc.
}

func (s *GoStmt) nodeMarker() {}
func (s *GoStmt) stmtNode()   {}

type SkipStmt struct {
	Count Expr
}

func (s *SkipStmt) nodeMarker() {}
func (s *SkipStmt) stmtNode()   {}

type LocateStmt struct {
	Scope string
	For   Expr
	While Expr
}

func (s *LocateStmt) nodeMarker() {}
func (s *LocateStmt) stmtNode()   {}

type ContinueStmt struct{}

func (s *ContinueStmt) nodeMarker() {}
func (s *ContinueStmt) stmtNode()   {}

type SeekStmt struct {
	Expr Expr
}

func (s *SeekStmt) nodeMarker() {}
func (s *SeekStmt) stmtNode()   {}

type CloseStmt struct {
	Databases bool
	All       bool
	Alias     string
}

func (s *CloseStmt) nodeMarker() {}
func (s *CloseStmt) stmtNode()   {}

type StoreStmt struct {
	Expr Expr
	Var  string
}

func (s *StoreStmt) nodeMarker() {}
func (s *StoreStmt) stmtNode()   {}

type InputStmt struct {
	Prompt string
	Var    string
}

func (s *InputStmt) nodeMarker() {}
func (s *InputStmt) stmtNode()   {}

type AcceptStmt struct {
	Prompt string
	Var    string
}

func (s *AcceptStmt) nodeMarker() {}
func (s *AcceptStmt) stmtNode()   {}

type WaitStmt struct {
	Prompt string
	Var    string
}

func (s *WaitStmt) nodeMarker() {}
func (s *WaitStmt) stmtNode()   {}

type ClearStmt struct{}

func (s *ClearStmt) nodeMarker() {}
func (s *ClearStmt) stmtNode()   {}

type CountStmt struct {
	Scope string
	For   Expr
	While Expr
	To    string
}

func (s *CountStmt) nodeMarker() {}
func (s *CountStmt) stmtNode()   {}

type SumStmt struct {
	Expr  Expr
	Scope string
	For   Expr
	While Expr
	To    string
}

func (s *SumStmt) nodeMarker() {}
func (s *SumStmt) stmtNode()   {}

type CallStmt struct {
	Name    string
	Args    []Expr
	With    bool
}

func (s *CallStmt) nodeMarker() {}
func (s *CallStmt) stmtNode()   {}

type ReturnStmt struct {
	Expr Expr
}

func (s *ReturnStmt) nodeMarker() {}
func (s *ReturnStmt) stmtNode()   {}

type AssignmentStmt struct {
	Target string
	Expr   Expr
}

func (s *AssignmentStmt) nodeMarker() {}
func (s *AssignmentStmt) stmtNode()   {}
func (s *AssignmentStmt) exprNode()   {}

type IfStmt struct {
	Condition  Expr
	ThenBody   []Stmt
	ElseBody   []Stmt
}

func (s *IfStmt) nodeMarker() {}
func (s *IfStmt) stmtNode()   {}

type WhileStmt struct {
	Condition Expr
	Body      []Stmt
}

func (s *WhileStmt) nodeMarker() {}
func (s *WhileStmt) stmtNode()   {}

type ForStmt struct {
	Var    string
	Start  Expr
	End    Expr
	Step   Expr
	Body   []Stmt
}

func (s *ForStmt) nodeMarker() {}
func (s *ForStmt) stmtNode()   {}

type ProcedureDef struct {
	Name string
	Body []Stmt
}

func (s *ProcedureDef) nodeMarker() {}
func (s *ProcedureDef) stmtNode()   {}

type FunctionDef struct {
	Name string
	Body []Stmt
}

func (s *FunctionDef) nodeMarker() {}
func (s *FunctionDef) stmtNode()   {}

type ReadStmt struct{}

type NavStmt struct {
	Entries map[string]string // choice -> procedure
}

func (s *NavStmt) nodeMarker() {}
func (s *NavStmt) stmtNode()   {}

type SetStmt struct {
	Parts []string
}

func (s *SetStmt) nodeMarker() {}
func (s *SetStmt) stmtNode()   {}

type RowActionDef struct {
	Label     string
	Procedure string
}

type ExecSQLStmt struct {
	Query   string
	Cols    []string
	Actions []RowActionDef
}

func (s *ExecSQLStmt) nodeMarker() {}
func (s *ExecSQLStmt) stmtNode()   {}

func (s *ReadStmt) nodeMarker() {}
func (s *ReadStmt) stmtNode()   {}

type SayGetStmt struct {
	Row     Expr
	Col     Expr
	SayExpr Expr
	GetVar  string
	GetExpr Expr
	Picture string
}

func (s *SayGetStmt) nodeMarker() {}
func (s *SayGetStmt) stmtNode()   {}

// ----- Expressions -----

type Expr interface {
	Node
	exprNode()
}

type NumberExpr struct {
	Value    float64
	IntValue int64
	IsInt    bool
}

func (e *NumberExpr) nodeMarker() {}
func (e *NumberExpr) exprNode()   {}

type StringExpr struct {
	Value string
}

func (e *StringExpr) nodeMarker() {}
func (e *StringExpr) exprNode()   {}

type BoolExpr struct {
	Value bool
}

func (e *BoolExpr) nodeMarker() {}
func (e *BoolExpr) exprNode()   {}

type IdentExpr struct {
	Name string
}

func (e *IdentExpr) nodeMarker() {}
func (e *IdentExpr) exprNode()   {}

type FieldRefExpr struct {
	Alias string
	Field string
}

func (e *FieldRefExpr) nodeMarker() {}
func (e *FieldRefExpr) exprNode()   {}

type UnaryExpr struct {
	Op    TokenType
	Inner Expr
}

func (e *UnaryExpr) nodeMarker() {}
func (e *UnaryExpr) exprNode()   {}

type BinaryExpr struct {
	Left  Expr
	Op    TokenType
	Right Expr
}

func (e *BinaryExpr) nodeMarker() {}
func (e *BinaryExpr) exprNode()   {}

type FuncCallExpr struct {
	Name string
	Args []Expr
}

func (e *FuncCallExpr) nodeMarker() {}
func (e *FuncCallExpr) exprNode()   {}
