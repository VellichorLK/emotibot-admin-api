package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"handlers"
	"html/template"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron"
)

var envs = make(map[string]string)

//key: appid-reportID
var cronEmail map[string]*handlers.CronEmail

var crontab = cron.New()

func fakeEnv() {
	envs["RABBITMQ_HOST"] = "127.0.0.1"
	envs["RABBITMQ_PORT"] = "5672"
	envs["DB_HOST"] = "127.0.0.1"
	envs["DB_PORT"] = "3306"
	envs["DB_USER"] = "root"
	envs["DB_PWD"] = "password"
	envs["FILE_PREFIX"] = "/Users/public/Documents"
	envs["LISTEN_PORT"] = ":8080"
	envs["RABBITMQ_USER"] = "guest"
	envs["RABBITMQ_PWD"] = "guest"
}

type JobRunner struct {
	Email  []string
	Appid  string
	Period string
}

type QueryTime struct {
	From           int64
	To             int64
	LastFrom       int64
	LastTo         int64
	TmplName       string
	TmplStruct     interface{}
	DataTimeFormat string
}

func getQueryStruct(base *time.Time, period string) (*QueryTime, error) {
	qt := &QueryTime{}

	switch period {
	case handlers.DAY:
		rt, _ := handlers.RoundUpTime(base.Unix(), handlers.Day)
		t := time.Unix(rt, 0)
		t = t.AddDate(0, 0, -1)
		qt.TmplStruct = &Tmpl1{TargetDate: t.Format("2006/01/02")}
		qt.From = t.Unix()
		qt.To = rt - 1
		t = t.AddDate(0, 0, -1)
		qt.LastFrom = t.Unix()
		qt.LastTo = qt.From - 1
		qt.TmplName = "email-day.tmpl"
		qt.DataTimeFormat = "15:04"

	case handlers.WEEK:
		subDay := base.Weekday()
		baset := base.AddDate(0, 0, -int(subDay))
		rt, _ := handlers.RoundUpTime(baset.Unix(), handlers.Day)
		t := time.Unix(rt, 0)
		t = t.AddDate(0, 0, -7)

		qt.From = t.Unix()
		qt.To = rt - 1

		qt.TmplStruct = &Tmpl2{TargetDateFrom: t.Format("2006/01/02"), TargetDateTo: time.Unix(qt.To, 0).Format("2006/01/02")}

		t = t.AddDate(0, 0, -7)
		qt.LastFrom = t.Unix()
		qt.LastTo = qt.From - 1
		qt.TmplName = "email-week.tmpl"
		qt.DataTimeFormat = "2006/01/02"

	case handlers.MONTH:
		rt, _ := handlers.RoundUpTime(base.Unix(), handlers.Month)
		t := time.Unix(rt, 0)
		t = t.AddDate(0, -1, 0)
		qt.TmplStruct = &Tmpl1{TargetDate: t.Format("2006/01/02")}
		qt.From = t.Unix()
		qt.To = rt - 1
		t = t.AddDate(0, -1, 0)
		qt.LastFrom = t.Unix()
		qt.LastTo = qt.From - 1
		qt.TmplName = "email-month.tmpl"
		qt.DataTimeFormat = "2006/01/02"
	case handlers.YEAR:
		rt, _ := handlers.RoundUpTime(base.Unix(), handlers.Year)
		t := time.Unix(rt, 0)
		t = t.AddDate(-1, 0, 0)
		qt.TmplStruct = &Tmpl1{TargetDate: t.Format("2006/01")}
		qt.From = t.Unix()
		qt.To = rt - 1
		t = t.AddDate(-1, 0, 0)
		qt.LastFrom = t.Unix()
		qt.LastTo = qt.From - 1
		qt.TmplName = "email-year.tmpl"
		qt.DataTimeFormat = "2006/01"

	default:
		return nil, errors.New("wrong period of time:" + period)
	}

	//fmt.Printf("from:%s\nto:%s\nlastFrom:%s\nlastTo:%s\n", time.Unix(qt.From, 0), time.Unix(qt.To, 0), time.Unix(qt.LastFrom, 0), time.Unix(qt.LastTo, 0))

	return qt, nil
}

//Run used as interface for cron lib
func (jr *JobRunner) Run() {
	log.Printf("send %s with %s query, %s period\n", jr.Email, jr.Appid, jr.Period)

	base := time.Now()
	qt, err := getQueryStruct(&base, jr.Period)
	if err != nil {
		log.Println(err)
	} else {

		qas := &handlers.QueryArgs{T1: qt.From, T2: qt.To}

		timeUnit, err := handlers.GetTimeUnit(qas)
		if err != nil {
			log.Println(err)
			return
		}

		nowS, _, err := handlers.QueryStat(qas, timeUnit, jr.Appid)
		if err != nil {
			log.Println(err)
			return
		}

		qas.T1 = qt.LastFrom
		qas.T2 = qt.LastTo
		lastS, _, err := handlers.QueryStat(qas, timeUnit, jr.Appid)
		if err != nil {
			log.Println(err)
			return
		}

		err = FillTmplStruct(lastS, nowS, qt.TmplStruct, qt.DataTimeFormat)
		if err != nil {
			log.Println(err)
			return
		}

		tmpl, err := template.ParseFiles("./template/" + qt.TmplName)
		if err != nil {
			log.Println(err)
			return
		}

		var buf bytes.Buffer

		err = tmpl.Execute(&buf, qt.TmplStruct)
		if err != nil {
			log.Println("execute template failed: ", err)
			return
		}

		es := &handlers.EmailService{From: "voice_emotion_sys@emotibot.com",
			To:      jr.Email,
			Subject: "Voice Emotion Report",
			Body:    buf.String(),
			Content: "text/html; charset=\"UTF-8\"",
		}

		es.SendEmail()

		//fmt.Printf("%s\n", buf.String())

	}

}

var variableLists = [...]string{"RABBITMQ_HOST", "RABBITMQ_PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PWD", "RABBITMQ_USER", "RABBITMQ_PWD"}

func parseEnv() {
	for _, v := range variableLists {
		if os.Getenv(v) == "" {
			log.Fatalf("%s is empty!", v)
		}
		envs[v] = os.Getenv(v)
	}
}
func startService() {

	//err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "mydb")

	err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"])
	if err != nil {
		log.Fatal("Can't connect to database!!!")
	}

	cronEmail, err = handlers.LoadCrontab()
	if err != nil {
		log.Fatal("Load cron table failed")
	}

	addToCrontab()

	crontab.Start()
	port, err := strconv.Atoi(envs["RABBITMQ_PORT"])
	if err != nil {
		log.Fatalf("Can't conver  RABBITMQ_PORT(%s) to int!!!", envs["RABBITMQ_PORT"])
	}

	handlers.StartReceiveTaskService(envs["RABBITMQ_HOST"], port, envs["RABBITMQ_USER"], envs["RABBITMQ_PWD"], handlers.QUEUEMAP["cronQueue"].Name, updateCron)

	crontab.Stop()
}

func addToCrontab() {
	for k, v := range cronEmail {
		addCrontab(v, k)
		//log.Printf("add one cron tab %s,%s,%s\n", v.Cron, v.Appid, v.Email)
	}
}

func addCrontab(ce *handlers.CronEmail, key string) error {
	jr := &JobRunner{Appid: ce.Appid, Email: ce.Email, Period: ce.Period}
	err := crontab.AddNamedJob("0 "+ce.Cron, jr, cron.EntryID(key))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func updateCron(task string) (string, string) {
	cd := &handlers.CronData{}
	err := json.Unmarshal([]byte(task), cd)
	if err != nil {
		log.Println(err)
		return "", ""
	}

	key := cd.Appid + "-" + cd.ID
	ce := &handlers.CronEmail{Cron: cd.Cron, Email: cd.Email, Appid: cd.Appid, Period: cd.Period}
	var okEmail bool
	for _, v := range cd.Email {
		if okEmail = handlers.ValidateEmail(v); !okEmail {
			break
		}
	}
	var nd handlers.NotifyData
	okCron := handlers.ParseCron(cd.Cron, &nd)
	oce, ok := cronEmail[key]

	switch cd.Method {
	case "PUT":
		if ok {
			log.Printf("[Warning] appid (%s) with report_id (%v) has already existed\n", cd.Appid, cd.ID)
		} else {
			if !okEmail || okCron != nil {
				log.Printf("[Warning] wrong email (%s) or cron (%s)\n", cd.Email, cd.Cron)
			} else {
				cronEmail[key] = ce
				addCrontab(ce, key)
			}
		}
	case "DELETE":
		if ok {
			delete(cronEmail, key)
			crontab.RemoveJob(cron.EntryID(key))
		} else {
			log.Printf("[Warning] appid (%s) with report_id (%v) doesn't exist\n", cd.Appid, cd.ID)
		}
	case "PATCH":
		if ok {
			if okEmail {
				oce.Email = cd.Email
			}
			if okCron == nil {
				oce.Cron = cd.Cron
			}
			crontab.RemoveJob(cron.EntryID(key))
			addCrontab(oce, key)
		} else {
			log.Printf("[Warning] appid (%s) with report_id (%v) doesn't exist\n", cd.Appid, cd.ID)
		}
	default:
		log.Printf("[Warning ]receive wrong method (%s)!\n", cd.Method)
	}

	return "", ""
}

func main() {

	//fakeEnv()
	parseEnv()
	startService()
	/*
		fakeEnv()
		err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"])
		if err != nil {
			log.Fatal("Can't connect to database!!!")
		}

		jr := &JobRunner{Email: []string{"taylorchung@emotibot.com", "deansu@emotibot.com"}, Appid: "fakeappid", Period: "week"}
		jr.Run()
	*/
}
