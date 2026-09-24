// Pure credited-minutes calculation. Integer minutes only — no float hours.
package attendance

import "time"

// BreakRule mirrors ojt_assignments.break_rule_* columns.
type BreakRule struct {
	Type             string // "none" | "fixed_after_threshold"
	ThresholdMinutes int
	DeductionMinutes int
}

// SessionTimes are the trusted effective timestamps.
type SessionTimes struct {
	In  time.Time
	Out time.Time
}

// CreditedMinutes computes the integer minutes credited for a session.
//
//	eligible = out - in (floored to whole minutes, never negative)
//	if break rule applies and eligible >= threshold: eligible -= deduction
//
// The deduction can push credited to zero but never below.
func CreditedMinutes(t SessionTimes, rule BreakRule) (elapsed, deducted, credited int) {
	elapsed = int(t.Out.Sub(t.In).Minutes())
	if elapsed < 0 {
		elapsed = 0
	}
	deducted = 0
	if rule.Type == "fixed_after_threshold" &&
		rule.ThresholdMinutes > 0 && rule.DeductionMinutes > 0 &&
		elapsed >= rule.ThresholdMinutes {
		deducted = min(rule.DeductionMinutes, elapsed)
	}
	credited = elapsed - deducted
	return elapsed, deducted, credited
}

// ProgressPercent maps completed/required minutes to 0–100.
func ProgressPercent(completed, required int) int {
	if required <= 0 {
		return 0
	}
	return min(completed*100/required, 100)
}
