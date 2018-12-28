// Package mathutil is the util functions that built-in math package does not provide.
// according to (this blog)[https://www.jianshu.com/p/4a833196b02c],
// we should implement small util(ex: MaxInt, MinInt) by ourself.
package mathutil

// MaxInt compare a and b, and return the maximum value.
func MaxInt(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

// MinInt compare a and b, and return the minimum value.
func MinInt(a int, b int) int {
	if a > b {
		return b
	}
	return a
}
