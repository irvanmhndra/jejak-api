package service

import (
	"math"
	"strings"

	"github.com/irvanmhndra/jejak-api/internal/model"
)

const earthRadiusM = 6371000.0

// haversine returns the great-circle distance in meters between two points.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// TrackStats is the summary derived from a GPS track.
type TrackStats struct {
	DistanceMeters float64
	ElapsedSeconds int
	MovingSeconds  int
	AvgSpeedMps    float64
	MaxSpeedMps    float64
	ElevationGainM float64
	ElevationLossM float64
}

// Movement below this speed (m/s) is treated as a stop for "moving time".
const movingSpeedThreshold = 0.6

// Segment speeds above this (m/s, ~144 km/h) are treated as GPS glitches.
const maxRealisticSpeed = 40.0

// computeTrackStats fills each point's cumulative Dist and returns the summary.
func computeTrackStats(points []model.TrackPoint) TrackStats {
	var st TrackStats
	if len(points) == 0 {
		return st
	}
	points[0].Dist = 0
	var cumulative float64
	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		d := haversine(prev.Lat, prev.Lng, cur.Lat, cur.Lng)
		cumulative += d
		points[i].Dist = cumulative

		dt := cur.T.Sub(prev.T).Seconds()
		if dt > 0 {
			speed := d / dt
			if speed <= maxRealisticSpeed && speed > st.MaxSpeedMps {
				st.MaxSpeedMps = speed
			}
			if speed >= movingSpeedThreshold {
				st.MovingSeconds += int(dt)
			}
		}

		if prev.Ele != nil && cur.Ele != nil {
			delta := *cur.Ele - *prev.Ele
			if delta >= 1.0 {
				st.ElevationGainM += delta
			} else if delta <= -1.0 {
				st.ElevationLossM += -delta
			}
		}
	}
	st.DistanceMeters = cumulative
	st.ElapsedSeconds = int(points[len(points)-1].T.Sub(points[0].T).Seconds())
	if st.MovingSeconds > 0 {
		st.AvgSpeedMps = st.DistanceMeters / float64(st.MovingSeconds)
	}
	return st
}

// encodePolyline encodes points using the Google Encoded Polyline Algorithm
// (precision 5) for compact map previews.
func encodePolyline(points []model.TrackPoint) string {
	var b strings.Builder
	var prevLat, prevLng int
	for _, p := range points {
		lat := int(math.Round(p.Lat * 1e5))
		lng := int(math.Round(p.Lng * 1e5))
		encodeSigned(&b, lat-prevLat)
		encodeSigned(&b, lng-prevLng)
		prevLat, prevLng = lat, lng
	}
	return b.String()
}

func encodeSigned(b *strings.Builder, v int) {
	shifted := v << 1
	if v < 0 {
		shifted = ^shifted
	}
	for shifted >= 0x20 {
		b.WriteByte(byte((0x20 | (shifted & 0x1f)) + 63))
		shifted >>= 5
	}
	b.WriteByte(byte(shifted + 63))
}
