package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
)

type GPSImportResult struct {
	Files      int
	RawRows    int
	GPSPoints  int
	SourceRows int
}

var gpsExportCoreHeader = []string{"Raw DT", "Normalised Time", "Lat", "Lng", "Altitude", "Angle", "Speed KPH"}

func ImportGPS(paths []string, loc *time.Location, cfg config.Config) (GPSImportResult, error) {
	if len(paths) == 0 {
		return GPSImportResult{}, fmt.Errorf("missing GPS input file")
	}
	db, err := Open(DefaultPath)
	if err != nil {
		return GPSImportResult{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return GPSImportResult{}, err
	}
	defer tx.Rollback()

	var result GPSImportResult
	for _, path := range paths {
		points, err := gps.ReadRawCSV(path, loc, cfg)
		if err != nil {
			return GPSImportResult{}, err
		}
		result.Files++
		result.RawRows += len(points)
		for _, point := range points {
			pointID, inserted, err := insertGPSPointTx(tx, point)
			if err != nil {
				return GPSImportResult{}, err
			}
			if inserted {
				result.GPSPoints++
			}
			sourceInserted, err := insertGPSPointSourceTx(tx, pointID, point)
			if err != nil {
				return GPSImportResult{}, err
			}
			if sourceInserted {
				result.SourceRows++
			}
			if sourceInserted && !inserted {
				if err := markGPSPointSeenTx(tx, pointID, point); err != nil {
					return GPSImportResult{}, err
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return GPSImportResult{}, err
	}
	return result, nil
}

func ExportGPS(path string) error {
	db, err := Open(DefaultPath)
	if err != nil {
		return err
	}
	defer db.Close()

	paramKeys, err := gpsExportParamKeys(db)
	if err != nil {
		return err
	}

	rows, err := db.Query(`SELECT raw_dt,normalised_time,lat,lng,altitude,angle,speed_kph,COALESCE(params_json,'') FROM gps_points ORDER BY normalised_time,id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = '\t'
	header := append([]string{}, gpsExportCoreHeader...)
	header = append(header, paramKeys...)
	if err := w.Write(header); err != nil {
		return err
	}

	for rows.Next() {
		var rawDT, normalisedTime, paramsJSON string
		var lat, lng float64
		var altitude, angle, speed sql.NullFloat64
		if err := rows.Scan(&rawDT, &normalisedTime, &lat, &lng, &altitude, &angle, &speed, &paramsJSON); err != nil {
			return err
		}
		params, err := decodeGPSParams(paramsJSON)
		if err != nil {
			return err
		}
		rec := []string{
			rawDT,
			normalisedTime,
			trimFloat(lat),
			trimFloat(lng),
			trimNullFloat(altitude),
			trimNullFloat(angle),
			trimNullFloat(speed),
		}
		for _, key := range paramKeys {
			rec = append(rec, params[key])
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func gpsExportParamKeys(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT COALESCE(params_json,'') FROM gps_points`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var paramsJSON string
		if err := rows.Scan(&paramsJSON); err != nil {
			return nil, err
		}
		params, err := decodeGPSParams(paramsJSON)
		if err != nil {
			return nil, err
		}
		for key := range params {
			seen[key] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ordered := make([]string, 0, len(seen))
	for _, key := range gps.ParamOrder {
		ordered = append(ordered, key)
		delete(seen, key)
	}

	unknown := make([]string, 0, len(seen))
	for key := range seen {
		unknown = append(unknown, key)
	}
	sort.Strings(unknown)
	ordered = append(ordered, unknown...)
	return ordered, nil
}

func decodeGPSParams(paramsJSON string) (map[string]string, error) {
	params := map[string]string{}
	if strings.TrimSpace(paramsJSON) == "" {
		return params, nil
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, fmt.Errorf("decode gps params: %w", err)
	}
	return params, nil
}

func insertGPSPointTx(tx *sql.Tx, point gps.RawPoint) (int64, bool, error) {
	hash, paramsJSON, err := gpsPointHash(point)
	if err != nil {
		return 0, false, err
	}

	res, err := tx.Exec(`INSERT OR IGNORE INTO gps_points(point_hash,first_source_file,first_raw_row,last_source_file,last_raw_row,seen_count,raw_dt,normalised_time,lat,lng,altitude,angle,speed_kph,params_raw,params_json)
VALUES(?,?,?,?,?,1,?,?,?,?,?,?,?,?,?)`,
		hash,
		point.SourceFile,
		point.RawRow,
		point.SourceFile,
		point.RawRow,
		point.RawDT,
		point.Time.Format(time.RFC3339Nano),
		point.Lat,
		point.Lng,
		point.Altitude,
		point.Angle,
		point.SpeedKPH,
		point.ParamsRaw,
		paramsJSON,
	)
	if err != nil {
		return 0, false, err
	}
	insertedRows, _ := res.RowsAffected()
	inserted := insertedRows > 0

	var id int64
	if err := tx.QueryRow(`SELECT id FROM gps_points WHERE point_hash=?`, hash).Scan(&id); err != nil {
		return 0, false, err
	}
	return id, inserted, nil
}

func markGPSPointSeenTx(tx *sql.Tx, gpsPointID int64, point gps.RawPoint) error {
	_, err := tx.Exec(`UPDATE gps_points SET last_source_file=?,last_raw_row=?,seen_count=seen_count+1 WHERE id=?`, point.SourceFile, point.RawRow, gpsPointID)
	return err
}

func insertGPSPointSourceTx(tx *sql.Tx, gpsPointID int64, point gps.RawPoint) (bool, error) {
	res, err := tx.Exec(`INSERT OR IGNORE INTO gps_point_sources(gps_point_id,source_file,raw_row) VALUES(?,?,?)`, gpsPointID, point.SourceFile, point.RawRow)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func gpsPointHash(point gps.RawPoint) (string, string, error) {
	paramsJSON := ""
	if len(point.Params) > 0 {
		b, err := json.Marshal(point.Params)
		if err != nil {
			return "", "", err
		}
		paramsJSON = string(b)
	}
	parts := []string{
		strings.TrimSpace(point.RawDT),
		point.Time.Format(time.RFC3339Nano),
		floatKey(point.Lat),
		floatKey(point.Lng),
		floatKey(point.Altitude),
		floatKey(point.Angle),
		floatKey(point.SpeedKPH),
		strings.TrimSpace(point.ParamsRaw),
		paramsJSON,
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:]), paramsJSON, nil
}

func floatKey(v float64) string {
	return fmt.Sprintf("%.10g", v)
}
