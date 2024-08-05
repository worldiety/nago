package ui

import "go.wdy.de/nago/presentation/ora"

type TKeyboardOptions struct {
	capitalization     bool
	autoCorrectEnabled bool
	keyboardType       KeyboardType
}

func KeyboardOptions() TKeyboardOptions {
	return TKeyboardOptions{}
}

func (opts TKeyboardOptions) ora() ora.KeyboardOptions {
	return ora.KeyboardOptions{
		Capitalization:     opts.capitalization,
		AutoCorrectEnabled: opts.autoCorrectEnabled,
		KeyboardType:       opts.keyboardType.ora(),
	}
}

func (opts TKeyboardOptions) Capitalization(capitalization bool) TKeyboardOptions {
	opts.capitalization = capitalization
	return opts
}

func (opts TKeyboardOptions) AutoCorrectEnabled(autoCorrectEnabled bool) TKeyboardOptions {
	opts.autoCorrectEnabled = autoCorrectEnabled
	return opts
}

// KeyboardType is a hint to the frontend. Technically, it is impossible to actually
// guarantee anything, and you have always to considers bugs and hacks:
//   - a malicious user may send you anything, which would otherwise not be possible (e.g. text instead of numbers)
//   - Android IME hints or keyboard types are never guaranteed. A user may install third-party keyboards which just ignore anything
//   - a user may inject anything using wrong autocompletion or the clipboard
func (opts TKeyboardOptions) KeyboardType(keyboardType KeyboardType) TKeyboardOptions {
	opts.keyboardType = keyboardType
	return opts
}

type KeyboardType int

func (k KeyboardType) ora() ora.KeyboardType {
	switch k {
	case KeyboardAscii:
		return ora.KeyboardAscii
	case KeyboardInteger:
		return ora.KeyboardInteger
	case KeyboardFloat:
		return ora.KeyboardFloat
	case KeyboardEMail:
		return ora.KeyboardEMail
	case KeyboardPhone:
		return ora.KeyboardPhone
	case KeyboardSearch:
		return ora.KeyboardSearch
	case KeyboardURL:
		return ora.KeyboardURL
	default:
		return ora.KeyboardDefault
	}
}

const (
	KeyboardDefault KeyboardType = iota
	KeyboardAscii
	KeyboardInteger
	KeyboardFloat
	KeyboardEMail
	KeyboardPhone
	KeyboardSearch
	KeyboardURL
)
