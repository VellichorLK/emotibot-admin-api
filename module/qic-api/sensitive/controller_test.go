package sensitive

import (
	"testing"
)

func TestTransformSensitiveWordInReqToSensitiveWord(t *testing.T) {
	inreq := &SensitiveWordInReq{
		Name:  "test",
		Score: 5,
		Exception: ExceptionInReq{
			Customer: []string{"1", "2"},
			Staff:    []string{"2, 3"},
		},
	}

	word := transformSensitiveWordInReqToSensitiveWord(inreq)
	if word == nil {
		t.Error("got nil word")
		return
	}

	if word.Name != inreq.Name {
		t.Errorf("parse sensitive word name failed, expect %s but got %s", inreq.Name, word.Name)
		return
	}

	if len(word.StaffException) != len(inreq.Exception.Staff) {
		t.Errorf(
			"parse sensitive staff execption failed, expect %+v but got %+v",
			inreq.Exception.Customer,
			word.CustomerException,
		)
		return
	}

	if len(word.CustomerException) != len(inreq.Exception.Customer) {
		t.Errorf(
			"parse sensitive customer execption failed, expect %+v but got %+v",
			inreq.Exception.Staff,
			word.CustomerException,
		)
		return
	}

}
