package proto

import (
	"bytes"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	t0 := &UpdateStateValueRequested{
		StatePointer:    1,
		FunctionPointer: 2,
	}

	var buf bytes.Buffer
	if err := Marshal(NewBinaryWriter(&buf), t0); err != nil {
		t.Error(err)
	}

	// just for comparison, we need 7 bytes, the most minimal json would be
	// {"t":1,"1":1,"2":2} = 19 byte

	obj, err := Unmarshal(NewBinaryReader(bytes.NewBuffer(buf.Bytes())))
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(obj, t0) {
		t.Errorf("expected: %v, actual: %v", t0, obj)
	}
}

func TestUnmarshal2(t *testing.T) {
	t0 := &AlignedComponent{
		Component: &Box{Frame: Frame{MaxWidth: "12"}},
		Alignment: 42,
	}

	var buf bytes.Buffer
	if err := Marshal(NewBinaryWriter(&buf), t0); err != nil {
		t.Error(err)
	}

	// just for comparison, we need 7 bytes, the most minimal json would be
	// {"t":1,"1":1,"2":2} = 19 byte

	obj, err := Unmarshal(NewBinaryReader(bytes.NewBuffer(buf.Bytes())))
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(obj, t0) {
		t.Errorf("expected: %v, actual: %v", t0, obj)
	}
}
