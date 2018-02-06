package handlers

import (
	"errors"
	"strconv"
	"time"
)

//time unit
const (
	Hour  = int64(time.Hour / time.Second)
	Day   = Hour * 24
	Week  = Day * 7
	Month = Day * 30
	Year  = Day * 365
)

func GetFloatPrecesion(v float64, precesion int) float64 {
	t := 10 * precesion
	return float64(int(v*float64(t))) / float64(t)
}

func RoundUpTime(t int64, unit int64) (int64, error) {

	t1 := time.Unix(t, 0)
	switch unit {
	case Hour:
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), 0, 0, 0, t1.Location())
	case Day:
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	case Month:
		t1 = time.Date(t1.Year(), t1.Month(), 1, 0, 0, 0, 0, t1.Location())
	case Year:
		t1 = time.Date(t1.Year(), 1, 1, 0, 0, 0, 0, t1.Location())
	default:
		return 0, errors.New("error time duration:" + strconv.FormatInt(unit, 10))
	}

	return t1.Unix(), nil
}

func AddTimeUnit(t int64, unit int64) int64 {

	t1 := time.Unix(t, 0)

	switch unit {
	case Hour:
		return t1.Unix() + Hour
	case Week:
		t1 = t1.AddDate(0, 0, 7)
	case Day:
		t1 = t1.AddDate(0, 0, 1)
	case Month:
		t1 = t1.AddDate(0, 1, 0)
	case Year:
		t1 = t1.AddDate(1, 0, 0)
	default:
		return 0
	}
	return t1.Unix()
}

//ParseTime return t1,t2 in unix epoch time
func ParseTime(_t1 string, _t2 string) (uint64, uint64, int, error) {

	layout := "20060102"
	//t1, err := time.Parse(layout, _t1)

	loc, err := time.LoadLocation("Local")
	if err != nil {
		return 0, 0, 0, err
	}

	t1, err := time.ParseInLocation(layout, _t1, loc)
	if err != nil {
		return 0, 0, 0, err
	}

	//t2, err := time.Parse(layout, _t2)
	t2, err := time.ParseInLocation(layout, _t2, loc)
	if err != nil {
		return 0, 0, 0, err
	}

	return CheckAndRoundTime(t1, t2)

}

func CheckAndRoundTime(t1 time.Time, t2 time.Time) (uint64, uint64, int, error) {
	if t1.Unix() > t2.Unix() {
		return 0, 0, 0, errors.New("t1 >= t2")
	}

	now := time.Now()

	if t2.Unix() > now.Unix() {
		return 0, 0, 0, errors.New("t2 > now. don't do future time. time traveler")
	}

	days := int(t2.Sub(t1).Hours()/24) + 1

	return uint64(t1.Unix()), uint64(t2.Unix()), days, nil
}
