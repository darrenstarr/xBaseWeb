package runtime

import (
	"fmt"
	"strings"
)

// WorkArea represents a database alias / work area, analogous to
// a SELECT work area in xBase.
type WorkArea struct {
	Alias      string
	TableName  string
	RecNo      int64
	LastRec    int64
	BOF        bool
	EOF        bool
	Found      bool
	Fields     map[string]Value
	Orders     []string
	OrderNum   int
	Deleted    bool
	Exclusive  bool
	ReadOnly   bool
	Filter     string
	Scope      string
}

func NewWorkArea(alias string) *WorkArea {
	return &WorkArea{
		Alias:  alias,
		Fields: make(map[string]Value),
		RecNo:  1,
		LastRec: 0,
		BOF:     true,
		EOF:     true,
	}
}

func (wa *WorkArea) GoTop() {
	wa.RecNo = 1
	wa.BOF = true
	wa.EOF = false
}

func (wa *WorkArea) GoBottom() {
	if wa.LastRec > 0 {
		wa.RecNo = wa.LastRec
	}
	wa.BOF = false
	wa.EOF = false
}

func (wa *WorkArea) GoTo(n int64) {
	if n < 1 {
		n = 1
	}
	if n > wa.LastRec && wa.LastRec > 0 {
		n = wa.LastRec
	}
	wa.RecNo = n
	wa.BOF = n == 1
	wa.EOF = n >= wa.LastRec
}

func (wa *WorkArea) Skip(n int64) {
	target := wa.RecNo + n
	if target < 1 {
		target = 1
	}
	if wa.LastRec > 0 && target > wa.LastRec {
		target = wa.LastRec + 1
	}
	wa.RecNo = target
	wa.BOF = target <= 1
	wa.EOF = target > wa.LastRec
	if target > wa.LastRec {
		wa.RecNo = wa.LastRec + 1
	}
}

func (wa *WorkArea) GetField(name string) Value {
	if v, ok := wa.Fields[name]; ok {
		return v
	}
	return Nil()
}

func (wa *WorkArea) SetField(name string, val Value) {
	wa.Fields[name] = val
}

func (wa *WorkArea) AppendBlank() {
	wa.LastRec++
	wa.RecNo = wa.LastRec
	wa.EOF = false
	wa.BOF = false
	wa.Fields = make(map[string]Value)
}

func (wa *WorkArea) Delete() {
	wa.Deleted = true
}

func (wa *WorkArea) Recall() {
	wa.Deleted = false
}

func (wa *WorkArea) Seek(val Value) bool {
	wa.Found = false
	return false
}

// ----- WorkAreaManager -----

type WorkAreaManager struct {
	areas     []*WorkArea
	current   int
	maxWorkAreas int
}

func NewWorkAreaManager() *WorkAreaManager {
	return &WorkAreaManager{
		areas:       make([]*WorkArea, 225),
		current:     1,
		maxWorkAreas: 225,
	}
}

func (m *WorkAreaManager) Select(n int) error {
	if n < 0 || n >= m.maxWorkAreas {
		return fmt.Errorf("invalid work area: %d", n)
	}
	if n == 0 {
		n = m.nextAvailable()
	}
	m.current = n
	if m.areas[n] == nil {
		m.areas[n] = NewWorkArea(fmt.Sprintf("A%d", n))
	}
	return nil
}

func (m *WorkAreaManager) SelectAlias(alias string) bool {
	for i, wa := range m.areas {
		if wa != nil && strings.EqualFold(wa.Alias, alias) {
			m.current = i
			return true
		}
	}
	return false
}

func (m *WorkAreaManager) Current() *WorkArea {
	if m.areas[m.current] == nil {
		m.areas[m.current] = NewWorkArea(fmt.Sprintf("A%d", m.current))
	}
	return m.areas[m.current]
}

func (m *WorkAreaManager) Get(alias string) *WorkArea {
	for _, wa := range m.areas {
		if wa != nil && strings.EqualFold(wa.Alias, alias) {
			return wa
		}
	}
	return nil
}

func (m *WorkAreaManager) Used() []*WorkArea {
	var used []*WorkArea
	for _, wa := range m.areas {
		if wa != nil {
			used = append(used, wa)
		}
	}
	return used
}

func (m *WorkAreaManager) CloseAll() {
	for i := range m.areas {
		m.areas[i] = nil
	}
	m.current = 1
}

func (m *WorkAreaManager) Close(alias string) {
	for i, wa := range m.areas {
		if wa != nil && strings.EqualFold(wa.Alias, alias) {
			m.areas[i] = nil
			if m.current == i {
				m.current = 1
			}
			break
		}
	}
}

func (m *WorkAreaManager) Use(alias string, table string) *WorkArea {
	wa := NewWorkArea(alias)
	wa.TableName = table
	m.areas[m.current] = wa
	return wa
}

func (m *WorkAreaManager) CurrentNum() int {
	return m.current
}

func (m *WorkAreaManager) nextAvailable() int {
	for i := 1; i < m.maxWorkAreas; i++ {
		if m.areas[i] == nil {
			return i
		}
	}
	return m.maxWorkAreas - 1
}
