package imagesManager

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

//InitDaemon should be called by server.go in vipshop-admin only
func InitDaemon() {
	periodStr := util.GetEnviroment(Envs, "SYNC_PERIOD_BY_SECONDS")
	period, err := strconv.Atoi(periodStr)
	if err != nil {
		// panic("Env SYNC_PERIOD_BY_SECONDS parsing error" + err.Error())
	}
	DefaultDaemon = NewDaemon(period, util.GetMainDB(), db)
	go DefaultDaemon.Sync()
}

//DefaultDaemon is the daemon that running at the start of package
var DefaultDaemon *Daemon

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
func (d *Daemon) Sync() {
	for {
		//TODO: Calibrate sleeping behavior at first time.
		time.Sleep(d.UpdatePeriod)
		util.LogTrace.Println("Start daemon syncing...")
		answers, err := d.FindImages()
		if err != nil {
			util.LogError.Printf("find image tag in answer failed, %v\n", err)
			continue
		}
		count, err := LinkImagesForAnswer(answers)
		if err != nil {
			util.LogError.Printf("Link image for answers failed, %v. \n", err)
			continue
		}
		util.LogTrace.Printf("daemon sync finished, num of %d inserted.\n", count)
	}

}

// FindImages scan answer's content and match the image tag in it
// Return Empty map if none of given image's file name is matched
func (d *Daemon) FindImages() (answers map[int][]int, err error) {
	rows, err := d.questionDB.Query("SELECT Answer_Id, Content FROM vipshop_answer WHERE Status = 1")
	if err != nil {
		return nil, fmt.Errorf("Query answer failed, %v", err)
	}
	defer rows.Close()
	rawQuery := "SELECT id FROM images WHERE fileName = ?"
	stmt, err := d.picDB.Prepare(rawQuery)
	defer stmt.Close()
	if err != nil {
		return nil, fmt.Errorf("sql prepared failed, %v", err)
	}
	answers = make(map[int][]int)
	for rows.Next() {
		var (
			id      int
			content string
		)
		err = rows.Scan(&id, &content)
		if err != nil {
			return nil, fmt.Errorf("scan failed, %v", err)
		}

		var imageGroup = make([]int, 0)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
				fmt.Println(strings.Index(content, "<img"))
				fmt.Println(content)
			}
		}()
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
				return nil, fmt.Errorf("sql queryRow failed, %v", err)
			}

			imageGroup = append(imageGroup, id)
		}
		//Only the answers with images will be added to map
		if len(imageGroup) > 0 {
			answers[id] = imageGroup
		}

	}

	return

}

// LinkImagesForAnswer will insert rows into middle table of image and answer.
// It have to clean up image_answer table first, because it is no way to sync the old info.
// Return count num of rows it have insert into image_answer table.
func LinkImagesForAnswer(answerImages map[int][]int) (int, error) {
	//count is a counter for how many rows we totally write.
	var count = 0
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("get transaction failed, %v", err)
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
		return 0, fmt.Errorf("clean up image_answer table failed, %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO image_answer (answer_id, image_id) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("sql prepare failed, %v", err)
	}
	for ansID, imgID := range answerImages {
		_, err := stmt.Exec(ansID, imgID)
		if err != nil {
			return 0, fmt.Errorf("sql insert failed, %v", err)
		}
		count++
	}

	return count, nil
}
