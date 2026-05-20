package gps

import "testing"

func TestParseParamsOrderIndependent(t *testing.T) {
	got := ParseParams("io24=1, pdop=1.1, g2=3, io1=0")
	if got["io24"] != "1" || got["pdop"] != "1.1" || got["g2"] != "3" || got["io1"] != "0" {
		t.Fatalf("unexpected params: %#v", got)
	}
}
