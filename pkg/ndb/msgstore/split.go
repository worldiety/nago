package msgstore

import "time"

// SegmentInfo describes the current state of a pending segment file.
// It is passed to SplitFunc to decide whether a new segment should be started.
type SegmentInfo struct {
	// MessageCount is the number of messages written to this segment so far.
	MessageCount uint64
	// ByteSize is the current file size in bytes (including header).
	ByteSize int64
	// FirstSeqID is the sequence ID of the first message in this segment.
	FirstSeqID uint64
	// FirstTimestamp is the unix-nano timestamp of the first message in this segment.
	FirstTimestamp int64
}

// SplitFunc decides whether the current pending segment should be finalized
// and a new one started. Return true to split.
type SplitFunc func(seg SegmentInfo) bool

// SplitBySize returns a SplitFunc that splits when the segment exceeds maxBytes.
func SplitBySize(maxBytes int64) SplitFunc {
	return func(seg SegmentInfo) bool {
		return seg.ByteSize >= maxBytes
	}
}

// SplitByCount returns a SplitFunc that splits when the segment reaches n messages.
func SplitByCount(n uint64) SplitFunc {
	return func(seg SegmentInfo) bool {
		return seg.MessageCount >= n
	}
}

// SplitByDay returns a SplitFunc that splits when the first message's day
// differs from the current UTC day.
func SplitByDay() SplitFunc {
	return func(seg SegmentInfo) bool {
		if seg.FirstTimestamp == 0 {
			return false
		}
		first := time.Unix(0, seg.FirstTimestamp).UTC()
		now := time.Now().UTC()
		return first.Year() != now.Year() || first.YearDay() != now.YearDay()
	}
}

// CombineSplits returns a SplitFunc that triggers a split if any of the
// provided functions returns true (OR combination).
func CombineSplits(funcs ...SplitFunc) SplitFunc {
	return func(seg SegmentInfo) bool {
		for _, fn := range funcs {
			if fn(seg) {
				return true
			}
		}
		return false
	}
}

// defaultSplit splits at 64 MiB or on day boundary.
var defaultSplit = CombineSplits(SplitBySize(64<<20), SplitByDay())

