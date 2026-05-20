package gps

import "time"

type RawPoint struct {
	RawRow            int
	RawDT             string
	Time              time.Time
	Lat               float64
	Lng               float64
	Altitude          float64
	Angle             float64
	SpeedKPH          float64
	ParamsRaw         string
	Params            map[string]string
	ParamNums         map[string]float64
	DistanceFromPrevM float64
	ImpliedSpeedKPH   float64
	Moving            bool
	Stationary        bool
	PDOPQuality       string
	Flags             []string
}

type Trip struct {
	Index                      int
	RawStartRow                int
	RawEndRow                  int
	RawPoints                  int
	Start                      time.Time
	End                        time.Time
	DepartLat                  float64
	DepartLng                  float64
	DestLat                    float64
	DestLng                    float64
	DistanceKM                 float64
	DurationHours              float64
	TopSpeedKPH                float64
	AverageSpeedKPH            float64
	SiteDurationHours          float64
	DepartureSite              string
	DepartureAddress           string
	DestinationSite            string
	DestinationAddress         string
	Jitter                     bool
	Flags                      []string
	RouteName                  string
	RouteConfidence            string
	RouteMatchStatus           string
	RouteExpectedDistanceRange string
	RouteExpectedDurationRange string
	RouteNotes                 string
	Points                     []RawPoint
}
