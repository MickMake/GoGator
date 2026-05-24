package sitematch

type NoopSiteMatcher struct{}

func (NoopSiteMatcher) Match(req SiteMatchRequest) (SiteMatchResult, error) {
	_ = req
	return SiteMatchResult{}, nil
}
