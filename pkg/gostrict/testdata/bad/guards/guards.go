package guards

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Item struct{ N int32 }

func Divide(a, b int32) int32 {
	return a / b // want GS030
}

func Guarded(a, b int32) int32 {
	if b != 0 {
		return a / b
	}
	return 0
}

func Early(a, b int32) int32 {
	if b == 0 {
		return 0
	}
	return a % b
}

func Positive(a, b int32) int32 {
	if b > 0 && a/b > 1 {
		return 1
	}
	return 0
}

func Loop(a, b int32) int32 {
	if b == 0 {
		return 0
	}
	s := int32(0)
	for i := int32(0); i < 3; i++ {
		s += a / b // want GS030
		b--
	}
	return s
}

func Closure(a, b int32) int32 {
	reset := func() { b = 0 }
	if b != 0 {
		reset()
		return a / b // want GS030
	}
	return 0
}

func Field(a int32, it Item) int32 {
	return a / it.N // want GS030
}

func Get(l list.List[*Item]) int32 {
	v, ok := l.Get(0)
	if ok {
		return v.N
	}
	w, _ := l.Get(1) // want GS033
	_ = w
	x, found := l.Get(2)
	_ = found
	return x.N // want GS033
}

func GetEarly(m dict.Map[string, *Item]) int32 {
	v, ok := m.Get("a")
	if !ok {
		return 0
	}
	return v.N
}

func Plain(l list.List[Item]) int32 {
	v, _ := l.Get(0)
	return v.N
}

func Errors(f func() error) string {
	err := f()
	if err != nil {
		return err.Error()
	}
	return err.Error() // want GS033
}

func Reassigned(f func() error) string {
	err := f()
	if err != nil {
		err = f()
		return err.Error() // want GS033
	}
	return ""
}
