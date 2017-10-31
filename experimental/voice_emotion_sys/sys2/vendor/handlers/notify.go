package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

//NotifyRegister used as user register the new notification
type NotifyRegister struct {
	Cycle string   `json:"period"`
	Time  string   `json:"time"`
	Day   int      `json:"day,omitempty"`
	Date  int      `json:"date,omitempty"`
	Month int      `json:"month,omitempty"`
	Email []string `json:"email"`
}

//NotifyID the notification id
type NotifyID struct {
	ID string `json:"id"`
}

//NotifyData notify data with id
type NotifyData struct {
	NotifyID
	NotifyRegister
}

//CronTask used to push to channel
type CronTask struct {
	FileID    string
	QueueName string
	Task      []byte
	Reply     chan bool
}

//CronData the data push to rabbitmq
type CronData struct {
	Cron   string   `json:"cron"`
	Email  []string `json:"email"`
	Appid  string   `json:"appid"`
	Method string   `json:"Method"`
	ID     string   `json:"report_id"`
	Period string   `json:"period"`
}

//CronEmail used as map structure to record current crontab and email
type CronEmail struct {
	Cron   string
	Email  []string
	Appid  string
	Period string
}

//name of period
const (
	DAY   = "day"
	MONTH = "month"
	YEAR  = "year"
	WEEK  = "week"
)

func NotifyOperation(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getNotifyList(w, r)
	case "PUT":
		addNotify(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func NotifyModifyOperation(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		removeNotify(w, r)
	case "PATCH":
		updateNotify(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func addNotify(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)

	var nid NotifyID
	var nr NotifyRegister
	var cron string
	var err error
	//nr.Day = -1

	if err = parseBodyToJSON(r, &nr); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if nr.Email == nil {
		http.Error(w, "no email assigned", http.StatusBadRequest)
		return
	}

	for _, e := range nr.Email {
		if !ValidateEmail(e) {
			http.Error(w, "bad email "+e, http.StatusBadRequest)
			return
		}
	}

	listEmail := strings.Join(nr.Email, ",")

	if cron, err = parseNotifyRegister(&nr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		sql := "insert into " + NotifyTable + " (" + NCRONTAB + "," + NEMAIL + "," + NAPPID + ") values(?,?,?)"
		res, err := ExecuteSQL(nil, sql, cron, listEmail, appid)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {

			id, err := res.LastInsertId()
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {

				nid.ID = strconv.FormatInt(id, 10)
				err = sendCronTask(nid.ID, cron, nr.Email, appid, r.Method, nr.Cycle)
				if err != nil {
					log.Println(err)
				}

				writeJSONResp(w, nid)

			}
		}
	}

	//nd := &NotifyData{ID: nid.ID, Cycle: nr.Cycle, Time: nr.Time, Day: nr.Day, Date: nr.Date, Month: nr.Month, Email: nr.Email}

}

func getNotifyList(w http.ResponseWriter, r *http.Request) {

	appid := r.Header.Get(HXAPPID)

	sql := "select " + NREPORTID + "," + NCRONTAB + "," + NEMAIL + " from " + NotifyTable + " where " + NAPPID + "=?"
	nds, err := QueryNotify(sql, appid)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		writeJSONResp(w, nds)
	}
}

func removeNotify(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	url := strings.SplitN(r.URL.Path, "/", MaxSlash)
	sql := "delete from " + NotifyTable + " where " + NAPPID + "=? and " + NREPORTID + "=?"
	res, err := ExecuteSQL(nil, sql, appid, url[len(url)-1])

	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	} else {
		ra, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if ra == 0 {
			http.Error(w, "no such id", http.StatusBadRequest)
		} else {

			//id, err := strconv.ParseUint(url[len(url)-1], 10, 64)
			id := url[len(url)-1]
			if err == nil {
				err = sendCronTask(id, "", nil, appid, r.Method, "")
				if err != nil {
					log.Println(err)
				}
			}

		}
	}

}

func parseNotifyRegister(nr *NotifyRegister) (string, error) {
	var cron string
	var err error

	t := strings.Split(nr.Time, ":")
	if err := parseHourMinute(t); err != nil {
		return "", err
	}

	switch nr.Cycle {
	case DAY:
		cron = t[1] + " " + t[0] + " * * *"
	case WEEK:
		if err := parseDay(nr.Day); err != nil {
			return "", err
		}
		nr.Day = nr.Day % 7
		cron = t[1] + " " + t[0] + " * * " + strconv.Itoa(nr.Day)
	case MONTH:
		if err = parseDate(nr.Date); err != nil {
			return "", err
		}
		cron = t[1] + " " + t[0] + " " + strconv.Itoa(nr.Date) + " * *"
	case YEAR:
		if err = parseDate(nr.Date); err != nil {
			return "", err
		}
		if err = parseMonth(nr.Month); err != nil {
			return "", err
		}
		if err = parseMonthDate(nr.Month, nr.Date); err != nil {
			return "", err
		}

		cron = t[1] + " " + t[0] + " " + strconv.Itoa(nr.Date) + " " + strconv.Itoa(nr.Month) + " *"
	default:
		return "", errors.New("wrong cycle")
	}

	return cron, nil
}

func updateNotify(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	url := strings.SplitN(r.URL.Path, "/", MaxSlash)

	var nr NotifyRegister
	var cron string
	var err error
	//nr.Day = -1

	params := make([]interface{}, 0)

	if err = parseBodyToJSON(r, &nr); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if nr.Cycle != "" {
		if cron, err = parseNotifyRegister(&nr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	sql := "update " + NotifyTable

	if nr.Email != nil {

		for _, v := range nr.Email {
			if !ValidateEmail(v) {
				http.Error(w, "bad email", http.StatusBadRequest)
				return
			}
		}

		listEmail := strings.Join(nr.Email, ",")

		sql += " set " + NEMAIL + "=?"
		params = append(params, listEmail)
	}
	if cron != "" {
		if nr.Email != nil {
			sql += ","
		} else {
			sql += " set "
		}
		sql += NCRONTAB + "=?"
		params = append(params, cron)
	}

	sql += " where " + NAPPID + "=? and " + NREPORTID + "=?"
	params = append(params, appid)
	params = append(params, url[len(url)-1])

	resp, err := ExecuteSQL(nil, sql, params...)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	} else {
		ra, err := resp.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if ra == 0 {
			http.Error(w, "no such id or no change", http.StatusBadRequest)
		} else {
			//id, err := strconv.ParseUint(url[len(url)-1], 10, 64)
			id := url[len(url)-1]
			if err == nil {
				err = sendCronTask(id, cron, nr.Email, appid, r.Method, "")
				if err != nil {
					log.Println(err)
				}
			}

		}
	}

}

func parseBodyToJSON(r *http.Request, structure interface{}) error {
	if r.Body != nil {
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(structure); err != nil {
			return err
		}
	}
	return nil
}

func parseMonthDate(month int, date int) error {
	var timeFormat string
	t := strconv.Itoa(month)
	d := strconv.Itoa(date)

	if len(t) == 1 {
		timeFormat = "0"
	}
	timeFormat += t
	timeFormat += "-"
	if len(d) == 1 {
		timeFormat += "0"
	}
	timeFormat += d

	_, err := time.Parse("01-02", timeFormat)
	return err
}

func parseDay(day int) error {
	if day < 1 || day > 7 {
		return errors.New("wrong day ")
	}
	return nil
}

func parseDate(date int) error {
	if date < 1 || date > 31 {
		return errors.New("wrong date")
	}
	return nil
}

func parseMonth(month int) error {
	if month < 1 || month > 12 {
		return errors.New("wrong month")
	}
	return nil
}

func parseHourMinute(t []string) error {
	if t == nil {
		return errors.New("No assigned time")
	}
	if len(t) != 2 {
		return errors.New("wrong time format")
	}

	hour, err := strconv.Atoi(t[0])
	if err != nil {
		return err
	}
	if hour < 0 || hour > 23 {
		return errors.New("wrong hour range")
	}

	minute, err := strconv.Atoi(t[1])
	if err != nil {
		return err
	}
	if minute < 0 || minute > 59 {
		return errors.New("wrong minute range")
	}

	return nil
}

func ParseCron(cron string, nd *NotifyData) error {
	crontab := strings.Split(cron, " ")
	if len(crontab) != 5 {
		return errors.New("wrong crontab")
	}

	var count int

	if crontab[0] != "*" && crontab[1] != "*" {
		m, err := strconv.Atoi(crontab[0])
		if err != nil {
			return err
		}
		if m < 0 || m > 59 {
			return errors.New("minute out of range")
		}
		h, err := strconv.Atoi(crontab[1])
		if err != nil {
			return err
		}
		if h < 0 || h > 23 {
			return errors.New("hour out of range")
		}
		nd.Time = crontab[1] + ":" + crontab[0]
		count++
	}

	if crontab[2] != "*" {

		d, err := strconv.Atoi(crontab[2])
		if err != nil {
			return err
		}
		if d < 1 || d > 31 {
			return errors.New("date out of range")
		}
		nd.Date = d
		count++
	} else {
		//nd.Date = -1
	}

	if crontab[3] != "*" {
		m, err := strconv.Atoi(crontab[3])
		if err != nil {
			return err
		}
		if m < 1 || m > 12 {
			return errors.New("month out of range")
		}
		nd.Month = m
		count++
	} else {
		//nd.Month = -1
	}

	if count > 0 {
		period := []string{DAY, MONTH, YEAR}
		nd.Cycle = period[count-1]
	}
	if crontab[4] != "*" {
		day, err := strconv.Atoi(crontab[4])
		if err != nil {
			return err
		}
		if day < 0 || day > 6 {
			return errors.New("day out of range")
		}
		nd.Day = day
		nd.Cycle = WEEK
	} else {
		//nd.Day = -1
	}
	return nil
}

func QueryNotify(query string, params ...interface{}) ([]*NotifyData, error) {
	rows, err := db.Query(query, params...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	nds := make([]*NotifyData, 0)
	for rows.Next() {
		var cron string
		var email string
		nd := &NotifyData{}
		err := rows.Scan(&nd.ID, &cron, &email)
		if err != nil {
			log.Println(err)
		} else {
			nd.Email = strings.Split(email, ",")
			if err = ParseCron(cron, nd); err != nil {
				log.Println(err)
				continue
			}
			nd.Day = nd.Day % 7
			nds = append(nds, nd)
		}

	}

	return nds, nil
}

func writeJSONResp(w http.ResponseWriter, resp interface{}) {
	b, _ := json.Marshal(resp)
	contentType := ContentTypeJSON
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func ValidateEmail(email string) bool {
	/*
		reg := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		return reg.MatchString(email)
	*/
	p := &mail.AddressParser{}
	_, err := p.Parse(email)
	if err != nil {
		return false
	}
	return true
}

func packageCronTask(id string, cron string, email []string, appid string, method string, period string) (*CronTask, error) {
	ct := &CronTask{}

	cd := &CronData{}

	cd.Cron = cron
	cd.Appid = appid
	cd.Method = method
	cd.Email = email
	cd.ID = id
	cd.Period = period

	b, err := json.Marshal(cd)
	if err != nil {
		return nil, err
	}

	ct.QueueName = QUEUEMAP["cronQueue"].Name
	ct.Task = b

	uuid := uuid.NewV4()
	corrID := hex.EncodeToString(uuid[:])
	ct.FileID = corrID

	return ct, nil
}

func sendCronTask(id string, cron string, email []string, appid string, method string, period string) error {
	reply := make(chan bool)
	cr, err := packageCronTask(id, cron, email, appid, method, period)
	if err != nil {
		return err
	}
	cr.Reply = reply
	defer close(reply)

	select {
	case CronQueue <- cr:
	case <-time.After(2 * time.Second):
		return errors.New("send cron task timeout")
	}

	select {
	case <-reply:
	case <-time.After(2 * time.Second):
		return errors.New("wait cron task ack timeout")
	}

	return nil
}

func parsePeriod(cron string) (string, error) {
	crons := strings.Split(cron, " ")
	if len(crons) != 5 {
		return "", errors.New("wrong crontab")
	}

	var stat int

	for i := 0; i < 3; i++ {
		if strings.Compare(crons[i+2], "*") != 0 {
			stat = stat | (1 << uint(i))
		}
	}
	var period string
	switch stat {
	case 0:
		period = DAY
	case 1:
		period = MONTH
	case 3:
		period = YEAR
	case 4:
		period = WEEK
	default:
		return "", errors.New("wrong crontab period")
	}

	return period, nil

}

//LoadCrontab load the crontask from database to the map structure
func LoadCrontab() (map[string]*CronEmail, error) {
	query := "select * from " + NotifyTable
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cronEmail := make(map[string]*CronEmail)
	nd := &NotifyData{}
	for rows.Next() {
		var id uint64
		var crontab, email, appid string

		if err = rows.Scan(&id, &crontab, &email, &appid); err != nil {
			return nil, err
		}

		emails := strings.Split(email, ",")

		for _, v := range emails {
			if !ValidateEmail(v) {
				log.Printf("[Warning] Invalid email in database at %v, %s, %s, %s\n", id, crontab, email, appid)
				continue
			}
		}

		if err = ParseCron(crontab, nd); err != nil {
			log.Printf("[Warning] Invalid crontab in database at %v, %s, %s, %s\n", id, crontab, email, appid)
			continue
		}
		period, err := parsePeriod(crontab)
		if err != nil {
			log.Printf("[Warning] Invalid crontab in database at %v, %s, %s, %s\n", id, crontab, email, appid)
			continue
		}

		key := appid + "-" + strconv.FormatUint(id, 10)
		ce := &CronEmail{Cron: crontab, Email: emails, Appid: appid, Period: period}
		cronEmail[key] = ce
	}

	return cronEmail, nil

}
