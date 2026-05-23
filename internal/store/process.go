package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
	processroutes "gogator/internal/routes"
	processsites "gogator/internal/sites"
)

func LoadGPSPointsForProcess() ([]gps.RawPoint, error) {
	db, err := Open(DefaultPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT COALESCE(first_source_file,''),COALESCE(first_raw_row,0),raw_dt,normalised_time,lat,lng,altitude,angle,speed_kph,COALESCE(params_raw,''),COALESCE(params_json,'') FROM gps_points ORDER BY normalised_time,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []gps.RawPoint
	for rows.Next() {
		var sourceFile, rawDT, normalisedTime, paramsRaw, paramsJSON string
		var rawRow int
		var lat, lng float64
		var altitude, angle, speed sql.NullFloat64
		if err := rows.Scan(&sourceFile, &rawRow, &rawDT, &normalisedTime, &lat, &lng, &altitude, &angle, &speed, &paramsRaw, &paramsJSON); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339Nano, normalisedTime)
		if err != nil {
			return nil, fmt.Errorf("gps_points normalised_time %q: %w", normalisedTime, err)
		}
		params, err := decodeGPSParams(paramsJSON)
		if err != nil {
			return nil, err
		}
		if sourceFile == "" {
			sourceFile = "gogator.sqlite"
		}
		out = append(out, gps.RawPoint{
			SourceFile: sourceFile,
			RawRow:     rawRow,
			RawDT:      rawDT,
			Time:       t,
			Lat:        lat,
			Lng:        lng,
			Altitude:   nullFloat(altitude),
			Angle:      nullFloat(angle),
			SpeedKPH:   nullFloat(speed),
			ParamsRaw:  paramsRaw,
			Params:     params,
			ParamNums:  gps.NumericParams(params),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	gps.RecalculatePointDeltas(out)
	return out, nil
}

func LoadSitesForProcess(cfg config.Config) ([]processsites.Site, error) {
	db, err := Open(DefaultPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT name,COALESCE(address,''),lat,lng,range_m,min_destination_minutes,COALESCE(type,''),important FROM sites ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []processsites.Site
	for rows.Next() {
		var name, address, siteType string
		var lat, lng float64
		var radius, minDestination sql.NullFloat64
		var importantInt int
		if err := rows.Scan(&name, &address, &lat, &lng, &radius, &minDestination, &siteType, &importantInt); err != nil {
			return nil, err
		}
		radiusM := nullFloat(radius)
		if radiusM <= 0 {
			radiusM = cfg.Site.DefaultRadiusM
		}
		if radiusM <= 0 {
			radiusM = 100
		}
		minDest := nullFloat(minDestination)
		if minDest < 0 {
			minDest = cfg.Site.DefaultMinDestinationMinutes
		}
		out = append(out, processsites.Site{
			Name:                  name,
			Address:               address,
			Lat:                   lat,
			Lng:                   lng,
			RadiusM:               radiusM,
			MinDestinationMinutes: minDest,
			SiteType:              siteType,
			Important:             importantInt != 0,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func LoadRoutesForProcess() ([]processroutes.Route, error) {
	db, err := Open(DefaultPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT fs.name,ts.name,r.name,COALESCE(r.confidence,''),COALESCE(r.notes,''),r.expected_distance_min_km,r.expected_distance_max_km,r.expected_duration_min_min,r.expected_duration_max_min FROM routes r JOIN sites fs ON fs.id=r.from_site_id JOIN sites ts ON ts.id=r.to_site_id ORDER BY fs.name,ts.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []processroutes.Route
	for rows.Next() {
		var from, to, name, confidence, notes string
		var dMin, dMax, tMin, tMax sql.NullFloat64
		if err := rows.Scan(&from, &to, &name, &confidence, &notes, &dMin, &dMax, &tMin, &tMax); err != nil {
			return nil, err
		}
		if strings.TrimSpace(name) == "" {
			name = from + " to " + to
		}
		out = append(out, processroutes.Route{
			Name:            name,
			FromSite:        from,
			ToSite:          to,
			DistanceMinKM:   nullFloat(dMin),
			DistanceMaxKM:   nullFloat(dMax),
			DurationMinMin:  nullFloat(tMin),
			DurationMaxMin:  nullFloat(tMax),
			ConfidenceBoost: confidence,
			Notes:           notes,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func nullFloat(v sql.NullFloat64) float64 {
	if !v.Valid {
		return 0
	}
	return v.Float64
}
