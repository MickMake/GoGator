package gps

import (
	"fmt"
	"math"
	"strings"

	"gogator/internal/config"
	"gogator/internal/sites"
)

func BuildTrips(points []RawPoint, cfg config.Config, siteList []sites.Site) (valid []Trip, jitter []Trip) {
	if len(points) == 0 {
		return nil, nil
	}
	points = filterStationaryTeleports(points, cfg, siteList)
	if len(points) == 0 {
		return nil, nil
	}
	inTrip := false
	firstMoving := 0
	lastMoving := 0

	finish := func(firstMove, lastMove, stopConfirm int) {
		if lastMove < firstMove {
			return
		}

		travelPts := points[firstMove : lastMove+1]

		// Departure should be matched from the stationary cluster immediately
		// before movement, not from the first moving point.
		depPts := clusterStationaryRunBefore(points, firstMove)

		depMatchPts := depPts
		if len(depPts) > 0 {
			depStart := firstMove - len(depPts)
			if depStart > 0 {
				depMatchPts = includePreviousPointForSilentStopGap(points, depStart, depPts, cfg)
			}
		}

		// Destination should be matched from the first stable stationary cluster
		// after movement, not the last moving/roadside point. This fixes cases
		// where the car has reached home but the final moving point is still
		// 100m+ away on the road.
		firstStationary := lastMove + 1
		var destPts []RawPoint
		if firstStationary < len(points) {
			// Use the full stationary dwell run after movement so site-specific
			// Min Destination Minutes can distinguish a real stop from crawling
			// through a site radius in traffic. If the tracker went silent after
			// entering a site and woke up stationary in the same area, use the
			// pre-gap point as the arrival time rather than treating the silent
			// parked interval as travel time.
			if isSilentStopGapStart(points, firstStationary, cfg) {
				destPts = points[firstStationary-1 : firstStationary]
			} else {
				destPts = clusterStationaryRunAfter(points, firstStationary)
			}
			if len(destPts) == 0 && stopConfirm >= firstStationary && stopConfirm < len(points) {
				destPts = points[firstStationary : stopConfirm+1]
			}
		}

		destMatchPts := destPts
		if len(destPts) > 0 && firstStationary > 0 {
			destMatchPts = includePreviousPointForSilentStopGap(points, firstStationary, destPts, cfg)
		}

		tr := makeTrip(len(valid)+len(jitter)+1, travelPts, depPts, destPts, depMatchPts, destMatchPts, cfg, siteList)
		if tr.Jitter {
			jitter = append(jitter, tr)
		} else {
			valid = append(valid, tr)
		}
	}

	for i, p := range points {
		if p.Moving {
			if !inTrip {
				inTrip = true
				firstMoving = i
			}
			lastMoving = i
			continue
		}
		if inTrip {
			stationarySecs := p.Time.Sub(points[lastMoving].Time).Seconds()
			if stationarySecs >= float64(cfg.Trip.MinStopDurationSeconds) {
				finish(firstMoving, lastMoving, i)
				inTrip = false
			}
		}
	}
	if inTrip {
		finish(firstMoving, len(points)-1, len(points)-1)
	}

	valid, jitter = suppressSameSiteMicroTrips(valid, jitter, cfg)
	valid = repairContinuity(valid, cfg)

	for i := range valid {
		valid[i].Index = i + 1
		if valid[i].ContinuityStatus == "" {
			valid[i].ContinuityStatus = "CONTINUITY_OK"
		}
	}
	for i := range jitter {
		jitter[i].Index = i + 1
		if jitter[i].ContinuityStatus == "" {
			jitter[i].ContinuityStatus = "CONTINUITY_OK"
		}
	}
	for i := range valid {
		if i+1 < len(valid) {
			valid[i].SiteDurationHours = valid[i+1].Start.Sub(valid[i].End).Hours()
		} else {
			valid[i].SiteDurationHours = points[len(points)-1].Time.Sub(valid[i].End).Hours()
		}
	}
	return valid, jitter
}

func filterStationaryTeleports(points []RawPoint, cfg config.Config, siteList []sites.Site) []RawPoint {
	if !cfg.Trip.StationaryTeleportGuardEnabled || len(points) == 0 || len(siteList) == 0 {
		return points
	}
	out := make([]RawPoint, 0, len(points))
	lastKnownIdx := -1
	lastKnownSite := ""
	lastKnownLat := 0.0
	lastKnownLng := 0.0
	lastKnownOdo := 0.0
	lastKnownHasOdo := false

	for _, p := range points {
		name, _, _, matched := sites.Match(siteList, p.Lat, p.Lng, cfg.Site.UnknownSiteLabel)
		odo, hasOdo := Num(p, "io14")

		if isStationaryTeleportCandidate(p, matched, lastKnownIdx >= 0, hasOdo, lastKnownHasOdo, odo, lastKnownOdo, lastKnownLat, lastKnownLng, cfg) {
			p.Flags = append(p.Flags, "stationary_teleport_filtered")
			continue
		}

		out = append(out, p)
		if matched && (p.Stationary || !p.Moving) {
			lastKnownIdx = len(out) - 1
			lastKnownSite = name
			_ = lastKnownSite
			lastKnownLat = p.Lat
			lastKnownLng = p.Lng
			lastKnownOdo = odo
			lastKnownHasOdo = hasOdo
		}
	}
	return out
}

func isStationaryTeleportCandidate(p RawPoint, matched, hasLast bool, hasOdo, lastHasOdo bool, odo, lastOdo, lastLat, lastLng float64, cfg config.Config) bool {
	if matched || !hasLast {
		return false
	}
	if !(p.Stationary || !p.Moving) {
		return false
	}
	if p.SpeedKPH >= cfg.Trip.MovingSpeedKPH {
		return false
	}
	if io24, ok := Num(p, "io24"); ok && io24 != 0 {
		return false
	}
	if io251, ok := Num(p, "io251"); ok && io251 != 1 {
		// io251 is not mandatory, but if it exists and says not idling, be conservative.
		return false
	}
	if cfg.Trip.StationaryTeleportRequiresOdometer {
		if !hasOdo || !lastHasOdo || math.Abs(odo-lastOdo) > 1 {
			return false
		}
	}
	jumpM := HaversineM(lastLat, lastLng, p.Lat, p.Lng)
	return jumpM >= cfg.Trip.StationaryTeleportMinJumpM
}

func suppressSameSiteMicroTrips(valid []Trip, jitter []Trip, cfg config.Config) ([]Trip, []Trip) {
	if len(valid) == 0 {
		return valid, jitter
	}
	out := make([]Trip, 0, len(valid))
	for _, tr := range valid {
		if len(out) > 0 && isSameSiteMicroTrip(out[len(out)-1], tr, cfg) {
			tr.Jitter = true
			tr.Flags = append(tr.Flags, "same_site_micro_trip_suppressed")
			jitter = append(jitter, tr)
			continue
		}
		out = append(out, tr)
	}
	return out, jitter
}

func isSameSiteMicroTrip(prev, cur Trip, cfg config.Config) bool {
	if cur.DistanceKM > cfg.Trip.SameSiteMicroTripMaxKM {
		return false
	}
	if cur.DurationHours*60 > cfg.Trip.SameSiteMicroTripMaxMinutes {
		return false
	}
	knownDest := cur.DestinationSite != "" && cur.DestinationSite != cfg.Site.UnknownSiteLabel
	if !knownDest {
		return false
	}
	if prev.DestinationSite != cur.DestinationSite {
		return false
	}
	if cur.DepartureSite == cur.DestinationSite {
		return true
	}
	if cur.DepartureSite == cfg.Site.UnknownSiteLabel {
		return HaversineM(prev.DestLat, prev.DestLng, cur.DepartLat, cur.DepartLng) <= cfg.Trip.SameSiteGuardRadiusM
	}
	return false
}

func repairContinuity(valid []Trip, cfg config.Config) []Trip {
	if !cfg.Site.ContinuityRepairEnabled || len(valid) < 2 {
		return valid
	}
	maxM := cfg.Site.ContinuityMatchMaxMetres
	if maxM <= 0 {
		maxM = 75
	}
	unknown := cfg.Site.UnknownSiteLabel
	for i := 0; i+1 < len(valid); i++ {
		prev := &valid[i]
		cur := &valid[i+1]
		d := HaversineM(prev.DestLat, prev.DestLng, cur.DepartLat, cur.DepartLng)
		if d > maxM {
			if prev.DestinationSite != cur.DepartureSite {
				cur.ContinuityStatus = "OOPS_GPS_GAP"
				cur.Flags = append(cur.Flags, "continuity_gps_gap")
			}
			continue
		}
		prevKnown := prev.DestinationSite != "" && prev.DestinationSite != unknown
		curKnown := cur.DepartureSite != "" && cur.DepartureSite != unknown
		switch {
		case !prevKnown && curKnown:
			prev.DestinationSite = cur.DepartureSite
			prev.DestinationAddress = cur.DepartureAddress
			prev.ContinuityStatus = "CONTINUITY_REPAIRED"
			prev.Flags = append(prev.Flags, "continuity_destination_repaired")
		case prevKnown && !curKnown:
			cur.DepartureSite = prev.DestinationSite
			cur.DepartureAddress = prev.DestinationAddress
			cur.ContinuityStatus = "CONTINUITY_REPAIRED"
			cur.Flags = append(cur.Flags, "continuity_departure_repaired")
		case prevKnown && curKnown && prev.DestinationSite != cur.DepartureSite:
			cur.ContinuityStatus = "OOPS_KNOWN_TO_KNOWN_CONFLICT"
			cur.Flags = append(cur.Flags, "continuity_known_site_conflict")
		case !prevKnown && !curKnown:
			cur.ContinuityStatus = "OOPS_UNRESOLVED_CHECK"
			cur.Flags = append(cur.Flags, "continuity_unresolved_check")
		}
	}
	return valid
}

func makeTrip(index int, travelPts, depPts, destPts, depMatchPts, destMatchPts []RawPoint, cfg config.Config, siteList []sites.Site) Trip {
	if len(travelPts) == 0 {
		return Trip{}
	}
	tr := Trip{
		Index:            index,
		RawStartRow:      travelPts[0].RawRow,
		RawEndRow:        travelPts[len(travelPts)-1].RawRow,
		RawPoints:        len(travelPts),
		Start:            travelPts[0].Time,
		End:              travelPts[len(travelPts)-1].Time,
		ContinuityStatus: "CONTINUITY_OK",
		Points:           travelPts,
	}
	if len(destPts) > 0 {
		tr.RawEndRow = destPts[0].RawRow
		tr.End = destPts[0].Time
	}

	if len(depPts) > 0 {
		tr.DepartLat, tr.DepartLng = stableCoordLimited(depPts, false, 20)
	} else {
		tr.DepartLat, tr.DepartLng = travelPts[0].Lat, travelPts[0].Lng
	}
	if len(destPts) > 0 {
		tr.DestLat, tr.DestLng = stableCoordLimited(destPts, true, 20)
	} else {
		tr.DestLat, tr.DestLng = travelPts[len(travelPts)-1].Lat, travelPts[len(travelPts)-1].Lng
	}

	var distM float64
	var top float64
	flags := map[string]bool{}
	for i, p := range travelPts {
		if i > 0 {
			distM += HaversineM(travelPts[i-1].Lat, travelPts[i-1].Lng, p.Lat, p.Lng)
		}
		if p.SpeedKPH > top {
			top = p.SpeedKPH
		}
		for _, f := range p.Flags {
			flags[f] = true
		}
	}
	durH := tr.End.Sub(tr.Start).Hours()
	tr.DistanceKM = distM / 1000
	tr.DurationHours = durH
	tr.TopSpeedKPH = top
	if durH > 0 {
		tr.AverageSpeedKPH = tr.DistanceKM / durH
	}
	for f := range flags {
		tr.Flags = append(tr.Flags, f)
	}

	tr.DepartureSite, tr.DepartureAddress, _, _ = matchSiteWithDwell(siteList, depMatchPts, tr.DepartLat, tr.DepartLng, cfg)
	tr.DestinationSite, tr.DestinationAddress, _, _ = matchSiteWithDwell(siteList, destMatchPts, tr.DestLat, tr.DestLng, cfg)

	if tr.DistanceKM*1000 < cfg.Trip.MinTripDistanceM || tr.End.Sub(tr.Start).Seconds() < float64(cfg.Trip.MinTripDurationSeconds) {
		tr.Jitter = true
		tr.Flags = append(tr.Flags, "below_minimum_trip")
	}
	if tr.DepartureSite == tr.DestinationSite && tr.DepartureSite != cfg.Site.UnknownSiteLabel && tr.DistanceKM*1000 < cfg.Trip.SameSiteJitterRadiusM*2 {
		tr.Jitter = true
		tr.Flags = append(tr.Flags, "same_site_short_movement")
	}
	return tr
}

func isSilentStopGapStart(points []RawPoint, firstStationary int, cfg config.Config) bool {
	if !cfg.Site.InferSilentStopGaps || firstStationary <= 0 || firstStationary >= len(points) {
		return false
	}
	gapSecs := points[firstStationary].Time.Sub(points[firstStationary-1].Time).Seconds()
	if gapSecs < cfg.Site.SilentStopMinGapMinutes*60 {
		return false
	}
	return isStationaryDwellPoint(points[firstStationary], cfg)
}

func includePreviousPointForSilentStopGap(points []RawPoint, firstStationary int, destPts []RawPoint, cfg config.Config) []RawPoint {
	if !cfg.Site.InferSilentStopGaps || firstStationary <= 0 || firstStationary >= len(points) || len(destPts) == 0 {
		return destPts
	}
	gapSecs := points[firstStationary].Time.Sub(points[firstStationary-1].Time).Seconds()
	if gapSecs < cfg.Site.SilentStopMinGapMinutes*60 {
		return destPts
	}
	if !isStationaryDwellPoint(points[firstStationary], cfg) {
		return destPts
	}
	if len(destPts) > 0 && destPts[0].RawRow == points[firstStationary-1].RawRow {
		out := make([]RawPoint, 0, len(destPts)+1)
		out = append(out, destPts[0])
		out = append(out, points[firstStationary])
		out = append(out, destPts[1:]...)
		return out
	}
	out := make([]RawPoint, 0, len(destPts)+1)
	out = append(out, points[firstStationary-1])
	out = append(out, destPts...)
	return out
}

func matchSiteWithDwell(siteList []sites.Site, pts []RawPoint, lat, lng float64, cfg config.Config) (name, address string, distM float64, matched bool) {
	best := sites.Site{}
	bestDist := 1e18
	for _, s := range siteList {
		d := HaversineM(lat, lng, s.Lat, s.Lng)
		if d <= s.RadiusM && d < bestDist {
			best = s
			bestDist = d
		}
	}
	if bestDist >= 1e18 {
		return cfg.Site.UnknownSiteLabel, "", 0, false
	}
	minMinutes := best.MinDestinationMinutes
	if minMinutes <= 0 {
		minMinutes = cfg.Site.DefaultMinDestinationMinutes
	}
	if len(pts) == 0 || minMinutes <= 0 {
		return best.Name, best.Address, bestDist, true
	}
	evidence := slidingDwellEvidence(best, pts, cfg)
	if evidence.insideSecs >= minMinutes*60 && evidence.insideRatio >= cfg.Site.DwellRequiredInsideRatio && evidence.stationaryRatio >= cfg.Site.DwellRequiredStationaryRatio {
		return best.Name, best.Address, bestDist, true
	}
	return cfg.Site.UnknownSiteLabel, "", bestDist, false
}

type dwellEvidence struct {
	insideSecs      float64
	totalSecs       float64
	stationarySecs  float64
	insideRatio     float64
	stationaryRatio float64
}

func slidingDwellEvidence(site sites.Site, pts []RawPoint, cfg config.Config) dwellEvidence {
	if len(pts) == 0 {
		return dwellEvidence{}
	}
	windowSecs := cfg.Site.DwellWindowMinutes * 60
	if windowSecs <= 0 {
		windowSecs = 180 * 60
	}
	maxGapSecs := cfg.Site.DwellMaxSampleGapMinutes * 60
	if maxGapSecs <= 0 {
		maxGapSecs = 90 * 60
	}
	best := dwellEvidence{}
	for start := 0; start < len(pts); start++ {
		cur := dwellEvidence{}
		for i := start; i+1 < len(pts); i++ {
			if pts[i+1].Time.Sub(pts[start].Time).Seconds() > windowSecs {
				break
			}
			secs := pts[i+1].Time.Sub(pts[i].Time).Seconds()
			if secs <= 0 {
				continue
			}
			if secs > maxGapSecs {
				break
			}
			cur.totalSecs += secs
			inside := HaversineM(pts[i].Lat, pts[i].Lng, site.Lat, site.Lng) <= site.RadiusM
			if inside {
				cur.insideSecs += secs
				if isStationaryDwellPoint(pts[i], cfg) || isSilentStopGapAt(site, pts, i, secs, cfg) {
					cur.stationarySecs += secs
				}
			}
		}
		cur.finish()
		if cur.insideSecs > best.insideSecs || (cur.insideSecs == best.insideSecs && cur.stationaryRatio > best.stationaryRatio) {
			best = cur
		}
	}
	return best
}

func (d *dwellEvidence) finish() {
	if d.totalSecs > 0 {
		d.insideRatio = d.insideSecs / d.totalSecs
	}
	if d.insideSecs > 0 {
		d.stationaryRatio = d.stationarySecs / d.insideSecs
	}
}

func isSilentStopGapAt(site sites.Site, pts []RawPoint, i int, secs float64, cfg config.Config) bool {
	if !cfg.Site.InferSilentStopGaps || i+1 >= len(pts) {
		return false
	}
	if secs < cfg.Site.SilentStopMinGapMinutes*60 {
		return false
	}
	next := pts[i+1]
	if HaversineM(next.Lat, next.Lng, site.Lat, site.Lng) > site.RadiusM {
		return false
	}
	return isStationaryDwellPoint(next, cfg)
}

func isStationaryDwellPoint(p RawPoint, cfg config.Config) bool {
	if p.SpeedKPH >= cfg.Trip.MovingSpeedKPH {
		return false
	}
	if io24, ok := Num(p, "io24"); ok && io24 == 1 {
		return false
	}
	odoNow, hasOdo := Num(p, "io14")
	_ = odoNow
	_ = hasOdo
	return p.Stationary || !p.Moving
}

func clusterStationaryRunAfter(points []RawPoint, idx int) []RawPoint {
	if idx < 0 || idx >= len(points) {
		return nil
	}
	end := idx
	for end < len(points) {
		p := points[end]
		if p.Moving && !p.Stationary && p.SpeedKPH >= 1 {
			break
		}
		end++
	}
	return points[idx:end]
}

func clusterStationaryRunBefore(points []RawPoint, idx int) []RawPoint {
	if idx <= 0 || idx > len(points) {
		return nil
	}
	start := idx - 1
	for start >= 0 {
		p := points[start]
		if p.Moving && !p.Stationary && p.SpeedKPH >= 1 {
			break
		}
		start--
	}
	return points[start+1 : idx]
}

func clusterBefore(points []RawPoint, idx, limit int) []RawPoint {
	start := idx - limit
	if start < 0 {
		start = 0
	}
	out := points[start:idx]
	// Prefer stationary points when present.
	var stat []RawPoint
	for _, p := range out {
		if p.Stationary || !p.Moving {
			stat = append(stat, p)
		}
	}
	if len(stat) > 0 {
		return stat
	}
	return out
}

func clusterAfter(points []RawPoint, idx, limit int) []RawPoint {
	end := idx + limit
	if end > len(points) {
		end = len(points)
	}
	out := points[idx:end]
	var stat []RawPoint
	for _, p := range out {
		if p.Stationary || !p.Moving {
			stat = append(stat, p)
		}
	}
	if len(stat) > 0 {
		return stat
	}
	return out
}

func stableCoord(pts []RawPoint, start bool) (float64, float64) {
	return stableCoordLimited(pts, start, 5)
}

func stableCoordLimited(pts []RawPoint, start bool, limit int) (float64, float64) {
	if len(pts) == 0 {
		return 0, 0
	}
	n := min(limit, len(pts))
	xs := make([]float64, 0, n)
	ys := make([]float64, 0, n)
	if start {
		for i := 0; i < n; i++ {
			xs = append(xs, pts[i].Lat)
			ys = append(ys, pts[i].Lng)
		}
	} else {
		for i := len(pts) - n; i < len(pts); i++ {
			xs = append(xs, pts[i].Lat)
			ys = append(ys, pts[i].Lng)
		}
	}
	return median(xs), median(ys)
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	a := append([]float64(nil), v...)
	for i := range a {
		for j := i + 1; j < len(a); j++ {
			if a[j] < a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
	mid := len(a) / 2
	if len(a)%2 == 1 {
		return a[mid]
	}
	return (a[mid-1] + a[mid]) / 2
}

func TripParamSummary(tr Trip) map[string]string {
	out := map[string]string{}
	countEq := func(key string, val float64) int {
		c := 0
		for _, p := range tr.Points {
			if v, ok := Num(p, key); ok && v == val {
				c++
			}
		}
		return c
	}
	nums := func(key string) []float64 {
		var a []float64
		for _, p := range tr.Points {
			if v, ok := Num(p, key); ok {
				a = append(a, v)
			}
		}
		return a
	}
	vals := func(key string) []string {
		seen := map[string]bool{}
		var a []string
		for _, p := range tr.Points {
			if v := p.Params[key]; v != "" && !seen[v] {
				seen[v] = true
				a = append(a, v)
			}
		}
		return a
	}
	maxAbs := func(a []float64) float64 {
		m := 0.0
		for _, v := range a {
			if math.Abs(v) > m {
				m = math.Abs(v)
			}
		}
		return m
	}
	avg := func(a []float64) float64 {
		if len(a) == 0 {
			return 0
		}
		s := 0.0
		for _, v := range a {
			s += v
		}
		return s / float64(len(a))
	}
	maxv := func(a []float64) float64 {
		if len(a) == 0 {
			return 0
		}
		m := a[0]
		for _, v := range a {
			if v > m {
				m = v
			}
		}
		return m
	}
	mags := []float64{}
	for _, p := range tr.Points {
		if m, ok := AccelMagnitude(p.ParamNums); ok {
			mags = append(mags, m)
		}
	}

	out["Ignition On Samples"] = itoa(countEq("io1", 1))
	out["Ignition Off Samples"] = itoa(countEq("io1", 0))
	out["Ignition Values"] = strings.Join(vals("io1"), ", ")
	if len(tr.Points) > 0 {
		out["Ignition Start"] = tr.Points[0].Params["io1"]
		out["Ignition End"] = tr.Points[len(tr.Points)-1].Params["io1"]
	}
	pdops := nums("pdop")
	poor := 0
	ideal := 0
	for _, v := range pdops {
		if v > 5 {
			poor++
		}
		if v <= 2 {
			ideal++
		}
	}
	out["PDOP Poor Samples"] = itoa(poor)
	out["PDOP Ideal Samples"] = itoa(ideal)
	out["GPS Level Max"] = ftoa(maxv(nums("gpslev")), 2)
	out["g0 Max Abs"] = ftoa(maxAbs(nums("g0")), 2)
	out["g1 Max Abs"] = ftoa(maxAbs(nums("g1")), 2)
	out["g2 Max Abs"] = ftoa(maxAbs(nums("g2")), 2)
	out["Accel Magnitude Max"] = ftoa(maxv(mags), 2)
	out["Accel Magnitude Avg"] = ftoa(avg(mags), 2)
	nz := 0
	for _, v := range mags {
		if v != 0 {
			nz++
		}
	}
	out["Accel Nonzero Samples"] = itoa(nz)
	out["Crash Detected Samples"] = itoa(countEq("io247", 1))
	bev := vals("io253")
	filtered := []string{}
	for _, v := range bev {
		if v != "0" && v != "0.0" {
			filtered = append(filtered, v)
		}
	}
	out["Behaviour Event Samples"] = itoa(len(filtered))
	out["Behaviour Event Values"] = strings.Join(filtered, ", ")
	out["Panic Trigger Samples"] = itoa(countEq("io3", 1))
	out["Power Cut Samples io252"] = itoa(countEq("io252", 1))
	out["Sleep Mode Values"] = strings.Join(vals("io200"), ", ")
	out["Network State Values io381"] = strings.Join(vals("io381"), ", ")
	out["Trip Status Values io254"] = strings.Join(vals("io254"), ", ")
	if len(tr.Points) > 0 {
		out["Odometer Raw Start io14"] = tr.Points[0].Params["io14"]
		out["Odometer Raw End io14"] = tr.Points[len(tr.Points)-1].Params["io14"]
		out["SIM ICCID Raw io11"] = tr.Points[0].Params["io11"]
	}
	return out
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }
func ftoa(f float64, places int) string {
	if f == 0 {
		return ""
	}
	return fmt.Sprintf("%.*f", places, f)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
