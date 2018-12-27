package model

import (
	"fmt"
	"testing"
)

func TestGetGroupsSQL(t *testing.T) {
	filter := GroupFilter{
		Deal:      -1,
		FileName:  "test.wav",
		Extension: "abcdefg",
	}

	targetStr := `SELECT g.%s, g.%s, g.%s, g.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s
	FROM (SELECT * FROM RuleGroupCondition WHERE %s = ? and %s = ?) as gc
	LEFT JOIN RuleGroup as rg ON gc.%s = rg.%s
	`
	targetStr = fmt.Sprintf(
		targetStr,
		RGID,
		RGName,
		RGLimitSpeed,
		RGLimitSilence,
		RGCFileName,
		RGCDeal,
		RGCSeries,
		RGCStaffID,
		RGCStaffName,
		RGCExtension,
		RGCDepartment,
		RGCCustomerID,
		RGCCustomerName,
		RGCCustomerPhone,
		RGCCallStart,
		RGCCallEnd,
		RGCLeftChannel,
		RGCRightChannel,
		RGCFileName,
		RGCExtension,
		RGCGroupID,
		RGID,
	)

	queryStr, values := getGroupsSQL(&filter)

	if len(values) != 2 {
		t.Error("expect values length 2")
		return
	}

	if targetStr != queryStr {
		t.Errorf("exptect %s but got %s", targetStr, queryStr)
		return
	}
}
