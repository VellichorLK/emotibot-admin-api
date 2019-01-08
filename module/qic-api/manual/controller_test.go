package manual

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
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

func TestInspectTaskToInspectTaskInRes(t *testing.T) {
	it := &model.InspectTask{
		ID:          int64(5),
		Name:        "testit",
		CreateTime:  int64(55688),
		PublishTime: -1,
		CallStart:   int64(20),
		CallEnd:     int64(200),
		Outlines: []model.Outline{
			model.Outline{
				Name: "outline1",
			},
			model.Outline{
				Name: "outline2",
			},
		},
		InspectNum:   5,
		InspectCount: 10,
		InspectTotal: 100,
		Reviewer:     "heelo",
		ReviewNum:    5,
		ReviewTotal:  10,
	}

	itInRes := inspectTaskToInspectTaskInRes(it)

	if itInRes.ID != it.ID || itInRes.Name != it.Name || itInRes.CreateTime != it.CreateTime || itInRes.PublishTime != it.PublishTime {
		t.Errorf("parse inspect task failed, expect %+v, but got: %+v", it, itInRes)
		return
	}

	if itInRes.TimeRange.StartTime != it.CallStart || itInRes.TimeRange.EndTime != it.CallEnd {
		t.Errorf("parse time range failed, expect: %+v, but got: %+v", it, itInRes.TimeRange)
		return
	}

	if itInRes.InspectNum != it.InspectNum || itInRes.InspectCount != it.InspectCount || itInRes.InspectTotal != it.InspectTotal {
		t.Errorf("parse inspect numbers failed, expect: total: %d count: %d num: %d, but got total: %d count: %d num: %d", it.InspectTotal, it.InspectCount, it.InspectNum, itInRes.InspectTotal, itInRes.InspectCount, itInRes.InspectNum)
		return
	}

	if itInRes.Reviewer != it.Reviewer || itInRes.ReviewNum != it.ReviewNum || itInRes.ReviewTotal != it.ReviewTotal {
		t.Errorf("parse reviewer failed, expect: %+v, but got: %+v", it, itInRes.TimeRange)
		return
	}
}

func TestInspectTaskToInspectTaskInResForNormalUser(t *testing.T) {
	it := &model.InspectTask{
		ID:          int64(5),
		Name:        "testit",
		CreateTime:  int64(55688),
		PublishTime: -1,
		CallStart:   int64(20),
		CallEnd:     int64(200),
		Outlines: []model.Outline{
			model.Outline{
				Name: "outline1",
			},
			model.Outline{
				Name: "outline2",
			},
		},
		InspectNum:   5,
		InspectCount: 10,
		InspectTotal: 100,
		Reviewer:     "heelo",
		ReviewNum:    5,
		ReviewTotal:  10,
	}

	itInRes := inspectTaskToInspectTaskInResForNormalUser(it)

	if itInRes.ID != it.ID || itInRes.Name != it.Name || itInRes.CreateTime != it.CreateTime {
		t.Errorf("parse inspect task failed, expect %+v, but got: %+v", it, itInRes)
		return
	}

	if itInRes.TimeRange.StartTime != it.CallStart || itInRes.TimeRange.EndTime != it.CallEnd {
		t.Errorf("parse time range failed, expect: %+v, but got: %+v", it, itInRes.TimeRange)
		return
	}

	if itInRes.Count != it.InspectCount || itInRes.Total != it.InspectTotal {
		t.Errorf("parse count failed, expect: %+v, but got: %+v", it, itInRes)
		return
	}

	if itInRes.Reviewer != it.Reviewer {
		t.Errorf("parse reviewer failed, expect: %s, but got: %s", it.Reviewer, itInRes.Reviewer)
		return
	}
}
