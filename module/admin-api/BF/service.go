package BF

func GetSSMCategories(appid string) (*Category, error) {
	return getSSMCategories(appid, false)
}
