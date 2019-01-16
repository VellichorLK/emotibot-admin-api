package v2

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func handleListRobotAudit(w http.ResponseWriter, r *http.Request) {
	filter := AuditInput{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, jsonErr.Error()), nil)
		return
	}

	locale := requestheader.GetLocale(r)
	records, err := GetRobotAuditRecord(&filter, locale)
	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, records)
	}
}
func handleListEnterpriseAudit(w http.ResponseWriter, r *http.Request) {
	filter := AuditInput{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, jsonErr.Error()), nil)
		return
	}

	locale := requestheader.GetLocale(r)
	records, err := GetEnterpriseAuditRecord(&filter, locale)
	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, records)
	}
}
func handleListSystemAudit(w http.ResponseWriter, r *http.Request) {
	filter := AuditInput{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, jsonErr.Error()), nil)
		return
	}

	locale := requestheader.GetLocale(r)
	records, err := GetSystemAuditRecord(&filter, locale)
	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, records)
	}
}
