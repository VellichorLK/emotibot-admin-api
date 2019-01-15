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

	targetStr := `SELECT rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s,
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s,
	rrr.%s
	FROM (SELECT * FROM %s WHERE %s=0 ) as rg
	INNER JOIN (SELECT * FROM %s WHERE file_name = ? and extension = ?) as gc on rg.%s = gc.%s
	LEFT JOIN %s as rrr ON rg.%s = rrr.%s
	`
	targetStr = fmt.Sprintf(
		targetStr,
		RGID,
		RGUUID,
		RGName,
		fldDescription,
		RGLimitSpeed,
		RGLimitSilence,
		fldCreateTime,
		fldGroupIsEnabled,
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
		RRRRuleID,
		tblRuleGroup,
		fldIsDelete,
		tblRGC,
		fldID,
		RGCGroupID,
		tblRelGrpRule,
		fldID,
		RRRGroupID,
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
