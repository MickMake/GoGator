package engine

import (
	"sort"
	"strings"

	"gogator/internal/gps"
	"gogator/internal/routes"
)

type LegacyAdapter struct{}

func (LegacyAdapter) SortAndRecalculate(points []gps.RawPoint) {
	sort.SliceStable(points, func(i, j int) bool {
		if points[i].Time.Equal(points[j].Time) {
			if points[i].SourceFile == points[j].SourceFile {
				return points[i].RawRow < points[j].RawRow
			}
			return points[i].SourceFile < points[j].SourceFile
		}
		return points[i].Time.Before(points[j].Time)
	})
	gps.RecalculatePointDeltas(points)
}
func (LegacyAdapter) Classify(points []gps.RawPoint, in Input) []gps.RawPoint {
	return gps.Classify(points, in.Config)
}
func (LegacyAdapter) BuildTrips(points []gps.RawPoint, in Input) ([]gps.Trip, []gps.Trip) {
	return gps.BuildTrips(points, in.Config, in.Sites)
}
func (LegacyAdapter) ApplyImportant(valid []gps.Trip, jitter []gps.Trip, in Input) ([]gps.Trip, []gps.Trip) {
	return gps.CollapseToImportantSites(valid, jitter, in.Config, in.Sites)
}
func (LegacyAdapter) ApplyRoutes(valid []gps.Trip, in Input) ([]gps.Trip, []routes.Observation, []routes.Anomaly) {
	return routes.Apply(valid, in.Routes, in.Config.Site.UnknownSiteLabel)
}
func (LegacyAdapter) SplitJitter(jitter []gps.Trip) (review []gps.Trip, sameSite []gps.Trip) {
	for _, t := range jitter {
		from := strings.TrimSpace(t.DepartureSite)
		to := strings.TrimSpace(t.DestinationSite)
		if from != "" && to != "" && strings.EqualFold(from, to) {
			sameSite = append(sameSite, t)
		} else {
			review = append(review, t)
		}
	}
	return review, sameSite
}
