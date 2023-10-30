package ui

import "strconv"

type Length interface {
	isLength()
}

type Px int

func (p Px) MarshalJSON() ([]byte, error) {
	tmp := make([]byte, 0, 8)
	tmp = append(tmp, '"')
	tmp = strconv.AppendInt(tmp, int64(p), 10)
	tmp = append(tmp, 'p', 'x', '"')
	return tmp, nil
}

func (Px) isLength() {}

type Rem float64

func (p Rem) MarshalJSON() ([]byte, error) {
	tmp := make([]byte, 0, 8)
	tmp = append(tmp, '"')
	tmp = strconv.AppendFloat(tmp, float64(p), 'f', 3, 64)
	tmp = append(tmp, 'r', 'e', 'm', '"')
	return tmp, nil
}

func (Rem) isLength() {}

// Fr is defined like the CSS fraction type
type Fr int

func (p Fr) MarshalJSON() ([]byte, error) {
	tmp := make([]byte, 0, 8)
	tmp = append(tmp, '"')
	tmp = strconv.AppendInt(tmp, int64(p), 10)
	tmp = append(tmp, 'f', 'r', '"')
	return tmp, nil
}

func (Fr) isLength() {}
