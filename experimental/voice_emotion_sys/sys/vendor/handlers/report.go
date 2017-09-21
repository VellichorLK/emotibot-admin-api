package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"time"
)

type ReportCtx struct {
	Format      string
	FileName    string
	WriteToFile bool

	finishCtx func(*Report, bool, *os.File) ([]byte, string, error)
	r         *Report
	f         *os.File
}

func (rc *ReportCtx) InitReportCtx(format string, fileName string, writeToFile bool) error {
	rc.Format = format
	rc.FileName = fileName
	rc.WriteToFile = writeToFile
	rc.r = new(Report)
	rc.r.Records = make([]*ReportRow, 0)
	switch rc.Format {
	case "":
		fallthrough
	case "json":
		rc.finishCtx = finishJSONReport
	case "csv":
		rc.finishCtx = finishCSVReport
	default:
		return errors.New("Wrong export format")
	}

	if rc.WriteToFile {
		err := rc.openReport()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (r *ReportCtx) CloseCtx() error {
	if r.f != nil {
		err := r.f.Close()
		if err != nil {
			log.Println(err)
		}
		/*
			err = os.Remove(r.FileName)
			if err != nil {
				log.Println(err)
			}
		*/
		return err
	}
	return nil
}

func (rc *ReportCtx) PutHeader(m map[string]int64) error {

	rc.r.From = m[NFROM]
	rc.r.To = m[NTO]

	t1T := time.Unix(rc.r.From, 0)
	t2T := time.Unix(rc.r.To, 0)

	rc.r.SFrom = t1T.Format(TimeFormat)
	rc.r.STo = t2T.Format(TimeFormat)
	return nil
	//return r.putHeader(m, r.f)
}

func (rc *ReportCtx) PutRecord(rr *ReportRow) error {
	_rr := new(ReportRow)
	*_rr = *rr
	rc.r.Records = append(rc.r.Records, _rr)
	return nil
}

func (rc *ReportCtx) FinishReport() ([]byte, string, error) {

	if rc.finishCtx != nil {
		return rc.finishCtx(rc.r, rc.WriteToFile, rc.f)
	}
	return nil, "", errors.New("No finish method assigned")
}

func (r *ReportCtx) openReport() error {
	if r.f != nil {
		return errors.New("File has already been opened")
	}
	if r.FileName == "" {
		return errors.New("No file name specified")
	}

	var err error
	r.f, err = os.Create(r.FileName)
	if err != nil {
		return errors.New("Open file error. Please check your appid is valid")
	}
	return nil
}

func finishJSONReport(r *Report, writeToFile bool, f *os.File) ([]byte, string, error) {

	for _, v := range r.Records {
		//duration := float64(v.Duration) / 1000
		duration := float64(v.Duration)
		r.CumulativeDuration += duration
		r.Count++

		v.FDuration = duration

		t := time.Unix(int64(v.UploadT), 0)
		v.SUploadT = t.Format(TimeFormat)

		t = time.Unix(int64(v.ProcessET)/1000, 0)
		v.SProcessET = t.Format(TimeFormat)
		v.ProcessET = v.ProcessET / 1000
	}

	pre := 1000.0

	r.CumulativeDuration = float64(int(r.CumulativeDuration*pre)) / pre

	b, err := json.Marshal(r)
	if err != nil {
		return nil, "", err
	}
	if writeToFile {
		_, err = f.Write(b)
		return nil, "", err
	}
	return b, ContentTypeJSON, nil
}

func finishCSVReport(r *Report, writeToFile bool, f *os.File) ([]byte, string, error) {
	//ctx := NFROM + "," + r.SFrom + "\n"
	//ctx += NTO + "," + r.STo + "\n"

	ctx := NFROM + "," + strconv.FormatInt(r.From, 10) + "\n"
	ctx += NTO + "," + strconv.FormatInt(r.To, 10) + "\n"

	ctx += "#," + NFILENAME + "," + NDURATION + "," + NTAG + "," + NTAG2 + "," + NUPT + "," + NFINT + "\n"

	for i := 0; i < len(r.Records); i++ {
		r.Count++
		row, err := makeCSVRow(r, i)
		if err != nil {
			return nil, "", err
		}
		ctx += row
	}

	ctx += NTOTFILES + "," + strconv.FormatUint(r.Count, 10) + "\n"
	ctx += NSUMD + "," + strconv.FormatFloat(r.CumulativeDuration, 'f', 3, 64) + "\n"

	if writeToFile {
		_, err := f.WriteString(ctx)
		return nil, "", err
	}

	return []byte(ctx), ContentTypeCSV, nil
}

func makeCSVRow(r *Report, idx int) (string, error) {
	row := strconv.FormatUint(r.Count, 10) + ","
	row += r.Records[idx].FileName + ","
	duration := float64(r.Records[idx].Duration) / 1000
	row += strconv.FormatFloat(duration, 'f', 3, 64) + ","
	row += r.Records[idx].Tag1 + ","
	row += r.Records[idx].Tag2 + ","

	//t := time.Unix(int64(r.Records[idx].UploadT), 0)
	//row += t.Format(TimeFormat) + ","
	row += strconv.FormatUint(r.Records[idx].UploadT, 10) + ","

	//t = time.Unix(int64(r.Records[idx].ProcessET)/1000, 0)
	//row += t.Format(TimeFormat) + "\n"
	r.Records[idx].ProcessET = r.Records[idx].ProcessET / 1000
	row += strconv.FormatUint(r.Records[idx].ProcessET, 10) + "\n"

	r.CumulativeDuration += duration
	return row, nil
}
