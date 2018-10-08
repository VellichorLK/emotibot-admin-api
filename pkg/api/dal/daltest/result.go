package daltest

// ExpectResult represent the result from dal module, which can provide find grind control over the mock result.
type ExpectResult struct {
	values []interface{}
	hasErr bool
}

// WillReturn store the expected value into ExpectResult, which will be used by injected command later.
// How the command uses the values is depending on command implementation.
func (r *ExpectResult) WillReturn(value ...interface{}) {
	r.values = value
}

// WillFail store the err into ExpectResult, which
func (r *ExpectResult) WillFail() {
	r.hasErr = true
}
