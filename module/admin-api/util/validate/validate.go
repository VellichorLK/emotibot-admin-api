package validate

func IsValidAppID(id string) bool {
	return len(id) > 0 && HasOnlyNumEngDash(id)
}
