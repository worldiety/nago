package msgstore

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// parseDSN parses a msgstore configuration DSN into [Options]. It covers the
// common, declarative case; settings that cannot be expressed as scalars
// (custom Compress/ShouldSplit functions, a shared FilePool) require the native
// [Options] struct instead.
//
// Accepted forms:
//
//	""                                      → defaults
//	"msgstore://?compress=s2&split=64mib"   → with scheme
//	"?compress=s2&split=count:5000"         → query only
//	"compress=s2&maxmsg=8mib"               → parameters only
//
// Recognised keys:
//
//	compress = none | s2 | default              → NoCompression / AlwaysS2 / DefaultCompression
//	split    = <size> | count:<n> | day | a,b   → e.g. "64mib", "count:5000", "64mib,day"
//	maxmsg   = <size>                           → MaxMessageSize, e.g. "16mib"
//	filepool = <n>                              → NewFilePool(n)
//
// Sizes accept the suffixes b, kb, kib, mb, mib, gb, gib (case-insensitive); a
// bare number is bytes. Unknown keys and malformed values are reported as an
// error so that typos fail fast.
func parseDSN(dsn string) (Options, error) {
	var opts Options

	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return opts, nil
	}

	query := dsn
	if i := strings.Index(dsn, "://"); i >= 0 {
		u, err := url.Parse(dsn)
		if err != nil {
			return opts, fmt.Errorf("parse dsn: %w", err)
		}
		query = u.RawQuery
	} else if strings.HasPrefix(dsn, "?") {
		query = dsn[1:]
	}

	if query == "" {
		return opts, nil
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		return opts, fmt.Errorf("parse dsn query: %w", err)
	}

	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		val := vals[len(vals)-1]

		switch key {
		case "compress":
			fn, err := parseCompress(val)
			if err != nil {
				return opts, err
			}
			opts.Compress = fn

		case "split":
			fn, err := parseSplit(val)
			if err != nil {
				return opts, err
			}
			opts.ShouldSplit = fn

		case "maxmsg":
			n, err := parseSize(val)
			if err != nil {
				return opts, fmt.Errorf("maxmsg: %w", err)
			}
			opts.MaxMessageSize = n

		case "filepool":
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return opts, fmt.Errorf("filepool: invalid handle count %q", val)
			}
			opts.FilePool = NewFilePool(n)

		default:
			return opts, fmt.Errorf("unknown config key %q", key)
		}
	}

	return opts, nil
}

func parseCompress(val string) (CompressFunc, error) {
	switch strings.ToLower(val) {
	case "none", "raw", "off":
		return NoCompression, nil
	case "s2", "always":
		return AlwaysS2, nil
	case "default", "auto":
		return DefaultCompression, nil
	default:
		return nil, fmt.Errorf("compress: unknown strategy %q", val)
	}
}

// parseSplit parses one or more split criteria joined by ',' into a SplitFunc.
// Each criterion is "count:<n>", "day", or a size (segment byte limit). A comma
// is used as separator rather than '+', because '+' decodes to a space in URL
// query strings.
func parseSplit(val string) (SplitFunc, error) {
	parts := strings.Split(val, ",")
	funcs := make([]SplitFunc, 0, len(parts))

	for _, raw := range parts {
		p := strings.TrimSpace(strings.ToLower(raw))
		if p == "" {
			continue
		}

		switch {
		case p == "day":
			funcs = append(funcs, SplitByDay())

		case strings.HasPrefix(p, "count:"):
			n, err := strconv.ParseUint(strings.TrimPrefix(p, "count:"), 10, 64)
			if err != nil || n == 0 {
				return nil, fmt.Errorf("split: invalid count %q", raw)
			}
			funcs = append(funcs, SplitByCount(n))

		default:
			n, err := parseSize(p)
			if err != nil {
				return nil, fmt.Errorf("split: %w", err)
			}
			funcs = append(funcs, SplitBySize(n))
		}
	}

	switch len(funcs) {
	case 0:
		return nil, fmt.Errorf("split: empty criterion")
	case 1:
		return funcs[0], nil
	default:
		return CombineSplits(funcs...), nil
	}
}

// parseSize parses a human-readable byte size such as "16mib", "512kb" or a
// bare byte count "1048576". It is case-insensitive and accepts both decimal
// (kb/mb/gb = 1000-based) and binary (kib/mib/gib = 1024-based) suffixes.
func parseSize(val string) (int64, error) {
	s := strings.TrimSpace(strings.ToLower(val))
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	type unit struct {
		suffix string
		mult   int64
	}
	// order matters: longer suffixes first so "mib" is matched before "mb"/"b".
	units := []unit{
		{"kib", 1 << 10}, {"mib", 1 << 20}, {"gib", 1 << 30},
		{"kb", 1000}, {"mb", 1000 * 1000}, {"gb", 1000 * 1000 * 1000},
		{"b", 1},
	}

	mult := int64(1)
	num := s
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			mult = u.mult
			num = strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
			break
		}
	}

	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", val)
	}
	if n < 0 {
		return 0, fmt.Errorf("negative size %q", val)
	}
	return n * mult, nil
}
