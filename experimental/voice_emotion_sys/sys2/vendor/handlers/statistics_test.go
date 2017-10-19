package handlers
/*
import (
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

func TestStatistic(t *testing.T) {
	src := "root:password@tcp(127.0.0.1:3306)/" + DataBase
	db, err := sql.Open("mysql", src)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	insertMainTmpl := "insert into " + MainTable + " (" +
		NFILEID + "," + NFILEPATH + "," + NFILENAME + "," + NFILETYPE + "," + NSIZE + "," +
		NDURATION + "," + NFILET + "," + NPRIORITY + "," + NAPPID + "," + NUPT + "," +
		NRDURATION +
		") values (?,?,?,?,?,?,?,?,?,?,?)"

	insertChanTmpl := "insert into " + ChannelTable + " (" +
		NID + "," + NCHANNEL + "," + NEMOTYPE + "," + NSCORE +
		") values(?,?,?,?)"

	const max = 6
	ch1Neu := [max]float64{13.34, 57.91, 20.19, 45.17, 55.88, 99.01}
	ch1Ang := [max]float64{40.99, 98.1, 78.3, 67.31, 78.77, 10.3}

	ch2Neu := [max]float64{14.3, 47.9, 87.3, 57.1, 49.3, 89.1}
	ch2Ang := [max]float64{87.1, 68.3, 29.1, 49.2, 87.88, 12.2}

	base := time.Unix(1507501283, 0)

	mainStmt, err := db.Prepare(insertMainTmpl)
	if err != nil {
		t.Error(err)
	}
	defer mainStmt.Close()

	chanStmt, err := db.Prepare(insertChanTmpl)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 5; i++ {

		uuid := uuid.NewV4()
		corrID := hex.EncodeToString(uuid[:])

		res, err := mainStmt.Exec(corrID, "/home/testpaath", "twst.wav", "wav", 312, 33, base.Unix(), 0, "testappid", base.Unix(), 33000)
		if err != nil {
			t.Error(err)
		}
		id, _ := res.LastInsertId()

		chanStmt.Exec(id, 1, 0, ch1Neu[i%max])
		chanStmt.Exec(id, 1, 1, ch1Ang[i%max])
		chanStmt.Exec(id, 2, 0, ch2Neu[i%max])
		chanStmt.Exec(id, 2, 1, ch2Ang[i%max])

		//base = base.AddDate(0, 0, 1)
		base = base.Add(time.Hour)
	}

}
*/