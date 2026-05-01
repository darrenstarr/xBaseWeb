package compiler

import (
	"testing"
)

func lex(input string) []Token {
	l := NewLexer(input)
	tokens, _ := l.Lex()
	// Strip EOF for convenience in most tests
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == T_EOF {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

func lexFull(input string) ([]Token, []string) {
	l := NewLexer(input)
	tokens, errs := l.Lex()
	return tokens, errs
}

func assertToken(t *testing.T, tok Token, typ TokenType, lexeme string) {
	t.Helper()
	if tok.Type != typ {
		t.Errorf("expected type %v, got %v (lexeme=%q)", typ, tok.Type, tok.Lexeme)
	}
	if tok.Lexeme != lexeme {
		t.Errorf("expected lexeme %q, got %q", lexeme, tok.Lexeme)
	}
}

// ---------- Basic tokens ----------

func TestEmptyInput(t *testing.T) {
	tokens, errs := lexFull("")
	if len(tokens) != 1 || tokens[0].Type != T_EOF {
		t.Errorf("expected single EOF token, got %d tokens", len(tokens))
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	tokens, errs := lexFull("   \t  \r  ")
	if len(tokens) != 1 || tokens[0].Type != T_EOF {
		t.Errorf("expected single EOF token, got %d tokens", len(tokens))
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestWhitespaceAndNewlines(t *testing.T) {
	tokens, errs := lexFull("   \t\n  \r\n  \n")
	// Each \n produces a ;, consecutive newlines produce one
	if len(tokens) != 2 || tokens[0].Type != T_SEMI || tokens[1].Type != T_EOF {
		t.Errorf("expected SEMI + EOF, got %d tokens", len(tokens))
		for i, tok := range tokens {
			t.Logf("  token %d: Type=%v Lexeme=%q", i, tok.Type, tok.Lexeme)
		}
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestPunctuation(t *testing.T) {
	input := "()[]{},;.::=->@#$%^*+-/\\"
	toks := lex(input)
	expected := []struct {
		typ    TokenType
		lexeme string
	}{
		{T_LPAREN, "("},
		{T_RPAREN, ")"},
		{T_LBRACKET, "["},
		{T_RBRACKET, "]"},
		{T_LBRACE, "{"},
		{T_RBRACE, "}"},
		{T_COMMA, ","},
		{T_SEMI, ";"},
		{T_DOT, "."},
		{T_COLON, ":"},
		{T_COLONEQ, ":="},
		{T_ARROW, "->"},
		{T_ATSIGN, "@"},
		{T_HASH, "#"},
		{T_DOLLAR, "$"},
		{T_PERCENT, "%"},
		{T_CARET, "^"},
		{T_STAR, "*"},
		{T_PLUS, "+"},
		{T_MINUS, "-"},
		{T_SLASH, "/"},
		{T_HASH, "\\"},
	}
	if len(toks) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(toks), toks)
	}
	for i, e := range expected {
		assertToken(t, toks[i], e.typ, e.lexeme)
	}
}

func TestNewlinesToSemicolons(t *testing.T) {
	input := "USE\nSELECT 1"
	toks := lex(input)
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_SEMI, ";")
	assertToken(t, toks[2], T_SELECT, "SELECT")
	assertToken(t, toks[3], T_NUMBER, "1")
}

func TestNewlineAfterIdentifier(t *testing.T) {
	input := "USE customers\nSELECT 1"
	toks := lex(input)
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_IDENTIFIER, "customers")
	assertToken(t, toks[2], T_SEMI, ";")
	assertToken(t, toks[3], T_SELECT, "SELECT")
	assertToken(t, toks[4], T_NUMBER, "1")
}

func TestMultipleNewlines(t *testing.T) {
	input := "a\n\n\nb"
	toks := lex(input)
	assertToken(t, toks[0], T_IDENTIFIER, "a")
	assertToken(t, toks[1], T_SEMI, ";")
	assertToken(t, toks[2], T_IDENTIFIER, "b")
	if len(toks) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(toks))
	}
}

// ---------- Comments ----------

func TestLineComment(t *testing.T) {
	input := "USE && this is a comment\nSELECT"
	toks := lex(input)
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_SEMI, ";")
	assertToken(t, toks[2], T_SELECT, "SELECT")
	if len(toks) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(toks))
	}
}

func TestLineCommentAtEOF(t *testing.T) {
	input := "USE && just a comment"
	toks := lex(input)
	assertToken(t, toks[0], T_USE, "USE")
	if len(toks) != 1 {
		t.Errorf("expected 1 token, got %d", len(toks))
	}
}

func TestBlockComment(t *testing.T) {
	input := "USE /* multi\nline */ SELECT"
	toks := lex(input)
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_SELECT, "SELECT")
}

func TestUnterminatedBlockComment(t *testing.T) {
	_, errs := lexFull("/* oops")
	if len(errs) == 0 {
		t.Error("expected error for unterminated block comment")
	}
}

// ---------- Strings ----------

func TestDoubleQuotedString(t *testing.T) {
	toks := lex(`"hello world"`)
	assertToken(t, toks[0], T_STRING, "hello world")
}

func TestSingleQuotedString(t *testing.T) {
	toks := lex(`'hello world'`)
	assertToken(t, toks[0], T_STRING, "hello world")
}

func TestStringWithNewline(t *testing.T) {
	toks := lex("\"hello\nworld\"")
	assertToken(t, toks[0], T_STRING, "hello\nworld")
}

func TestEmptyString(t *testing.T) {
	toks := lex(`""`)
	assertToken(t, toks[0], T_STRING, "")
}

func TestUnterminatedString(t *testing.T) {
	_, errs := lexFull(`"no end`)
	if len(errs) == 0 {
		t.Error("expected error for unterminated string")
	}
}

func TestStringWithEscapes(t *testing.T) {
	toks := lex(`"it's fine"`)
	assertToken(t, toks[0], T_STRING, "it's fine")
}

// ---------- Numbers ----------

func TestInteger(t *testing.T) {
	toks := lex("12345")
	assertToken(t, toks[0], T_NUMBER, "12345")
	if toks[0].IntVal != 12345 {
		t.Errorf("expected IntVal=12345, got %d", toks[0].IntVal)
	}
	if toks[0].FloatVal != 12345.0 {
		t.Errorf("expected FloatVal=12345, got %f", toks[0].FloatVal)
	}
}

func TestFloat(t *testing.T) {
	toks := lex("123.45")
	assertToken(t, toks[0], T_NUMBER, "123.45")
	if toks[0].FloatVal != 123.45 {
		t.Errorf("expected FloatVal=123.45, got %f", toks[0].FloatVal)
	}
}

func TestLeadingDot(t *testing.T) {
	toks := lex("0.5")
	assertToken(t, toks[0], T_NUMBER, "0.5")
}

func TestScientificNotation(t *testing.T) {
	toks := lex("1.5e10")
	assertToken(t, toks[0], T_NUMBER, "1.5e10")
	if toks[0].FloatVal != 1.5e10 {
		t.Errorf("expected 1.5e10, got %f", toks[0].FloatVal)
	}
}

func TestScientificNotationNegative(t *testing.T) {
	toks := lex("1.5e-3")
	assertToken(t, toks[0], T_NUMBER, "1.5e-3")
	if toks[0].FloatVal != 0.0015 {
		t.Errorf("expected 0.0015, got %f", toks[0].FloatVal)
	}
}

func TestHexNumber(t *testing.T) {
	toks := lex("0xFF")
	assertToken(t, toks[0], T_NUMBER, "0xFF")
	if toks[0].IntVal != 255 {
		t.Errorf("expected IntVal=255, got %d", toks[0].IntVal)
	}
}

func TestZero(t *testing.T) {
	toks := lex("0")
	assertToken(t, toks[0], T_NUMBER, "0")
	if toks[0].IntVal != 0 {
		t.Errorf("expected IntVal=0, got %d", toks[0].IntVal)
	}
}

func TestNegativeNumber(t *testing.T) {
	toks := lex("-42")
	assertToken(t, toks[0], T_MINUS, "-")
	assertToken(t, toks[1], T_NUMBER, "42")
}

// ---------- Identifiers and keywords ----------

func TestIdentifier(t *testing.T) {
	toks := lex("myVar")
	assertToken(t, toks[0], T_IDENTIFIER, "myVar")
}

func TestIdentifierWithUnderscore(t *testing.T) {
	toks := lex("my_var_123")
	assertToken(t, toks[0], T_IDENTIFIER, "my_var_123")
}

func TestKeywordAsIdentifier(t *testing.T) {
	toks := lex("USE")
	assertToken(t, toks[0], T_USE, "USE")
}

func TestKeywordCaseInsensitive(t *testing.T) {
	toks := lex("use Use uSe")
	assertToken(t, toks[0], T_USE, "use")
	assertToken(t, toks[1], T_USE, "Use")
	assertToken(t, toks[2], T_USE, "uSe")
}

func TestAllKeywords(t *testing.T) {
	inputs := []struct {
		input string
		typ   TokenType
	}{
		{"USE", T_USE},
		{"SELECT", T_SELECT},
		{"SET", T_SET},
		{"REPLACE", T_REPLACE},
		{"APPEND", T_APPEND},
		{"BLANK", T_BLANK},
		{"DELETE", T_DELETE},
		{"PACK", T_PACK},
		{"ZAP", T_ZAP},
		{"GO", T_GO},
		{"GOTO", T_GOTO},
		{"SKIP", T_SKIP},
		{"LOCATE", T_LOCATE},
		{"CONTINUE", T_CONTINUE},
		{"SEEK", T_SEEK},
		{"FIND", T_FIND},
		{"COUNT", T_COUNT},
		{"SUM", T_SUM},
		{"AVERAGE", T_AVERAGE},
		{"AVG", T_AVG},
		{"SORT", T_SORT},
		{"INDEX", T_INDEX},
		{"DO", T_DO},
		{"WHILE", T_WHILE},
		{"ENDDO", T_ENDDO},
		{"FOR", T_FOR},
		{"ENDFOR", T_ENDFOR},
		{"IF", T_IF},
		{"ELSE", T_ELSE},
		{"ENDIF", T_ENDIF},
		{"PROCEDURE", T_PROCEDURE},
		{"FUNCTION", T_FUNCTION},
		{"RETURN", T_RETURN},
		{"PARAMETERS", T_PARAMETERS},
		{"PRIVATE", T_PRIVATE},
		{"PUBLIC", T_PUBLIC},
		{"LOCAL", T_LOCAL},
		{"STATIC", T_STATIC},
		{"LOOP", T_LOOP},
		{"EXIT", T_EXIT},
		{"STORE", T_STORE},
		{"INPUT", T_INPUT},
		{"ACCEPT", T_ACCEPT},
		{"WAIT", T_WAIT},
		{"CLEAR", T_CLEAR},
		{"CLOSE", T_CLOSE},
		{"QUIT", T_QUIT},
		{"CANCEL", T_CANCEL},
		{"DEFINE", T_DEFINE},
		{"CREATE", T_CREATE},
		{"MODIFY", T_MODIFY},
		{"ERASE", T_ERASE},
		{"RENAME", T_RENAME},
		{"TYPE", T_TYPE},
		{"DIR", T_DIR},
		{"LIST", T_LIST},
		{"DISPLAY", T_DISPLAY},
		{"SAY", T_SAY},
		{"GET", T_GET},
		{"READ", T_READ},
		{"WITH", T_WITH},
		{"OTHERWISE", T_OTHERWISE},
		{"CASE", T_CASE},
		{"MEMVAR", T_MEMVAR},
		{"FIELD", T_FIELD},
		{"IN", T_IN},
		{"TO", T_TO},
		{"ALL", T_ALL},
		{"NEXT", T_NEXT},
		{"RECORD", T_RECORD},
		{"REST", T_REST},
		{"TEXT", T_TEXT},
		{"ENDTEXT", T_ENDTEXT},
		{"SCATTER", T_SCATTER},
		{"GATHER", T_GATHER},
		{"NOT", T_NOT},
		{"AND", T_AND},
		{"OR", T_OR},
	}
	for _, tc := range inputs {
		t.Run(tc.input, func(t *testing.T) {
			toks := lex(tc.input)
			if len(toks) != 1 {
				t.Fatalf("expected 1 token, got %d", len(toks))
			}
			assertToken(t, toks[0], tc.typ, tc.input)
		})
	}
}

// ---------- Dot operators and logical literals ----------

func TestDotAnd(t *testing.T) {
	toks := lex(".AND.")
	assertToken(t, toks[0], T_AND, ".AND.")
}

func TestDotOr(t *testing.T) {
	toks := lex(".OR.")
	assertToken(t, toks[0], T_OR, ".OR.")
}

func TestDotNot(t *testing.T) {
	toks := lex(".NOT.")
	assertToken(t, toks[0], T_NOT, ".NOT.")
}

func TestLogicalTrue(t *testing.T) {
	toks := lex(".T.")
	assertToken(t, toks[0], T_LOGICAL, ".T.")
}

func TestLogicalFalse(t *testing.T) {
	toks := lex(".F.")
	assertToken(t, toks[0], T_LOGICAL, ".F.")
}

func TestLogicalYes(t *testing.T) {
	toks := lex(".Y.")
	assertToken(t, toks[0], T_LOGICAL, ".Y.")
}

func TestLogicalNo(t *testing.T) {
	toks := lex(".N.")
	assertToken(t, toks[0], T_LOGICAL, ".N.")
}

func TestLogicalNull(t *testing.T) {
	toks := lex(".NULL.")
	assertToken(t, toks[0], T_LOGICAL, ".NULL.")
}

func testDotOperatorCaseInsensitive(t *testing.T) {
	toks := lex(".and. .And. .AND.")
	assertToken(t, toks[0], T_AND, ".and.")
	assertToken(t, toks[1], T_AND, ".And.")
	assertToken(t, toks[2], T_AND, ".AND.")
}

func TestDotAsFieldSeparator(t *testing.T) {
	toks := lex("Customer.Phone")
	assertToken(t, toks[0], T_IDENTIFIER, "Customer")
	assertToken(t, toks[1], T_DOT, ".")
	assertToken(t, toks[2], T_IDENTIFIER, "Phone")
}

// ---------- Comparison operators ----------

func TestComparisons(t *testing.T) {
	input := "= == != <> < > <= >="
	toks := lex(input)
	assertToken(t, toks[0], T_ASSIGN, "=")
	assertToken(t, toks[1], T_EQ, "==")
	assertToken(t, toks[2], T_NE, "!=")
	assertToken(t, toks[3], T_NE, "<>")
	assertToken(t, toks[4], T_LT, "<")
	assertToken(t, toks[5], T_GT, ">")
	assertToken(t, toks[6], T_LE, "<=")
	assertToken(t, toks[7], T_GE, ">=")
}

// ---------- Arrow and concat ----------

func TestArrow(t *testing.T) {
	toks := lex("Customer->Phone")
	assertToken(t, toks[0], T_IDENTIFIER, "Customer")
	assertToken(t, toks[1], T_ARROW, "->")
	assertToken(t, toks[2], T_IDENTIFIER, "Phone")
}

func TestConcat(t *testing.T) {
	toks := lex("'hello' ++ 'world'")
	assertToken(t, toks[0], T_STRING, "hello")
	assertToken(t, toks[1], T_CONCAT, "++")
	assertToken(t, toks[2], T_STRING, "world")
}

// ---------- Date literals via curly braces ----------

func TestDateLiteral(t *testing.T) {
	toks := lex("{^2024-01-15}")
	assertToken(t, toks[0], T_DATELITERAL, "{^2024-01-15}")
}

func TestCurlyNonDate(t *testing.T) {
	toks := lex("{hello}")
	assertToken(t, toks[0], T_LBRACE, "{")
}

func TestUnterminatedDateLiteral(t *testing.T) {
	_, errs := lexFull("{^2024-01-15")
	if len(errs) == 0 {
		t.Error("expected error for unterminated date literal")
	}
}

// ---------- Error handling ----------

func TestUnknownCharacter(t *testing.T) {
	_, errs := lexFull("`backtick`")
	if len(errs) == 0 {
		t.Error("expected error for backtick")
	}
}

func TestLexerErrorCollectsAll(t *testing.T) {
	_, errs := lexFull("`a` `b`")
	if len(errs) != 4 {
		t.Errorf("expected 4 errors (one per backtick), got %d: %v", len(errs), errs)
	}
}

// ---------- Line and column tracking ----------

func TestLineColumnTracking(t *testing.T) {
	input := "a\nb\nc"
	l := NewLexer(input)
	tokens, _ := l.Lex()
	if len(tokens) < 4 {
		t.Fatalf("expected at least 4 tokens, got %d", len(tokens))
	}
	// a at line 1, col 1
	if tokens[0].Line != 1 || tokens[0].Column != 1 {
		t.Errorf("expected 'a' at 1:1, got %d:%d", tokens[0].Line, tokens[0].Column)
	}
	// \n -> ; at line 1, col 2
	if tokens[1].Type != T_SEMI {
		t.Errorf("expected SEMI after 'a', got %v", tokens[1].Type)
	}
	// b at line 2, col 1
	if tokens[2].Line != 2 || tokens[2].Column != 1 {
		t.Errorf("expected 'b' at 2:1, got %d:%d", tokens[2].Line, tokens[2].Column)
	}
}

// ---------- Real xBase snippets ----------

func TestUseStatement(t *testing.T) {
	input := "USE customers ALIAS cust"
	toks := lex(input)
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_IDENTIFIER, "customers")
	assertToken(t, toks[2], T_IDENTIFIER, "ALIAS")
	assertToken(t, toks[3], T_IDENTIFIER, "cust")
}

func TestReplaceStatement(t *testing.T) {
	input := "REPLACE Name WITH 'John'"
	toks := lex(input)
	assertToken(t, toks[0], T_REPLACE, "REPLACE")
	assertToken(t, toks[1], T_IDENTIFIER, "Name")
	assertToken(t, toks[2], T_WITH, "WITH")
	assertToken(t, toks[3], T_STRING, "John")
}

func TestIfBlock(t *testing.T) {
	input := "IF x > 5\n  y = 10\nENDIF"
	toks := lex(input)
	assertToken(t, toks[0], T_IF, "IF")
	assertToken(t, toks[1], T_IDENTIFIER, "x")
	assertToken(t, toks[2], T_GT, ">")
	assertToken(t, toks[3], T_NUMBER, "5")
	assertToken(t, toks[4], T_SEMI, ";")
	assertToken(t, toks[5], T_IDENTIFIER, "y")
	assertToken(t, toks[6], T_ASSIGN, "=")
	assertToken(t, toks[7], T_NUMBER, "10")
	assertToken(t, toks[8], T_SEMI, ";")
	assertToken(t, toks[9], T_ENDIF, "ENDIF")
}

func TestDoWhileLoop(t *testing.T) {
	input := "DO WHILE .NOT. EOF()\n  SKIP\nENDDO"
	toks := lex(input)
	assertToken(t, toks[0], T_DO, "DO")
	assertToken(t, toks[1], T_WHILE, "WHILE")
	assertToken(t, toks[2], T_NOT, ".NOT.")
	assertToken(t, toks[3], T_IDENTIFIER, "EOF")
	assertToken(t, toks[4], T_LPAREN, "(")
	assertToken(t, toks[5], T_RPAREN, ")")
	assertToken(t, toks[6], T_SEMI, ";")
	assertToken(t, toks[7], T_SKIP, "SKIP")
	assertToken(t, toks[8], T_SEMI, ";")
	assertToken(t, toks[9], T_ENDDO, "ENDDO")
}

func TestProcedureDefinition(t *testing.T) {
	input := "PROCEDURE MyProc\n  ? 'hello'\nRETURN"
	toks := lex(input)
	assertToken(t, toks[0], T_PROCEDURE, "PROCEDURE")
	assertToken(t, toks[1], T_IDENTIFIER, "MyProc")
	assertToken(t, toks[2], T_SEMI, ";")
	assertToken(t, toks[3], T_ATSIGN, "?")
	assertToken(t, toks[4], T_STRING, "hello")
	assertToken(t, toks[5], T_SEMI, ";")
	assertToken(t, toks[6], T_RETURN, "RETURN")
}

func TestAtSayGet(t *testing.T) {
	input := "@ 10,5 SAY 'Name:' GET mName"
	toks := lex(input)
	assertToken(t, toks[0], T_ATSIGN, "@")
	assertToken(t, toks[1], T_NUMBER, "10")
	assertToken(t, toks[2], T_COMMA, ",")
	assertToken(t, toks[3], T_NUMBER, "5")
	assertToken(t, toks[4], T_SAY, "SAY")
	assertToken(t, toks[5], T_STRING, "Name:")
	assertToken(t, toks[6], T_GET, "GET")
	assertToken(t, toks[7], T_IDENTIFIER, "mName")
}

func TestAppendBlank(t *testing.T) {
	toks := lex("APPEND BLANK")
	assertToken(t, toks[0], T_APPEND, "APPEND")
	assertToken(t, toks[1], T_BLANK, "BLANK")
}

func TestSelectStatement(t *testing.T) {
	toks := lex("SELECT 0")
	assertToken(t, toks[0], T_SELECT, "SELECT")
	assertToken(t, toks[1], T_NUMBER, "0")
}

func TestGoToStatement(t *testing.T) {
	toks := lex("GO TOP")
	assertToken(t, toks[0], T_GO, "GO")
	assertToken(t, toks[1], T_IDENTIFIER, "TOP")
}

func TestSkipStatement(t *testing.T) {
	toks := lex("SKIP -1")
	assertToken(t, toks[0], T_SKIP, "SKIP")
	assertToken(t, toks[1], T_MINUS, "-")
	assertToken(t, toks[2], T_NUMBER, "1")
}

func TestSetStatement(t *testing.T) {
	toks := lex("SET TALK OFF")
	assertToken(t, toks[0], T_SET, "SET")
	assertToken(t, toks[1], T_IDENTIFIER, "TALK")
	assertToken(t, toks[2], T_IDENTIFIER, "OFF")
}

// ---------- Edge cases ----------

func TestTabCharacters(t *testing.T) {
	toks := lex("\tUSE\tcustomers\t")
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_IDENTIFIER, "customers")
}

func TestCarriageReturn(t *testing.T) {
	toks := lex("USE\r\nSELECT")
	assertToken(t, toks[0], T_USE, "USE")
	assertToken(t, toks[1], T_SEMI, ";")
	assertToken(t, toks[2], T_SELECT, "SELECT")
}

func TestLeadingWhitespace(t *testing.T) {
	toks := lex("   USE")
	assertToken(t, toks[0], T_USE, "USE")
}

func TestTrailingWhitespace(t *testing.T) {
	toks := lex("USE   ")
	assertToken(t, toks[0], T_USE, "USE")
}

func TestOnlyComments(t *testing.T) {
	tokens, errs := lexFull("&& this is a comment\n&& another comment")
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	// Each line comment ends with \n which produces a ;, then EOF
	if len(tokens) != 2 || tokens[0].Type != T_SEMI || tokens[1].Type != T_EOF {
		t.Errorf("expected SEMI + EOF, got %d tokens", len(tokens))
	}
}

func TestExponentWithoutDigits(t *testing.T) {
	_, errs := lexFull("1.5e")
	if len(errs) != 0 {
		t.Errorf("expected no errors for incomplete exponent, got %v", errs)
	}
}

func TestVeryLongIdentifier(t *testing.T) {
	long := ""
	for i := 0; i < 1000; i++ {
		long += "a"
	}
	toks := lex(long)
	if len(toks) != 1 || toks[0].Type != T_IDENTIFIER {
		t.Errorf("expected single identifier token")
	}
	if len(toks[0].Lexeme) != 1000 {
		t.Errorf("expected 1000 chars, got %d", len(toks[0].Lexeme))
	}
}

func TestStringContainingKeyword(t *testing.T) {
	toks := lex(`"USE SELECT IF ENDIF"`)
	assertToken(t, toks[0], T_STRING, "USE SELECT IF ENDIF")
}

func TestMixedOperators(t *testing.T) {
	input := "a + b * c - d / e ^ f % g"
	toks := lex(input)
	expectedTypes := []TokenType{
		T_IDENTIFIER, T_PLUS, T_IDENTIFIER, T_STAR, T_IDENTIFIER,
		T_MINUS, T_IDENTIFIER, T_SLASH, T_IDENTIFIER, T_CARET,
		T_IDENTIFIER, T_PERCENT, T_IDENTIFIER,
	}
	if len(toks) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTypes), len(toks))
	}
	for i, typ := range expectedTypes {
		if toks[i].Type != typ {
			t.Errorf("token %d: expected %v, got %v", i, typ, toks[i].Type)
		}
	}
}

func TestIdentifierWithDigits(t *testing.T) {
	toks := lex("var123")
	assertToken(t, toks[0], T_IDENTIFIER, "var123")
}

func TestIdentifierStartingWithUnderscore(t *testing.T) {
	toks := lex("_private")
	assertToken(t, toks[0], T_IDENTIFIER, "_private")
}

func TestMultipleDotsInRow(t *testing.T) {
	toks := lex("a..b")
	assertToken(t, toks[0], T_IDENTIFIER, "a")
	assertToken(t, toks[1], T_DOT, ".")
	assertToken(t, toks[2], T_DOT, ".")
	assertToken(t, toks[3], T_IDENTIFIER, "b")
}

func TestDoubleAsterisk(t *testing.T) {
	toks := lex("a ** b")
	assertToken(t, toks[0], T_IDENTIFIER, "a")
	assertToken(t, toks[1], T_STAR, "**")
	assertToken(t, toks[2], T_IDENTIFIER, "b")
}

func TestPlusEquals(t *testing.T) {
	toks := lex("x += 1")
	assertToken(t, toks[0], T_IDENTIFIER, "x")
	assertToken(t, toks[1], T_ASSIGN, "+=")
	assertToken(t, toks[2], T_NUMBER, "1")
}

func TestMinusEquals(t *testing.T) {
	toks := lex("x -= 1")
	assertToken(t, toks[0], T_IDENTIFIER, "x")
	assertToken(t, toks[1], T_ASSIGN, "-=")
	assertToken(t, toks[2], T_NUMBER, "1")
}

func TestStartsWithMinusDash(t *testing.T) {
	toks := lex("---a")
	assertToken(t, toks[0], T_MINUS, "-")
	assertToken(t, toks[1], T_MINUS, "-")
	assertToken(t, toks[2], T_MINUS, "-")
	assertToken(t, toks[3], T_IDENTIFIER, "a")
}

func TestQuestionQuery(t *testing.T) {
	toks := lex("?\n??")
	assertToken(t, toks[0], T_ATSIGN, "?")
	assertToken(t, toks[1], T_SEMI, ";")
	assertToken(t, toks[2], T_ATSIGN, "??")
}

func TestNotEquals(t *testing.T) {
	toks := lex("x != y")
	assertToken(t, toks[0], T_IDENTIFIER, "x")
	assertToken(t, toks[1], T_NE, "!=")
	assertToken(t, toks[2], T_IDENTIFIER, "y")
}

func TestBangNot(t *testing.T) {
	toks := lex("!x")
	assertToken(t, toks[0], T_NOT, "!")
	assertToken(t, toks[1], T_IDENTIFIER, "x")
}

func TestStoreStatement(t *testing.T) {
	toks := lex("STORE 0 TO x")
	assertToken(t, toks[0], T_STORE, "STORE")
	assertToken(t, toks[1], T_NUMBER, "0")
	assertToken(t, toks[2], T_TO, "TO")
	assertToken(t, toks[3], T_IDENTIFIER, "x")
}

func TestForLoop(t *testing.T) {
	toks := lex("FOR i = 1 TO 10\n  ? i\nENDFOR")
	assertToken(t, toks[0], T_FOR, "FOR")
	assertToken(t, toks[1], T_IDENTIFIER, "i")
	assertToken(t, toks[2], T_ASSIGN, "=")
	assertToken(t, toks[3], T_NUMBER, "1")
	assertToken(t, toks[4], T_TO, "TO")
	assertToken(t, toks[5], T_NUMBER, "10")
	assertToken(t, toks[6], T_SEMI, ";")
	assertToken(t, toks[7], T_ATSIGN, "?")
	assertToken(t, toks[8], T_IDENTIFIER, "i")
	assertToken(t, toks[9], T_SEMI, ";")
	assertToken(t, toks[10], T_ENDFOR, "ENDFOR")
}

func TestParameters(t *testing.T) {
	toks := lex("PARAMETERS a, b, c")
	assertToken(t, toks[0], T_PARAMETERS, "PARAMETERS")
	assertToken(t, toks[1], T_IDENTIFIER, "a")
	assertToken(t, toks[2], T_COMMA, ",")
	assertToken(t, toks[3], T_IDENTIFIER, "b")
	assertToken(t, toks[4], T_COMMA, ",")
	assertToken(t, toks[5], T_IDENTIFIER, "c")
}

func TestPrivatePublicLocalStatic(t *testing.T) {
	inputs := []string{"PRIVATE x", "PUBLIC y", "LOCAL z", "STATIC w"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			toks := lex(input)
			if len(toks) != 2 {
				t.Fatalf("expected 2 tokens, got %d: %v", len(toks), toks)
			}
		})
	}
}
