package mapmatch

import "fmt"

type NoopMapMatcher struct{}

func (NoopMapMatcher) Match(req MatchRequest) (MatchResult, error) {
	if len(req.Points) == 0 {
		return MatchResult{}, nil
	}
	for i, p := range req.Points {
		if p.Lat < -90 || p.Lat > 90 || p.Lng < -180 || p.Lng > 180 {
			return MatchResult{}, fmt.Errorf("invalid point at index %d", i)
		}
	}
	return MatchResult{}, nil
}
