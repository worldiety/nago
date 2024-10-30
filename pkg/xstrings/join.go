package xstrings

import "strings"

func Join[T ~string](s []T, sep string) T {
	tmp := make([]string, 0, len(s))
	for i := range s {
		tmp = append(tmp, string(s[i]))
	}

	return T(strings.Join(tmp, sep))
}
