package engine

type CandidateTripComparison struct {
	CandidateTripCount           int
	LegacyValidTripCount         int
	LegacyJitterTripCount        int
	ApproxMatchedTrips           int
	UnmatchedCandidateTrips      []int
	UnmatchedLegacyTrips         []int
	MajorTimeBoundaryDifferences int
	SiteDifferenceCount          int
	ShadowSummary                ShadowSummary
}
