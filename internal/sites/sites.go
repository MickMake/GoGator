package sites

import (
	"bytes"
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"gogator/internal/config"
	"math"
)

type Site struct {
	Name                  string
	Address               string
	Lat                   float64
	Lng                   float64
	RadiusM               float64
	MinDestinationMinutes float64
	SiteType              string
	Important             bool
}

func Load(path string, cfg config.Config) ([]Site, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	delimiter := detectDelimiter(data)
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	r.Comma = delimiter
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	headerIdx := -1
	for i, row := range rows {
		if hasHeader(row) {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil, nil
	}
	header := rows[headerIdx]
	idx := map[string]int{}
	for i, h := range header {
		idx[norm(h)] = i
	}
	get := func(row []string, keys ...string) string {
		for _, k := range keys {
			if n, ok := idx[norm(k)]; ok && n < len(row) {
				return strings.TrimSpace(row[n])
			}
		}
		return ""
	}
	hasImportant := false
	if _, ok := idx[norm("Important")]; ok {
		hasImportant = true
	}
	var out []Site
	for _, row := range rows[headerIdx+1:] {
		name := get(row, "Site", "Customer Name", "Site Name", "Name")
		addr := get(row, "Real Address", "Address")
		gpsText := get(row, "GPS", "Coordinates", "Lat,Lng", "Latitude,Longitude")
		lat, lng, ok := parseGPS(gpsText)
		if !ok {
			lat, _ = strconv.ParseFloat(cleanNumber(get(row, "Latitude", "Lat")), 64)
			lng, _ = strconv.ParseFloat(cleanNumber(get(row, "Longitude", "Lng", "Lon")), 64)
			if lat == 0 && lng == 0 {
				continue
			}
		}
		radius := parseRadius(get(row, "Range", "Radius", "Match Radius Metres", "Match Radius M"), cfg.Site.DefaultRadiusM)
		if radius <= 0 {
			radius = 100
		}
		minDest := parseRadius(get(row, "Min Destination Minutes", "Min Stop Minutes", "Min Dwell Minutes", "Destination Min Dwell Minutes"), cfg.Site.DefaultMinDestinationMinutes)
		if minDest < 0 {
			minDest = cfg.Site.DefaultMinDestinationMinutes
		}
		siteType := get(row, "Type", "Site Type")
		important := true
		if hasImportant {
			important = parseImportant(get(row, "Important"))
		}
		if name == "" {
			name = addr
		}
		out = append(out, Site{Name: name, Address: addr, Lat: lat, Lng: lng, RadiusM: radius, MinDestinationMinutes: minDest, SiteType: siteType, Important: important})
	}
	return out, nil
}

func Match(all []Site, lat, lng float64, unknown string) (name, address string, distM float64, matched bool) {
	best := Site{}
	bestDist := 1e18
	for _, s := range all {
		d := haversineM(lat, lng, s.Lat, s.Lng)
		if d <= s.RadiusM && d < bestDist {
			best, bestDist = s, d
		}
	}
	if bestDist < 1e18 {
		return best.Name, best.Address, bestDist, true
	}
	return unknown, "", 0, false
}

func IsImportant(all []Site, name, unknown string) bool {
	if name == "" || name == unknown {
		return false
	}
	for _, s := range all {
		if s.Name == name {
			return s.Important
		}
	}
	// Backwards-compatible fallback for sites not found in metadata.
	return true
}

func detectDelimiter(data []byte) rune {
	// Prefer the delimiter visible in the header line. Counting the whole file is
	// unreliable because address fields and GPS coordinates contain many commas,
	// even when the actual export is TSV.
	for _, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
		if line == "" {
			continue
		}
		tabs := strings.Count(line, "\t")
		commas := strings.Count(line, ",")
		semis := strings.Count(line, ";")
		if tabs >= commas && tabs >= semis && tabs > 0 {
			return '\t'
		}
		if semis > commas && semis > 0 {
			return ';'
		}
		return ','
	}
	return ','
}

func hasHeader(row []string) bool {
	hasSite := false
	hasGPS := false
	for _, c := range row {
		n := norm(c)
		if n == "site" || n == "site name" || n == "customer name" || n == "name" {
			hasSite = true
		}
		if n == "gps" || n == "coordinates" || n == "latitude" || n == "lat" {
			hasGPS = true
		}
	}
	return hasSite && hasGPS
}
func norm(s string) string {
	s = strings.TrimPrefix(s, "\ufeff")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}
func cleanNumber(s string) string {
	s = strings.TrimSpace(strings.TrimPrefix(s, "\ufeff"))
	s = strings.TrimSuffix(s, "m")
	s = strings.TrimSuffix(s, "M")
	s = strings.TrimSpace(s)
	return s
}
func parseRadius(s string, fallback float64) float64 {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(cleanNumber(s), 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
func parseGPS(s string) (float64, float64, bool) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "\ufeff"))
	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	lat, e1 := strconv.ParseFloat(cleanNumber(parts[0]), 64)
	lng, e2 := strconv.ParseFloat(cleanNumber(parts[1]), 64)
	return lat, lng, e1 == nil && e2 == nil
}
func parseImportant(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "yes" || s == "y" || s == "true" || s == "1"
}

func haversineM(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0
	p1 := lat1 * math.Pi / 180
	p2 := lat2 * math.Pi / 180
	dp := (lat2 - lat1) * math.Pi / 180
	dl := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return 2 * R * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
