package compiler

type TokenType int

const (
	T_EOF TokenType = iota
	T_ERROR
	T_IDENTIFIER
	T_NUMBER
	T_STRING
	T_DATELITERAL
	T_LOGICAL

	// Keywords
	T_USE
	T_SELECT
	T_SET
	T_REPLACE
	T_APPEND
	T_BLANK
	T_DELETE
	T_PACK
	T_ZAP
	T_GO
	T_GOTO
	T_SKIP
	T_LOCATE
	T_CONTINUE
	T_SEEK
	T_FIND
	T_COUNT
	T_SUM
	T_AVERAGE
	T_SORT
	T_INDEX
	T_DO
	T_WHILE
	T_ENDDO
	T_FOR
	T_ENDFOR
	T_IF
	T_ELSE
	T_ENDIF
	T_PROCEDURE
	T_FUNCTION
	T_RETURN
	T_PARAMETERS
	T_PRIVATE
	T_PUBLIC
	T_LOCAL
	T_STATIC
	T_LOOP
	T_EXIT
	T_STORE
	T_INPUT
	T_ACCEPT
	T_WAIT
	T_CLEAR
	T_CLOSE
	T_QUIT
	T_CANCEL
	T_DEFINE
	T_CREATE
	T_MODIFY
	T_ERASE
	T_RENAME
	T_TYPE
	T_DIR
	T_LIST
	T_DISPLAY
	T_SAY
	T_GET
	T_READ
	T_WITH
	T_OTHERWISE
	T_CASE
	T_ENDCASE
	T_MENU
	T_NOT
	T_AND
	T_OR
	T_MEMVAR
	T_FIELD
	T_IN
	T_TO
	T_ALL
	T_NEXT
	T_RECORD
	T_REST
	T_DATABASE
	T_DBTITLE
	T_SKIP2
	T_MOVE
	T_STEP
	T_RUNSQL
	T_NAV
	T_CONFIRM
	T_TEXT
	T_ENDTEXT
	T_SCATTER
	T_GATHER
	T_AVG
	T_CALCULATE

	// Parser support only (not keyword-matched)
	T_CALL

	// Operators
	T_PLUS
	T_MINUS
	T_STAR
	T_SLASH
	T_CARET
	T_PERCENT
	T_EQ
	T_NE
	T_LT
	T_GT
	T_LE
	T_GE
	T_CONCAT
	T_ASSIGN
	T_DOLLAR
	T_DEREF

	// Delimiters
	T_LPAREN
	T_RPAREN
	T_LBRACKET
	T_RBRACKET
	T_LBRACE
	T_RBRACE
	T_COMMA
	T_SEMI
	T_DOT
	T_COLON
	T_ARROW
	T_COLONEQ
	T_HASH
	T_ATSIGN
	T_ATSAY
	T_ATGET

	MAX_TOKEN
)

var tokenNames = map[TokenType]string{
	T_EOF:           "EOF",
	T_ERROR:         "ERROR",
	T_IDENTIFIER:    "IDENTIFIER",
	T_NUMBER:        "NUMBER",
	T_STRING:        "STRING",
	T_DATELITERAL:   "DATELITERAL",
	T_LOGICAL:       "LOGICAL",
	T_USE:           "USE",
	T_SELECT:        "SELECT",
	T_SET:           "SET",
	T_REPLACE:       "REPLACE",
	T_APPEND:        "APPEND",
	T_BLANK:         "BLANK",
	T_DELETE:        "DELETE",
	T_PACK:          "PACK",
	T_ZAP:           "ZAP",
	T_GO:            "GO",
	T_GOTO:          "GOTO",
	T_SKIP:          "SKIP",
	T_LOCATE:        "LOCATE",
	T_CONTINUE:      "CONTINUE",
	T_SEEK:          "SEEK",
	T_FIND:          "FIND",
	T_COUNT:         "COUNT",
	T_SUM:           "SUM",
	T_AVERAGE:        "AVERAGE",
	T_SORT:          "SORT",
	T_INDEX:         "INDEX",
	T_DO:            "DO",
	T_WHILE:         "WHILE",
	T_ENDDO:         "ENDDO",
	T_FOR:           "FOR",
	T_ENDFOR:        "ENDFOR",
	T_IF:            "IF",
	T_ELSE:          "ELSE",
	T_ENDIF:         "ENDIF",
	T_PROCEDURE:     "PROCEDURE",
	T_FUNCTION:      "FUNCTION",
	T_RETURN:        "RETURN",
	T_PARAMETERS:    "PARAMETERS",
	T_PRIVATE:       "PRIVATE",
	T_PUBLIC:        "PUBLIC",
	T_LOCAL:         "LOCAL",
	T_STATIC:        "STATIC",
	T_LOOP:          "LOOP",
	T_EXIT:          "EXIT",
	T_STORE:         "STORE",
	T_INPUT:         "INPUT",
	T_ACCEPT:        "ACCEPT",
	T_WAIT:          "WAIT",
	T_CLEAR:         "CLEAR",
	T_CLOSE:         "CLOSE",
	T_QUIT:          "QUIT",
	T_CANCEL:        "CANCEL",
	T_DEFINE:        "DEFINE",
	T_CREATE:        "CREATE",
	T_MODIFY:        "MODIFY",
	T_ERASE:         "ERASE",
	T_RENAME:        "RENAME",
	T_TYPE:          "TYPE",
	T_DIR:           "DIR",
	T_LIST:          "LIST",
	T_DISPLAY:       "DISPLAY",
	T_SAY:           "SAY",
	T_GET:           "GET",
	T_READ:          "READ",
	T_WITH:          "WITH",
	T_OTHERWISE:     "OTHERWISE",
	T_CASE:          "CASE",
	T_ENDCASE:       "ENDCASE",
	T_MENU:          "MENU",
	T_NOT:           "NOT",
	T_AND:           "AND",
	T_OR:            "OR",
	T_MEMVAR:        "MEMVAR",
	T_FIELD:         "FIELD",
	T_IN:            "IN",
	T_TO:            "TO",
	T_ALL:           "ALL",
	T_NEXT:          "NEXT",
	T_RECORD:        "RECORD",
	T_REST:          "REST",
	T_STEP:          "STEP",
	T_RUNSQL:        "RUNSQL",
	T_NAV:           "NAV",
	T_CONFIRM:       "CONFIRM",
	T_PLUS:          "+",
	T_MINUS:         "-",
	T_STAR:          "*",
	T_SLASH:         "/",
	T_CARET:         "^",
	T_PERCENT:       "%",
	T_EQ:            "=",
	T_NE:            "!=",
	T_LT:            "<",
	T_GT:            ">",
	T_LE:            "<=",
	T_GE:            ">=",
	T_CONCAT:        "++",
	T_ASSIGN:        ":=",
	T_DOLLAR:        "$",
	T_LPAREN:        "(",
	T_RPAREN:        ")",
	T_LBRACKET:      "[",
	T_RBRACKET:      "]",
	T_LBRACE:        "{",
	T_RBRACE:        "}",
	T_COMMA:         ",",
	T_SEMI:          ";",
	T_DOT:           ".",
	T_COLON:         ":",
	T_ARROW:         "->",
	T_HASH:          "#",
	T_ATSIGN:        "@",
	T_TEXT:          "TEXT",
	T_ENDTEXT:       "ENDTEXT",
	T_SCATTER:       "SCATTER",
	T_GATHER:        "GATHER",
	T_CALCULATE:     "CALCULATE",
	T_CALL:          "CALL",
}

func (t TokenType) String() string {
	if s, ok := tokenNames[t]; ok {
		return s
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Lexeme  string
	Line    int
	Column  int
	FloatVal float64
	IntVal   int64
}

func (t Token) String() string {
	return t.Lexeme
}
