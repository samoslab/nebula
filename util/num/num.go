package num

import "strconv"

func FixLength(val uint32, length int) string {
	s := strconv.Itoa(int(val))
	for len(s) < length {
		s = "0" + s
	}
	return s
}
