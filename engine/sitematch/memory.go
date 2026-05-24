package sitematch

import "math"

type MemorySite struct {
	Name    string
	Address string
	Lat     float64
	Lng     float64
}

type MemorySiteMatcher struct {
	sites []MemorySite
}

func NewMemorySiteMatcher(sites []MemorySite) *MemorySiteMatcher {
	copied := append([]MemorySite(nil), sites...)
	return &MemorySiteMatcher{sites: copied}
}

func (m *MemorySiteMatcher) Match(req SiteMatchRequest) (SiteMatchResult, error) {
	if m == nil || len(m.sites) == 0 {
		return SiteMatchResult{}, nil
	}
	radius := req.RadiusMeters
	if radius <= 0 {
		return SiteMatchResult{}, nil
	}
	result := SiteMatchResult{}
	for _, site := range m.sites {
		d := haversineMeters(req.Latitude, req.Longitude, site.Lat, site.Lng)
		if d > radius {
			continue
		}
		result.Candidates = append(result.Candidates, SiteMatchCandidate{SiteName: site.Name, SiteAddress: site.Address, SiteLatitude: site.Lat, SiteLongitude: site.Lng, DistanceMeters: d})
	}
	if len(result.Candidates) > 0 {
		result.Confidence.Score = 1
	}
	return result, nil
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000.0
	dLat := radians(lat2 - lat1)
	dLng := radians(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(radians(lat1))*math.Cos(radians(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func radians(v float64) float64 { return v * math.Pi / 180 }
