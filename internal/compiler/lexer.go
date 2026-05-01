package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	start  int
	startCol int
	width  int
	tokens []Token
	errors []string
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) Lex() ([]Token, []string) {
	l.tokens = nil
	l.errors = nil
	for {
		tok := l.next()
		l.tokens = append(l.tokens, tok)
		if tok.Type == T_EOF {
			break
		}
		if tok.Type == T_ERROR && l.pos < len(l.input) {
			continue
		}
		if tok.Type == T_ERROR {
			break
		}
	}
	return l.tokens, l.errors
}

func (l *Lexer) next() Token {
	l.skipWhitespace()
	l.start = l.pos
	l.startCol = l.col

	if l.pos >= len(l.input) {
		return l.makeToken(T_EOF, "")
	}

	r := l.peek()

	switch {
	case r == '\n':
		for l.pos < len(l.input) && (l.peek() == '\n' || l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\r') {
			if l.peek() == '\n' {
				l.line++
				l.col = 1
			}
			l.pos++
		}
		return l.makeToken(T_SEMI, ";")
	case r == '&':
		return l.lexAmpersand()
	case r == '/' && l.peekNext() == '*':
		return l.lexBlockComment()
	case r == '*' && l.peekNext() == '*':
		l.advance()
		l.advance()
		return l.makeToken(T_STAR, "**")
	case r == '(':
		l.advance()
		return l.makeToken(T_LPAREN, "(")
	case r == ')':
		l.advance()
		return l.makeToken(T_RPAREN, ")")
	case r == '[':
		l.advance()
		return l.makeToken(T_LBRACKET, "[")
	case r == ']':
		l.advance()
		return l.makeToken(T_RBRACKET, "]")
	case r == ',':
		l.advance()
		return l.makeToken(T_COMMA, ",")
	case r == '#':
		l.advance()
		return l.makeToken(T_HASH, "#")
	case r == '.':
		return l.lexDotOperator()
	case r == '+' && l.peekNext() == '+':
		l.advance()
		l.advance()
		return l.makeToken(T_CONCAT, "++")
	case r == '+' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return l.makeToken(T_ASSIGN, "+=")
	case r == '-':
		if l.peekNext() == '>' {
			l.advance()
			l.advance()
			return l.makeToken(T_ARROW, "->")
		}
		if l.peekNext() == '=' {
			l.advance()
			l.advance()
			return l.makeToken(T_ASSIGN, "-=")
		}
		l.advance()
		return l.makeToken(T_MINUS, "-")
	case r == '0' && (l.peekNext() == 'x' || l.peekNext() == 'X'):
		return l.lexHexNumber()
	case r == '{':
		if l.peekNext() == '^' {
			return l.lexDateLiteral()
		}
		l.advance()
		return l.makeToken(T_LBRACE, "{")
	case r == '}':
		l.advance()
		return l.makeToken(T_RBRACE, "}")
	case r == '"' || r == '\'':
		return l.lexString()
	case r == '@':
		l.advance()
		return l.makeToken(T_ATSIGN, "@")
	case r == ';':
		l.advance()
		return l.makeToken(T_SEMI, ";")
	case r == ':':
		if l.peekNext() == '=' {
			l.advance()
			l.advance()
			return l.makeToken(T_COLONEQ, ":=")
		}
		l.advance()
		return l.makeToken(T_COLON, ":")
	case r == '=':
		if l.peekNext() == '=' {
			l.advance()
			l.advance()
			return l.makeToken(T_EQ, "==")
		}
		l.advance()
		return l.makeToken(T_ASSIGN, "=")
	case r == '<':
		if l.peekNext() == '=' {
			l.advance()
			l.advance()
			return l.makeToken(T_LE, "<=")
		}
		if l.peekNext() == '>' {
			l.advance()
			l.advance()
			return l.makeToken(T_NE, "<>")
		}
		l.advance()
		return l.makeToken(T_LT, "<")
	case r == '>':
		if l.peekNext() == '=' {
			l.advance()
			l.advance()
			return l.makeToken(T_GE, ">=")
		}
		l.advance()
		return l.makeToken(T_GT, ">")
	case r == '!' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return l.makeToken(T_NE, "!=")
	case r == '!':
		l.advance()
		return l.makeToken(T_NOT, "!")
	case r == '$':
		l.advance()
		return l.makeToken(T_DOLLAR, "$")
	case r == '*':
		l.advance()
		return l.makeToken(T_STAR, "*")
	case r == '/':
		l.advance()
		return l.makeToken(T_SLASH, "/")
	case r == '%':
		l.advance()
		return l.makeToken(T_PERCENT, "%")
	case r == '^':
		l.advance()
		return l.makeToken(T_CARET, "^")
	case r == '~':
		l.advance()
		return l.makeToken(T_MINUS, "~")
	case r == '+':
		l.advance()
		return l.makeToken(T_PLUS, "+")
	case r == '\\':
		l.advance()
		return l.makeToken(T_HASH, "\\")
	default:
		if unicode.IsLetter(rune(r)) || r == '_' {
			return l.lexIdentifierOrKeyword()
		}
		if unicode.IsDigit(rune(r)) {
			return l.lexNumber()
		}
		if r == '?' {
			return l.lexQuestion()
		}
		l.advance()
		return l.errorToken("unexpected character '%c'", r)
	}
}

func (l *Lexer) lexQuestion() Token {
	l.advance()
	if l.peek() == '?' {
		l.advance()
		return l.makeToken(T_ATSIGN, "??")
	}
	return l.makeToken(T_ATSIGN, "?")
}

func (l *Lexer) lexAmpersand() Token {
	l.advance()
	if l.peek() == '&' {
		l.advance()
		for l.pos < len(l.input) && l.peek() != '\n' {
			l.advance()
		}
		return l.next()
	}
	return l.makeToken(T_CONCAT, "&")
}

func (l *Lexer) lexBlockComment() Token {
	l.advance()
	l.advance()
	for l.pos < len(l.input) {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance()
			l.advance()
			return l.next()
		}
		if l.peek() == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
	return l.errorToken("unterminated block comment")
}

func (l *Lexer) lexDotOperator() Token {
	start := l.pos
	l.advance()

	for l.pos < len(l.input) && unicode.IsLetter(rune(l.peek())) {
		l.advance()
	}

	word := l.input[start+1 : l.pos]

	if l.pos >= len(l.input) || l.peek() != '.' || len(word) == 0 {
		l.pos = start + 1
		l.col++
		return l.makeToken(T_DOT, ".")
	}

	l.advance()

	switch strings.ToUpper(word) {
	case "AND":
		return l.makeToken(T_AND, ".AND.")
	case "OR":
		return l.makeToken(T_OR, ".OR.")
	case "NOT":
		return l.makeToken(T_NOT, ".NOT.")
	case "T":
		return l.makeToken(T_LOGICAL, ".T.")
	case "F":
		return l.makeToken(T_LOGICAL, ".F.")
	case "Y":
		return l.makeToken(T_LOGICAL, ".Y.")
	case "N":
		return l.makeToken(T_LOGICAL, ".N.")
	case "NULL":
		return l.makeToken(T_LOGICAL, ".NULL.")
	default:
		lexeme := l.input[start:l.pos]
		return l.makeToken(T_IDENTIFIER, lexeme)
	}
}

func (l *Lexer) lexDateLiteral() Token {
	l.advance() // {
	l.advance() // ^
	start := l.pos

	for l.pos < len(l.input) && l.peek() != '}' {
		if l.peek() == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}

	if l.pos >= len(l.input) {
		return l.errorToken("unterminated date literal")
	}

	content := l.input[start:l.pos]
	l.advance() // }

	return l.makeToken(T_DATELITERAL, "{^"+content+"}")
}

func (l *Lexer) lexString() Token {
	quote := l.peek()
	l.advance()
	var buf strings.Builder
	for {
		if l.pos >= len(l.input) {
			return l.errorToken("unterminated string")
		}
		r := l.peek()
		if r == '\n' {
			l.line++
			l.col = 1
			buf.WriteByte(r)
			l.pos++
			continue
		}
		if r == quote {
			l.advance()
			return l.makeToken(T_STRING, buf.String())
		}
		buf.WriteByte(r)
		l.advance()
	}
}

func (l *Lexer) lexHexNumber() Token {
	l.advance()
	l.advance()
	start := l.pos
	for l.pos < len(l.input) && isHexDigit(l.peek()) {
		l.advance()
	}
	lexeme := l.input[start:l.pos]
	val, _ := strconv.ParseInt(lexeme, 16, 64)
	tok := l.makeToken(T_NUMBER, "0x"+lexeme)
	tok.IntVal = val
	tok.FloatVal = float64(val)
	return tok
}

func (l *Lexer) lexNumber() Token {
	start := l.pos
	isFloat := false
	for l.pos < len(l.input) && (unicode.IsDigit(rune(l.peek())) || l.peek() == '.') {
		if l.peek() == '.' {
			if isFloat {
				break
			}
			isFloat = true
		}
		l.advance()
	}

	if l.pos < len(l.input) && (l.peek() == 'e' || l.peek() == 'E') {
		l.advance()
		if l.pos < len(l.input) && (l.peek() == '+' || l.peek() == '-') {
			l.advance()
		}
		for l.pos < len(l.input) && unicode.IsDigit(rune(l.peek())) {
			l.advance()
		}
		isFloat = true
	}

	lexeme := l.input[start:l.pos]
	tok := l.makeToken(T_NUMBER, lexeme)
	if isFloat {
		tok.FloatVal, _ = strconv.ParseFloat(lexeme, 64)
		tok.IntVal = int64(tok.FloatVal)
	} else {
		tok.IntVal, _ = strconv.ParseInt(lexeme, 10, 64)
		tok.FloatVal = float64(tok.IntVal)
	}
	return tok
}

var keywords = map[string]TokenType{
	"USE":        T_USE,
	"SELECT":     T_SELECT,
	"SET":        T_SET,
	"REPLACE":    T_REPLACE,
	"APPEND":     T_APPEND,
	"BLANK":      T_BLANK,
	"DELETE":     T_DELETE,
	"PACK":       T_PACK,
	"ZAP":        T_ZAP,
	"GO":         T_GO,
	"GOTO":       T_GOTO,
	"SKIP":       T_SKIP,
	"LOCATE":     T_LOCATE,
	"CONTINUE":   T_CONTINUE,
	"SEEK":       T_SEEK,
	"FIND":       T_FIND,
	"COUNT":      T_COUNT,
	"SUM":        T_SUM,
	"AVERAGE":    T_AVERAGE,
	"AVG":        T_AVG,
	"SORT":       T_SORT,
	"INDEX":      T_INDEX,
	"DO":         T_DO,
	"WHILE":      T_WHILE,
	"ENDDO":      T_ENDDO,
	"FOR":        T_FOR,
	"ENDFOR":     T_ENDFOR,
	"IF":         T_IF,
	"ELSE":       T_ELSE,
	"ENDIF":      T_ENDIF,
	"PROCEDURE":  T_PROCEDURE,
	"FUNCTION":   T_FUNCTION,
	"RETURN":     T_RETURN,
	"PARAMETERS": T_PARAMETERS,
	"PRIVATE":    T_PRIVATE,
	"PUBLIC":     T_PUBLIC,
	"LOCAL":      T_LOCAL,
	"STATIC":     T_STATIC,
	"LOOP":       T_LOOP,
	"EXIT":       T_EXIT,
	"STORE":      T_STORE,
	"INPUT":      T_INPUT,
	"ACCEPT":     T_ACCEPT,
	"WAIT":       T_WAIT,
	"CLEAR":      T_CLEAR,
	"CLOSE":      T_CLOSE,
	"QUIT":       T_QUIT,
	"CANCEL":     T_CANCEL,
	"DEFINE":     T_DEFINE,
	"CREATE":     T_CREATE,
	"MODIFY":     T_MODIFY,
	"ERASE":      T_ERASE,
	"RENAME":     T_RENAME,
	"TYPE":       T_TYPE,
	"DIR":        T_DIR,
	"LIST":       T_LIST,
	"DISPLAY":    T_DISPLAY,
	"SAY":        T_SAY,
	"GET":        T_GET,
	"READ":       T_READ,
	"WITH":       T_WITH,
	"OTHERWISE":  T_OTHERWISE,
	"CASE":       T_CASE,
	"ENDCASE":    T_ENDCASE,
	"MEMVAR":     T_MEMVAR,
	"FIELD":      T_FIELD,
	"IN":         T_IN,
	"TO":         T_TO,
	"ALL":        T_ALL,
	"NEXT":       T_NEXT,
	"RECORD":     T_RECORD,
	"REST":       T_REST,
	"STEP":       T_STEP,
	"RUNSQL":     T_RUNSQL,
	"NAV":        T_NAV,
	"TEXT":       T_TEXT,
	"ENDTEXT":    T_ENDTEXT,
	"SCATTER":    T_SCATTER,
	"GATHER":     T_GATHER,
	"CALCULATE":  T_CALCULATE,
	"CALL":       T_CALL,
	"NOT":        T_NOT,
	"AND":        T_AND,
	"OR":         T_OR,
}

func (l *Lexer) lexIdentifierOrKeyword() Token {
	for l.pos < len(l.input) {
		r := l.peek()
		if unicode.IsLetter(rune(r)) || unicode.IsDigit(rune(r)) || r == '_' {
			l.advance()
		} else {
			break
		}
	}

	lexeme := l.input[l.start:l.pos]
	upper := strings.ToUpper(lexeme)

	if t, ok := keywords[upper]; ok {
		return l.makeToken(t, lexeme)
	}
	return l.makeToken(T_IDENTIFIER, lexeme)
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) advance() {
	if l.pos >= len(l.input) {
		return
	}
	r := l.input[l.pos]
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.pos++
}

func (l *Lexer) makeToken(typ TokenType, lexeme string) Token {
	return Token{
		Type:   typ,
		Lexeme: lexeme,
		Line:   l.line,
		Column: l.startCol,
	}
}

func (l *Lexer) errorToken(format string, args ...interface{}) Token {
	msg := fmt.Sprintf(format, args...)
	l.errors = append(l.errors, fmt.Sprintf("line %d:%d: %s", l.line, l.col, msg))
	return Token{
		Type:   T_ERROR,
		Lexeme: msg,
		Line:   l.line,
		Column: l.col,
	}
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
