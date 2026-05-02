package compiler

import (
	"testing"
)

func parse(input string) (*Program, []string) {
	l := NewLexer(input)
	tokens, lexErrs := l.Lex()
	if len(lexErrs) > 0 {
		return nil, lexErrs
	}
	p := NewParser(tokens)
	return p.Parse()
}

func assertNoErrors(t *testing.T, errs []string) {
	t.Helper()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func assertStmtCount(t *testing.T, prog *Program, n int) {
	t.Helper()
	if prog == nil {
		t.Fatalf("program is nil")
	}
	if len(prog.Stmts) != n {
		t.Fatalf("expected %d statements, got %d", n, len(prog.Stmts))
	}
}

// ---------- Basic statements ----------

func TestParseUse(t *testing.T) {
	prog, errs := parse("USE customers")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
	_, ok := prog.Stmts[0].(*UseStmt)
	if !ok {
		t.Fatalf("expected *UseStmt, got %T", prog.Stmts[0])
	}
}

func TestParseUseWithAlias(t *testing.T) {
	prog, errs := parse("USE customers ALIAS cust")
	assertNoErrors(t, errs)
	use := prog.Stmts[0].(*UseStmt)
	if use.Table != "customers" {
		t.Errorf("expected Table=customers, got %q", use.Table)
	}
	if use.Alias != "cust" {
		t.Errorf("expected Alias=cust, got %q", use.Alias)
	}
}

func TestParseUseExclusive(t *testing.T) {
	prog, errs := parse("USE mytable EXCLUSIVE")
	assertNoErrors(t, errs)
	use := prog.Stmts[0].(*UseStmt)
	if !use.Exclusive {
		t.Error("expected Exclusive=true")
	}
}

func TestParseSelect(t *testing.T) {
	prog, errs := parse("SELECT 1")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
	_, ok := prog.Stmts[0].(*SelectStmt)
	if !ok {
		t.Fatalf("expected *SelectStmt, got %T", prog.Stmts[0])
	}
}

func TestParseReplace(t *testing.T) {
	prog, errs := parse("REPLACE Name WITH \"John\"")
	assertNoErrors(t, errs)
	rep, ok := prog.Stmts[0].(*ReplaceStmt)
	if !ok {
		t.Fatalf("expected *ReplaceStmt, got %T", prog.Stmts[0])
	}
	if rep.Field != "Name" {
		t.Errorf("expected Field=Name, got %q", rep.Field)
	}
	if rep.Alias != "" {
		t.Errorf("expected empty Alias, got %q", rep.Alias)
	}
}

func TestParseReplaceWithAlias(t *testing.T) {
	prog, errs := parse("REPLACE Cust->Name WITH \"Jane\"")
	assertNoErrors(t, errs)
	rep := prog.Stmts[0].(*ReplaceStmt)
	if rep.Alias != "Cust" {
		t.Errorf("expected Alias=Cust, got %q", rep.Alias)
	}
	if rep.Field != "Name" {
		t.Errorf("expected Field=Name, got %q", rep.Field)
	}
}

func TestParseAppendBlank(t *testing.T) {
	prog, errs := parse("APPEND BLANK")
	assertNoErrors(t, errs)
	app := prog.Stmts[0].(*AppendStmt)
	if !app.Blank {
		t.Error("expected Blank=true")
	}
}

func TestParseAppend(t *testing.T) {
	prog, errs := parse("APPEND")
	assertNoErrors(t, errs)
	app := prog.Stmts[0].(*AppendStmt)
	if app.Blank {
		t.Error("expected Blank=false for plain APPEND")
	}
}

func TestParsePack(t *testing.T) {
	prog, errs := parse("PACK")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*PackStmt)
	if !ok {
		t.Fatalf("expected *PackStmt, got %T", prog.Stmts[0])
	}
}

func TestParseZap(t *testing.T) {
	prog, errs := parse("ZAP")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*ZapStmt)
	if !ok {
		t.Fatalf("expected *ZapStmt, got %T", prog.Stmts[0])
	}
}

func TestParseGoTop(t *testing.T) {
	prog, errs := parse("GO TOP")
	assertNoErrors(t, errs)
	goStmt := prog.Stmts[0].(*GoStmt)
	if goStmt.Pos != "TOP" {
		t.Errorf("expected Pos=TOP, got %q", goStmt.Pos)
	}
}

func TestParseGoRecord(t *testing.T) {
	prog, errs := parse("GO 100")
	assertNoErrors(t, errs)
	goStmt := prog.Stmts[0].(*GoStmt)
	if goStmt.Expr == nil {
		t.Errorf("expected non-nil Expr")
	} else if num, ok := goStmt.Expr.(*NumberExpr); !ok {
		t.Errorf("expected NumberExpr, got %T", goStmt.Expr)
	} else if num.IntValue != 100 {
		t.Errorf("expected IntValue=100, got %d", num.IntValue)
	}
}

func TestParseSkip(t *testing.T) {
	prog, errs := parse("SKIP 5")
	assertNoErrors(t, errs)
	skip := prog.Stmts[0].(*SkipStmt)
	if skip.Count == nil {
		t.Fatal("expected non-nil Count")
	}
}

func TestParseSkipNoCount(t *testing.T) {
	prog, errs := parse("SKIP")
	assertNoErrors(t, errs)
	skip := prog.Stmts[0].(*SkipStmt)
	if skip.Count != nil {
		t.Error("expected nil Count for plain SKIP")
	}
}

func TestParseSkipNegative(t *testing.T) {
	prog, errs := parse("SKIP -1")
	assertNoErrors(t, errs)
	skip := prog.Stmts[0].(*SkipStmt)
	if skip.Count == nil {
		t.Fatal("expected non-nil Count")
	}
}

func TestParseLocate(t *testing.T) {
	prog, errs := parse("LOCATE FOR Name = \"John\"")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*LocateStmt)
	if !ok {
		t.Fatalf("expected *LocateStmt, got %T", prog.Stmts[0])
	}
}

func TestParseContinue(t *testing.T) {
	prog, errs := parse("CONTINUE")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*ContinueStmt)
	if !ok {
		t.Fatalf("expected *ContinueStmt, got %T", prog.Stmts[0])
	}
}

func TestParseSeek(t *testing.T) {
	prog, errs := parse("SEEK \"Smith\"")
	assertNoErrors(t, errs)
	seek := prog.Stmts[0].(*SeekStmt)
	if seek.Expr == nil {
		t.Fatal("expected non-nil Expr")
	}
}

func TestParseCloseDatabases(t *testing.T) {
	prog, errs := parse("CLOSE DATABASES")
	assertNoErrors(t, errs)
	cls := prog.Stmts[0].(*CloseStmt)
	if !cls.All {
		t.Error("expected All=true for CLOSE DATABASES")
	}
}

func TestParseCloseAlias(t *testing.T) {
	prog, errs := parse("CLOSE myAlias")
	assertNoErrors(t, errs)
	cls := prog.Stmts[0].(*CloseStmt)
	if cls.Alias != "myAlias" {
		t.Errorf("expected Alias=myAlias, got %q", cls.Alias)
	}
}

func TestParseStore(t *testing.T) {
	prog, errs := parse("STORE 0 TO x")
	assertNoErrors(t, errs)
	store := prog.Stmts[0].(*StoreStmt)
	if store.Var != "x" {
		t.Errorf("expected Var=x, got %q", store.Var)
	}
}

func TestParseClear(t *testing.T) {
	prog, errs := parse("CLEAR")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*ClearStmt)
	if !ok {
		t.Fatalf("expected *ClearStmt, got %T", prog.Stmts[0])
	}
}

// ---------- Control flow ----------

func TestParseIf(t *testing.T) {
	prog, errs := parse("IF x > 5\n  y = 10\nENDIF")
	assertNoErrors(t, errs)
	ifStmt, ok := prog.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected *IfStmt, got %T", prog.Stmts[0])
	}
	if len(ifStmt.ThenBody) != 1 {
		t.Errorf("expected 1 then-body statement, got %d", len(ifStmt.ThenBody))
	}
	if ifStmt.ElseBody != nil {
		t.Errorf("expected nil else-body, got %d statements", len(ifStmt.ElseBody))
	}
}

func TestParseIfElse(t *testing.T) {
	prog, errs := parse("IF x > 5\n  y = 10\nELSE\n  y = 20\nENDIF")
	assertNoErrors(t, errs)
	ifStmt := prog.Stmts[0].(*IfStmt)
	if len(ifStmt.ThenBody) != 1 {
		t.Errorf("expected 1 then-body statement, got %d", len(ifStmt.ThenBody))
	}
	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("expected 1 else-body statement, got %d", len(ifStmt.ElseBody))
	}
}

func TestParseNestedIf(t *testing.T) {
	input := `IF a > 0
  IF b > 0
    c = 1
  ENDIF
ENDIF`
	prog, errs := parse(input)
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected *IfStmt, got %T", prog.Stmts[0])
	}
}

func TestParseDoWhile(t *testing.T) {
	prog, errs := parse("DO WHILE .NOT. EOF()\n  SKIP\nENDDO")
	assertNoErrors(t, errs)
	whileStmt, ok := prog.Stmts[0].(*WhileStmt)
	if !ok {
		t.Fatalf("expected *WhileStmt, got %T", prog.Stmts[0])
	}
	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestParseDoWhileNested(t *testing.T) {
	input := `DO WHILE a > 0
  DO WHILE b > 0
    c = c + 1
  ENDDO
ENDDO`
	prog, errs := parse(input)
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestParseForLoop(t *testing.T) {
	prog, errs := parse("FOR i = 1 TO 10\n  ? i\nENDFOR")
	assertNoErrors(t, errs)
	forStmt, ok := prog.Stmts[0].(*ForStmt)
	if !ok {
		t.Fatalf("expected *ForStmt, got %T", prog.Stmts[0])
	}
	if forStmt.Var != "i" {
		t.Errorf("expected Var=i, got %q", forStmt.Var)
	}
	if len(forStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(forStmt.Body))
	}
}

func TestParseForLoopWithStep(t *testing.T) {
	prog, errs := parse("FOR i = 1 TO 100 STEP 2\n  ? i\nENDFOR")
	assertNoErrors(t, errs)
	forStmt := prog.Stmts[0].(*ForStmt)
	if forStmt.Step == nil {
		t.Fatal("expected non-nil Step")
	}
}

func TestParseReturn(t *testing.T) {
	prog, errs := parse("RETURN")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*ReturnStmt)
	if !ok {
		t.Fatalf("expected *ReturnStmt, got %T", prog.Stmts[0])
	}
}

func TestParseReturnWithValue(t *testing.T) {
	prog, errs := parse("RETURN x + 1")
	assertNoErrors(t, errs)
	ret := prog.Stmts[0].(*ReturnStmt)
	if ret.Expr == nil {
		t.Fatal("expected non-nil Expr")
	}
}

func TestParseProcedure(t *testing.T) {
	prog, errs := parse("PROCEDURE MyProc\n  ? \"hello\"\nRETURN")
	assertNoErrors(t, errs)
	proc, ok := prog.Stmts[0].(*ProcedureDef)
	if !ok {
		t.Fatalf("expected *ProcedureDef, got %T", prog.Stmts[0])
	}
	if proc.Name != "MyProc" {
		t.Errorf("expected Name=MyProc, got %q", proc.Name)
	}
	if len(proc.Body) != 2 { // the ? statement + RETURN
		t.Errorf("expected 2 body statements, got %d", len(proc.Body))
	}
}

func TestParseFunction(t *testing.T) {
	prog, errs := parse("FUNCTION Add(a, b)\n  RETURN a + b")
	assertNoErrors(t, errs)
	fn, ok := prog.Stmts[0].(*FunctionDef)
	if !ok {
		t.Fatalf("expected *FunctionDef, got %T", prog.Stmts[0])
	}
	if fn.Name != "Add" {
		t.Errorf("expected Name=Add, got %q", fn.Name)
	}
}

func TestParseDelete(t *testing.T) {
	prog, errs := parse("DELETE ALL FOR Name = \"Old\"")
	assertNoErrors(t, errs)
	del := prog.Stmts[0].(*DeleteStmt)
	if del.Scope != "ALL" {
		t.Errorf("expected Scope=ALL, got %q", del.Scope)
	}
}

func TestParseCount(t *testing.T) {
	prog, errs := parse("COUNT FOR Active TO nCount")
	assertNoErrors(t, errs)
	cnt := prog.Stmts[0].(*CountStmt)
	if cnt.To != "nCount" {
		t.Errorf("expected To=nCount, got %q", cnt.To)
	}
}

func TestParseWait(t *testing.T) {
	prog, errs := parse("WAIT \"Press a key\" TO mKey")
	assertNoErrors(t, errs)
	wait := prog.Stmts[0].(*WaitStmt)
	if wait.Prompt != "Press a key" {
		t.Errorf("expected Prompt='Press a key', got %q", wait.Prompt)
	}
	if wait.Var != "mKey" {
		t.Errorf("expected Var=mKey, got %q", wait.Var)
	}
}

func TestParseInput(t *testing.T) {
	prog, errs := parse("INPUT \"Enter value: \" TO mVal")
	assertNoErrors(t, errs)
	inp := prog.Stmts[0].(*InputStmt)
	if inp.Var != "mVal" {
		t.Errorf("expected Var=mVal, got %q", inp.Var)
	}
}

func TestParseAccept(t *testing.T) {
	prog, errs := parse("ACCEPT \"Enter name: \" TO mName")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*AcceptStmt)
	if !ok {
		t.Fatalf("expected *AcceptStmt, got %T", prog.Stmts[0])
	}
}

// ---------- Expressions ----------

func TestParseExpression(t *testing.T) {
	prog, errs := parse("x = 5 + 3")
	assertNoErrors(t, errs)
	ass, ok := prog.Stmts[0].(*AssignmentStmt)
	if !ok {
		t.Fatalf("expected *AssignmentStmt, got %T", prog.Stmts[0])
	}
	if ass.Target != "x" {
		t.Errorf("expected Target=x, got %q", ass.Target)
	}
}

func TestParseComplexExpression(t *testing.T) {
	prog, errs := parse("x = (a + b) * c / d")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
	_, ok := prog.Stmts[0].(*AssignmentStmt)
	if !ok {
		t.Fatalf("expected *AssignmentStmt, got %T", prog.Stmts[0])
	}
}

func TestParseStringConcat(t *testing.T) {
	prog, errs := parse("full = first ++ \" \" ++ last")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestParseComparison(t *testing.T) {
	prog, errs := parse("result = x > y")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestParseLogicalExpr(t *testing.T) {
	prog, errs := parse("flag = a .AND. b .OR. .NOT. c")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestParseFieldRef(t *testing.T) {
	prog, errs := parse("x = Customer->Phone")
	assertNoErrors(t, errs)
	ass := prog.Stmts[0].(*AssignmentStmt)
	ref, ok := ass.Expr.(*FieldRefExpr)
	if !ok {
		t.Fatalf("expected *FieldRefExpr as assigned expression, got %T", ass.Expr)
	}
	if ref.Alias != "Customer" {
		t.Errorf("expected Alias=Customer, got %q", ref.Alias)
	}
	if ref.Field != "Phone" {
		t.Errorf("expected Field=Phone, got %q", ref.Field)
	}
}

func TestParseFunctionCall(t *testing.T) {
	prog, errs := parse("result = MyFunc(a, b + 1, \"hello\")")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestParseFunctionCallNoArgs(t *testing.T) {
	prog, errs := parse("x = EOF()")
	assertNoErrors(t, errs)
	ass := prog.Stmts[0].(*AssignmentStmt)
	call, ok := ass.Expr.(*FuncCallExpr)
	if !ok {
		t.Fatalf("expected *FuncCallExpr, got %T", ass.Expr)
	}
	if call.Name != "EOF" {
		t.Errorf("expected Name=EOF, got %q", call.Name)
	}
	if len(call.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(call.Args))
	}
}

func TestParseParenthesizedExpr(t *testing.T) {
	prog, errs := parse("x = (1 + 2) * 3")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestParseBoolLiterals(t *testing.T) {
	prog, errs := parse("x = .T.\ny = .F.\nz = .Y.\nw = .N.")
	assertNoErrors(t, errs)
	if len(prog.Stmts) < 4 {
		t.Fatalf("expected at least 4 statements, got %d", len(prog.Stmts))
	}
	// Each newline between statements becomes a SEMI
}

// ---------- @ SAY/GET ----------

func TestParseSayGet(t *testing.T) {
	prog, errs := parse("@ 10,5 SAY \"Name:\" GET mName")
	assertNoErrors(t, errs)
	sg, ok := prog.Stmts[0].(*SayGetStmt)
	if !ok {
		t.Fatalf("expected *SayGetStmt, got %T", prog.Stmts[0])
	}
	if sg.SayExpr == nil {
		t.Error("expected non-nil SayExpr")
	}
	if sg.GetVar != "mName" {
		t.Errorf("expected GetVar=mName, got %q", sg.GetVar)
	}
}

func TestParseSayOnly(t *testing.T) {
	prog, errs := parse("@ 1,1 SAY \"hello\"")
	assertNoErrors(t, errs)
	_, ok := prog.Stmts[0].(*SayGetStmt)
	if !ok {
		t.Fatalf("expected *SayGetStmt, got %T", prog.Stmts[0])
	}
}

// ---------- Multi-statement programs ----------

func TestMultipleStatements(t *testing.T) {
	input := `USE customers
SELECT 1
REPLACE Name WITH "Test"
APPEND BLANK`
	prog, errs := parse(input)
	assertNoErrors(t, errs)
	if len(prog.Stmts) < 4 {
		t.Errorf("expected at least 4 statements, got %d", len(prog.Stmts))
	}
}

func TestMultipleProcedures(t *testing.T) {
	input := `PROCEDURE One
  ? 1
RETURN

PROCEDURE Two
  ? 2
RETURN`
	prog, errs := parse(input)
	assertNoErrors(t, errs)
	if len(prog.Stmts) != 2 {
		t.Errorf("expected 2 procedures, got %d", len(prog.Stmts))
	}
}

// ---------- Error handling ----------

func TestParserError(t *testing.T) {
	_, errs := parse("USE")
	// USE with nothing after it — should work since we just need a table name, but parser may try next()
	// Actually USE expects IDENTIFIER, but peek will be EOF. The expect() will produce an error.
	if len(errs) == 0 {
		t.Errorf("expected parser error for USE with no table")
	}
}

func TestParserErrorsCollected(t *testing.T) {
	_, errs := parse("IF x > 5\n  y = 10\nIF")
	// Should get an error about ENDIF
	if len(errs) == 0 {
		t.Errorf("expected parser error for unmatched IF")
	}
}

// ---------- Edge cases ----------

func TestEmptyProgram(t *testing.T) {
	prog, errs := parse("")
	assertNoErrors(t, errs)
	if prog == nil {
		t.Fatal("expected non-nil program")
	}
	if len(prog.Stmts) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Stmts))
	}
}

func TestWhitespaceOnlyProgram(t *testing.T) {
	prog, errs := parse("   \n  \n  ")
	assertNoErrors(t, errs)
	if len(prog.Stmts) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Stmts))
	}
}

func TestCommentsOnly(t *testing.T) {
	prog, errs := parse("&& comment\n&& another")
	assertNoErrors(t, errs)
	if len(prog.Stmts) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Stmts))
	}
}

func TestOperatorPrecedence(t *testing.T) {
	prog, errs := parse("x = 1 + 2 * 3")
	assertNoErrors(t, errs)
	ass := prog.Stmts[0].(*AssignmentStmt)
	bin, ok := ass.Expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected *BinaryExpr, got %T", ass.Expr)
	}
	if bin.Op != T_PLUS {
		t.Errorf("expected outer op to be + (1 + (2*3)), got %v", bin.Op)
	}
}

func TestComparisonChain(t *testing.T) {
	prog, errs := parse("x = a > b .AND. c < d")
	assertNoErrors(t, errs)
	assertStmtCount(t, prog, 1)
}

func TestDoCall(t *testing.T) {
	prog, errs := parse("DO MyProc WITH a, b, c")
	assertNoErrors(t, errs)
	call, ok := prog.Stmts[0].(*CallStmt)
	if !ok {
		t.Fatalf("expected *CallStmt, got %T", prog.Stmts[0])
	}
	if call.Name != "MyProc" {
		t.Errorf("expected Name=MyProc, got %q", call.Name)
	}
	if len(call.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(call.Args))
	}
}

func TestDoProcNoArgs(t *testing.T) {
	prog, errs := parse("DO MyProc")
	assertNoErrors(t, errs)
	call, ok := prog.Stmts[0].(*CallStmt)
	if !ok {
		t.Fatalf("expected *CallStmt, got %T", prog.Stmts[0])
	}
	if call.Name != "MyProc" {
		t.Errorf("expected Name=MyProc, got %q", call.Name)
	}
}

func TestAssignmentWithColonEquals(t *testing.T) {
	prog, errs := parse("x := 42")
	assertNoErrors(t, errs)
	ass := prog.Stmts[0].(*AssignmentStmt)
	if ass.Target != "x" {
		t.Errorf("expected Target=x, got %q", ass.Target)
	}
}

// ---------- Real-world xBase script ----------

func TestComplexProgram(t *testing.T) {
	input := `PROCEDURE Main
  USE customers ALIAS cust
  SELECT 0
  USE orders ALIAS ord
  SELECT cust
  DO WHILE .NOT. EOF()
    IF cust->Balance > 1000
      REPLACE Status WITH "Premium"
    ELSE
      REPLACE Status WITH "Regular"
    ENDIF
    SKIP
  ENDDO
  CLOSE DATABASES
RETURN

FUNCTION IsActive(custName)
  LOCAL result
  result = .F.
  USE customers
  LOCATE FOR Name = custName
  IF FOUND()
    result = .T.
  ENDIF
  RETURN result`

	prog, errs := parse(input)
	assertNoErrors(t, errs)
	if prog == nil || len(prog.Stmts) == 0 {
		t.Fatal("expected parsed program with statements")
	}
	// Should have a PROCEDURE and a FUNCTION
	foundProc, foundFunc := false, false
	for _, s := range prog.Stmts {
		switch s.(type) {
		case *ProcedureDef:
			foundProc = true
		case *FunctionDef:
			foundFunc = true
		}
	}
	if !foundProc {
		t.Error("expected PROCEDURE definition")
	}
	if !foundFunc {
		t.Error("expected FUNCTION definition")
	}
}
