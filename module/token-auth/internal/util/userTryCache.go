package util

import (
	"time"
)

// BanInfo record ban info of user
type BanInfo struct {
	Duration time.Duration `json:"duration"`
	Expired  time.Time     `json:"expired"`
}

// BanMap is a map from userID to ban info
type BanMap map[string]*BanInfo

// UpdateExpire will set expired to now + duration * 2
func (info *BanInfo) UpdateExpire() {
	info.Duration *= 2
	info.Expired = time.Now().Add(info.Duration)
}

var (
	// UserBanInfos map user to it's baninfo
	UserBanInfos BanMap
)

func init() {
	UserBanInfos = map[string]*BanInfo{}
}

// IsUserBanned will check user is banned or not
func (infos BanMap) IsUserBanned(userID string) bool {
	if info, ok := UserBanInfos[userID]; ok {
		return time.Now().Unix() <= info.Expired.Unix()
	}
	return false
}

// BanUser will add user into ban list, default will ban 30 seconds
func (infos BanMap) BanUser(userID string) {
	if info, ok := UserBanInfos[userID]; ok {
		info.UpdateExpire()
	} else {
		banTime := time.Second * 30
		info = &BanInfo{
			Duration: banTime,
			Expired:  time.Now().Add(banTime),
		}
		UserBanInfos[userID] = info
	}
}

// ClearBanInfo will remove ban info of provided userID
func (infos BanMap) ClearBanInfo(userID string) {
	if _, ok := UserBanInfos[userID]; ok {
		delete(UserBanInfos, userID)
	}
}
