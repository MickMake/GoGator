package engine

type Diagnostics struct {
	Points         []PointEvidence
	Motion         []MotionSample
	Stays          []Stay
	Visits         []Visit
	Excursions     []Excursion
	CandidateTrips []CandidateTrip
	TripComparison CandidateTripComparison
}

func (r Result) Diagnostics() Diagnostics {
	return Diagnostics{
		Points:         append([]PointEvidence(nil), r.Evidence.Points...),
		Motion:         append([]MotionSample(nil), r.Motion.Samples...),
		Stays:          append([]Stay(nil), r.Stays.Stays...),
		Visits:         append([]Visit(nil), r.Visits.Visits...),
		Excursions:     append([]Excursion(nil), r.Excursions.Excursions...),
		CandidateTrips: append([]CandidateTrip(nil), r.CandidateTrips.Trips...),
		TripComparison: r.TripComparison,
	}
}
