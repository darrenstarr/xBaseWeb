package runtime

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type ValueType int

const (
	TypeNil ValueType = iota
	TypeNumber
	TypeString
	TypeLogical
	TypeDate
	TypeArray
	TypeObject
)

func (t ValueType) String() string {
	switch t {
	case TypeNil:
		return "NIL"
	case TypeNumber:
		return "N"
	case TypeString:
		return "C"
	case TypeLogical:
		return "L"
	case TypeDate:
		return "D"
	case TypeArray:
		return "A"
	case TypeObject:
		return "O"
	default:
		return "?"
	}
}

type Value struct {
	Type    ValueType
	Number  float64
	Str     string
	Logical bool
	Date    time.Time
	Array   []Value
	Fields  map[string]Value
}

// ----- Constructors -----

func Nil() Value {
	return Value{Type: TypeNil}
}

func NewNumber(n float64) Value {
	return Value{Type: TypeNumber, Number: n}
}

func NewString(s string) Value {
	return Value{Type: TypeString, Str: s}
}

func NewLogical(b bool) Value {
	return Value{Type: TypeLogical, Logical: b}
}

func NewDate(t time.Time) Value {
	return Value{Type: TypeDate, Date: t}
}

func NewArray(vals []Value) Value {
	return Value{Type: TypeArray, Array: vals}
}

func NewObject() Value {
	return Value{Type: TypeObject, Fields: make(map[string]Value)}
}

// ----- Accessors -----

func (v Value) AsBool() bool {
	switch v.Type {
	case TypeLogical:
		return v.Logical
	case TypeNumber:
		return v.Number != 0
	case TypeString:
		s := strings.TrimSpace(v.Str)
		return strings.EqualFold(s, ".T.") || strings.EqualFold(s, ".Y.") || s == "1"
	case TypeNil:
		return false
	default:
		return false
	}
}

func (v Value) AsInt() int64 {
	if v.Type == TypeNumber {
		return int64(v.Number)
	}
	return 0
}

func (v Value) IsTruthy() bool {
	return v.AsBool()
}

// ----- Arithmetic -----

func (a Value) Add(b Value) Value {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		return NewNumber(a.Number + b.Number)
	}
	if a.Type == TypeString && b.Type == TypeString {
		return NewString(a.Str + b.Str)
	}
	if a.Type == TypeDate && b.Type == TypeNumber {
		return NewDate(a.Date.AddDate(0, 0, int(b.Number)))
	}
	if a.Type == TypeNumber && b.Type == TypeDate {
		return NewDate(b.Date.AddDate(0, 0, int(a.Number)))
	}
	return Nil()
}

func (a Value) Sub(b Value) Value {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		return NewNumber(a.Number - b.Number)
	}
	if a.Type == TypeDate && b.Type == TypeNumber {
		return NewDate(a.Date.AddDate(0, 0, -int(b.Number)))
	}
	if a.Type == TypeDate && b.Type == TypeDate {
		return NewNumber(float64(a.Date.Sub(b.Date).Hours() / 24))
	}
	return Nil()
}

func (a Value) Mul(b Value) Value {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		return NewNumber(a.Number * b.Number)
	}
	return Nil()
}

func (a Value) Div(b Value) Value {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		if b.Number == 0 {
			return Nil()
		}
		return NewNumber(a.Number / b.Number)
	}
	return Nil()
}

func (a Value) Mod(b Value) Value {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		if b.Number == 0 {
			return Nil()
		}
		return NewNumber(float64(int64(a.Number) % int64(b.Number)))
	}
	return Nil()
}

func (a Value) Pow(b Value) Value {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		return NewNumber(math.Pow(a.Number, b.Number))
	}
	return Nil()
}

func (a Value) Concat(b Value) Value {
	if a.Type == TypeString && b.Type == TypeString {
		return NewString(strings.TrimRight(a.Str, " ") + b.Str)
	}
	return a.Add(b)
}

// ----- Comparison -----

func (a Value) Eq(b Value) Value {
	// In xBase, Nil equals empty string
	if a.Type == TypeNil && b.Type == TypeString && b.Str == "" {
		return NewLogical(true)
	}
	if a.Type == TypeString && a.Str == "" && b.Type == TypeNil {
		return NewLogical(true)
	}
	if a.Type != b.Type {
		return NewLogical(false)
	}
	switch a.Type {
	case TypeNil:
		return NewLogical(true)
	case TypeNumber:
		return NewLogical(a.Number == b.Number)
	case TypeString:
		return NewLogical(a.Str == b.Str)
	case TypeLogical:
		return NewLogical(a.Logical == b.Logical)
	case TypeDate:
		return NewLogical(a.Date.Equal(b.Date))
	default:
		return NewLogical(false)
	}
}

func (a Value) Neq(b Value) Value {
	r := a.Eq(b)
	if r.Type == TypeLogical {
		return NewLogical(!r.Logical)
	}
	return NewLogical(true)
}

func (a Value) Lt(b Value) Value {
	switch a.Type {
	case TypeNumber:
		if b.Type == TypeNumber {
			return NewLogical(a.Number < b.Number)
		}
	case TypeString:
		if b.Type == TypeString {
			return NewLogical(a.Str < b.Str)
		}
	case TypeDate:
		if b.Type == TypeDate {
			return NewLogical(a.Date.Before(b.Date))
		}
	}
	return NewLogical(false)
}

func (a Value) Le(b Value) Value {
	switch a.Type {
	case TypeNumber:
		if b.Type == TypeNumber {
			return NewLogical(a.Number <= b.Number)
		}
	case TypeString:
		if b.Type == TypeString {
			return NewLogical(a.Str <= b.Str)
		}
	case TypeDate:
		if b.Type == TypeDate {
			return NewLogical(!a.Date.After(b.Date))
		}
	}
	return NewLogical(false)
}

func (a Value) Gt(b Value) Value {
	r := a.Le(b)
	if r.Type == TypeLogical {
		return NewLogical(!r.Logical && a.Eq(b).Logical == false)
	}
	return NewLogical(false)
}

func (a Value) Ge(b Value) Value {
	r := a.Lt(b)
	if r.Type == TypeLogical {
		return NewLogical(!r.Logical)
	}
	return NewLogical(false)
}

func (a Value) Contains(b Value) Value {
	if a.Type == TypeString && b.Type == TypeString {
		return NewLogical(strings.Contains(a.Str, b.Str))
	}
	return NewLogical(false)
}

// ----- Logical -----

func (a Value) And(b Value) Value {
	return NewLogical(a.IsTruthy() && b.IsTruthy())
}

func (a Value) Or(b Value) Value {
	return NewLogical(a.IsTruthy() || b.IsTruthy())
}

func (a Value) Not() Value {
	return NewLogical(!a.IsTruthy())
}

// ----- Conversion -----

func (v Value) String() string {
	switch v.Type {
	case TypeNil:
		return "NIL"
	case TypeNumber:
		return fmt.Sprintf("%v", v.Number)
	case TypeString:
		return v.Str
	case TypeLogical:
		if v.Logical {
			return ".T."
		}
		return ".F."
	case TypeDate:
		return v.Date.Format("2006-01-02")
	case TypeArray:
		parts := make([]string, len(v.Array))
		for i, e := range v.Array {
			parts[i] = e.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return "?"
	}
}
