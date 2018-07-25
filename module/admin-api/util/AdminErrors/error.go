package AdminErrors

import "fmt"

func New(errno int, text string) AdminError {
	return &errorStruct{
		errno, text,
	}
}

type AdminError interface {
	Error() string
	Errno() int
	String(input ...string) string
}

type errorStruct struct {
	errno  int
	errMsg string
}

func (e *errorStruct) Error() string {
	return e.errMsg
}
func (e *errorStruct) Errno() int {
	return e.errno
}
func (e *errorStruct) String(input ...string) string {
	locale := "ch-cn"
	if len(input) > 0 && msg[input[0]] != nil {
		locale = input[0]
	}
	if e.Errno() == ErrnoSuccess {
		return ""
	}
	errnoMsg := msg[locale][e.Errno()]
	if errnoMsg == "" {
		errnoMsg = unknownMsg[locale]
	}

	return fmt.Sprintf("%s: %s", errnoMsg, e.Error())
}
