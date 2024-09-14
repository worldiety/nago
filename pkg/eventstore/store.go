package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xiter"
	"io"
	"io/fs"
	"iter"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// ID contains an encoded timestamp used for lexicographical ordering and simple prefix queries.
// A valid ID always starts with <year>/<month>/<day>/<hour>/<min>/<sec>/<milliseconds>/... with additional data
// appended, like a value from the EVENTSTORE_INSTANCE_NAME environment variable or a sequence number or
// even some appended random data.
type ID string

// Max returns the maximum year to search for. We are mostly using milliseconds internally, however, if you
// calculate in Nanoseconds, keep in mind that it wraps around in 2262.
const Max ID = "9999"

func (id ID) Time(loc *time.Location) (time.Time, error) {
	if len(id) < 24 {
		return time.Time{}, fmt.Errorf("empty id %s", id)
	}

	str := string(id[:19])
	secTime, err := time.ParseInLocation("2006/01/02/15/04/05", str, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse time from id %s: %w", id, err)
	}

	ms := string(id[20:24])
	msp, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return secTime, fmt.Errorf("cannot parse ms fraction from id %s: %w", id, err)
	}

	return secTime.Add(time.Duration(msp) * time.Millisecond), nil
}

var lastUnixMilli atomic.Int64
var lastSeqNo atomic.Int64

// instanceName can be influenced from the environment
var instanceName = os.Getenv("EVENTSTORE_INSTANCE_NAME")

// NewID returns an ID which is constructed as follows:
//
//	<year>/<month>/<day>/<hour>/<min>/<sec>/<milliseconds>/(<EVENTSTORE_INSTANCE_NAME>/)?<seq number in millisecond>
//
// Important: this is only suitable for a single machine use case. If you need to distribute across multiple process
// instances (or machines) you have to set the EVENTSTORE_INSTANCE_NAME environment variable to something unique.
// If no EVENTSTORE_INSTANCE_NAME is defined, the path segment is omitted.
func NewID() ID {
	return timeIntoID(time.Now())
}

// timeIntoID expects the given time to be monotonic. Otherwise, the generated IDs may cause collisions and may cause
// broken ids in the future due to a global state.
func timeIntoID(now time.Time) ID {
	nowMilli := now.UnixMilli()
	var seqNo int64
	if lastUnixMilli.Load() == nowMilli {
		seqNo = lastSeqNo.Add(1)
	} else {
		lastSeqNo.Store(0)
	}

	lastUnixMilli.Store(nowMilli)

	year, month, day := now.Date()
	hour := now.Hour()
	minute := now.Minute()
	sec := now.Second()
	ms := now.Nanosecond() / 1e6
	if instanceName == "" {
		// shorter version
		return ID(fmt.Sprintf("%d/%02d/%02d/%02d/%02d/%02d/%04d/%d", year, month, day, hour, minute, sec, ms, seqNo))
	}

	// include node instance id
	return ID(fmt.Sprintf("%d/%02d/%02d/%02d/%02d/%02d/%04d/%s/%d", year, month, day, hour, minute, sec, ms, instanceName, seqNo))
}

type ConsumerID [32]byte

type Message struct {
	ID   ID
	Data []byte
}

type jsonMessage struct {
	ContentType string `json:"c"`
	Data        []byte `json:"d"`
}

type Store struct {
	store blob.Store2
	ctx   context.Context
}

// Replay returns a sequence which iterates over all messages which occurred after the given offset ID. Thus,
// it is an open-range search with exclusive semantics of the offset value.
func (s *Store) Replay(offsetExcl ID) iter.Seq2[Message, error] {
	return xiter.Map2(func(key string, err error) (Message, error) {
		if err != nil {
			return Message{}, err
		}

		optReader, err := s.store.NewReader(s.ctx, key)
		if err != nil {
			return Message{}, err
		}

		if !optReader.IsNone() {
			return Message{}, fs.ErrNotExist // TODO better omit the entire entry, but we delay closing when used with filter?
		}

		return fromEntry(key, optReader.Unwrap())
	}, s.store.List(s.ctx, blob.ListOptions{
		MinInc: string(offsetExcl),
		MaxInc: string(Max),
	}))
}

// Save persists the given data to the underlying store and returns the generated unique ID to identify it.
// Without considering edge cases like manipulating the unix time clock, the returned ID is strictly monotonic and
// will never cause any collision. If used in a cluster context, you must ensure that the EVENTSTORE_INSTANCE_NAME
// environment variable is unique for each node, to guarantee that no collisions arise.
func (s *Store) Save(data []byte) (ID, error) {
	id := NewID()
	w, err := s.store.NewWriter(s.ctx, string(id))
	if err != nil {
		return "", fmt.Errorf("cannot open writer %w", err)
	}

	if _, err := w.Write(data); err != nil {
		_ = w.Close() // suppressing any followup failures
		return "", fmt.Errorf("cannot write message %w", err)
	}

	if err = w.Close(); err != nil {
		return "", fmt.Errorf("cannot commit writer %w", err)
	}

	return id, nil
}

func consumerNameToStoreKey(consumer string) string {
	return fmt.Sprintf("$%s", consumer)
}

func (s *Store) Offset(consumer string, accepted ID) (msg Message, err error) {
	w, err := s.store.NewWriter(s.ctx, consumerNameToStoreKey(consumer))
	if err != nil {
		return Message{}, fmt.Errorf("cannot open writer %w", err)
	}

	defer std.Try(w.Close, &err)
	return Message{}, nil
}

func fromEntry(key string, r io.ReadCloser) (Message, error) {
	defer r.Close()

	var msg jsonMessage
	dec := json.NewDecoder(r)
	if err := dec.Decode(&msg); err != nil {
		return Message{}, fmt.Errorf("cannot decode %v: %w", key, err)
	}

	return Message{
		ID:   ID(key),
		Data: msg.Data,
	}, nil
}

// Unmarshal loads a distinct message by id and calls the given callback for decoding. The message is owned by
// the store and must not escape.
func Unmarshal[T any](store Store, seq int64, fn func(Message) (T, error)) (std.Option[Message], error) {
	panic("@")
}
