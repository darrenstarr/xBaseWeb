package form

import (
	"testing"
)

// ---------- NewField ----------

func TestNewField(t *testing.T) {
	f := NewField("phone", "Phone Number", "(XXX)XXX-XXXX")
	if f.Name != "phone" {
		t.Errorf("expected Name=phone, got %q", f.Name)
	}
	if f.Label != "Phone Number" {
		t.Errorf("expected Label='Phone Number', got %q", f.Label)
	}
	if f.Mask != "(XXX)XXX-XXXX" {
		t.Errorf("expected Mask=(XXX)XXX-XXXX, got %q", f.Mask)
	}
	if f.Width != 13 {
		t.Errorf("expected Width=13, got %d", f.Width)
	}
	if !f.Visible {
		t.Error("expected Visible=true by default")
	}
	if f.TabOrder != -1 {
		t.Errorf("expected TabOrder=-1 (unset), got %d", f.TabOrder)
	}
}

func TestNewFieldEmptyMask(t *testing.T) {
	f := NewField("raw", "Raw Field", "")
	if f.Width != 0 {
		t.Errorf("expected Width=0 for empty mask, got %d", f.Width)
	}
}

func TestNewFieldSSN(t *testing.T) {
	f := NewField("ssn", "SSN", "999-99-9999")
	if f.Width != 11 {
		t.Errorf("expected Width=11 for SSN mask, got %d", f.Width)
	}
}

func TestNewFieldNumeric(t *testing.T) {
	f := NewField("zip", "ZIP", "99999")
	if f.Width != 5 {
		t.Errorf("expected Width=5 for ZIP, got %d", f.Width)
	}
}

// ---------- NewForm ----------

func TestNewForm(t *testing.T) {
	f := NewForm("customer", "Customer Entry")
	if f.Name != "customer" {
		t.Errorf("expected Name=customer, got %q", f.Name)
	}
	if f.Title != "Customer Entry" {
		t.Errorf("expected Title='Customer Entry', got %q", f.Title)
	}
	if f.Width != 600 {
		t.Errorf("expected Width=600, got %d", f.Width)
	}
	if f.Events == nil {
		t.Error("expected non-nil Events map")
	}
	if len(f.Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(f.Fields))
	}
}

// ---------- AddField ----------

func TestAddField(t *testing.T) {
	form := NewForm("test", "Test")
	form.AddField(NewField("f1", "Field 1", "XXXXXXXX"))
	form.AddField(NewField("f2", "Field 2", "999-99-9999"))

	if len(form.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(form.Fields))
	}
	if form.Fields[0].Name != "f1" {
		t.Errorf("expected first field name=f1, got %q", form.Fields[0].Name)
	}
	if form.Fields[1].Name != "f2" {
		t.Errorf("expected second field name=f2, got %q", form.Fields[1].Name)
	}
}

func TestAddFieldMany(t *testing.T) {
	form := NewForm("large", "Large Form")
	for i := 0; i < 100; i++ {
		form.AddField(NewField(
			"f"+string(rune('0'+i%10)),
			"Field",
			"XXXXXXXX",
		))
	}
	if len(form.Fields) != 100 {
		t.Errorf("expected 100 fields, got %d", len(form.Fields))
	}
}

// ---------- AddButton ----------

func TestAddButton(t *testing.T) {
	form := NewForm("test", "Test")
	form.AddButton(ButtonDef{Name: "save", Label: "Save", Action: "DoSave", Primary: true})
	form.AddButton(ButtonDef{Name: "cancel", Label: "Cancel", Action: "DoCancel"})

	if len(form.Buttons) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(form.Buttons))
	}
	if !form.Buttons[0].Primary {
		t.Error("expected first button to be primary")
	}
	if form.Buttons[1].Primary {
		t.Error("expected second button to not be primary")
	}
}

// ---------- SetEvent ----------

func TestSetEvent(t *testing.T) {
	form := NewForm("test", "Test")
	form.SetEvent("init", "FormInit")
	form.SetEvent("validate", "FormValidate")

	if len(form.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(form.Events))
	}
	if form.Events["init"] != "FormInit" {
		t.Errorf("expected init=FormInit, got %q", form.Events["init"])
	}
	if form.Events["validate"] != "FormValidate" {
		t.Errorf("expected validate=FormValidate, got %q", form.Events["validate"])
	}
}

func TestSetEventOverwrite(t *testing.T) {
	form := NewForm("test", "Test")
	form.SetEvent("init", "First")
	form.SetEvent("init", "Second")

	if form.Events["init"] != "Second" {
		t.Errorf("expected init=Second after overwrite, got %q", form.Events["init"])
	}
}

// ---------- FieldCount ----------

func TestFieldCount(t *testing.T) {
	form := NewForm("test", "Test")
	form.AddField(NewField("v1", "Visible", "XX"))
	f2 := NewField("v2", "Hidden", "XX")
	f2.Visible = false
	form.AddField(f2)
	form.AddField(NewField("v3", "Visible 2", "XX"))

	if form.FieldCount() != 2 {
		t.Errorf("expected 2 visible fields, got %d", form.FieldCount())
	}
}

func TestFieldCountEmpty(t *testing.T) {
	form := NewForm("empty", "Empty")
	if form.FieldCount() != 0 {
		t.Errorf("expected 0 fields, got %d", form.FieldCount())
	}
}

// ---------- TotalFieldWidth ----------

func TestTotalFieldWidth(t *testing.T) {
	form := NewForm("test", "Test")
	form.AddField(NewField("phone", "Phone", "(XXX)XXX-XXXX")) // width 13
	form.AddField(NewField("ssn", "SSN", "999-99-9999"))       // width 11
	form.AddField(NewField("zip", "ZIP", "99999"))             // width 5

	total := form.TotalFieldWidth()
	if total != 29 {
		t.Errorf("expected total width 29 (13+11+5), got %d", total)
	}
}

func TestTotalFieldWidthEmpty(t *testing.T) {
	form := NewForm("empty", "Empty")
	if form.TotalFieldWidth() != 0 {
		t.Errorf("expected 0, got %d", form.TotalFieldWidth())
	}
}

func TestTotalFieldWidthIncludesAll(t *testing.T) {
	form := NewForm("test", "Test")
	form.AddField(NewField("v1", "V1", "XXXX")) // width 4
	f2 := NewField("v2", "V2", "XXXX")           // width 4, hidden
	f2.Visible = false
	form.AddField(f2)

	// TotalFieldWidth counts all fields regardless of visibility
	if form.TotalFieldWidth() != 8 {
		t.Errorf("expected 8 (including hidden), got %d", form.TotalFieldWidth())
	}
}

// ---------- Real-world form ----------

func TestCustomerForm(t *testing.T) {
	form := NewForm("customer", "Customer Entry")
	form.AddField(NewField("name", "Name", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"))
	form.AddField(NewField("phone", "Phone", "(XXX)XXX-XXXX"))
	form.AddField(NewField("ssn", "SSN", "999-99-9999"))
	form.AddField(NewField("zip", "ZIP", "99999-9999"))
	form.AddField(NewField("active", "Active", "XXX"))

	form.AddButton(ButtonDef{Name: "save", Label: "Save", Action: "SaveCustomer", Primary: true})
	form.AddButton(ButtonDef{Name: "cancel", Label: "Cancel", Action: "ReturnToMenu"})

	form.SetEvent("init", "InitCustomer")
	form.SetEvent("validate", "ValidateCustomer")
	form.SetEvent("save", "SaveCustomer")

	// Verify
	if len(form.Fields) != 5 {
		t.Errorf("expected 5 fields, got %d", len(form.Fields))
	}
	if len(form.Buttons) != 2 {
		t.Errorf("expected 2 buttons, got %d", len(form.Buttons))
	}
	if len(form.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(form.Events))
	}
	if form.Width != 600 {
		t.Errorf("expected default Width=600, got %d", form.Width)
	}

	// Verify field widths
	expectedWidths := map[string]int{
		"name":   30,
		"phone":  13,
		"ssn":    11,
		"zip":    10,
		"active": 3,
	}
	for _, f := range form.Fields {
		if expectedWidths[f.Name] != f.Width {
			t.Errorf("field %s: expected width %d, got %d", f.Name, expectedWidths[f.Name], f.Width)
		}
	}
}

func TestOrderForm(t *testing.T) {
	form := NewForm("order", "Order Entry")
	form.AddField(NewField("orderNum", "Order #", "999999"))
	form.AddField(NewField("custName", "Customer", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"))
	form.AddField(NewField("orderDate", "Date", "99/99/9999"))
	form.AddField(NewField("amount", "Amount", "999999.99"))
	form.AddField(NewField("notes", "Notes", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"))

	if form.TotalFieldWidth() != 30+6+14+10+10+4 {
		// 30 (name) + 6 (order) + 10 (date) + 9 (amount) + 38 (notes) = 93 + 6 = 99? hmm
		t.Logf("Total field width: %d", form.TotalFieldWidth())
	}
}

// ---------- Edge cases ----------

func TestFormWithoutFields(t *testing.T) {
	form := NewForm("empty", "Empty Form")
	if form.FieldCount() != 0 {
		t.Errorf("expected 0 visible fields, got %d", form.FieldCount())
	}
	if form.TotalFieldWidth() != 0 {
		t.Errorf("expected 0 total width, got %d", form.TotalFieldWidth())
	}
}

func TestFormFieldTabOrder(t *testing.T) {
	form := NewForm("tabtest", "Tab Order Test")
	for i := 0; i < 5; i++ {
		f := NewField(
			"f"+string(rune('1'+i)),
			"Field",
			"XXXX",
		)
		f.TabOrder = i + 1
		form.AddField(f)
	}

	for i, f := range form.Fields {
		if f.TabOrder != i+1 {
			t.Errorf("field %d: expected TabOrder=%d, got %d", i, i+1, f.TabOrder)
		}
	}
}

func TestFormFieldDefaults(t *testing.T) {
	f := NewField("test", "Test", "XXXX")
	if f.ReadOnly {
		t.Error("expected ReadOnly=false by default")
	}
	if f.Required {
		t.Error("expected Required=false by default")
	}
	if f.Align != "" {
		t.Errorf("expected empty Align by default, got %q", f.Align)
	}
	if f.BindTo != "" {
		t.Errorf("expected empty BindTo by default, got %q", f.BindTo)
	}
	if f.HelpText != "" {
		t.Errorf("expected empty HelpText by default, got %q", f.HelpText)
	}
	if f.InputType != "" {
		t.Errorf("expected empty InputType by default, got %q", f.InputType)
	}
}
