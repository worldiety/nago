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

type Rem int

func (p Rem) MarshalJSON() ([]byte, error) {
	tmp := make([]byte, 0, 8)
	tmp = append(tmp, '"')
	tmp = strconv.AppendInt(tmp, int64(p), 10)
	tmp = append(tmp, 'r', 'e', 'm', '"')
	return tmp, nil
}

func (Rem) isLength() {}
