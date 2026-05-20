package gps

import (
	"math"
	"strconv"
	"strings"
)

var ParamOrder = []string{"gpslev", "gsmlev", "pdop", "io1", "io3", "io6", "io11", "io14", "io24", "io66", "io67", "io113", "io175", "io180", "io200", "io236", "io239", "io240", "io241", "io246", "io247", "io251", "io252", "io253", "io254", "io303", "io310", "io380", "io381", "g0", "g1", "g2"}

func ParseParams(s string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" || !strings.Contains(part, "=") {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return out
}

func NumericParams(params map[string]string) map[string]float64 {
	out := map[string]float64{}
	for k, v := range params {
		if strings.ContainsAny(strings.ToUpper(v), "ABCDEF") && k != "pdop" {
			continue
		}
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			out[k] = n
		}
	}
	return out
}

func Num(p RawPoint, key string) (float64, bool)               { v, ok := p.ParamNums[key]; return v, ok }
func NumFrom(m map[string]float64, key string) (float64, bool) { v, ok := m[key]; return v, ok }

func AccelMagnitude(nums map[string]float64) (float64, bool) {
	g0, ok0 := nums["g0"]
	g1, ok1 := nums["g1"]
	g2, ok2 := nums["g2"]
	if !(ok0 && ok1 && ok2) {
		return 0, false
	}
	return math.Sqrt(g0*g0 + g1*g1 + g2*g2), true
}

func PDOPQuality(pdop, ideal, poor float64, ok bool) string {
	if !ok {
		return ""
	}
	if pdop <= ideal {
		return "ideal"
	}
	if pdop > poor {
		return "poor"
	}
	return "usable"
}
