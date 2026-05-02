package compiler

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	tokens []Token
	pos    int
	errors []string
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (*Program, []string) {
	prog := &Program{}
	p.errors = nil
	for !p.atEnd() {
		if stmt := p.parseStmt(); stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
	}
	return prog, p.errors
}

func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens) || p.peek().Type == T_EOF
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: T_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) next() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) check(typ TokenType) bool {
	return !p.atEnd() && p.peek().Type == typ
}

func (p *Parser) expect(typ TokenType) (Token, bool) {
	if p.check(typ) {
		return p.next(), true
	}
	tok := p.peek()
	p.error(tok, "expected %s, got %s", typ, tok.Type)
	return Token{}, false
}

func (p *Parser) expectOneOf(types ...TokenType) (Token, bool) {
	for _, typ := range types {
		if p.check(typ) {
			return p.next(), true
		}
	}
	tok := p.peek()
	p.error(tok, "expected one of %v, got %s", types, tok.Type)
	return Token{}, false
}

func (p *Parser) match(typ TokenType) bool {
	if p.check(typ) {
		p.pos++
		return true
	}
	return false
}

func (p *Parser) matchOneOf(types ...TokenType) bool {
	for _, typ := range types {
		if p.check(typ) {
			p.pos++
			return true
		}
	}
	return false
}

func (p *Parser) error(tok Token, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errors = append(p.errors, fmt.Sprintf("line %d:%d: %s", tok.Line, tok.Column, msg))
}

func (p *Parser) parseStmt() Stmt {
	for p.check(T_SEMI) {
		p.next()
	}
	if p.atEnd() {
		return nil
	}
	tok := p.peek()

	switch tok.Type {
	case T_PROCEDURE:
		return p.parseProcedure()
	case T_FUNCTION:
		return p.parseFunction()
	case T_USE:
		return p.parseUse()
	case T_SELECT:
		return p.parseSelect()
	case T_REPLACE:
		return p.parseReplace()
	case T_APPEND:
		return p.parseAppend()
	case T_DELETE:
		return p.parseDelete()
	case T_PACK:
		p.next()
		return &PackStmt{}
	case T_ZAP:
		p.next()
		return &ZapStmt{}
	case T_GO, T_GOTO:
		return p.parseGo()
	case T_SKIP:
		return p.parseSkip()
	case T_LOCATE:
		return p.parseLocate()
	case T_CONTINUE:
		p.next()
		return &ContinueStmt{}
	case T_SEEK:
		return p.parseSeek()
	case T_CLOSE:
		return p.parseClose()
	case T_STORE:
		return p.parseStore()
	case T_INPUT:
		return p.parseInput()
	case T_ACCEPT:
		return p.parseAccept()
	case T_WAIT:
		return p.parseWait()
	case T_CLEAR:
		p.next()
		return &ClearStmt{}
	case T_COUNT:
		return p.parseCount()
	case T_SUM:
		return p.parseSum()
	case T_IF:
		return p.parseIf()
	case T_DO:
		return p.parseDo()
	case T_FOR:
		return p.parseFor()
	case T_RETURN:
		return p.parseReturn()
	case T_QUIT:
		p.next()
		return &ExprStmt{Expr: &IdentExpr{Name: "QUIT"}}
	case T_READ:
		p.next()
		return &ReadStmt{}
	case T_RUNSQL:
		return p.parseRunSQL()
	case T_NAV:
		return p.parseNav()
	case T_CONFIRM:
		return p.parseConfirm()
	case T_MENU:
		return p.parseMenu()
	case T_CALL:
		return p.parseCall()
	case T_SET:
		return p.parseSet()
	case T_LOCAL, T_PRIVATE, T_PUBLIC, T_STATIC:
		return p.parseVarDecl()
	case T_ENDIF, T_ENDDO, T_ENDFOR, T_ELSE, T_OTHERWISE, T_CASE, T_ENDCASE, T_ENDTEXT:
		return nil
	default:
		return p.parseExpressionStmt()
	}
}

func (p *Parser) parseProcedure() *ProcedureDef {
	p.next() // consume PROCEDURE
	nameTok, ok := p.expect(T_IDENTIFIER)
	if !ok {
		return nil
	}
	p.consumeNewlines()
	body := p.parseBodyUntil(T_PROCEDURE, T_FUNCTION, T_RETURN)
	return &ProcedureDef{Name: nameTok.Lexeme, Body: body}
}

func (p *Parser) parseFunction() *FunctionDef {
	p.next()
	nameTok, ok := p.expect(T_IDENTIFIER)
	if !ok {
		return nil
	}
	if p.check(T_LPAREN) {
		p.next()
		for !p.check(T_RPAREN) && !p.atEnd() {
			p.next()
		}
		p.expect(T_RPAREN)
	}
	p.consumeNewlines()
	body := p.parseBodyUntil(T_PROCEDURE, T_FUNCTION, T_RETURN)
	return &FunctionDef{Name: nameTok.Lexeme, Body: body}
}

func (p *Parser) parseBodyUntil(terminals ...TokenType) []Stmt {
	var body []Stmt
	for !p.atEnd() {
		for p.check(T_SEMI) {
			p.next()
		}
		if p.atEnd() {
			break
		}
		tok := p.peek()
		if tok.Type == T_RETURN || isOneOf(tok.Type, terminals) {
			if tok.Type == T_RETURN {
				p.next()
				if !p.atEnd() && p.peek().Type != T_SEMI && p.peek().Type != T_EOF && !p.isTerminal(p.peek().Type) {
					body = append(body, &ReturnStmt{Expr: p.parseExpr()})
				} else {
					body = append(body, &ReturnStmt{})
				}
			}
			break
		}
		if tok.Type == T_ENDIF || tok.Type == T_ENDDO || tok.Type == T_ENDFOR || tok.Type == T_ELSE {
			break
		}
		if stmt := p.parseStmt(); stmt != nil {
			body = append(body, stmt)
		} else if !p.isTerminal(p.peek().Type) && p.peek().Type != T_RETURN {
			p.next()
		}
	}
	return body
}

func (p *Parser) parseUse() *UseStmt {
	p.next()
	stmt := &UseStmt{}
	if tableTok, ok := p.expect(T_IDENTIFIER); ok {
		stmt.Table = tableTok.Lexeme
	}
	p.consumeNewlines()
	if p.atEnd() || p.peek().Type == T_SEMI || p.isTerminal(p.peek().Type) {
		return stmt
	}
	nextTyp := p.peek().Type
	for isUseModifier(nextTyp) {
		switch nextTyp {
		case T_IDENTIFIER:
			ident := p.next().Lexeme
			upper := strings.ToUpper(ident)
			switch upper {
			case "ALIAS":
				if aliasTok, ok := p.expect(T_IDENTIFIER); ok {
					stmt.Alias = aliasTok.Lexeme
				}
			case "EXCLUSIVE":
				stmt.Exclusive = true
			case "SHARED":
				stmt.Shared = true
			case "NOUPDATE":
				stmt.NoUpdate = true
			}
		case T_IN:
			p.next()
			if p.check(T_NUMBER) {
				stmt.In = p.next().IntVal
			} else if p.check(T_IDENTIFIER) {
				n, _ := strconv.ParseInt(p.next().Lexeme, 10, 64)
				stmt.In = n
			}
		}
		p.consumeNewlines()
		if p.atEnd() || p.peek().Type == T_SEMI || p.isTerminal(p.peek().Type) {
			break
		}
		nextTyp = p.peek().Type
	}
	return stmt
}

func isUseModifier(typ TokenType) bool {
	return typ == T_IDENTIFIER || typ == T_IN
}

func (p *Parser) parseSelect() *SelectStmt {
	p.next()
	stmt := &SelectStmt{}
	if p.check(T_NUMBER) || p.check(T_IDENTIFIER) {
		stmt.Expr = p.parseExpr()
	} else {
		p.error(p.peek(), "expected work area number or alias")
	}
	return stmt
}

func (p *Parser) parseReplace() *ReplaceStmt {
	p.next()
	stmt := &ReplaceStmt{}
	fieldTok, ok := p.expect(T_IDENTIFIER)
	if !ok {
		return stmt
	}
	stmt.Field = fieldTok.Lexeme
	if p.check(T_ARROW) {
		p.next()
		if f, ok := p.expect(T_IDENTIFIER); ok {
			stmt.Alias = stmt.Field
			stmt.Field = f.Lexeme
		}
	}
	p.consumeNewlines()
	if !p.matchOneOf(T_WITH, T_ASSIGN, T_EQ) {
		p.error(p.peek(), "expected WITH after REPLACE field")
		return stmt
	}
	stmt.Expr = p.parseExpr()
	return stmt
}

func (p *Parser) parseAppend() *AppendStmt {
	p.next()
	stmt := &AppendStmt{}
	if p.match(T_BLANK) {
		stmt.Blank = true
	}
	return stmt
}

func (p *Parser) parseDelete() *DeleteStmt {
	p.next()
	stmt := &DeleteStmt{}
	// Check scope keywords (ALL, NEXT, RECORD, REST)
	switch p.peek().Type {
	case T_ALL, T_NEXT, T_RECORD, T_REST:
		stmt.Scope = p.next().Lexeme
	case T_IDENTIFIER:
		ident := strings.ToUpper(p.peek().Lexeme)
		if ident == "ALL" || ident == "NEXT" || ident == "RECORD" || ident == "REST" {
			stmt.Scope = p.next().Lexeme
		}
	}
	p.consumeNewlines()
	for !p.atEnd() && p.peek().Type != T_SEMI {
		switch p.peek().Type {
		case T_FOR:
			p.next()
			stmt.For = p.parseExpr()
		case T_WHILE:
			p.next()
			stmt.While = p.parseExpr()
		case T_IDENTIFIER:
			ident := strings.ToUpper(p.peek().Lexeme)
			if ident == "FOR" {
				p.next()
				stmt.For = p.parseExpr()
			} else if ident == "WHILE" {
				p.next()
				stmt.While = p.parseExpr()
			} else {
				break
			}
		default:
			break
		}
		break
	}
	return stmt
}

func (p *Parser) parseGo() *GoStmt {
	p.next()
	stmt := &GoStmt{}
	if p.check(T_IDENTIFIER) {
		upper := strings.ToUpper(p.peek().Lexeme)
		if upper == "TOP" || upper == "BOTTOM" {
			stmt.Pos = p.next().Lexeme
			return stmt
		}
	}
	// Handle expressions: GO 5, GO VAL(mId), etc.
	stmt.Expr = p.parseExpr()
	stmt.Pos = ""
	return stmt
}

func (p *Parser) parseSkip() *SkipStmt {
	p.next()
	stmt := &SkipStmt{}
	if p.check(T_NUMBER) || p.check(T_MINUS) || p.check(T_PLUS) {
		stmt.Count = p.parseExpr()
	}
	return stmt
}

func (p *Parser) parseLocate() *LocateStmt {
	p.next()
	stmt := &LocateStmt{}
	for !p.atEnd() && p.peek().Type != T_SEMI {
		ident := strings.ToUpper(p.peek().Lexeme)
		if ident == "ALL" || ident == "NEXT" || ident == "REST" {
			stmt.Scope = p.next().Lexeme
			continue
		}
		if ident == "FOR" || p.peek().Type == T_FOR {
			p.next()
			stmt.For = p.parseExpr()
			continue
		}
		if ident == "WHILE" {
			p.next()
			stmt.While = p.parseExpr()
			continue
		}
		break
	}
	return stmt
}

func (p *Parser) parseSeek() *SeekStmt {
	p.next()
	return &SeekStmt{Expr: p.parseExpr()}
}

func (p *Parser) parseClose() *CloseStmt {
	p.next()
	stmt := &CloseStmt{}
	if p.check(T_IDENTIFIER) {
		ident := strings.ToUpper(p.peek().Lexeme)
		if ident == "DATABASES" || ident == "ALL" {
			stmt.All = true
			p.next()
		} else if ident == "DATABASE" {
			stmt.Databases = true
			p.next()
		} else {
			stmt.Alias = p.next().Lexeme
		}
	}
	return stmt
}

func (p *Parser) parseStore() *StoreStmt {
	p.next()
	expr := p.parseExpr()
	p.consumeNewlines()
	p.expect(T_TO)
	varTok, ok := p.expect(T_IDENTIFIER)
	if !ok {
		return &StoreStmt{Expr: expr}
	}
	return &StoreStmt{Expr: expr, Var: varTok.Lexeme}
}

func (p *Parser) parseInput() *InputStmt {
	p.next()
	stmt := &InputStmt{}
	if p.check(T_STRING) {
		stmt.Prompt = p.next().Lexeme
	}
	if p.check(T_TO) {
		p.next()
	}
	if p.check(T_IDENTIFIER) {
		stmt.Var = p.next().Lexeme
	}
	return stmt
}

func (p *Parser) parseAccept() *AcceptStmt {
	p.next()
	stmt := &AcceptStmt{}
	if p.check(T_STRING) {
		stmt.Prompt = p.next().Lexeme
	}
	if p.check(T_TO) {
		p.next()
	}
	if p.check(T_IDENTIFIER) {
		stmt.Var = p.next().Lexeme
	}
	return stmt
}

func (p *Parser) parseWait() *WaitStmt {
	p.next()
	stmt := &WaitStmt{}
	if p.check(T_STRING) {
		stmt.Prompt = p.next().Lexeme
	}
	if p.check(T_TO) {
		p.next()
	}
	if p.check(T_IDENTIFIER) {
		stmt.Var = p.next().Lexeme
	}
	return stmt
}

func (p *Parser) parseCount() *CountStmt {
	p.next()
	stmt := &CountStmt{}
	for !p.atEnd() && p.peek().Type != T_SEMI {
		ident := strings.ToUpper(p.peek().Lexeme)
		if ident == "ALL" || ident == "NEXT" || ident == "REST" {
			stmt.Scope = p.next().Lexeme
			continue
		}
		if ident == "FOR" || p.peek().Type == T_FOR {
			p.next()
			stmt.For = p.parseExpr()
			continue
		}
		if ident == "WHILE" {
			p.next()
			stmt.While = p.parseExpr()
			continue
		}
		if ident == "TO" {
			p.next()
			if p.check(T_IDENTIFIER) {
				stmt.To = p.next().Lexeme
			}
			continue
		}
		break
	}
	return stmt
}

func (p *Parser) parseSum() *SumStmt {
	p.next()
	stmt := &SumStmt{}
	stmt.Expr = p.parseExpr()
	for !p.atEnd() && p.peek().Type != T_SEMI {
		ident := strings.ToUpper(p.peek().Lexeme)
		if ident == "ALL" || ident == "NEXT" || ident == "REST" {
			stmt.Scope = p.next().Lexeme
			continue
		}
		if ident == "FOR" || p.peek().Type == T_FOR {
			p.next()
			stmt.For = p.parseExpr()
			continue
		}
		if ident == "WHILE" {
			p.next()
			stmt.While = p.parseExpr()
			continue
		}
		if ident == "TO" {
			p.next()
			if p.check(T_IDENTIFIER) {
				stmt.To = p.next().Lexeme
			}
			continue
		}
		break
	}
	return stmt
}

func (p *Parser) parseIf() *IfStmt {
	p.next() // IF
	stmt := &IfStmt{Condition: p.parseExpr()}
	p.consumeNewlines()
	for !p.atEnd() && p.peek().Type != T_ELSE && p.peek().Type != T_ENDIF {
		if s := p.parseStmt(); s != nil {
			stmt.ThenBody = append(stmt.ThenBody, s)
		} else if !p.isTerminal(p.peek().Type) {
			p.next()
		}
	}
	if p.match(T_ELSE) {
		p.consumeNewlines()
		for !p.atEnd() && p.peek().Type != T_ENDIF {
			if s := p.parseStmt(); s != nil {
				stmt.ElseBody = append(stmt.ElseBody, s)
			} else if !p.isTerminal(p.peek().Type) {
				p.next()
			}
		}
	}
	p.expect(T_ENDIF)
	return stmt
}

func (p *Parser) parseDo() Stmt {
	p.next()
	if p.check(T_WHILE) {
		return p.parseDoWhile()
	}
	if p.check(T_CASE) {
		return p.parseDoCase()
	}
	if p.check(T_IDENTIFIER) {
		return p.parseDoCall()
	}
	p.error(p.peek(), "expected WHILE, CASE, or procedure name after DO")
	return nil
}

func (p *Parser) parseDoWhile() *WhileStmt {
	p.next()
	stmt := &WhileStmt{Condition: p.parseExpr()}
	p.consumeNewlines()
	for !p.atEnd() && p.peek().Type != T_ENDDO {
		if s := p.parseStmt(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else if !p.isTerminal(p.peek().Type) {
			p.next()
		}
	}
	p.expect(T_ENDDO)
	return stmt
}

// simple DO CASE - not fully nested CASE expressions
func (p *Parser) parseDoCase() Stmt {
	p.next()
	// Build a chain of IF statements
	var result Stmt
	p.consumeNewlines()
	for !p.atEnd() && p.peek().Type != T_ENDCASE && p.peek().Type != T_OTHERWISE {
		if !p.match(T_CASE) {
			break
		}
		cond := p.parseExpr()
		p.consumeNewlines()
		var body []Stmt
		for !p.atEnd() && p.peek().Type != T_CASE && p.peek().Type != T_OTHERWISE && p.peek().Type != T_ENDCASE {
			if s := p.parseStmt(); s != nil {
				body = append(body, s)
			} else if !p.isTerminal(p.peek().Type) && p.peek().Type != T_RETURN {
				p.next()
			}
		}
		if result == nil {
			result = &IfStmt{Condition: cond, ThenBody: body}
		} else {
			ifStmt := result.(*IfStmt)
			for ifStmt.ElseBody != nil {
				if nested, ok := ifStmt.ElseBody[0].(*IfStmt); ok && len(ifStmt.ElseBody) == 1 {
					ifStmt = nested
				} else {
					break
				}
			}
			ifStmt.ElseBody = []Stmt{&IfStmt{Condition: cond, ThenBody: body}}
		}
	}
	if p.match(T_OTHERWISE) {
		p.consumeNewlines()
		var otherwiseBody []Stmt
		for !p.atEnd() && p.peek().Type != T_ENDCASE {
			if s := p.parseStmt(); s != nil {
				otherwiseBody = append(otherwiseBody, s)
			} else if !p.isTerminal(p.peek().Type) && p.peek().Type != T_RETURN {
				p.next()
			}
		}
		if result == nil {
			result = &IfStmt{ThenBody: otherwiseBody}
		} else {
			ifStmt := result.(*IfStmt)
			for ifStmt.ElseBody != nil {
				if nested, ok := ifStmt.ElseBody[0].(*IfStmt); ok && len(ifStmt.ElseBody) == 1 {
					ifStmt = nested
				} else {
					break
				}
			}
			ifStmt.ElseBody = otherwiseBody
		}
	}
	if p.atEnd() || p.peek().Type != T_ENDCASE {
		return result
	}
	p.next()
	return result
}

func (p *Parser) parseDoCall() *CallStmt {
	name := p.next().Lexeme
	stmt := &CallStmt{Name: name}
	if p.check(T_LPAREN) {
		p.next()
		if !p.check(T_RPAREN) {
			stmt.Args = p.parseExprList()
		}
		p.expect(T_RPAREN)
	} else if p.check(T_WITH) {
		p.next()
		stmt.With = true
		stmt.Args = p.parseExprList()
	} else {
		p.consumeNewlines()
		for !p.atEnd() && p.peek().Type != T_SEMI && !p.isTerminal(p.peek().Type) && p.peek().Type != T_WITH {
			if p.check(T_IDENTIFIER) {
				ident := p.next()
				stmt.Args = append(stmt.Args, &IdentExpr{Name: ident.Lexeme})
			} else {
				break
			}
		}
	}
	return stmt
}

func (p *Parser) parseFor() *ForStmt {
	p.next()
	varTok, ok := p.expect(T_IDENTIFIER)
	if !ok {
		return nil
	}
	stmt := &ForStmt{Var: varTok.Lexeme}
	p.expectOneOf(T_ASSIGN, T_EQ)
	stmt.Start = p.parseExpr()
	p.consumeNewlines()
	p.expect(T_TO)
	stmt.End = p.parseExpr()
	if p.check(T_STEP) {
		p.next()
		stmt.Step = p.parseExpr()
	}
	p.consumeNewlines()
	for !p.atEnd() && p.peek().Type != T_ENDFOR {
		if s := p.parseStmt(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else if !p.isTerminal(p.peek().Type) {
			p.next()
		}
	}
	p.expect(T_ENDFOR)
	return stmt
}

func (p *Parser) parseReturn() *ReturnStmt {
	p.next()
	stmt := &ReturnStmt{}
	if !p.atEnd() && p.peek().Type != T_SEMI && !p.isTerminal(p.peek().Type) {
		stmt.Expr = p.parseExpr()
	}
	return stmt
}

func (p *Parser) parseCall() *CallStmt {
	p.next()
	nameTok, ok := p.expect(T_IDENTIFIER)
	if !ok {
		return nil
	}
	stmt := &CallStmt{Name: nameTok.Lexeme}
	if p.check(T_LPAREN) {
		p.next()
		if !p.check(T_RPAREN) {
			stmt.Args = p.parseExprList()
		}
		p.expect(T_RPAREN)
	}
	return stmt
}

func (p *Parser) parseConfirm() Stmt {
	p.next()
	stmt := &ConfirmStmt{}
	if p.check(T_STRING) {
		stmt.Message = p.next().Lexeme
	}
	return stmt
}

// MENU "Title" "Label" -> "Procedure", "Label2" -> "Proc2"
func (p *Parser) parseMenu() Stmt {
	p.next()
	stmt := &MenuStmt{}
	if p.check(T_STRING) {
		stmt.Title = p.next().Lexeme
	}
	for !p.atEnd() && p.peek().Type == T_STRING {
		item := MenuItemNode{}
		item.Label = p.next().Lexeme
		if p.check(T_ARROW) {
			p.next()
		} else if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "TO" {
			p.next()
		}
		if p.check(T_STRING) {
			item.Procedure = p.next().Lexeme
		} else if p.check(T_IDENTIFIER) {
			item.Procedure = p.next().Lexeme
		}
		stmt.Items = append(stmt.Items, item)
		if p.check(T_COMMA) {
			p.next()
		}
	}
	return stmt
}

func (p *Parser) parseNav() Stmt {
	p.next()
	stmt := &NavStmt{Entries: make(map[string]string)}
	for !p.atEnd() && p.peek().Type == T_STRING {
		choice := p.next().Lexeme
		if p.check(T_ARROW) {
			p.next()
		} else if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "TO" {
			p.next()
		}
		if p.check(T_STRING) {
			stmt.Entries[choice] = p.next().Lexeme
		} else if p.check(T_IDENTIFIER) {
			stmt.Entries[choice] = p.next().Lexeme
		}
		if p.check(T_COMMA) {
			p.next()
		}
	}
	return stmt
}

func (p *Parser) parseRunSQL() Stmt {
	p.next()
	stmt := &ExecSQLStmt{}
	if p.check(T_STRING) {
		stmt.Query = p.next().Lexeme
	}
	// Optional column headers: COLUMNS "ID", "Name", ...
	if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "COLUMNS" {
		p.next()
		for !p.atEnd() && p.peek().Type == T_STRING {
			stmt.Cols = append(stmt.Cols, p.next().Lexeme)
			if p.check(T_COMMA) {
				p.next()
			}
		}
	}
	// Optional search columns: SEARCH "Name", "Email"
	if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "SEARCH" {
		p.next()
		for !p.atEnd() && p.peek().Type == T_STRING {
			stmt.SearchCols = append(stmt.SearchCols, p.next().Lexeme)
			if p.check(T_COMMA) {
				p.next()
			}
		}
	}
	// Optional row actions: ACTIONS "Edit" -> "EditProc", "Delete" -> "DelProc"
	if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "ACTIONS" {
		p.next()
		for !p.atEnd() && p.peek().Type == T_STRING {
			label := p.next().Lexeme
			if p.check(T_ARROW) {
				p.next()
			} else if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "TO" {
				p.next()
			}
			if p.check(T_STRING) {
				stmt.Actions = append(stmt.Actions, RowActionDef{Label: label, Procedure: p.next().Lexeme})
			} else if p.check(T_IDENTIFIER) {
				stmt.Actions = append(stmt.Actions, RowActionDef{Label: label, Procedure: p.next().Lexeme})
			}
			if p.check(T_COMMA) {
				p.next()
			}
		}
	}
	return stmt
}

func (p *Parser) parseSet() Stmt {
	p.next()
	parts := []string{}
	for !p.atEnd() && p.peek().Type != T_SEMI && !p.isTerminal(p.peek().Type) {
		parts = append(parts, p.next().Lexeme)
	}
	return &SetStmt{Parts: parts}
}

func (p *Parser) parseVarDecl() Stmt {
	p.next()
	stmt := &ExprStmt{Expr: &IdentExpr{Name: "DECLARE"}}
	for !p.atEnd() && p.peek().Type != T_SEMI && p.peek().Type != T_EOF && !p.isTerminal(p.peek().Type) {
		if p.check(T_IDENTIFIER) {
			p.next()
		} else if p.check(T_COMMA) {
			p.next()
		} else {
			break
		}
	}
	return stmt
}

// ----- Expression statements -----

func (p *Parser) parseExpressionStmt() Stmt {
	if p.check(T_ATSIGN) {
		return p.parseAtSayGet()
	}
	expr := p.parseExpr()
	if assign, ok := expr.(*AssignmentStmt); ok {
		return assign
	}
	if call, ok := expr.(*FuncCallExpr); ok {
		return &CallStmt{Name: call.Name, Args: call.Args}
	}
	return p.makeExprStmt(expr)
}

func (p *Parser) parseAtSayGet() Stmt {
	p.next() // @
	stmt := &SayGetStmt{}
	stmt.Row = p.parseExpr()
	if !p.check(T_COMMA) {
		return stmt
	}
	p.next()
	stmt.Col = p.parseExpr()
	if !p.check(T_SAY) && !p.check(T_GET) {
		return stmt
	}
	if p.match(T_SAY) {
		stmt.SayExpr = p.parseExpr()
	}
	if p.match(T_GET) {
		if p.check(T_IDENTIFIER) {
			stmt.GetVar = p.next().Lexeme
			// Handle Alias->Field reference
			if p.check(T_ARROW) {
				p.next()
				if p.check(T_IDENTIFIER) {
					stmt.GetVar = stmt.GetVar + "->" + p.next().Lexeme
				}
			}
		} else {
			stmt.GetExpr = p.parseExpr()
		}
	}
	// PICTURE clause
	if p.check(T_IDENTIFIER) && strings.ToUpper(p.peek().Lexeme) == "PICTURE" {
		p.next()
		if picExpr, ok := p.parseExpr().(*StringExpr); ok {
			stmt.Picture = picExpr.Value
		}
	}
	return stmt
}

func (p *Parser) makeExprStmt(expr Expr) Stmt {
	return &ExprStmt{Expr: expr}
}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) nodeMarker() {}
func (e *ExprStmt) stmtNode()   {}

// ----- Expression parsing (precedence climbing) -----

var precedences = map[TokenType]int{
	T_OR:      1,
	T_AND:     2,
	T_NOT:     3,
	T_EQ:      4,
	T_ASSIGN:  4,
	T_NE:      4,
	T_LT:      4,
	T_GT:      4,
	T_LE:      4,
	T_GE:      4,
	T_DOLLAR:  4,
	T_CONCAT:  5,
	T_PLUS:    5,
	T_MINUS:   5,
	T_STAR:    6,
	T_SLASH:   6,
	T_PERCENT: 6,
	T_CARET:   7,
}

func (p *Parser) parseExpr() Expr {
	return p.parseBinary(0)
}

func (p *Parser) parseBinary(minPrec int) Expr {
	var left Expr
	if p.match(T_MINUS) {
		inner := p.parsePrimary()
		left = &UnaryExpr{Op: T_MINUS, Inner: inner}
	} else if p.matchOneOf(T_NOT, T_DOT) {
		inner := p.parsePrimary()
		left = &UnaryExpr{Op: T_NOT, Inner: inner}
	} else if p.match(T_PLUS) {
		left = p.parsePrimary()
	} else {
		left = p.parsePrimary()
	}

	for {
		tok := p.peek()
		prec, isOp := precedences[tok.Type]
		if !isOp || prec < minPrec {
			break
		}
		p.next()
		right := p.parseBinary(prec + 1)
		left = &BinaryExpr{Left: left, Op: tok.Type, Right: right}
	}
	return left
}

func (p *Parser) parsePrimary() Expr {
	tok := p.next()
	switch tok.Type {
	case T_NUMBER:
		if tok.FloatVal != float64(int64(tok.FloatVal)) || strings.Contains(tok.Lexeme, ".") || strings.Contains(tok.Lexeme, "e") || strings.Contains(tok.Lexeme, "E") {
			return &NumberExpr{Value: tok.FloatVal, IsInt: false}
		}
		return &NumberExpr{Value: tok.FloatVal, IntValue: tok.IntVal, IsInt: true}
	case T_STRING:
		return &StringExpr{Value: tok.Lexeme}
	case T_LOGICAL:
		upper := strings.ToUpper(tok.Lexeme)
		val := upper == ".T." || upper == ".Y."
		return &BoolExpr{Value: val}
	case T_IDENTIFIER:
		// Check for function call
		if p.check(T_LPAREN) {
			p.next()
			var args []Expr
			if !p.check(T_RPAREN) {
				args = p.parseExprList()
			}
			p.expect(T_RPAREN)
			return &FuncCallExpr{Name: tok.Lexeme, Args: args}
		}
		// Check for assignment
		if p.checkOneOf(T_ASSIGN, T_COLONEQ) {
			p.next()
			val := p.parseExpr()
			return &AssignmentStmt{Target: tok.Lexeme, Expr: val}
		}
		// Check for field reference: Alias->Field
		if p.check(T_ARROW) {
			p.next()
			if fieldTok, ok := p.expect(T_IDENTIFIER); ok {
				return &FieldRefExpr{Alias: tok.Lexeme, Field: fieldTok.Lexeme}
			}
		}
		return &IdentExpr{Name: tok.Lexeme}
	case T_LPAREN:
		expr := p.parseExpr()
		p.expect(T_RPAREN)
		return expr
	case T_DOT:
		nextTok := p.peek()
		if nextTok.Type == T_IDENTIFIER {
			p.next()
			if p.check(T_DOT) {
				p.next()
				return &IdentExpr{Name: "." + nextTok.Lexeme + "."}
			}
		}
		return &IdentExpr{Name: "."}
	default:
		p.error(tok, "unexpected token in expression: %s", tok.Type)
		return &IdentExpr{Name: tok.Lexeme}
	}
}

func (p *Parser) parseExprList() []Expr {
	var exprs []Expr
	exprs = append(exprs, p.parseExpr())
	for p.match(T_COMMA) {
		p.consumeNewlines()
		exprs = append(exprs, p.parseExpr())
	}
	return exprs
}

func (p *Parser) consumeNewlines() {
	for p.check(T_SEMI) {
		p.next()
	}
}

func (p *Parser) isTerminal(typ TokenType) bool {
	switch typ {
	case T_ENDIF, T_ENDDO, T_ENDFOR, T_ELSE, T_OTHERWISE, T_CASE, T_ENDCASE, T_ENDTEXT:
		return true
	}
	return false
}

func (p *Parser) checkOneOf(types ...TokenType) bool {
	for _, typ := range types {
		if p.check(typ) {
			return true
		}
	}
	return false
}

func isOneOf(typ TokenType, types []TokenType) bool {
	for _, t := range types {
		if typ == t {
			return true
		}
	}
	return false
}
