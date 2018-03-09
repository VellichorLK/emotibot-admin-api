package imagesManager

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/hashicorp/consul/api"
)

//InitDaemon should be called by server.go in vipshop-admin only
func InitDaemon() {
	periodStr := util.GetEnviroment(Envs, "SYNC_PERIOD_BY_SECONDS")
	period, err := strconv.Atoi(periodStr)
	if err != nil {
		// panic("Env SYNC_PERIOD_BY_SECONDS parsing error" + err.Error())
	}
	DefaultDaemon = NewDaemon(period, db, util.GetMainDB())
	go func() {
		for {
			//TODO: Calibrate sleeping behavior at first time.
			util.LogTrace.Printf("sleep %s", DefaultDaemon.UpdatePeriod.String())
			time.Sleep(DefaultDaemon.UpdatePeriod)

			lock, err := util.DefaultConsulClient.Lock(daemonKey)
			if err == api.ErrLockHeld {
				util.LogError.Println("Consul lock acquired by other admin-api, give up this time")
				continue
			} else if err != nil {
				util.LogError.Printf("acquiring consul lock failed, %v\n", err)
				continue
			}
			stop, err := lock.Lock(make(chan struct{}))
			if err != nil {
				util.LogError.Printf("lock acquiring failed, %v\n", err)
			}

			err = DefaultDaemon.Sync(stop)
			if err != nil {
				util.LogError.Println("sync failed, " + err.Error())
			}
		}
	}()
}

//DefaultDaemon is the daemon that running at the start of package
var DefaultDaemon *Daemon

const (
	daemonKey = "imageSync"
)

// Daemon will run in background to sync
type Daemon struct {
	picDB        *sql.DB
	questionDB   *sql.DB
	UpdatePeriod time.Duration
}

func NewDaemon(updatePeriod int, pictureDB, questionDB *sql.DB) *Daemon {
	return &Daemon{
		picDB:        pictureDB,
		questionDB:   questionDB,
		UpdatePeriod: time.Duration(updatePeriod) * time.Second,
	}

}

//Sync will make sure imageDB and questionDB have data consistency every UpdatePeriod seconds.
func (d *Daemon) Sync(stop <-chan struct{}) error {
	util.LogTrace.Println("Start daemon syncing...")
	findJob := &FindImageJob{
		questionDB: d.questionDB,
		picDB:      d.picDB,
	}

	err := findJob.Do(stop)
	if err != nil {
		return fmt.Errorf("find image tag in answer failed, %v", err)
	}
	linkJob := &LinkImageJob{
		input: findJob.Result,
	}

	err = linkJob.Do(stop)
	if err != nil {
		return fmt.Errorf("link image for answers failed, %v", err)
	}
	util.LogTrace.Printf("daemon sync finished, num of %d inserted.\n", linkJob.AffecttedRows)
	return nil
}

// interruptableJob can be execute by Do, return nil if work is done normally.
// It also may be cancled by other factor
// So It should return ErrInterrutted if job are cancelled
type interruptableJob interface {
	Do(signal <-chan struct{}) error
}

//ErrInterruptted should be return when interruptableJob are interruptted
var ErrInterruptted = errors.New("interruptted error")

// FindImagesJob is the job unitwill scan answer's content and match the image tag in it
type FindImageJob struct {
	Result     map[int][]int
	questionDB *sql.DB
	picDB      *sql.DB
}

// Do FindImagesJob will scan answer's content and match the image tag in it
// Return Empty map if none of given image's file name is matched
func (j *FindImageJob) Do(signal <-chan struct{}) error {
	rows, err := j.questionDB.Query("SELECT Answer_Id, Content FROM vipshop_answer WHERE Status = 1")
	if err != nil {
		return fmt.Errorf("Query answer failed, %v", err)
	}
	defer rows.Close()
	rawQuery := "SELECT id FROM images WHERE fileName = ?"
	stmt, err := j.picDB.Prepare(rawQuery)
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("sql prepared failed, %v", err)
	}
	j.Result = make(map[int][]int)
	for rows.Next() {
		var (
			id      int
			content string
		)
		err = rows.Scan(&id, &content)
		if err != nil {
			return fmt.Errorf("scan failed, %v", err)
		}

		var imageGroup = make([]int, 0)

		//Find img tag and trim the content
		for currentIndex := strings.Index(content, "<img"); currentIndex >= 0; currentIndex = strings.Index(content, "<img") {
			content = content[currentIndex+4:]
			//Find src attribute
			var start, end int
			start = strings.Index(content, "src=\"") + 5
			if start == -1 || len(content) < start {
				//Skip the bad formatted img tag to avoid runtime error
				continue
			}
			for i, char := range content[start:] {
				if char == '"' {
					end = start + i
					break
				}
			}
			//Find image's ID
			srcContent := strings.TrimPrefix(content[start:end], "http://")
			names := strings.Split(srcContent, "/")
			name := names[len(names)-1]
			var id int
			err := stmt.QueryRow(name).Scan(&id)
			if err == sql.ErrNoRows {
				//Skip the unknown img tag
				continue
			} else if err != nil {
				return fmt.Errorf("sql queryRow failed, %v", err)
			}

			imageGroup = append(imageGroup, id)
		}
		//Only the answers with images will be added to map
		if len(imageGroup) > 0 {
			j.Result[id] = imageGroup
		}

	}

	return nil
}

// LinkImageJob is
type LinkImageJob struct {
	IsDone        bool
	AffecttedRows int //AffecttedRows count how many rows we totally write.
	input         map[int][]int
}

// Do LinkImageJob will insert rows into middle table of image and answer.
// It have to clean up image_answer table first, because it is no way to sync the old info.
// Return count num of rows it have insert into image_answer table.
func (j *LinkImageJob) Do(signal <-chan struct{}) error {

	j.AffecttedRows = 0
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("get transaction failed, %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	_, err = tx.Exec("DELETE FROM image_answer")
	if err != nil {
		return fmt.Errorf("clean up image_answer table failed, %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO image_answer (answer_id, image_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("sql prepare failed, %v", err)
	}
	for ansID, images := range j.input {
		for _, imgID := range images {
			_, err := stmt.Exec(ansID, imgID)
			if err != nil {
				return fmt.Errorf("sql insert failed, %v", err)
			}
			j.AffecttedRows++
		}

	}
	return nil
}
