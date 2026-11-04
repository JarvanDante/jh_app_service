package public

import (
	"math"
	"testing"
)

func TestFormatAmountModeEdges(t *testing.T) {
	cases := []struct {
		v    float64
		d    int
		mode string
		want string
	}{
		// negatives in every mode
		{-0.005, 2, RoundingHalfUp, "-0.01"},
		{-2.5, 0, RoundingHalfEven, "-2"},
		{-3.5, 0, RoundingHalfEven, "-4"},
		{-1.001, 2, RoundingUp, "-1.01"},
		{-0.004, 2, RoundingHalfUp, "0.00"},
		{-0.0001, 2, RoundingUp, "-0.01"},
		// HALF_EVEN ties at several precisions
		{0.125, 2, RoundingHalfEven, "0.12"},
		{0.135, 2, RoundingHalfEven, "0.14"},
		{1.00005, 4, RoundingHalfEven, "1.0000"},
		{1.00015, 4, RoundingHalfEven, "1.0002"},
		{0.5, 0, RoundingHalfEven, "0"},
		{1.5, 0, RoundingHalfEven, "2"},
		// 0 and 4 decimals
		{12860.5, 0, RoundingHalfUp, "12861"},
		{12860.5, 0, RoundingDown, "12860"},
		{1.23456, 4, RoundingHalfUp, "1.2346"},
		{1.23454, 4, RoundingHalfUp, "1.2345"},
		{0, 0, RoundingHalfUp, "0"},
		{0, 4, RoundingHalfUp, "0.0000"},
		// tiny and large values (no exponent notation, carry across the integer part)
		{1e-7, 4, RoundingHalfUp, "0.0000"},
		{1e-7, 4, RoundingUp, "0.0001"},
		{99999999.995, 2, RoundingHalfUp, "100000000.00"},
		{12345678901234.5, 0, RoundingHalfUp, "12345678901235"},
		{1e21, 2, RoundingHalfUp, "1000000000000000000000.00"},
		{2147483647, 2, RoundingHalfUp, "2147483647.00"},
		// invalid input
		{math.NaN(), 2, RoundingHalfUp, "0.00"},
		{math.Inf(1), 2, RoundingHalfUp, "0.00"},
		{math.Inf(-1), 0, RoundingHalfUp, "0"},
		{1.5, -3, RoundingHalfUp, "2"},
	}
	for _, c := range cases {
		if got := formatAmountMode(c.v, c.d, c.mode); got != c.want {
			t.Errorf("formatAmountMode(%v,%d,%q)=%s want %s", c.v, c.d, c.mode, got, c.want)
		}
	}
}
