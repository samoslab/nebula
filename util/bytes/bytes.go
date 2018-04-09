package bytes

//bytes.Equal
// func SameBytes(b1 []byte, b2 []byte) bool {
// 	if b1 == nil {
// 		if b2 == nil {
// 			return true
// 		} else {
// 			return false
// 		}
// 	} else {
// 		if b2 == nil {
// 			return false
// 		} else {
// 			if len(b1) != len(b2) {
// 				return false
// 			} else {
// 				for idx, v1 := range b1 {
// 					if v1 != b2[idx] {
// 						return false
// 					}
// 				}
// 				return true
// 			}
// 		}
// 	}
// }

func FillUint32(b []byte, startIdx int, val uint32) {
	for i := 3; i >= 0; i-- {
		b[startIdx+i] = byte(val & 255)
		val = val >> 8
		if val == 0 {
			break
		}
	}
}

func FromUint64(val uint64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(val & 255)
		val = val >> 8
		if val == 0 {
			break
		}
	}
	return b
}

func FromUint32(val uint32) []byte {
	b := make([]byte, 4)
	for i := 3; i >= 0; i-- {
		b[i] = byte(val & 255)
		val = val >> 8
		if val == 0 {
			break
		}
	}
	return b
}

func ToUint32(b []byte, startIdx int) uint32 {
	return uint32(b[startIdx+3]) | uint32(b[startIdx+2])<<8 | uint32(b[startIdx+1])<<16 | uint32(b[startIdx])<<24
}
