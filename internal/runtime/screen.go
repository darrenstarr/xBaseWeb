package runtime

// Screen captures the display state produced by @ SAY / GET / CLEAR.
type Screen struct {
	Lines   []ScreenLine     `json:"lines"`
	Fields  []ScreenField    `json:"fields"`
	Prompt  string           `json:"prompt,omitempty"`
	Confirm string           `json:"confirm,omitempty"` // confirmation message
	Wait    bool             `json:"wait,omitempty"`
	Done    bool             `json:"done,omitempty"`
	Result  string           `json:"result,omitempty"`
	SQL     string           `json:"sql,omitempty"`
	Cols    []string         `json:"cols,omitempty"`
	Table   *TableData       `json:"table,omitempty"`
	Title   string           `json:"title,omitempty"`
	Tagline string           `json:"tagline,omitempty"`
	Nav     map[string]string `json:"nav,omitempty"`
}

type ScreenLine struct {
	Row  int    `json:"row"`
	Col  int    `json:"col"`
	Text string `json:"text"`
}

type ScreenField struct {
	Row     int    `json:"row"`
	Col     int    `json:"col"`
	Var     string `json:"var"`
	Picture string `json:"picture,omitempty"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type"`
}

type TableData struct {
	Columns    []TableColumn `json:"columns"`
	Rows       [][]string    `json:"rows"`
	Actions    []RowAction   `json:"actions,omitempty"`
	KeyCol     int           `json:"keyCol,omitempty"`
	Query      string        `json:"query,omitempty"`
	Limit      int           `json:"limit,omitempty"`
	Offset     int           `json:"offset,omitempty"`
	Total      int           `json:"total,omitempty"`
	SearchCols []string      `json:"searchCols,omitempty"` // columns searchable via SEARCH clause
}

type TableColumn struct {
	Name  string `json:"name"`
	Align string `json:"align,omitempty"`
}

type RowAction struct {
	Label     string `json:"label"`
	Procedure string `json:"procedure"`
}
