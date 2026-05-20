package output

import "testing"

func TestPrefixRemovesExtension(t *testing.T) {
	cases := map[string]string{
		"april.csv":      "april",
		"april.CSV":      "april",
		"/tmp/april.csv": "/tmp/april",
	}
	for in, want := range cases {
		if got := Prefix(in); got != want {
			t.Fatalf("Prefix(%q) = %q, want %q", in, got, want)
		}
	}
}
