package engine

import (
	"context"

	"gogator/internal/gps"
)

func Run(ctx context.Context, in Input) (Result, error) {
	_ = ctx
	adapter := LegacyAdapter{}
	points := append([]gps.RawPoint(nil), in.Points...)
	enableQuality := in.EngineConfig.Quality
	motionCfg := in.EngineConfig.Motion
	stayCfg := in.EngineConfig.StayDetection
	if in.EngineConfig == (EngineConfig{}) {
		enableQuality = in.Config.Engine.Quality.Enabled
		motionCfg = MotionConfig{
			Enabled:                     in.Config.Engine.Motion.Enabled,
			StationarySpeedThresholdKPH: in.Config.Engine.Motion.StationarySpeedThresholdKPH,
			MovingSpeedThresholdKPH:     in.Config.Engine.Motion.MovingSpeedThresholdKPH,
			GapThresholdMinutes:         in.Config.Engine.Motion.GapThresholdMinutes,
			MinConsecutiveSamples:       in.Config.Engine.Motion.MinConsecutiveSamples,
		}
		stayCfg = StayConfig{
			Enabled:                in.Config.Engine.StayDetection.Enabled,
			MinDurationMinutes:     in.Config.Engine.StayDetection.MinDurationMinutes,
			MaxRadiusMeters:        in.Config.Engine.StayDetection.MaxRadiusMeters,
			MinPoints:              in.Config.Engine.StayDetection.MinPoints,
			SiteMatchRadiusMeters:  in.Config.Engine.StayDetection.SiteMatchRadiusMeters,
			GapInferredStopEnabled: in.Config.Engine.StayDetection.GapInferredStopEnabled,
		}
	}
	evidence := buildEvidence(points, enableQuality)
	motion := MotionEvidence{}
	if motionCfg.Enabled {
		motion = classifyMotion(evidence.Points, motionCfg)
	}
	stays := detectStays(evidence.Points, motion, in.Sites, stayCfg)
	adapter.SortAndRecalculate(points)
	points = adapter.Classify(points, in)
	valid, jitter := adapter.BuildTrips(points, in)
	valid, jitter = adapter.ApplyImportant(valid, jitter, in)
	valid, observations, anomalies := adapter.ApplyRoutes(valid, in)
	review, sameSite := adapter.SplitJitter(jitter)
	return Result{Points: points, Evidence: evidence, Motion: motion, Stays: stays, Valid: valid, Jitter: jitter, JitterReview: review, JitterSameSite: sameSite, RouteObservations: observations, RouteAnomalies: anomalies, SiteCount: len(in.Sites), RouteCount: len(in.Routes)}, nil
}
