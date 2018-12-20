package BF

func GetSSMCategories(appid string) (*Category, error) {
	return getSSMCategories(appid, false)
}

func GetSSMLabels(appid string) ([]*SSMLabel, error) {
	return getSSMLabels(appid)
}

func GetBFAccessToken(userid string) (string, error) {
	return getBFAccessToken(userid)
}
