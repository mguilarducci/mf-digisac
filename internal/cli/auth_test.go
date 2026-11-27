package cli

import "testing"

func TestMask(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"abc", "***"},
		{"12345678", "********"},
		{"123456789", "1234*6789"},
		{"abcd0000000000wxyz", "abcd**********wxyz"},
	}
	for _, c := range cases {
		if got := mask(c.in); got != c.want {
			t.Errorf("mask(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
