package attendance

import (
	"testing"
	"time"
)

func TestLocalDate(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		tz   string
		want string
	}{
		// 23:30 UTC = 07:30 next day in Asia/Manila (+8)
		{"utc late evening becomes next day Manila", time.Date(2026, 3, 1, 23, 30, 0, 0, time.UTC), "Asia/Manila", "2026-03-02"},
		// 15:59 UTC = 23:59 same day Manila
		{"utc afternoon same day Manila", time.Date(2026, 3, 1, 15, 59, 0, 0, time.UTC), "Asia/Manila", "2026-03-01"},
		// 16:00 UTC = 00:00 next day Manila — midnight boundary
		{"exactly midnight Manila", time.Date(2026, 3, 1, 16, 0, 0, 0, time.UTC), "Asia/Manila", "2026-03-02"},
		{"utc noon stays utc day", time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), "UTC", "2026-03-01"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LocalDate(tc.in, tc.tz)
			if err != nil {
				t.Fatalf("LocalDate: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
	if _, err := LocalDate(time.Now(), "Not/AZone"); err == nil {
		t.Fatal("expected error for bogus timezone")
	}
}

func TestHaversineM(t *testing.T) {
	// ~111.2 km per degree of latitude near the equator.
	d := HaversineM(0, 0, 1, 0)
	if d < 110_000 || d > 112_000 {
		t.Fatalf("1° lat ≈ 111km, got %.0fm", d)
	}
	if d := HaversineM(14.5995, 120.9842, 14.5995, 120.9842); d != 0 {
		t.Fatalf("same point should be 0, got %v", d)
	}
	// Manila city hall → Intramuros ≈ 700m
	d = HaversineM(14.5907, 120.9810, 14.5896, 120.9750)
	if d < 550 || d > 850 {
		t.Fatalf("expected ~700m, got %.0fm", d)
	}
}

func TestEvaluateLocation(t *testing.T) {
	lat, lon := 14.5995, 120.9842
	f := func(v float64) *float64 { return &v }
	i := func(v int) *int { return &v }

	cases := []struct {
		name      string
		in        LocationInput
		radius    int
		threshold float64
		want      LocationStatus
		wantDist  bool
	}{
		{"no fix is unavailable", LocationInput{Reason: "denied"}, 100, 100, LocationUnavailable, false},
		{"invalid coords unavailable", LocationInput{Latitude: f(91), Longitude: f(0), AccuracyM: f(5)}, 100, 100, LocationUnavailable, false},
		{"inside radius verified", LocationInput{Latitude: f(lat), Longitude: f(lon), AccuracyM: f(15)}, 100, 100, LocationVerified, true},
		{"outside radius flagged", LocationInput{Latitude: f(lat + 0.02), Longitude: f(lon), AccuracyM: f(15)}, 100, 100, LocationOutsideRadius, true},
		{"poor accuracy flagged", LocationInput{Latitude: f(lat), Longitude: f(lon), AccuracyM: f(500)}, 100, 100, LocationLowAccuracy, true},
		{"missing accuracy flagged", LocationInput{Latitude: f(lat), Longitude: f(lon)}, 100, 100, LocationLowAccuracy, true},
		{"accuracy boundary passes", LocationInput{Latitude: f(lat), Longitude: f(lon), AccuracyM: f(100)}, 100, 100, LocationVerified, true},
		{"edge: distance exactly radius verified", LocationInput{Latitude: f(lat), Longitude: f(lon), AccuracyM: f(10)}, 100, 100, LocationVerified, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateLocation(tc.in, lat, lon, tc.radius, tc.threshold)
			if got.Status != tc.want {
				t.Fatalf("status: got %s want %s", got.Status, tc.want)
			}
			if tc.wantDist && (got.DistanceM == nil || got.RadiusUsedM == nil) {
				t.Fatal("expected distance and radius recorded")
			}
			if !tc.wantDist && (got.DistanceM != nil || got.RadiusUsedM != nil) {
				t.Fatal("unavailable fix must not record distance/radius")
			}
			if tc.wantDist {
				_ = i // silence helper
			}
		})
	}
}

func TestCreditedMinutes(t *testing.T) {
	in := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	at := func(mins int) time.Time { return in.Add(time.Duration(mins) * time.Minute) }

	cases := []struct {
		name                        string
		out                         time.Time
		rule                        BreakRule
		elapsed, deducted, credited int
	}{
		{"8h no break", at(480), BreakRule{Type: "none"}, 480, 0, 480},
		{"8h with 30m break after 4h", at(480), BreakRule{Type: "fixed_after_threshold", ThresholdMinutes: 240, DeductionMinutes: 30}, 480, 30, 450},
		{"3h below threshold no deduction", at(180), BreakRule{Type: "fixed_after_threshold", ThresholdMinutes: 240, DeductionMinutes: 30}, 180, 0, 180},
		{"exactly threshold deducts", at(240), BreakRule{Type: "fixed_after_threshold", ThresholdMinutes: 240, DeductionMinutes: 30}, 240, 30, 210},
		{"30m session with 60m deduction floors at 0", at(30), BreakRule{Type: "fixed_after_threshold", ThresholdMinutes: 1, DeductionMinutes: 60}, 30, 30, 0},
		{"sub-minute session floors to 0", in.Add(45 * time.Second), BreakRule{Type: "none"}, 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, d, c := CreditedMinutes(SessionTimes{In: in, Out: tc.out}, tc.rule)
			if e != tc.elapsed || d != tc.deducted || c != tc.credited {
				t.Fatalf("got elapsed=%d deducted=%d credited=%d, want %d/%d/%d",
					e, d, c, tc.elapsed, tc.deducted, tc.credited)
			}
		})
	}
}

func TestProgressPercent(t *testing.T) {
	if ProgressPercent(0, 100) != 0 {
		t.Fatal("0")
	}
	if ProgressPercent(50, 100) != 50 {
		t.Fatal("50")
	}
	if ProgressPercent(200, 100) != 100 {
		t.Fatal("capped at 100")
	}
	if ProgressPercent(10, 0) != 0 {
		t.Fatal("required=0 → 0")
	}
}
