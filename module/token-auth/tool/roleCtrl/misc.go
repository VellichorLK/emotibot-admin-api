package main

import "emotibot.com/emotigo/module/token-auth/service"
import "emotibot.com/emotigo/module/token-auth/internal/data"

func checkEnterprise(enterpriseID string) (enterprise *data.Enterprise, retCode int, err error) {
	if enterpriseID == "" {
		retCode = errEnterpriseID
		return
	}

	enterprise, err = service.GetEnterprise(enterpriseID)
	if err != nil || enterprise == nil {
		retCode = errEnterpriseID
		return
	}
	return
}
