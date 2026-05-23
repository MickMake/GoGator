package engine

import (
	"math"

	"gogator/internal/gps"
)

type CandidateTripComparison struct {
	CandidateTripCount           int
	LegacyValidTripCount         int
	LegacyJitterTripCount        int
	ApproxMatchedTrips           int
	UnmatchedCandidateTrips      []int
	UnmatchedLegacyTrips         []int
	MajorTimeBoundaryDifferences int
	SiteDifferenceCount          int
}

func compareCandidateTrips(candidates CandidateTripEvidence, legacyValid, legacyJitter []gps.Trip) CandidateTripComparison {
	cmp := CandidateTripComparison{CandidateTripCount: len(candidates.Trips), LegacyValidTripCount: len(legacyValid), LegacyJitterTripCount: len(legacyJitter)}
	used := map[int]bool{}
	for ci, c := range candidates.Trips {
		best := -1
		for li, l := range legacyValid {
			if used[li] {
				continue
			}
			if approxMatch(c, l) {
				best = li
				break
			}
		}
		if best == -1 {
			cmp.UnmatchedCandidateTrips = append(cmp.UnmatchedCandidateTrips, ci)
			continue
		}
		cmp.ApproxMatchedTrips++
		used[best] = true
		l := legacyValid[best]
		if math.Abs(c.StartTime.Sub(l.Start).Minutes()) > 10 || math.Abs(c.EndTime.Sub(l.End).Minutes()) > 10 {
			cmp.MajorTimeBoundaryDifferences++
		}
		if (c.OriginLabel != "" && c.OriginLabel != l.DepartureSite) || (c.DestinationLabel != "" && c.DestinationLabel != l.DestinationSite) {
			cmp.SiteDifferenceCount++
		}
	}
	for i := range legacyValid {
		if !used[i] {
			cmp.UnmatchedLegacyTrips = append(cmp.UnmatchedLegacyTrips, i)
		}
	}
	return cmp
}

func approxMatch(c CandidateTrip, l gps.Trip) bool {
	if math.Abs(c.StartTime.Sub(l.Start).Minutes()) > 20 {
		return false
	}
	if math.Abs(c.EndTime.Sub(l.End).Minutes()) > 20 {
		return false
	}
	if c.OriginLabel != "" && l.DepartureSite != "" && c.OriginLabel != l.DepartureSite {
		return false
	}
	if c.DestinationLabel != "" && l.DestinationSite != "" && c.DestinationLabel != l.DestinationSite {
		return false
	}
	return true
}
