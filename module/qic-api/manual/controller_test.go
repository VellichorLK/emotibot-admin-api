package manual

import (
	"net/url"
	"testing"
)

func TestInspectTaskInReqToInspectTask(t *testing.T) {
	timeRange := CallTimeRange{
		StartTime: 55688,
		EndTime:   77889,
	}

	sampling := Sampling{
		Percentage: 50,
		ByPerson:   5,
	}

	inreq := InspectTaskInReq{
		Name:      "inreq",
		TimeRange: timeRange,
		Outlines: []int64{
			1,
			2,
			3,
		},
		Staffs: []string{
			"a",
			"b",
			"c",
		},
		Form:         55688,
		Sampling:     sampling,
		IsInspecting: 1,
	}

	it := inspectTaskInReqToInspectTask(&inreq)

	if inreq.Name != it.Name || inreq.IsInspecting != it.ExcludeInspected || inreq.Form != it.Form.ID {
		t.Errorf("transfrom inreq to task failed, expect %+v, but got: %+v\n", inreq, it)
		return
	}

	if len(inreq.Outlines) != len(it.Outlines) || len(inreq.Staffs) != len(it.Staffs) {
		t.Errorf("transfrom inreq to task failed, expect %+v, but got: %+v\n", inreq, it)
		return
	}

	if inreq.Sampling.Percentage != it.InspectPercentage || inreq.Sampling.ByPerson != it.InspectByPerson {
		t.Errorf("transform inreq sampling to task failed, expect %+v, but got: %d, %d", inreq.Sampling, it.InspectPercentage, it.InspectByPerson)
		return
	}

	if inreq.TimeRange.StartTime != it.CallStart || inreq.TimeRange.EndTime != it.CallEnd {
		t.Errorf("transform inreq time range to task failed, expect %+v, but got: %d, %d", inreq.TimeRange, it.CallStart, it.CallEnd)
		return
	}

	if inreq.Form != it.Form.ID {
		t.Errorf("transform inreq score form to task failed, expect %d, but got: %+v", inreq.Form, it.Form)
		return
	}
}

func TestParseTaskFilter(t *testing.T) {
	values := url.Values{}
	values.Add("page", "5")
	values.Add("limit", "12")

	filter := parseTaskFilter(&values)
	if filter.Page != 5 || filter.Limit != 12 {
		t.Errorf("error while parse filter from values, expect %s, %s, but got %d, %d", values["page"], values["limit"], filter.Page, filter.Limit)
		return
	}
}
