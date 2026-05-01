package form

import (
	"testing"
)

// ---------- ParseMask ----------

func TestParseMaskEmpty(t *testing.T) {
	m := ParseMask("")
	if m.Width != 0 {
		t.Errorf("expected Width=0, got %d", m.Width)
	}
	if len(m.Patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(m.Patterns))
	}
}

func TestParseMaskAllAny(t *testing.T) {
	m := ParseMask("XXXXXXXXXX")
	if m.Width != 10 {
		t.Errorf("expected Width=10, got %d", m.Width)
	}
	if len(m.Patterns) != 10 {
		t.Errorf("expected 10 patterns, got %d", len(m.Patterns))
	}
	for i, mc := range m.Patterns {
		if mc != MaskAny {
			t.Errorf("pattern[%d]: expected MaskAny, got %v", i, mc)
		}
	}
}

func TestParseMaskPhone(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	if m.Width != 13 {
		t.Errorf("expected Width=13 for phone mask '(XXX)XXX-XXXX', got %d", m.Width)
	}
	if len(m.Patterns) != 13 {
		t.Errorf("expected 13 patterns, got %d", len(m.Patterns))
	}
}

func TestParseMaskSSN(t *testing.T) {
	m := ParseMask("999-99-9999")
	if m.Width != 11 {
		t.Errorf("expected Width=11 for SSN mask, got %d", m.Width)
	}
	// Check that digits have proper positions
	if m.Patterns[0] != MaskDigit {
		t.Error("expected MaskDigit at position 0")
	}
	if m.Patterns[3] != MaskLiteral {
		t.Errorf("expected MaskLiteral at position 3 (-), got %v", m.Patterns[3])
	}
}

func TestParseMaskAllTypes(t *testing.T) {
	m := ParseMask("X9A!")
	if m.Width != 4 {
		t.Errorf("expected Width=4, got %d", m.Width)
	}
	if m.Patterns[0] != MaskAny {
		t.Errorf("pattern[0]: expected MaskAny, got %v", m.Patterns[0])
	}
	if m.Patterns[1] != MaskDigit {
		t.Errorf("pattern[1]: expected MaskDigit, got %v", m.Patterns[1])
	}
	if m.Patterns[2] != MaskAlpha {
		t.Errorf("pattern[2]: expected MaskAlpha, got %v", m.Patterns[2])
	}
	if m.Patterns[3] != MaskUppercase {
		t.Errorf("pattern[3]: expected MaskUppercase, got %v", m.Patterns[3])
	}
}

func TestParseMaskLowerCasePatterns(t *testing.T) {
	m := ParseMask("xa9!")
	// Lowercase x and a should be treated as valid patterns
	if m.Patterns[0] != MaskAny {
		t.Errorf("pattern[0]: expected MaskAny for 'x', got %v", m.Patterns[0])
	}
	if m.Patterns[1] != MaskAlpha {
		t.Errorf("pattern[1]: expected MaskAlpha for 'a', got %v", m.Patterns[1])
	}
	if m.Patterns[2] != MaskDigit {
		t.Errorf("pattern[2]: expected MaskDigit for 9, got %v", m.Patterns[2])
	}
}

func TestParseMaskOnlyLiterals(t *testing.T) {
	m := ParseMask("()- ")
	if m.Width != 4 {
		t.Errorf("expected Width=4, got %d", m.Width)
	}
	for i, mc := range m.Patterns {
		if mc != MaskLiteral {
			t.Errorf("pattern[%d]: expected MaskLiteral, got %v", i, mc)
		}
	}
}

func TestParseMaskZipCode(t *testing.T) {
	m := ParseMask("99999-9999")
	if m.Width != 10 {
		t.Errorf("expected Width=10, got %d", m.Width)
	}
}

func TestParseMaskDate(t *testing.T) {
	m := ParseMask("99/99/9999")
	if m.Width != 10 {
		t.Errorf("expected Width=10, got %d", m.Width)
	}
}

func TestParseMaskComplex(t *testing.T) {
	m := ParseMask("!XXXXXXXXX!")
	if m.Width != 11 {
		t.Errorf("expected Width=11, got %d", m.Width)
	}
}

// ---------- Width is the key metric ----------

func TestWidthFromMask(t *testing.T) {
	tests := []struct {
		mask  string
		width int
	}{
		{"", 0},
		{"X", 1},
		{"XX", 2},
		{"999", 3},
		{"AAA", 3},
		{"!!!", 3},
		{"(XXX)XXX-XXXX", 13},
		{"999-99-9999", 11},
		{"99/99/9999", 10},
		{"99999-9999", 10},
		{"X!A9", 4},
		{"@#$%^&*", 7},
		{"X X X", 5},
		{"   ", 3},
	}

	for _, tc := range tests {
		t.Run(tc.mask, func(t *testing.T) {
			m := ParseMask(tc.mask)
			if m.Width != tc.width {
				t.Errorf("expected width %d, got %d", tc.width, m.Width)
			}
		})
	}
}

// ---------- Validate ----------

func TestValidateExactMatch(t *testing.T) {
	m := ParseMask("999-99-9999")
	if !m.Validate("123-45-6789") {
		t.Error("expected valid SSN")
	}
}

func TestValidateWrongLength(t *testing.T) {
	m := ParseMask("999-99-9999")
	if m.Validate("123-45-678") {
		t.Error("expected invalid (too short)")
	}
	if m.Validate("123-45-67890") {
		t.Error("expected invalid (too long)")
	}
}

func TestValidateDigitPosition(t *testing.T) {
	m := ParseMask("999-99-9999")
	if m.Validate("abc-de-fghi") {
		t.Error("expected invalid (letters where digits expected)")
	}
}

func TestValidatePhone(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	if !m.Validate("(800)555-0100") {
		t.Error("expected valid phone")
	}
	if m.Validate("(800)555-01") {
		t.Error("expected invalid (too short)")
	}
}

func TestValidateAlpha(t *testing.T) {
	m := ParseMask("AAA")
	if !m.Validate("abc") {
		t.Error("expected valid alpha")
	}
	if m.Validate("123") {
		t.Error("expected invalid (digits where alpha expected)")
	}
}

func TestValidateAny(t *testing.T) {
	m := ParseMask("XXX")
	if !m.Validate("a1!") {
		t.Error("expected valid any chars")
	}
	if !m.Validate("999") {
		t.Error("expected valid digits where X")
	}
}

func TestValidateLiteralMismatch(t *testing.T) {
	m := ParseMask("999-99-9999")
	if m.Validate("999/99/9999") {
		t.Error("expected invalid (literal - vs /)")
	}
}

func TestValidateEmptyMask(t *testing.T) {
	m := ParseMask("")
	if m.Validate("anything") {
		t.Error("expected invalid for empty mask")
	}
	if !m.Validate("") {
		t.Error("expected valid for empty mask with empty input (zero-length match)")
	}
}

func TestValidateUnicode(t *testing.T) {
	m := ParseMask("XXX")
	if !m.Validate("世A界") {
		t.Error("expected valid unicode characters with X mask")
	}
}

// ---------- InputMask ----------

func TestInputMaskPhone(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	im := m.InputMask()
	expected := "(___)___-____"
	if im != expected {
		t.Errorf("expected %q, got %q", expected, im)
	}
}

func TestInputMaskSSN(t *testing.T) {
	m := ParseMask("999-99-9999")
	im := m.InputMask()
	expected := "___-__-____"
	if im != expected {
		t.Errorf("expected %q, got %q", expected, im)
	}
}

func TestInputMaskAllLiteral(t *testing.T) {
	m := ParseMask("()-")
	im := m.InputMask()
	expected := "()-"
	if im != expected {
		t.Errorf("expected %q, got %q", expected, im)
	}
}

func TestInputMaskEmpty(t *testing.T) {
	m := ParseMask("")
	if m.InputMask() != "" {
		t.Errorf("expected empty string for empty mask")
	}
}

// ---------- Format ----------

func TestFormatPhone(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	result := m.Format("8005550100")
	expected := "(800)555-0100"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatSSN(t *testing.T) {
	m := ParseMask("999-99-9999")
	result := m.Format("123456789")
	expected := "123-45-6789"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatShortInput(t *testing.T) {
	m := ParseMask("999-99-9999")
	result := m.Format("123")
	expected := "123-  -    "
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEmptyInput(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	result := m.Format("")
	expected := "(   )   -    " // 13 chars: literals + spaces for X
	if result != expected {
		t.Errorf("expected %q (len=%d), got %q (len=%d)", expected, len(expected), result, len(result))
	}
}

func TestFormatUppercase(t *testing.T) {
	m := ParseMask("!!!")
	result := m.Format("abc")
	expected := "ABC"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatMixed(t *testing.T) {
	m := ParseMask("!X9")
	result := m.Format("ab1")
	if result[0] != 'A' {
		t.Errorf("expected uppercase A, got %c", result[0])
	}
}

func TestFormatExactLength(t *testing.T) {
	m := ParseMask("XXX")
	result := m.Format("abc")
	if result != "abc" {
		t.Errorf("expected abc, got %q", result)
	}
}

func TestFormatExcessInput(t *testing.T) {
	m := ParseMask("XXX")
	result := m.Format("abcdef")
	if result != "abc" {
		t.Errorf("expected abc (truncated), got %q", result)
	}
}

// ---------- Edge cases ----------

func TestUnicodeMaskWidth(t *testing.T) {
	m := ParseMask("X界X")
	if m.Width != 3 {
		t.Errorf("expected Width=3 for unicode mask, got %d", m.Width)
	}
	if m.Patterns[1] != MaskLiteral {
		t.Errorf("expected MaskLiteral at position 1 for '界', got %v", m.Patterns[1])
	}
}

func TestValidateSSNWithCorrectFormat(t *testing.T) {
	m := ParseMask("999-99-9999")
	validSSNs := []string{
		"123-45-6789",
		"000-00-0000",
		"999-99-9999",
	}
	for _, ssn := range validSSNs {
		if !m.Validate(ssn) {
			t.Errorf("expected valid SSN: %q", ssn)
		}
	}
}

func TestValidateSSNWithIncorrectFormat(t *testing.T) {
	m := ParseMask("999-99-9999")
	invalidSSNs := []string{
		"123-45-678",
		"123-45-67890",
		"123.45.6789",
		"abc-de-fghi",
		"123-45-678a",
		"",
	}
	for _, ssn := range invalidSSNs {
		if m.Validate(ssn) {
			t.Errorf("expected invalid SSN: %q", ssn)
		}
	}
}

func TestValidatePhoneNumbers(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	valid := []string{
		"(800)555-0100",
		"(aaa)bbb-cccc",
		"(123)456-7890",
	}
	invalid := []string{
		"800-555-0100",
		"(800)555-010",
		"(800)555-01000",
		" (800)555-0100",
		"(800) 555-0100",
		"",
	}
	for _, p := range valid {
		if !m.Validate(p) {
			t.Errorf("expected valid phone: %q", p)
		}
	}
	for _, p := range invalid {
		if m.Validate(p) {
			t.Errorf("expected invalid phone: %q", p)
		}
	}
}

func TestFormatAndValidateRoundTrip(t *testing.T) {
	m := ParseMask("(XXX)XXX-XXXX")
	raw := "8005550100"
	formatted := m.Format(raw)
	if !m.Validate(formatted) {
		t.Errorf("round-trip failed: format(%q) = %q, but validate returned false", raw, formatted)
	}
}

func TestMultipleMasks(t *testing.T) {
	masks := []string{
		"999-99-9999",
		"(XXX)XXX-XXXX",
		"99/99/9999",
		"AAAAA",
		"!!!!!",
		"99999",
	}
	for _, mask := range masks {
		t.Run(mask, func(t *testing.T) {
			m := ParseMask(mask)
			if m.Width <= 0 {
				t.Errorf("expected positive width, got %d", m.Width)
			}
			if len(m.Patterns) != len([]rune(mask)) {
				t.Errorf("expected %d patterns, got %d", len([]rune(mask)), len(m.Patterns))
			}
		})
	}
}
