package msgstore

import (
	"iter"
	"log/slog"
)

// replayCursor tracks the iteration state for a single event type during
// a merged replay. Each cursor wraps a pull-style iterator over the chained
// segments of one type.
//
// The cursor's current Message (c.msg) contains a Payload that is a zero-copy
// view into the underlying read buffer. It is invalidated when advance() is
// called, because the buffer is reused for the next message. The Replay method
// always yields a message before advancing, so the caller can safely read the
// payload within each iteration step.
type replayCursor struct {
	typeID TypeID
	msg    Message
	next   func() (Message, error, bool) // pull iterator from iter.Pull2
	stop   func()                         // release goroutine
}

// advance moves the cursor to the next valid (non-tombstone, no error) message.
// Returns false when the iterator is exhausted.
func (c *replayCursor) advance() bool {
	for {
		msg, err, ok := c.next()
		if !ok {
			return false
		}
		if err != nil {
			slog.Warn("msgstore: replay cursor error", "type", c.typeID, "err", err)
			continue
		}
		if msg.IsTombstone() {
			continue
		}
		c.msg = msg
		return true
	}
}

// cursorHeap is a min-heap of replayCursors ordered by ascending SequenceID.
// This ensures the merged output is in strict global sequence order.
type cursorHeap []*replayCursor

func (h cursorHeap) Len() int            { return len(h) }
func (h cursorHeap) Less(i, j int) bool  { return h[i].msg.SequenceID < h[j].msg.SequenceID }
func (h cursorHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *cursorHeap) Push(x any)         { *h = append(*h, x.(*replayCursor)) }

func (h *cursorHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

// shouldSkipSegment returns true if a segment is entirely outside [minSeq, maxSeq].
// Pending segments are never skipped because their upper bound is unknown.
func shouldSkipSegment(seg segmentFile, minSeq, maxSeq uint64) bool {
	if seg.isPending() {
		// pending segments must always be scanned – we don't know their max sequence ID
		return false
	}
	return seg.maxSeq < minSeq || seg.minSeq > maxSeq
}

// replayType returns an iterator that chains through all segments of a single
// event type, yielding messages with SequenceID in [minSeq, maxSeq].
// Tombstones are NOT filtered here – that is done in replayCursor.advance().
func replayType(pool *FilePool, segments []segmentFile, maxMsgSize int64, minSeq, maxSeq uint64) iter.Seq2[Message, error] {
	return func(yield func(Message, error) bool) {
		for _, seg := range segments {
			if shouldSkipSegment(seg, minSeq, maxSeq) {
				continue
			}

			for msg, err := range readMessages(pool, seg.path, maxMsgSize) {
				if err != nil {
					if !yield(Message{}, err) {
						return
					}
					break // skip rest of this corrupt segment
				}
				if msg.SequenceID < minSeq {
					continue
				}
				if msg.SequenceID > maxSeq {
					return // remaining messages in this type are beyond range
				}
				if !yield(msg, nil) {
					return
				}
			}
		}
	}
}

