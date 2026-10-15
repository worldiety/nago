// Package wordstats computes simple statistics about the words of a text.
package wordstats

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Counter counts occurrences of keys and remembers their first occurrence.
type Counter[K dict.Key] struct {
	counts dict.Map[K, int32]
}

func NewCounter[K dict.Key]() Counter[K] {
	return Counter[K]{counts: dict.New[K, int32]()}
}

func (c Counter[K]) Add(k K) {
	n, _ := c.counts.Get(k)
	c.counts.Set(k, n+1)
}

type entry[K dict.Key] struct {
	Key   K
	Count int32
}

// Top returns the n most frequent keys; ties keep the order of first occurrence.
func (c Counter[K]) Top(n int32) list.List[entry[K]] {
	entries := list.New[entry[K]]()
	for k, v := range c.counts.All() {
		entries.Append(entry[K]{Key: k, Count: v})
	}
	list.SortFunc(entries, func(a, b entry[K]) int32 {
		return b.Count - a.Count
	})
	res := list.New[entry[K]]()
	for i, e := range entries.All() {
		if i >= n {
			break
		}
		res.Append(e)
	}
	return res
}

func isWordChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '\''
}

func words(s string) list.List[string] {
	res := list.New[string]()
	cur := ""
	for c := range text.CodePoints(text.LowerASCII(s)).Values() {
		if isWordChar(c) {
			cur += text.FromCodePoint(c)
			continue
		}
		if cur != "" {
			res.Append(cur)
			cur = ""
		}
	}
	if cur != "" {
		res.Append(cur)
	}
	return res
}

func Run(s string) string {
	ws := words(s)
	counter := NewCounter[string]()
	var total int32
	longest := ""
	for w := range ws.Values() {
		counter.Add(w)
		n := text.RuneCount(w)
		total += n
		if n > text.RuneCount(longest) {
			longest = w
		}
	}
	count := ws.Len()
	avg := int32(0)
	if count != 0 {
		avg = total / count
	}

	top := list.New[string]()
	for e := range counter.Top(3).Values() {
		top.Append(e.Key + "=" + text.FromInt(e.Count))
	}
	return "words=" + text.FromInt(count) + " avg=" + text.FromInt(avg) +
		" longest=" + longest + " top=" + text.Join(top, ",")
}
