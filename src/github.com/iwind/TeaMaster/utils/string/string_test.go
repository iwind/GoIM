package stringutil

import (
	"testing"
	"strconv"
)

func TestRandString(t *testing.T) {
	t.Log(Rand(10))
}

func TestConvertID(t *testing.T) {
	t.Log(ConvertID(1234567890))
}

func TestConvertIntToString(t *testing.T)  {
	t.Log(strconv.Itoa(123))
	t.Log(123)
}