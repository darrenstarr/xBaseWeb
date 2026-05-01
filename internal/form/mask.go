package form

// FieldMask represents a format mask string like "(XXX)XXX-XXXX"
// that determines the visual width and input constraints of a form field.
type FieldMask struct {
	Raw      string
	Width    int
	Patterns []MaskChar
}

type MaskChar int

const (
	MaskAny           MaskChar = iota // X - any character
	MaskDigit                         // 9 - digit only
	MaskAlpha                         // A - alpha only
	MaskUppercase                     // ! - uppercase alpha
	MaskLiteral                       // literal character (displayed as-is)
)

// ParseMask parses a format mask string and returns the computed FieldMask.
// Width is derived from the number of X/9/A/! characters plus any literal
// characters between them.
//
// Examples:
//
//	"999-99-9999"       → Width 11 (SSN)
//	"(XXX)XXX-XXXX"     → Width 14 (phone, literals included)
//	"!!!"               → Width 3
//	"@!"                → Width 2 (uppercase following any char)
//	""                  → Width 0
//	"XXXXXXXXXX"        → Width 10
func ParseMask(mask string) FieldMask {
	fm := FieldMask{Raw: mask}
	if mask == "" {
		return fm
	}

	runes := []rune(mask)
	fm.Patterns = make([]MaskChar, len(runes))

	for i, r := range runes {
		switch r {
		case 'X', 'x':
			fm.Patterns[i] = MaskAny
			fm.Width++
		case '9':
			fm.Patterns[i] = MaskDigit
			fm.Width++
		case 'A', 'a':
			fm.Patterns[i] = MaskAlpha
			fm.Width++
		case '!':
			fm.Patterns[i] = MaskUppercase
			fm.Width++
		default:
			fm.Patterns[i] = MaskLiteral
			fm.Width++
		}
	}
	return fm
}

// Validate checks whether a given input string matches the mask constraints.
func (fm FieldMask) Validate(input string) bool {
	inputRunes := []rune(input)
	if len(inputRunes) != len(fm.Patterns) {
		return false
	}
	for i, mc := range fm.Patterns {
		if i >= len(inputRunes) {
			return false
		}
		r := inputRunes[i]
		switch mc {
		case MaskLiteral:
			if r != rune(fm.Raw[i]) {
				return false
			}
		case MaskAny:
			// any character accepted
		case MaskDigit:
			if r < '0' || r > '9' {
				return false
			}
		case MaskAlpha:
			if !isAlpha(r) {
				return false
			}
		case MaskUppercase:
			if !isAlpha(r) {
				return false
			}
		}
	}
	return true
}

// InputMask returns a string showing the mask pattern for UI display,
// with literals in place and X/9/A/! converted to underscore placeholders.
func (fm FieldMask) InputMask() string {
	if len(fm.Patterns) == 0 {
		return ""
	}
	result := make([]rune, len(fm.Patterns))
	for i, mc := range fm.Patterns {
		switch mc {
		case MaskAny, MaskDigit, MaskAlpha, MaskUppercase:
			result[i] = '_'
		case MaskLiteral:
			if i < len(fm.Raw) {
				result[i] = rune(fm.Raw[i])
			} else {
				result[i] = ' '
			}
		}
	}
	return string(result)
}

// Format applies the mask to format a raw input value.
// For example, with mask "(XXX)XXX-XXXX", raw "8005550100" becomes "(800)555-0100".
func (fm FieldMask) Format(raw string) string {
	rawRunes := []rune(raw)
	result := make([]rune, 0, len(fm.Patterns))
	rawIdx := 0
	for _, mc := range fm.Patterns {
		switch mc {
		case MaskLiteral:
			if len(fm.Raw) > len(result) {
				result = append(result, rune(fm.Raw[len(result)]))
			}
		default:
			if rawIdx < len(rawRunes) {
				r := rawRunes[rawIdx]
				if mc == MaskUppercase {
					r = toUpper(r)
				}
				result = append(result, r)
				rawIdx++
			} else {
				result = append(result, ' ')
			}
		}
	}
	return string(result)
}

func isAlpha(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}
