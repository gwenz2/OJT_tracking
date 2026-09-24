// Pure geofence evaluation: Haversine distance + location-status rules.
// The server decides everything — client GPS values are evidence, never truth.
package attendance

import "math"

const earthRadiusM = 6_371_000

// LocationStatus is the server-computed geofence verdict persisted on
// attendance_evidence.location_status.
type LocationStatus string

const (
	LocationVerified      LocationStatus = "verified"
	LocationOutsideRadius LocationStatus = "outside_radius"
	LocationLowAccuracy   LocationStatus = "low_accuracy"
	LocationUnavailable   LocationStatus = "unavailable"
)

// LocationInput is the untrusted client-reported fix.
type LocationInput struct {
	Latitude  *float64
	Longitude *float64
	AccuracyM *float64
	// Reason is required when the client could not obtain a fix.
	Reason string
}

// LocationResult is what evidence rows persist.
type LocationResult struct {
	Status      LocationStatus
	DistanceM   *float64 // nil when no fix
	RadiusUsedM *int     // nil when no fix
}

// HaversineM returns the great-circle distance in meters between two
// WGS84 points.
func HaversineM(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadiusM * math.Asin(math.Sqrt(a))
}

// EvaluateLocation classifies the fix against the site geofence.
//
// Rules (spec FR-016..019):
//   - no fix                -> unavailable (reason required at handler level)
//   - fix out of lat/lon bounds -> unavailable (nonsense coords)
//   - accuracy missing or > lowAccuracyThresholdM -> low_accuracy
//   - distance > site radius  -> outside_radius
//   - otherwise               -> verified
//
// Low accuracy wins over outside-radius: with a poor fix we cannot trust the
// distance, so the honest flag is low_accuracy. Distance is still recorded
// (it is evidence) but does not drive the status.
func EvaluateLocation(in LocationInput, siteLat, siteLon float64, radiusM int, lowAccuracyThresholdM float64) LocationResult {
	if in.Latitude == nil || in.Longitude == nil {
		return LocationResult{Status: LocationUnavailable}
	}
	lat, lon := *in.Latitude, *in.Longitude
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return LocationResult{Status: LocationUnavailable}
	}

	dist := HaversineM(lat, lon, siteLat, siteLon)
	res := LocationResult{
		DistanceM:   &dist,
		RadiusUsedM: &radiusM,
	}

	accuracyOK := in.AccuracyM != nil && *in.AccuracyM >= 0 && *in.AccuracyM <= lowAccuracyThresholdM
	switch {
	case !accuracyOK:
		res.Status = LocationLowAccuracy
	case dist > float64(radiusM):
		res.Status = LocationOutsideRadius
	default:
		res.Status = LocationVerified
	}
	return res
}
