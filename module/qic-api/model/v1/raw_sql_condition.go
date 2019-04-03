package model

import "sort"

// RangeCondition is the abstract condition for query some number in Between the ub and lb.
// If ub & lb is nil, then this RangeCondition should be considered as invalid.
// You should generate a valid RangeCondition by NewRangeCondition or Upper or Lower.
type RangeCondition struct {
	ub *int64
	lb *int64
}

func NewUpperRange(upperBound int64) RangeCondition {
	return RangeCondition{
		ub: &upperBound,
	}
}

func NewLowerRange(lowerBound int64) RangeCondition {
	return RangeCondition{
		lb: &lowerBound,
	}
}

// NewRangeCondition can accept numbers for range condition.
// the numbers will be sorted, and it use the smallest and biggest number for the RangeCondition.
// If only one was given, it will become a upper bound condition.
// If none was given, it return an invalid RangeCondition.
func NewRangeCondition(num ...int64) RangeCondition {
	var rc RangeCondition
	if length := len(num); length == 1 {
		return NewUpperRange(num[0])
	} else if length > 1 {
		sort.Slice(num, func(i, j int) bool {
			return num[i] < num[j]
		})
		rc.lb = &num[0]
		rc.ub = &num[len(num)-1]
	}
	return rc

}

//SetLowerBound set the lower bound of the r.
func (r *RangeCondition) SetLowerBound(lowerBound int64) *RangeCondition {
	r.lb = &lowerBound
	return r
}

//SetUpperBound set the upper bound of the r.
func (r *RangeCondition) SetUpperBound(upperBound int64) *RangeCondition {
	r.ub = &upperBound
	return r
}

//IsValid If any of theRangeCondition's upper and lower bound is not nil, it considered as valid.
func (r *RangeCondition) IsValid() bool {
	return r.ub != nil && r.lb != nil
}
