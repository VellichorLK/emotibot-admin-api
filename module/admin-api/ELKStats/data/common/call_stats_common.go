package common

const (
	CallStatusComplete       = 1
	CallStatusOngoing        = 0
	CallStatusTranserToHuman = -1
	CallStatusTimeout        = -2
	CallStatusCancel         = -3
)

const (
	CallStatsTypeTime    = "time"
	CallStatsTypeAnswers = "answers"
)

const (
	CallsMetricTotals        = "totals"
	CallsMetricCompletes     = "completes"
	CallsMetricCompletesRate = "completes_rate"
	CallsMetricToHumans      = "to_humans"
	CallsMetricToHumansRate  = "to_humans_rate"
	CallsMetricTimeouts      = "timeouts"
	CallsMetricTimeoutsRate  = "timeouts_rate"
	CallsMetricCancels       = "cancels"
	CallsMetricCancelsRate   = "cancels_rate"
	CallsMetricUnknowns      = "unknowns"
)
