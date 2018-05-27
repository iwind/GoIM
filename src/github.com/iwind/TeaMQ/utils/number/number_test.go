package numberutil

import "testing"

func TestRangeInt(t *testing.T) {
	t.Log(RangeInt(30, 10, 4))
	t.Log(RangeInt(10, 30, 4))
}

func TestTimes(t *testing.T) {
	Times(10, func(i uint) {
		t.Log(i)
	})
}
