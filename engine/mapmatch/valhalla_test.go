package mapmatch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNoopMapMatcherSafeEmpty(t *testing.T) {
	m := NoopMapMatcher{}
	res, err := m.Match(MatchRequest{})
	if err != nil || len(res.MatchedEdges) != 0 || len(res.MatchedShape) != 0 {
		t.Fatalf("expected empty safe result, got res=%+v err=%v", res, err)
	}
}

func TestValhallaSendsExpectedRequestAndParsesResponse(t *testing.T) {
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trace_route" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"trip":{"summary":{"length":1.23,"time":321},"legs":[{"shape":"abc","maneuvers":[{"street_names":["Main St"]}]}]}}`))
	}))
	defer ts.Close()
	m := NewValhallaMapMatcher(ValhallaConfig{BaseURL: ts.URL, Timeout: time.Second})
	res, err := m.Match(MatchRequest{Points: []MatchPoint{{Lat: -33, Lng: 151, Time: time.Unix(100, 0)}}})
	if err != nil {
		t.Fatal(err)
	}
	if got["shape"] == nil {
		t.Fatalf("expected shape in request")
	}
	if len(res.MatchedEdges) != 1 || res.MatchedEdges[0].Name != "Main St" {
		t.Fatalf("unexpected edges %+v", res.MatchedEdges)
	}
	if res.DistanceM != 1230 || res.DurationS != 321 {
		t.Fatalf("unexpected summary %+v", res)
	}
}

func TestValhallaHTTP500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer ts.Close()
	m := NewValhallaMapMatcher(ValhallaConfig{BaseURL: ts.URL})
	_, err := m.Match(MatchRequest{Points: []MatchPoint{{Lat: 1, Lng: 1}}})
	if err == nil || !strings.Contains(err.Error(), "HTTP error") {
		t.Fatalf("err=%v", err)
	}
}

func TestValhallaMalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{")) }))
	defer ts.Close()
	m := NewValhallaMapMatcher(ValhallaConfig{BaseURL: ts.URL})
	_, err := m.Match(MatchRequest{Points: []MatchPoint{{Lat: 1, Lng: 1}}})
	if err == nil || !strings.Contains(err.Error(), "decode valhalla response") {
		t.Fatalf("err=%v", err)
	}
}

func TestValhallaEmptyInputSafe(t *testing.T) {
	m := NewValhallaMapMatcher(ValhallaConfig{BaseURL: "http://example"})
	res, err := m.Match(MatchRequest{})
	if err != nil || res.DistanceM != 0 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}
