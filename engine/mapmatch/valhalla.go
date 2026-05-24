package mapmatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ValhallaConfig struct {
	BaseURL             string
	Timeout             time.Duration
	Endpoint            string
	MaxPointsPerRequest int
}

type ValhallaMapMatcher struct {
	baseURL             string
	endpoint            string
	maxPointsPerRequest int
	client              *http.Client
}

func NewValhallaMapMatcher(cfg ValhallaConfig) *ValhallaMapMatcher {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = "trace_route"
	}
	return &ValhallaMapMatcher{baseURL: strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"), endpoint: endpoint, maxPointsPerRequest: cfg.MaxPointsPerRequest, client: &http.Client{Timeout: timeout}}
}

func (v *ValhallaMapMatcher) Match(req MatchRequest) (MatchResult, error) {
	if len(req.Points) == 0 {
		return MatchResult{}, nil
	}
	maxPts := v.maxPointsPerRequest
	if req.MaxPointsPerRequest > 0 {
		maxPts = req.MaxPointsPerRequest
	}
	if maxPts > 0 && len(req.Points) > maxPts {
		return MatchResult{}, fmt.Errorf("too many points: got %d max %d", len(req.Points), maxPts)
	}
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		endpoint = v.endpoint
	}
	payload := map[string]any{"costing": "auto", "shape_match": "map_snap", "shape": buildShape(req.Points)}
	buf, err := json.Marshal(payload)
	if err != nil {
		return MatchResult{}, fmt.Errorf("marshal valhalla request: %w", err)
	}
	httpReq, err := http.NewRequest(http.MethodPost, v.baseURL+"/"+strings.TrimLeft(endpoint, "/"), bytes.NewReader(buf))
	if err != nil {
		return MatchResult{}, fmt.Errorf("create valhalla request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := v.client.Do(httpReq)
	if err != nil {
		return MatchResult{}, fmt.Errorf("send valhalla request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return MatchResult{}, fmt.Errorf("valhalla HTTP error: status=%d", resp.StatusCode)
	}
	var body struct {
		Trip struct {
			Summary struct {
				Length float64 `json:"length"`
				Time   float64 `json:"time"`
			} `json:"summary"`
			Legs []struct {
				Shape     string `json:"shape"`
				Maneuvers []struct {
					StreetNames []string `json:"street_names"`
				} `json:"maneuvers"`
			} `json:"legs"`
		} `json:"trip"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return MatchResult{}, fmt.Errorf("decode valhalla response: %w", err)
	}
	result := MatchResult{DistanceM: body.Trip.Summary.Length * 1000, DurationS: body.Trip.Summary.Time}
	if body.Error != "" {
		result.Warnings = append(result.Warnings, MatchWarning{Code: "valhalla_error", Message: body.Error})
	}
	for _, leg := range body.Trip.Legs {
		if leg.Shape != "" {
			result.Warnings = append(result.Warnings, MatchWarning{Code: "shape_present", Message: "encoded polyline returned"})
		}
		for _, m := range leg.Maneuvers {
			for _, s := range m.StreetNames {
				if strings.TrimSpace(s) != "" {
					result.MatchedEdges = append(result.MatchedEdges, MatchedEdge{Name: s})
				}
			}
		}
	}
	return result, nil
}

func buildShape(points []MatchPoint) []map[string]any {
	shape := make([]map[string]any, 0, len(points))
	for _, p := range points {
		item := map[string]any{"lat": p.Lat, "lon": p.Lng}
		if !p.Time.IsZero() {
			item["time"] = p.Time.Unix()
		}
		shape = append(shape, item)
	}
	return shape
}
