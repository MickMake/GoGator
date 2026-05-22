package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
			pointID, inserted, err := upsertGPSPointTx(tx, point)
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
		}
	}
	if err := tx.Commit(); err != nil {
		return GPSImportResult{}, err
	}
	return result, nil
}

func upsertGPSPointTx(tx *sql.Tx, point gps.RawPoint) (int64, bool, error) {
	hash, paramsJSON, err := gpsPointHash(point)
	if err != nil {
		return 0, false, err
	}

	res, err := tx.Exec(`INSERT OR IGNORE INTO gps_points(point_hash,first_source_file,first_raw_row,last_source_file,last_raw_row,seen_count,raw_dt,normalised_time,lat,lng,altitude,angle,speed_kph,params_raw,params_json)
VALUES(?,?,?,?,1,?, ?,?,?,?,?,?,?,?,?)`,
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
	if !inserted {
		if _, err := tx.Exec(`UPDATE gps_points SET last_source_file=?,last_raw_row=?,seen_count=seen_count+1 WHERE point_hash=?`, point.SourceFile, point.RawRow, hash); err != nil {
			return 0, false, err
		}
	}

	var id int64
	if err := tx.QueryRow(`SELECT id FROM gps_points WHERE point_hash=?`, hash).Scan(&id); err != nil {
		return 0, false, err
	}
	return id, inserted, nil
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
