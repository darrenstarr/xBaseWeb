package form

// FieldDefinition describes a single field on a form.
type FieldDefinition struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Mask      string `json:"mask"`
	Width     int    `json:"width"`
	Align     string `json:"align"`
	Visible   bool   `json:"visible"`
	ReadOnly  bool   `json:"readOnly"`
	Required  bool   `json:"required"`
	Default   string `json:"default"`
	TabOrder  int    `json:"tabOrder"`
	BindTo    string `json:"bindTo"`    // ALIAS->FIELD
	HelpText  string `json:"helpText"`
	InputType string `json:"inputType"` // text, number, date, password
}

// FormDefinition describes a complete form layout.
type FormDefinition struct {
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Width       int               `json:"width"`
	Fields      []FieldDefinition `json:"fields"`
	Buttons     []ButtonDef       `json:"buttons"`
	TableSource string            `json:"tableSource"`
	Events      map[string]string `json:"events"` // event name -> .prg function name
}

// ButtonDef describes a form button.
type ButtonDef struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Action   string `json:"action"`
	Primary  bool   `json:"primary"`
}

// NewField creates a FieldDefinition from a name and mask.
func NewField(name, label, mask string) FieldDefinition {
	fm := ParseMask(mask)
	return FieldDefinition{
		Name:     name,
		Label:    label,
		Mask:     mask,
		Width:    fm.Width,
		Visible:  true,
		TabOrder: -1,
	}
}

// NewForm creates a FormDefinition with sensible defaults.
func NewForm(name, title string) FormDefinition {
	return FormDefinition{
		Name:   name,
		Title:  title,
		Width:  600,
		Fields: nil,
		Events: make(map[string]string),
	}
}

// AddField appends a field to the form.
func (f *FormDefinition) AddField(fd FieldDefinition) {
	f.Fields = append(f.Fields, fd)
}

// AddButton appends a button to the form.
func (f *FormDefinition) AddButton(bd ButtonDef) {
	f.Buttons = append(f.Buttons, bd)
}

// SetEvent associates an event (e.g. "init", "validate") with a .prg handler.
func (f *FormDefinition) SetEvent(event, handler string) {
	f.Events[event] = handler
}

// TotalFieldWidth returns the sum of all field widths for layout calculations.
func (f *FormDefinition) TotalFieldWidth() int {
	total := 0
	for _, fd := range f.Fields {
		total += fd.Width
	}
	return total
}

// FieldCount returns the number of visible fields.
func (f *FormDefinition) FieldCount() int {
	count := 0
	for _, fd := range f.Fields {
		if fd.Visible {
			count++
		}
	}
	return count
}
