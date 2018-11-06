package filecheck

import (
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

func TestGenMetadata(t *testing.T) {
	for i := 0; i < 1000; i++ {
		GenMetadata("/Users/lijt/Downloads/MenuMeters_1.9.7.zip", 32768)
		// GenMetadata("/Users/lijt/Downloads/ep.tar.gz", 32768)
	}

}
