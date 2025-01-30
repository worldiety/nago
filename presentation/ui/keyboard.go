package ui

import "go.wdy.de/nago/presentation/proto"

type TKeyboardOptions struct {
	capitalization     bool
	autoCorrectEnabled bool
	keyboardType       KeyboardType
}

func KeyboardOptions() TKeyboardOptions {
	return TKeyboardOptions{}
}

func (opts TKeyboardOptions) ora() proto.KeyboardOptions {
	return proto.KeyboardOptions{
		Capitalization:     proto.Bool(opts.capitalization),
		AutoCorrectEnabled: proto.Bool(opts.autoCorrectEnabled),
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

func (k KeyboardType) ora() proto.KeyboardType {
	switch k {
	case KeyboardAscii:
		return proto.KeyboardAscii
	case KeyboardInteger:
		return proto.KeyboardInteger
	case KeyboardFloat:
		return proto.KeyboardFloat
	case KeyboardEMail:
		return proto.KeyboardEMail
	case KeyboardPhone:
		return proto.KeyboardPhone
	case KeyboardSearch:
		return proto.KeyboardSearch
	case KeyboardURL:
		return proto.KeyboardURL
	default:
		return proto.KeyboardDefault
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
