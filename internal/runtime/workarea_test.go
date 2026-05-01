package runtime

import (
	"testing"
)

// ---------- WorkArea ----------

func TestNewWorkArea(t *testing.T) {
	wa := NewWorkArea("cust")
	if wa.Alias != "cust" {
		t.Errorf("expected Alias=cust, got %q", wa.Alias)
	}
	if wa.RecNo != 1 {
		t.Errorf("expected RecNo=1, got %d", wa.RecNo)
	}
	if !wa.BOF {
		t.Error("expected BOF=true for new work area")
	}
	if !wa.EOF {
		t.Error("expected EOF=true for new work area")
	}
	if wa.LastRec != 0 {
		t.Errorf("expected LastRec=0, got %d", wa.LastRec)
	}
	if wa.Fields == nil {
		t.Error("expected non-nil Fields map")
	}
}

func TestWorkAreaOpen(t *testing.T) {
	wa := NewWorkArea("test")
	if wa.Alias != "test" || wa.TableName != "" {
		t.Errorf("expected empty TableName, got %q", wa.TableName)
	}
}

func TestGoTop(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.RecNo = 5
	wa.GoTop()
	if wa.RecNo != 1 {
		t.Errorf("expected RecNo=1, got %d", wa.RecNo)
	}
	if !wa.BOF {
		t.Error("expected BOF=true after GoTop")
	}
}

func TestGoBottom(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.GoBottom()
	if wa.RecNo != 10 {
		t.Errorf("expected RecNo=10, got %d", wa.RecNo)
	}
}

func TestGoBottomEmpty(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 0
	wa.GoBottom()
	if wa.RecNo != 1 {
		t.Errorf("expected RecNo=1 for empty table, got %d", wa.RecNo)
	}
}

func TestGoTo(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.GoTo(5)
	if wa.RecNo != 5 {
		t.Errorf("expected RecNo=5, got %d", wa.RecNo)
	}
}

func TestGoToBelowOne(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.GoTo(0)
	if wa.RecNo != 1 {
		t.Errorf("expected RecNo=1, got %d", wa.RecNo)
	}
}

func TestGoToBeyondEnd(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.GoTo(100)
	if wa.RecNo != 10 {
		t.Errorf("expected RecNo=10, got %d", wa.RecNo)
	}
}

func TestSkipForward(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.RecNo = 3
	wa.Skip(2)
	if wa.RecNo != 5 {
		t.Errorf("expected RecNo=5, got %d", wa.RecNo)
	}
}

func TestSkipBackward(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.RecNo = 5
	wa.Skip(-2)
	if wa.RecNo != 3 {
		t.Errorf("expected RecNo=3, got %d", wa.RecNo)
	}
}

func TestSkipPastEOF(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.RecNo = 9
	wa.Skip(5)
	if wa.RecNo != 11 {
		t.Errorf("expected RecNo=11 (past EOF), got %d", wa.RecNo)
	}
	if !wa.EOF {
		t.Error("expected EOF=true after skipping past end")
	}
}

func TestSkipBeforeBOF(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.RecNo = 3
	wa.Skip(-5)
	if wa.RecNo != 1 {
		t.Errorf("expected RecNo=1, got %d", wa.RecNo)
	}
}

func TestSkipZero(t *testing.T) {
	wa := NewWorkArea("test")
	wa.LastRec = 10
	wa.RecNo = 5
	wa.Skip(0)
	if wa.RecNo != 5 {
		t.Errorf("expected RecNo=5, got %d", wa.RecNo)
	}
}

func TestAppendBlank(t *testing.T) {
	wa := NewWorkArea("test")
	wa.AppendBlank()
	if wa.LastRec != 1 {
		t.Errorf("expected LastRec=1, got %d", wa.LastRec)
	}
	if wa.RecNo != 1 {
		t.Errorf("expected RecNo=1, got %d", wa.RecNo)
	}
	if wa.EOF {
		t.Error("expected EOF=false after append")
	}
	if wa.BOF {
		t.Error("expected BOF=false after append")
	}
}

func TestAppendBlanks(t *testing.T) {
	wa := NewWorkArea("test")
	for i := 0; i < 5; i++ {
		wa.AppendBlank()
	}
	if wa.LastRec != 5 {
		t.Errorf("expected LastRec=5, got %d", wa.LastRec)
	}
	if wa.RecNo != 5 {
		t.Errorf("expected RecNo=5, got %d", wa.RecNo)
	}
}

func TestFieldGetSet(t *testing.T) {
	wa := NewWorkArea("test")
	wa.SetField("name", NewString("John"))
	wa.SetField("age", NewNumber(30))
	wa.SetField("active", NewLogical(true))

	if wa.GetField("name").Str != "John" {
		t.Errorf("expected name=John, got %v", wa.GetField("name"))
	}
	if wa.GetField("age").Number != 30 {
		t.Errorf("expected age=30, got %v", wa.GetField("age"))
	}
	if !wa.GetField("active").Logical {
		t.Error("expected active=true")
	}
}

func TestFieldGetUnknown(t *testing.T) {
	wa := NewWorkArea("test")
	v := wa.GetField("nonexistent")
	if v.Type != TypeNil {
		t.Errorf("expected Nil for unknown field, got %v", v)
	}
}

func TestDelete(t *testing.T) {
	wa := NewWorkArea("test")
	if wa.Deleted {
		t.Error("expected Deleted=false initially")
	}
	wa.Delete()
	if !wa.Deleted {
		t.Error("expected Deleted=true after Delete()")
	}
}

func TestRecall(t *testing.T) {
	wa := NewWorkArea("test")
	wa.Delete()
	wa.Recall()
	if wa.Deleted {
		t.Error("expected Deleted=false after Recall()")
	}
}

// ---------- WorkAreaManager ----------

func TestNewManager(t *testing.T) {
	m := NewWorkAreaManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	if m.CurrentNum() != 1 {
		t.Errorf("expected current=1, got %d", m.CurrentNum())
	}
}

func TestSelectNumeric(t *testing.T) {
	m := NewWorkAreaManager()
	err := m.Select(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.CurrentNum() != 1 {
		t.Errorf("expected current=1, got %d", m.CurrentNum())
	}
}

func TestSelectZero(t *testing.T) {
	m := NewWorkAreaManager()
	err := m.Select(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.CurrentNum() < 1 {
		t.Errorf("expected current >= 1, got %d", m.CurrentNum())
	}
}

func TestSelectInvalid(t *testing.T) {
	m := NewWorkAreaManager()
	err := m.Select(-1)
	if err == nil {
		t.Error("expected error for negative work area")
	}
}

func TestSelectOutOfRange(t *testing.T) {
	m := NewWorkAreaManager()
	err := m.Select(999)
	if err == nil {
		t.Error("expected error for out-of-range work area")
	}
}

func TestSelectAndUse(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	wa := m.Current()
	if wa.Alias != "cust" {
		t.Errorf("expected Alias=cust, got %q", wa.Alias)
	}
	if wa.TableName != "customers" {
		t.Errorf("expected TableName=customers, got %q", wa.TableName)
	}
}

func TestSelectAlias(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	m.Select(2)
	m.Use("ord", "orders")

	if !m.SelectAlias("cust") {
		t.Fatal("expected SelectAlias to succeed")
	}
	if m.CurrentNum() != 1 {
		t.Errorf("expected current=1, got %d", m.CurrentNum())
	}

	if !m.SelectAlias("ORD") {
		t.Fatal("expected SelectAlias to be case-insensitive")
	}
	if m.CurrentNum() != 2 {
		t.Errorf("expected current=2, got %d", m.CurrentNum())
	}
}

func TestSelectAliasNotFound(t *testing.T) {
	m := NewWorkAreaManager()
	if m.SelectAlias("nonexistent") {
		t.Error("expected false for nonexistent alias")
	}
}

func TestGetByAlias(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")

	wa := m.Get("cust")
	if wa == nil {
		t.Fatal("expected non-nil WorkArea")
	}
	if wa.TableName != "customers" {
		t.Errorf("expected TableName=customers, got %q", wa.TableName)
	}
}

func TestGetByAliasNotFound(t *testing.T) {
	m := NewWorkAreaManager()
	if wa := m.Get("ghost"); wa != nil {
		t.Error("expected nil for unknown alias")
	}
}

func TestGetByAliasCaseInsensitive(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")

	wa := m.Get("CUST")
	if wa == nil {
		t.Fatal("expected case-insensitive lookup to succeed")
	}
}

func TestUsed(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	m.Select(2)
	m.Use("ord", "orders")

	used := m.Used()
	if len(used) != 2 {
		t.Errorf("expected 2 used work areas, got %d", len(used))
	}
}

func TestUsedNone(t *testing.T) {
	m := NewWorkAreaManager()
	used := m.Used()
	if len(used) != 0 {
		t.Errorf("expected 0 used, got %d", len(used))
	}
}

func TestCloseAll(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	m.Select(2)
	m.Use("ord", "orders")

	m.CloseAll()
	used := m.Used()
	if len(used) != 0 {
		t.Errorf("expected 0 used after CloseAll, got %d", len(used))
	}
	if m.CurrentNum() != 1 {
		t.Errorf("expected current reset to 1, got %d", m.CurrentNum())
	}
}

func TestCloseByAlias(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	m.Select(2)
	m.Use("ord", "orders")

	m.Close("cust")
	used := m.Used()
	if len(used) != 1 {
		t.Errorf("expected 1 used after closing one, got %d", len(used))
	}
	if used[0].Alias != "ord" {
		t.Errorf("expected remaining area to be 'ord', got %q", used[0].Alias)
	}
}

func TestCloseCurrentResets(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	m.Close("cust")
	if m.CurrentNum() != 1 {
		t.Errorf("expected current=1 after closing current area, got %d", m.CurrentNum())
	}
}

func TestMultipleAreasIndependent(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(1)
	m.Use("cust", "customers")
	m.Current().SetField("name", NewString("John"))

	m.Select(2)
	m.Use("ord", "orders")
	m.Current().SetField("item", NewString("Widget"))

	// Verify cust still has its data
	m.SelectAlias("cust")
	if m.Current().GetField("name").Str != "John" {
		t.Errorf("expected name=John in cust area")
	}

	// Verify ord has its data
	m.SelectAlias("ord")
	if m.Current().GetField("item").Str != "Widget" {
		t.Errorf("expected item=Widget in ord area")
	}
}

func TestSelectWithExistingAlias(t *testing.T) {
	m := NewWorkAreaManager()
	m.Select(5)
	m.Use("cust", "customers")

	m.Select(5)
	wa := m.Current()
	if wa.Alias != "cust" {
		t.Errorf("expected Alias=cust, got %q", wa.Alias)
	}
}
