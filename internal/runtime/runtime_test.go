package runtime

import (
	"testing"
)

func TestNewRuntime(t *testing.T) {
	rt := New()
	if rt == nil {
		t.Fatal("expected non-nil Runtime")
	}
	if rt.Variables == nil {
		t.Error("expected non-nil Variables map")
	}
	if rt.WorkAreas == nil {
		t.Error("expected non-nil WorkAreas")
	}
	if rt.CallStack == nil {
		t.Error("expected non-nil CallStack")
	}
}

func TestGetVarDefault(t *testing.T) {
	rt := New()
	v := rt.GetVar("undefined")
	if v.Type != TypeNil {
		t.Errorf("expected Nil for undefined var, got %v", v)
	}
}

func TestSetAndGetVar(t *testing.T) {
	rt := New()
	rt.SetVar("name", NewString("test"))
	rt.SetVar("count", NewNumber(42))
	rt.SetVar("active", NewLogical(true))

	if rt.GetVar("name").Str != "test" {
		t.Errorf("expected name=test, got %v", rt.GetVar("name"))
	}
	if rt.GetVar("count").Number != 42 {
		t.Errorf("expected count=42, got %v", rt.GetVar("count"))
	}
	if !rt.GetVar("active").Logical {
		t.Error("expected active=true")
	}
}

func TestSetVarOverwrites(t *testing.T) {
	rt := New()
	rt.SetVar("x", NewNumber(1))
	rt.SetVar("x", NewNumber(2))
	if rt.GetVar("x").Number != 2 {
		t.Errorf("expected x=2 after overwrite, got %v", rt.GetVar("x"))
	}
}

func TestSetVarChangeType(t *testing.T) {
	rt := New()
	rt.SetVar("x", NewNumber(42))
	rt.SetVar("x", NewString("hello"))
	if rt.GetVar("x").Str != "hello" {
		t.Errorf("expected x=hello, got %v", rt.GetVar("x"))
	}
}

func TestPushCall(t *testing.T) {
	rt := New()
	rt.PushCall("Main")
	if len(rt.CallStack) != 1 {
		t.Errorf("expected 1 call on stack, got %d", len(rt.CallStack))
	}
	if rt.CallStack[0] != "Main" {
		t.Errorf("expected Main on stack, got %q", rt.CallStack[0])
	}
}

func TestCallStackOrder(t *testing.T) {
	rt := New()
	rt.PushCall("Main")
	rt.PushCall("Process")
	rt.PushCall("Validate")

	expected := []string{"Main", "Process", "Validate"}
	for i, name := range expected {
		if rt.CallStack[i] != name {
			t.Errorf("position %d: expected %q, got %q", i, name, rt.CallStack[i])
		}
	}
}

func TestPopCall(t *testing.T) {
	rt := New()
	rt.PushCall("Main")
	rt.PushCall("Sub")
	rt.PopCall()
	if len(rt.CallStack) != 1 {
		t.Errorf("expected 1 call after pop, got %d", len(rt.CallStack))
	}
	if rt.CallStack[0] != "Main" {
		t.Errorf("expected Main remaining, got %q", rt.CallStack[0])
	}
}

func TestPopCallEmptyStack(t *testing.T) {
	rt := New()
	rt.PopCall() // should not panic
	if len(rt.CallStack) != 0 {
		t.Errorf("expected 0 after popping empty stack, got %d", len(rt.CallStack))
	}
}

func TestRuntimeIsolation(t *testing.T) {
	rt1 := New()
	rt2 := New()

	rt1.SetVar("x", NewNumber(100))
	rt2.SetVar("x", NewNumber(200))

	if rt1.GetVar("x").Number != 100 {
		t.Errorf("expected rt1.x=100, got %v", rt1.GetVar("x"))
	}
	if rt2.GetVar("x").Number != 200 {
		t.Errorf("expected rt2.x=200, got %v", rt2.GetVar("x"))
	}
}

func TestWorkAreaIsolation(t *testing.T) {
	rt1 := New()
	rt2 := New()

	rt1.WorkAreas.Select(1)
	rt1.WorkAreas.Use("cust1", "customers")
	rt1.WorkAreas.Current().SetField("name", NewString("Alice"))

	rt2.WorkAreas.Select(1)
	rt2.WorkAreas.Use("cust2", "customers")
	rt2.WorkAreas.Current().SetField("name", NewString("Bob"))

	if rt1.WorkAreas.Get("cust1").GetField("name").Str != "Alice" {
		t.Error("expected rt1 field name=Alice")
	}
	if rt2.WorkAreas.Get("cust2").GetField("name").Str != "Bob" {
		t.Error("expected rt2 field name=Bob")
	}
}

func TestMixedVarAndWorkAreas(t *testing.T) {
	rt := New()
	rt.SetVar("temp", NewString("variable value"))
	rt.WorkAreas.Select(1)
	rt.WorkAreas.Use("test", "test_table")
	rt.WorkAreas.Current().SetField("fld", NewString("field value"))

	if rt.GetVar("temp").Str != "variable value" {
		t.Errorf("expected var temp='variable value', got %v", rt.GetVar("temp"))
	}
	if rt.WorkAreas.Current().GetField("fld").Str != "field value" {
		t.Errorf("expected field fld='field value', got %v", rt.WorkAreas.Current().GetField("fld"))
	}
}
