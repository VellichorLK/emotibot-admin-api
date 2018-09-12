package v2

import (
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

// GetRobotAuditRecord will get audit record of specific appid
func GetRobotAuditRecord(filter *AuditInput) ([]*AuditLog, AdminErrors.AdminError) {
	var userIDPtr *string
	if filter.UserID != "" {
		userIDPtr = &filter.UserID
	}
	modulePtr, opPtr := getModuleOpPtr(filter.Filter)
	ret, err := getAuditList(nil, &filter.RobotID, userIDPtr, modulePtr, opPtr, filter.Start, filter.End, filter.Page, filter.ListPerPage)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return ret, nil
}

// GetEnterpriseAuditRecord will get audit record of specific enterprise
func GetEnterpriseAuditRecord(filter *AuditInput) ([]*AuditLog, AdminErrors.AdminError) {
	var userIDPtr *string
	if filter.UserID != "" {
		userIDPtr = &filter.UserID
	}
	modulePtr, opPtr := getModuleOpPtr(filter.Filter)
	ret, err := getAuditList(&filter.EnterpriseID, nil, userIDPtr, modulePtr, opPtr, filter.Start, filter.End, filter.Page, filter.ListPerPage)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return ret, nil
}

// GetSystemAuditRecord will get audit record of specific enterprise
func GetSystemAuditRecord(filter *AuditInput) ([]*AuditLog, AdminErrors.AdminError) {
	var userIDPtr *string
	if filter.UserID != "" {
		userIDPtr = &filter.UserID
	}
	modulePtr, opPtr := getModuleOpPtr(filter.Filter)
	ret, err := getAuditList(nil, nil, userIDPtr, modulePtr, opPtr, filter.Start, filter.End, filter.Page, filter.ListPerPage)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return ret, nil
}

func getModuleOpPtr(filter *AuditFilter) (*string, *string) {
	var modulePtr *string
	var opPtr *string
	if filter != nil {
		if filter.Module != "" {
			modulePtr = &filter.Module
		}
		if filter.Operation != "" {
			opPtr = &filter.Operation
		}
	}
	return modulePtr, opPtr
}
