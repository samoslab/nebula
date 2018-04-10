package bytes

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// func TestSameBytes(t *testing.T) {
// 	if !SameBytes(nil, nil) {
// 		t.Errorf("failed")
// 	}
// 	bytes, _ := hex.DecodeString("c925852911756e6d4b14b425188f5cf67d1d3cfc")
// 	if SameBytes(nil, bytes) {
// 		t.Errorf("failed")
// 	}
// 	if SameBytes(bytes, nil) {
// 		t.Errorf("failed")
// 	}
// 	bytes2, _ := hex.DecodeString("c925852911756e6d4b14b425188f5cf67d1d3cfc")
// 	if !SameBytes(bytes, bytes2) {
// 		t.Errorf("failed")
// 	}
// }

func TestFromUint32(t *testing.T) {
	var i uint32 = 123456789
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	if !bytes.Equal(FromUint32(i), b) {
		t.Errorf("failed")
	}
	i = 987654321
	b = make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	if !bytes.Equal(FromUint32(i), b) {
		t.Errorf("failed")
	}
}

func TestFromUint64(t *testing.T) {
	var i uint64 = 123456789987654321
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	if !bytes.Equal(FromUint64(i), b) {
		t.Errorf("failed")
	}
	i = 987654321123456789
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	if !bytes.Equal(FromUint64(i), b) {
		t.Errorf("failed")
	}
}

func TestFillUint32(t *testing.T) {
	b := make([]byte, 8)
	var v uint32
	v = 123456789
	for i := 0; i < 5; i++ {
		FillUint32(b, i, v)
		if !bytes.Equal(FromUint32(v), b[i:i+4]) {
			t.Errorf("failed")
		}
	}
	v = 987654321
	for i := 0; i < 5; i++ {
		FillUint32(b, i, v)
		if !bytes.Equal(FromUint32(v), b[i:i+4]) {
			t.Errorf("failed")
		}
	}
}

func TestToUint32(t *testing.T) {
	var i uint32 = 123456789
	if ToUint32(FromUint32(i), 0) != i {
		t.Errorf("failed")
	}
	i = 987654321
	if ToUint32(FromUint32(i), 0) != i {
		t.Errorf("failed")
	}
}
