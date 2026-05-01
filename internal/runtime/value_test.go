package runtime

import (
	"math"
	"testing"
	"time"
)

func assertValue(t *testing.T, v Value, typ ValueType, msg string) {
	t.Helper()
	if v.Type != typ {
		t.Errorf("%s: expected type %v, got %v (value=%v)", msg, typ, v.Type, v)
	}
}

func assertNumber(t *testing.T, v Value, expected float64) {
	t.Helper()
	if v.Type != TypeNumber {
		t.Fatalf("expected TypeNumber, got %v (value=%v)", v.Type, v)
	}
	if v.Number != expected {
		t.Errorf("expected number %v, got %v", expected, v.Number)
	}
}

func assertString(t *testing.T, v Value, expected string) {
	t.Helper()
	if v.Type != TypeString {
		t.Fatalf("expected TypeString, got %v (value=%v)", v.Type, v)
	}
	if v.Str != expected {
		t.Errorf("expected string %q, got %q", expected, v.Str)
	}
}

func assertLogical(t *testing.T, v Value, expected bool) {
	t.Helper()
	if v.Type != TypeLogical {
		t.Fatalf("expected TypeLogical, got %v (value=%v)", v.Type, v)
	}
	if v.Logical != expected {
		t.Errorf("expected logical %v, got %v", expected, v.Logical)
	}
}

func assertDate(t *testing.T, v Value, expected time.Time) {
	t.Helper()
	if v.Type != TypeDate {
		t.Fatalf("expected TypeDate, got %v (value=%v)", v.Type, v)
	}
	if !v.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, v.Date)
	}
}

// ---------- Constructors ----------

func TestNilValue(t *testing.T) {
	v := Nil()
	assertValue(t, v, TypeNil, "Nil()")
}

func TestNewNumber(t *testing.T) {
	v := NewNumber(42.5)
	assertValue(t, v, TypeNumber, "NewNumber")
	assertNumber(t, v, 42.5)
}

func TestNewNumberInt(t *testing.T) {
	v := NewNumber(100)
	assertNumber(t, v, 100)
}

func TestNewNumberZero(t *testing.T) {
	v := NewNumber(0)
	assertNumber(t, v, 0)
}

func TestNewNumberNegative(t *testing.T) {
	v := NewNumber(-3.14)
	assertNumber(t, v, -3.14)
}

func TestNewString(t *testing.T) {
	v := NewString("hello")
	assertValue(t, v, TypeString, "NewString")
	assertString(t, v, "hello")
}

func TestNewStringEmpty(t *testing.T) {
	v := NewString("")
	assertString(t, v, "")
}

func TestNewStringUnicode(t *testing.T) {
	v := NewString("héllo 世界")
	assertString(t, v, "héllo 世界")
}

func TestNewLogicalTrue(t *testing.T) {
	v := NewLogical(true)
	assertLogical(t, v, true)
}

func TestNewLogicalFalse(t *testing.T) {
	v := NewLogical(false)
	assertLogical(t, v, false)
}

func TestNewDate(t *testing.T) {
	d := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	v := NewDate(d)
	assertDate(t, v, d)
}

func TestNewDateZero(t *testing.T) {
	var zero time.Time
	v := NewDate(zero)
	assertDate(t, v, zero)
}

func TestNewArray(t *testing.T) {
	v := NewArray([]Value{NewNumber(1), NewString("two")})
	assertValue(t, v, TypeArray, "NewArray")
	if len(v.Array) != 2 {
		t.Fatalf("expected array of length 2, got %d", len(v.Array))
	}
}

func TestNewArrayEmpty(t *testing.T) {
	v := NewArray(nil)
	assertValue(t, v, TypeArray, "NewArray empty")
	if len(v.Array) != 0 {
		t.Errorf("expected empty array, got len %d", len(v.Array))
	}
}

func TestNewObject(t *testing.T) {
	v := NewObject()
	assertValue(t, v, TypeObject, "NewObject")
	if v.Fields == nil {
		t.Fatal("expected non-nil Fields map")
	}
}

// ---------- AsBool ----------

func TestAsBoolLogical(t *testing.T) {
	if !NewLogical(true).AsBool() {
		t.Error("expected true")
	}
	if NewLogical(false).AsBool() {
		t.Error("expected false")
	}
}

func TestAsBoolNumber(t *testing.T) {
	if !NewNumber(1).AsBool() {
		t.Error("expected number 1 to be truthy")
	}
	if !NewNumber(-1).AsBool() {
		t.Error("expected number -1 to be truthy")
	}
	if NewNumber(0).AsBool() {
		t.Error("expected number 0 to be falsy")
	}
}

func TestAsBoolString(t *testing.T) {
	if !NewString(".T.").AsBool() {
		t.Error("expected .T. to be truthy")
	}
	if NewString("").AsBool() {
		t.Error("expected empty string to be falsy")
	}
	if NewString("hello").AsBool() {
		t.Error("expected non-empty non-logical string to be falsy")
	}
}

func TestAsBoolNil(t *testing.T) {
	if Nil().AsBool() {
		t.Error("expected nil to be falsy")
	}
}

func TestIsTruthy(t *testing.T) {
	if !NewNumber(42).IsTruthy() {
		t.Error("42 should be truthy")
	}
	if NewNumber(0).IsTruthy() {
		t.Error("0 should be falsy")
	}
	if !NewLogical(true).IsTruthy() {
		t.Error("true should be truthy")
	}
}

// ---------- Arithmetic ----------

func TestAddNumbers(t *testing.T) {
	assertNumber(t, NewNumber(5).Add(NewNumber(3)), 8)
}

func TestAddNegative(t *testing.T) {
	assertNumber(t, NewNumber(5).Add(NewNumber(-3)), 2)
}

func TestAddZero(t *testing.T) {
	assertNumber(t, NewNumber(5).Add(NewNumber(0)), 5)
}

func TestAddDecimals(t *testing.T) {
	assertNumber(t, NewNumber(1.5).Add(NewNumber(2.5)), 4.0)
}

func TestAddStrings(t *testing.T) {
	assertString(t, NewString("hello ").Add(NewString("world")), "hello world")
}

func TestAddStringEmpty(t *testing.T) {
	assertString(t, NewString("a").Add(NewString("")), "a")
}

func TestAddDateAndNumber(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	v := NewDate(d).Add(NewNumber(10))
	assertDate(t, v, time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC))
}

func TestAddNumberAndDate(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	v := NewNumber(10).Add(NewDate(d))
	assertDate(t, v, time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC))
}

func TestAddNil(t *testing.T) {
	v := NewNumber(5).Add(Nil())
	assertValue(t, v, TypeNil, "number + nil")
}

func TestAddTypeMismatch(t *testing.T) {
	v := NewNumber(5).Add(NewString("x"))
	assertValue(t, v, TypeNil, "number + string")
}

func TestSubNumbers(t *testing.T) {
	assertNumber(t, NewNumber(10).Sub(NewNumber(3)), 7)
}

func TestSubNegativeResult(t *testing.T) {
	assertNumber(t, NewNumber(3).Sub(NewNumber(10)), -7)
}

func TestSubDateAndNumber(t *testing.T) {
	d := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	v := NewDate(d).Sub(NewNumber(5))
	assertDate(t, v, time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC))
}

func TestSubDates(t *testing.T) {
	d1 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	v := NewDate(d1).Sub(NewDate(d2))
	assertNumber(t, v, 9) // 9 days
}

func TestSubDateZeroDay(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	v := NewDate(d).Sub(NewDate(d))
	assertNumber(t, v, 0)
}

func TestMulNumbers(t *testing.T) {
	assertNumber(t, NewNumber(5).Mul(NewNumber(3)), 15)
}

func TestMulByZero(t *testing.T) {
	assertNumber(t, NewNumber(5).Mul(NewNumber(0)), 0)
}

func TestMulNegative(t *testing.T) {
	assertNumber(t, NewNumber(5).Mul(NewNumber(-2)), -10)
}

func TestMulTypeMismatch(t *testing.T) {
	v := NewNumber(5).Mul(NewString("x"))
	assertValue(t, v, TypeNil, "number * string")
}

func TestDivNumbers(t *testing.T) {
	assertNumber(t, NewNumber(10).Div(NewNumber(2)), 5)
}

func TestDivNonInteger(t *testing.T) {
	assertNumber(t, NewNumber(7).Div(NewNumber(2)), 3.5)
}

func TestDivByZero(t *testing.T) {
	v := NewNumber(5).Div(NewNumber(0))
	assertValue(t, v, TypeNil, "division by zero")
}

func TestDivTypeMismatch(t *testing.T) {
	v := NewNumber(5).Div(NewString("x"))
	assertValue(t, v, TypeNil, "number / string")
}

func TestModNumbers(t *testing.T) {
	assertNumber(t, NewNumber(10).Mod(NewNumber(3)), 1)
}

func TestModExact(t *testing.T) {
	assertNumber(t, NewNumber(10).Mod(NewNumber(5)), 0)
}

func TestModByZero(t *testing.T) {
	v := NewNumber(5).Mod(NewNumber(0))
	assertValue(t, v, TypeNil, "mod by zero")
}

func TestPowNumbers(t *testing.T) {
	assertNumber(t, NewNumber(2).Pow(NewNumber(3)), 8)
}

func TestPowZeroExponent(t *testing.T) {
	assertNumber(t, NewNumber(5).Pow(NewNumber(0)), 1)
}

func TestPowNegativeExponent(t *testing.T) {
	v := NewNumber(2).Pow(NewNumber(-1))
	assertNumber(t, v, 0.5)
}

func TestConcatStrings(t *testing.T) {
	assertString(t, NewString("abc").Concat(NewString("def")), "abcdef")
}

func TestConcatWithSpaces(t *testing.T) {
	assertString(t, NewString("abc   ").Concat(NewString("def")), "abcdef")
}

// ---------- Comparison ----------

func TestEqNumbers(t *testing.T) {
	assertLogical(t, NewNumber(5).Eq(NewNumber(5)), true)
	assertLogical(t, NewNumber(5).Eq(NewNumber(6)), false)
}

func TestEqStrings(t *testing.T) {
	assertLogical(t, NewString("hello").Eq(NewString("hello")), true)
	assertLogical(t, NewString("hello").Eq(NewString("world")), false)
}

func TestEqCaseSensitive(t *testing.T) {
	assertLogical(t, NewString("Hello").Eq(NewString("hello")), false)
}

func TestEqLogical(t *testing.T) {
	assertLogical(t, NewLogical(true).Eq(NewLogical(true)), true)
	assertLogical(t, NewLogical(true).Eq(NewLogical(false)), false)
}

func TestEqDates(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assertLogical(t, NewDate(d).Eq(NewDate(d)), true)
	d2 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	assertLogical(t, NewDate(d).Eq(NewDate(d2)), false)
}

func TestEqNil(t *testing.T) {
	assertLogical(t, Nil().Eq(Nil()), true)
}

func TestEqNilVsNumber(t *testing.T) {
	assertLogical(t, Nil().Eq(NewNumber(0)), false)
}

func TestEqTypeMismatch(t *testing.T) {
	assertLogical(t, NewNumber(5).Eq(NewString("5")), false)
}

func TestNeq(t *testing.T) {
	assertLogical(t, NewNumber(5).Neq(NewNumber(6)), true)
	assertLogical(t, NewNumber(5).Neq(NewNumber(5)), false)
}

func TestLtNumbers(t *testing.T) {
	assertLogical(t, NewNumber(3).Lt(NewNumber(5)), true)
	assertLogical(t, NewNumber(5).Lt(NewNumber(3)), false)
	assertLogical(t, NewNumber(5).Lt(NewNumber(5)), false)
}

func TestLtStrings(t *testing.T) {
	assertLogical(t, NewString("abc").Lt(NewString("def")), true)
	assertLogical(t, NewString("def").Lt(NewString("abc")), false)
}

func TestLtDates(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	assertLogical(t, NewDate(d1).Lt(NewDate(d2)), true)
	assertLogical(t, NewDate(d2).Lt(NewDate(d1)), false)
}

func TestLeNumbers(t *testing.T) {
	assertLogical(t, NewNumber(3).Le(NewNumber(5)), true)
	assertLogical(t, NewNumber(5).Le(NewNumber(5)), true)
	assertLogical(t, NewNumber(6).Le(NewNumber(5)), false)
}

func TestGtNumbers(t *testing.T) {
	assertLogical(t, NewNumber(5).Gt(NewNumber(3)), true)
	assertLogical(t, NewNumber(3).Gt(NewNumber(5)), false)
	assertLogical(t, NewNumber(5).Gt(NewNumber(5)), false)
}

func TestGeNumbers(t *testing.T) {
	assertLogical(t, NewNumber(5).Ge(NewNumber(3)), true)
	assertLogical(t, NewNumber(5).Ge(NewNumber(5)), true)
	assertLogical(t, NewNumber(3).Ge(NewNumber(5)), false)
}

func TestContains(t *testing.T) {
	assertLogical(t, NewString("hello world").Contains(NewString("world")), true)
	assertLogical(t, NewString("hello").Contains(NewString("xyz")), false)
	assertLogical(t, NewString("").Contains(NewString("")), true)
	assertLogical(t, NewString("hello").Contains(NewString("")), true)
}

func TestContainsTypeMismatch(t *testing.T) {
	assertLogical(t, NewString("hello").Contains(NewNumber(5)), false)
}

// ---------- Logical operators ----------

func TestAndTrueTrue(t *testing.T) {
	assertLogical(t, NewLogical(true).And(NewLogical(true)), true)
}

func TestAndTrueFalse(t *testing.T) {
	assertLogical(t, NewLogical(true).And(NewLogical(false)), false)
}

func TestAndFalseTrue(t *testing.T) {
	assertLogical(t, NewLogical(false).And(NewLogical(true)), false)
}

func TestAndFalseFalse(t *testing.T) {
	assertLogical(t, NewLogical(false).And(NewLogical(false)), false)
}

func TestAndMixedTypes(t *testing.T) {
	assertLogical(t, NewNumber(1).And(NewLogical(true)), true)
	assertLogical(t, NewNumber(0).And(NewLogical(true)), false)
}

func TestOrTrueFalse(t *testing.T) {
	assertLogical(t, NewLogical(true).Or(NewLogical(false)), true)
}

func TestOrFalseTrue(t *testing.T) {
	assertLogical(t, NewLogical(false).Or(NewLogical(true)), true)
}

func TestOrFalseFalse(t *testing.T) {
	assertLogical(t, NewLogical(false).Or(NewLogical(false)), false)
}

func TestNotTrue(t *testing.T) {
	assertLogical(t, NewLogical(true).Not(), false)
}

func TestNotFalse(t *testing.T) {
	assertLogical(t, NewLogical(false).Not(), true)
}

func TestNotNumber(t *testing.T) {
	assertLogical(t, NewNumber(42).Not(), false)
	assertLogical(t, NewNumber(0).Not(), true)
}

func TestNotNil(t *testing.T) {
	assertLogical(t, Nil().Not(), true)
}

// ---------- String representation ----------

func TestValueString(t *testing.T) {
	if Nil().String() != "NIL" {
		t.Errorf("expected NIL, got %s", Nil().String())
	}
	if NewNumber(42.5).String() != "42.5" {
		t.Errorf("expected 42.5, got %s", NewNumber(42.5).String())
	}
	if NewString("hello").String() != "hello" {
		t.Errorf("expected hello, got %s", NewString("hello").String())
	}
	if NewLogical(true).String() != ".T." {
		t.Errorf("expected .T., got %s", NewLogical(true).String())
	}
	if NewLogical(false).String() != ".F." {
		t.Errorf("expected .F., got %s", NewLogical(false).String())
	}
}

func TestArrayString(t *testing.T) {
	v := NewArray([]Value{NewNumber(1), NewString("two")})
	s := v.String()
	if s != "[1, two]" {
		t.Errorf("expected [1, two], got %s", s)
	}
}

func TestEmptyArrayString(t *testing.T) {
	v := NewArray(nil)
	if v.String() != "[]" {
		t.Errorf("expected [], got %s", v.String())
	}
}

// ---------- Edge cases ----------

func TestVeryLargeNumber(t *testing.T) {
	v := NewNumber(1e308)
	assertNumber(t, v, 1e308)
}

func TestInfinity(t *testing.T) {
	v := NewNumber(math.Inf(1))
	if !math.IsInf(v.Number, 1) {
		t.Error("expected +Inf")
	}
}

func TestArrayOfArrays(t *testing.T) {
	inner := NewArray([]Value{NewNumber(1), NewNumber(2)})
	outer := NewArray([]Value{inner, NewString("end")})
	if len(outer.Array) != 2 {
		t.Errorf("expected 2 elements, got %d", len(outer.Array))
	}
}

func TestObjectSetAndGet(t *testing.T) {
	v := NewObject()
	v.Fields["name"] = NewString("test")
	v.Fields["count"] = NewNumber(42)
	if v.Fields["name"].Str != "test" {
		t.Errorf("expected name=test, got %v", v.Fields["name"])
	}
	if v.Fields["count"].Number != 42 {
		t.Errorf("expected count=42, got %v", v.Fields["count"])
	}
}
