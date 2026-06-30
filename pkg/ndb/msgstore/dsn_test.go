package msgstore

import (
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"", 0, true},
		{"0", 0, false},
		{"1024", 1024, false},
		{"512b", 512, false},
		{"1kib", 1 << 10, false},
		{"16mib", 16 << 20, false},
		{"2gib", 2 << 30, false},
		{"1kb", 1000, false},
		{"3mb", 3 * 1000 * 1000, false},
		{"1gb", 1000 * 1000 * 1000, false},
		{"16MiB", 16 << 20, false},
		{"  8mib  ", 8 << 20, false},
		{"-5", 0, true},
		{"abc", 0, true},
		{"12x", 0, true},
	}

	for _, tt := range tests {
		got, err := parseSize(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseSize(%q): expected error, got %d", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSize(%q): unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSize(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestParseDSNDefaults(t *testing.T) {
	for _, in := range []string{"", "   ", "msgstore://", "?"} {
		opts, err := parseDSN(in)
		if err != nil {
			t.Errorf("parseDSN(%q): unexpected error: %v", in, err)
			continue
		}
		if opts.Compress != nil || opts.ShouldSplit != nil || opts.FilePool != nil || opts.MaxMessageSize != 0 {
			t.Errorf("parseDSN(%q): expected zero Options, got %+v", in, opts)
		}
	}
}

func TestParseDSNForms(t *testing.T) {
	// scheme, query-only and bare-parameter forms must all be accepted.
	for _, in := range []string{
		"msgstore://?maxmsg=8mib",
		"?maxmsg=8mib",
		"maxmsg=8mib",
	} {
		opts, err := parseDSN(in)
		if err != nil {
			t.Fatalf("parseDSN(%q): %v", in, err)
		}
		if opts.MaxMessageSize != 8<<20 {
			t.Errorf("parseDSN(%q): maxmsg = %d, want %d", in, opts.MaxMessageSize, 8<<20)
		}
	}
}

func TestParseDSNCompress(t *testing.T) {
	cases := map[string]Encoding{
		"compress=none":    EncodingRaw,
		"compress=s2":      EncodingS2,
		"compress=default": EncodingRaw, // small payload below the 512B threshold
	}
	for dsn, wantEnc := range cases {
		opts, err := parseDSN(dsn)
		if err != nil {
			t.Fatalf("parseDSN(%q): %v", dsn, err)
		}
		if opts.Compress == nil {
			t.Fatalf("parseDSN(%q): Compress is nil", dsn)
		}
		enc, _ := opts.Compress("tiny-type", []byte("tiny"))
		if enc != wantEnc {
			t.Errorf("parseDSN(%q): encoding for tiny payload = %d, want %d", dsn, enc, wantEnc)
		}
	}
}

func TestParseDSNSplit(t *testing.T) {
	// count:3 — splits at the 3rd message.
	opts, err := parseDSN("split=count:3")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if opts.ShouldSplit(SegmentInfo{MessageCount: 2}) {
		t.Error("split=count:3 should not split at 2 messages")
	}
	if !opts.ShouldSplit(SegmentInfo{MessageCount: 3}) {
		t.Error("split=count:3 should split at 3 messages")
	}

	// size — splits at the byte threshold.
	opts, err = parseDSN("split=64mib")
	if err != nil {
		t.Fatalf("size: %v", err)
	}
	if opts.ShouldSplit(SegmentInfo{ByteSize: (64 << 20) - 1}) {
		t.Error("split=64mib should not split below the limit")
	}
	if !opts.ShouldSplit(SegmentInfo{ByteSize: 64 << 20}) {
		t.Error("split=64mib should split at the limit")
	}

	// combined a,b — OR semantics.
	opts, err = parseDSN("split=64mib,count:3")
	if err != nil {
		t.Fatalf("combined: %v", err)
	}
	if !opts.ShouldSplit(SegmentInfo{MessageCount: 3}) {
		t.Error("combined split should trigger on count")
	}
	if !opts.ShouldSplit(SegmentInfo{ByteSize: 64 << 20}) {
		t.Error("combined split should trigger on size")
	}
	if opts.ShouldSplit(SegmentInfo{MessageCount: 1, ByteSize: 10}) {
		t.Error("combined split should not trigger below both limits")
	}
}

func TestParseDSNFilePool(t *testing.T) {
	opts, err := parseDSN("filepool=256")
	if err != nil {
		t.Fatalf("filepool: %v", err)
	}
	if opts.FilePool == nil {
		t.Fatal("expected a FilePool")
	}
}

func TestParseDSNCombinedKeys(t *testing.T) {
	opts, err := parseDSN("?compress=s2&split=count:5&maxmsg=8mib&filepool=128")
	if err != nil {
		t.Fatalf("combined keys: %v", err)
	}
	if opts.Compress == nil || opts.ShouldSplit == nil || opts.FilePool == nil {
		t.Fatalf("expected all fields set, got %+v", opts)
	}
	if opts.MaxMessageSize != 8<<20 {
		t.Errorf("maxmsg = %d, want %d", opts.MaxMessageSize, 8<<20)
	}
}

func TestParseDSNErrors(t *testing.T) {
	for _, in := range []string{
		"unknown=1",          // unknown key
		"compress=lz4",       // unknown compress strategy
		"split=count:0",      // invalid count
		"split=count:abc",    // non-numeric count
		"split=12x",          // bad size in split
		"maxmsg=banana",      // bad size
		"filepool=0",         // non-positive handle count
		"filepool=-1",        // negative
		"filepool=abc",       // non-numeric
		"split=",             // empty split criterion
	} {
		if _, err := parseDSN(in); err == nil {
			t.Errorf("parseDSN(%q): expected error, got nil", in)
		}
	}
}
