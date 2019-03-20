package util

import (
	"testing"
)

func TestUserTryCache(t *testing.T) {
	if UserBanInfos == nil {
		t.Errorf("UserBanInfos is not init successfully")
	}
	UserBanInfos.BanUser("test")
	if !UserBanInfos.IsUserBanned("test") {
		t.Errorf("User is not banned successfully")
	}
	UserBanInfos.ClearBanInfo("test")
	if UserBanInfos.IsUserBanned("test") {
		t.Errorf("User ban info is not cleared successfully")
	}
}
