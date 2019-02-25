package feedback

import (
	"errors"
)

var ErrReasonNotExists = errors.New("Reason not exists")
var ErrDuplicateContent = errors.New("Duplicate content")
