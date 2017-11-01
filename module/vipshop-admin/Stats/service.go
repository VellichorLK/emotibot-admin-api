package Stats

import (
	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
)

func GetAuditList(appid string, input *AuditInput) (*AuditRet, int, error) {
	list, totalCnt, err := getAuditListPage(appid, input, input.Page, input.ListPerPage)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := &AuditRet{
		TotalCount: totalCnt,
		Data:       list,
	}

	return ret, ApiError.SUCCESS, nil
}
