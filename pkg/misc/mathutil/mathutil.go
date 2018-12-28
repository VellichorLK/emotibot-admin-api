// Package mathutil is the util functions that built-in math package does not provide.
// according to these sources:
// 	- https://www.jianshu.com/p/4a833196b02c
//  - https://mrekucci.blogspot.com/2015/07/dont-abuse-mathmax-mathmin.html
// we should implement small util(ex: MaxInt, MinInt) by ourself.
package mathutil

// MaxInt will return the max of nums.
// If nums if empty, return 0.
func MaxInt(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, n := range nums[1:] {
		if n > max {
			max = n
		}
	}
	return max
}

// MinInt will return the min of nums.
// If nums if empty, return 0.
func MinInt(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	min := nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
	}
	return min
}
