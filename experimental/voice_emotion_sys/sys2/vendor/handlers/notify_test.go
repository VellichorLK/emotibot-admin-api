package handlers

import (
	"strconv"
	"testing"
)

func TestParseDay(t *testing.T) {
	//boundary test
	day := 0
	err := parseDay(day)
	if err == nil {
		t.Error(err)
	}

	day = 8
	err = parseDay(day)
	if err == nil {
		t.Error("day = -1, error")
	}

	//normal test
	days := []int{1, 2, 3, 4, 5, 6, 7}
	for _, v := range days {
		err = parseDay(v)
		if err != nil {
			t.Error("day = " + strconv.Itoa(v) + " error")
		}
	}

}

func TestParseDate(t *testing.T) {
	//boundary test
	date := 0
	err := parseDate(date)
	if err == nil {
		t.Error("wrong date=0")
	}

	date = 32
	err = parseDate(date)
	if err == nil {
		t.Error("wrong date=32")
	}

	//normal test
	dateList := []int{1, 3, 21, 31, 9, 7, 6, 13, 19}

	for _, v := range dateList {
		err = parseDate(v)
		if err != nil {
			t.Error(err)
		}
	}

}

func TestParseMonth(t *testing.T) {

	var err error

	fails := []int{0, 13, 41, -1, 100}
	for _, v := range fails {
		err = parseMonth(v)
		if err == nil {
			t.Error("month=" + strconv.Itoa(v))
		}
	}

	success := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	for _, v := range success {
		err = parseMonth(v)
		if err != nil {
			t.Error("month=" + strconv.Itoa(v))
		}
	}
}

func TestParseHourMinute(t *testing.T) {
	var err error

	fails := [][]string{{"13", "99"}, {"41", "01"}, {"-1", "30"}, {"100", "-3"}, {"23", "61"}}

	for _, v := range fails {
		err = parseHourMinute(v)
		if err == nil {
			t.Error("error " + v[0] + ":" + v[1])
		}
	}

	success := [][]string{{"0", "0"}, {"13", "01"}, {"23", "59"}, {"1", "1"}, {"08", "9"}}
	for _, v := range success {
		err = parseHourMinute(v)
		if err != nil {
			t.Error("error " + v[0] + ":" + v[1])
		}
	}
}

func TestParseMothDate(t *testing.T) {
	var err error
	failsM := []int{-1, 0, 13, 1, 2, 9, 8}
	failsD := []int{13, 2, 0, 0, 31, 31, 32}
	for idx := range failsM {
		err = parseMonthDate(failsM[idx], failsD[idx])
		if err == nil {
			t.Error("error " + strconv.Itoa(failsM[idx]) + "-" + strconv.Itoa(failsD[idx]))
		}
	}

	successM := []int{1, 2, 3, 4, 5, 8, 9, 12}
	successD := []int{1, 28, 31, 5, 5, 31, 30, 31}
	for idx := range successM {
		err = parseMonthDate(successM[idx], successD[idx])
		if err != nil {
			t.Error("error " + strconv.Itoa(successM[idx]) + "-" + strconv.Itoa(successD[idx]))
		}
	}
}

func TestParseCron(t *testing.T) {
	nd := &NotifyData{}

	cron := "* * * * 4"
	if err := ParseCron(cron, nd); err != nil {
		t.Error(err)
	}
	if nd.Cycle != WEEK || nd.Day != 4 {
		t.Error()
	}

	cron = "59 12 * * *"
	if err := ParseCron(cron, nd); err != nil {
		t.Error(err)
	}
	if nd.Cycle != DAY || nd.Time != "12:59" {
		t.Error("12:59 check error")
	}

	cron = "59 13 3 * *"
	if err := ParseCron(cron, nd); err != nil {
		t.Error(err)
	}
	if nd.Cycle != MONTH || nd.Time != "13:59" || nd.Date != 3 {
		t.Error("12:59 mothn 3 check error")
	}

	cron = "59 13 3 9 *"
	if err := ParseCron(cron, nd); err != nil {
		t.Error(err)
	}
	if nd.Cycle != YEAR || nd.Time != "13:59" || nd.Date != 3 || nd.Month != 9 {
		t.Error("12:59 date 3 month 9 check error")
	}

}

func TestPeriodCrontab(t *testing.T) {
	cron := "* * * * 4"
	period, err := parsePeriod(cron)
	if err != nil {
		t.Error(err)
	}
	if period != WEEK {
		t.Error("wrong period parsing at " + cron + ", expecting " + WEEK + " but get " + period)
	}

	cron = "12 13 21 * *"
	period, err = parsePeriod(cron)
	if err != nil {
		t.Error(err)
	}
	if period != MONTH {
		t.Error("wrong period parsing at " + cron + ", expecting " + MONTH + " but get " + period)
	}

	cron = "12 13 * * *"
	period, err = parsePeriod(cron)
	if err != nil {
		t.Error(err)
	}
	if period != DAY {
		t.Error("wrong period parsing at " + cron + ", expecting " + DAY + " but get " + period)
	}

	cron = "12 13 20 1 *"
	period, err = parsePeriod(cron)
	if err != nil {
		t.Error(err)
	}
	if period != YEAR {
		t.Error("wrong period parsing at " + cron + ", expecting " + DAY + " but get " + period)
	}

	cron = "12 13 20 1 4"
	period, err = parsePeriod(cron)
	if err == nil {
		t.Error("wrong crontab " + cron + " doesn't get error")
	}

}
