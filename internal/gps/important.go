package gps

import (
	"gogator/internal/config"
	"gogator/internal/sites"
)

// CollapseToImportantSites turns the raw trip chain into the business-useful
// ledger: travel between sites explicitly marked Important=yes in sites.csv.
// Unimportant sites and CHECK rows are treated as pass-through noise and are
// merged into the next important destination where possible.
func CollapseToImportantSites(valid []Trip, jitter []Trip, cfg config.Config, siteList []sites.Site) ([]Trip, []Trip) {
	if len(valid) == 0 || len(siteList) == 0 {
		return valid, jitter
	}

	for i := range valid {
		repairEndpointToImportant(&valid[i], cfg, siteList)
	}

	out := make([]Trip, 0, len(valid))
	var pending *Trip

	suppressPending := func(flag string) {
		if pending == nil {
			return
		}
		pending.Jitter = true
		pending.Flags = append(pending.Flags, flag)
		pending.ContinuityStatus = "UNIMPORTANT_SUPPRESSED"
		jitter = append(jitter, *pending)
		pending = nil
	}

	emitMerged := func(tr Trip) {
		if tr.DepartureSite == tr.DestinationSite && isImportantName(tr.DepartureSite, cfg, siteList) {
			tr.Jitter = true
			tr.Flags = append(tr.Flags, "same_important_site_loop_suppressed")
			tr.ContinuityStatus = "IMPORTANT_LOOP_SUPPRESSED"
			jitter = append(jitter, tr)
			return
		}
		out = append(out, tr)
	}

	for _, tr := range valid {
		depImp := isImportantName(tr.DepartureSite, cfg, siteList)
		destImp := isImportantName(tr.DestinationSite, cfg, siteList)

		switch {
		case depImp && destImp:
			suppressPending("unimportant_chain_superseded")
			emitMerged(tr)

		case depImp && !destImp:
			if pending == nil {
				p := tr
				p.Flags = append(p.Flags, "important_chain_pending")
				pending = &p
			} else {
				mergeTrip(pending, tr)
			}

		case !depImp && destImp:
			if pending != nil {
				mergeTrip(pending, tr)
				pending.DestinationSite = tr.DestinationSite
				pending.DestinationAddress = tr.DestinationAddress
				pending.DestLat = tr.DestLat
				pending.DestLng = tr.DestLng
				pending.Flags = append(pending.Flags, "unimportant_chain_collapsed")
				pending.ContinuityStatus = "IMPORTANT_CHAIN_COLLAPSED"
				emitMerged(*pending)
				pending = nil
			} else {
				tr.Flags = append(tr.Flags, "starts_from_unimportant_site")
				emitMerged(tr)
			}

		default:
			if pending != nil {
				mergeTrip(pending, tr)
				pending.Flags = append(pending.Flags, "unimportant_chain_absorbed")
			} else {
				tr.Jitter = true
				tr.Flags = append(tr.Flags, "unimportant_trip_suppressed")
				tr.ContinuityStatus = "UNIMPORTANT_SUPPRESSED"
				jitter = append(jitter, tr)
			}
		}
	}
	suppressPending("unimportant_tail_suppressed")

	reindexAndDurations(out)
	for i := range jitter {
		jitter[i].Index = i + 1
	}
	return out, jitter
}

func repairEndpointToImportant(tr *Trip, cfg config.Config, siteList []sites.Site) {
	if !isImportantName(tr.DepartureSite, cfg, siteList) {
		if s, ok := nearestImportantSite(siteList, tr.DepartLat, tr.DepartLng); ok {
			tr.DepartureSite = s.Name
			tr.DepartureAddress = s.Address
			tr.Flags = append(tr.Flags, "important_departure_repaired")
		}
	}
	if !isImportantName(tr.DestinationSite, cfg, siteList) {
		if s, ok := nearestImportantSite(siteList, tr.DestLat, tr.DestLng); ok {
			tr.DestinationSite = s.Name
			tr.DestinationAddress = s.Address
			tr.Flags = append(tr.Flags, "important_destination_repaired")
		}
	}
}

func nearestImportantSite(siteList []sites.Site, lat, lng float64) (sites.Site, bool) {
	best := sites.Site{}
	bestDist := 1e18
	for _, s := range siteList {
		if !s.Important {
			continue
		}
		d := HaversineM(lat, lng, s.Lat, s.Lng)
		if d <= s.RadiusM && d < bestDist {
			best = s
			bestDist = d
		}
	}
	return best, bestDist < 1e18
}

func isImportantName(name string, cfg config.Config, siteList []sites.Site) bool {
	return sites.IsImportant(siteList, name, cfg.Site.UnknownSiteLabel)
}

func mergeTrip(dst *Trip, src Trip) {
	dst.RawEndRow = src.RawEndRow
	dst.RawPoints += src.RawPoints
	dst.End = src.End
	dst.DestLat = src.DestLat
	dst.DestLng = src.DestLng
	dst.DestinationSite = src.DestinationSite
	dst.DestinationAddress = src.DestinationAddress
	dst.DistanceKM += src.DistanceKM
	if src.TopSpeedKPH > dst.TopSpeedKPH {
		dst.TopSpeedKPH = src.TopSpeedKPH
	}
	dst.Points = append(dst.Points, src.Points...)
	dst.Flags = append(dst.Flags, src.Flags...)
	dst.DurationHours = dst.End.Sub(dst.Start).Hours()
	if dst.DurationHours > 0 {
		dst.AverageSpeedKPH = dst.DistanceKM / dst.DurationHours
	}
}

func reindexAndDurations(trips []Trip) {
	for i := range trips {
		trips[i].Index = i + 1
		if trips[i].ContinuityStatus == "" {
			trips[i].ContinuityStatus = "CONTINUITY_OK"
		}
		if i+1 < len(trips) {
			trips[i].SiteDurationHours = trips[i+1].Start.Sub(trips[i].End).Hours()
		}
	}
}
