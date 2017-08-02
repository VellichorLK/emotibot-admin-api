package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"handlers"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const v1basePath = "/voice/emotion/v1"

const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_Y = "\x1b[33;1m"
const CLR_B = "\x1b[34;1m"
const CLR_N = "\x1b[0m"

type Profile struct {
	url         string
	concurrency int
	period      int
	file        string
	config      string
	params      map[string]string
	headers     map[string]string
	insecure    bool
	fileType    string
	duration    int
	mode        string
	save        string
	createT     int64
	tag1        string
	tag2        string
}

var uploadSuccessCount uint64
var totalUploadSuccessCount uint64
var uploadFailCount uint64
var finishedCount uint64
var totalFinishedCount uint64

func failOnError(err error) {
	if err != nil {
		log.Println(err)
		panic(1)
	}
}

func checkDetail(p *Profile, appidChan chan string, drb *handlers.DetailReturnBlock) {

	detailURL := p.url + v1basePath + "/files/"
	var mydrb *handlers.DetailReturnBlock
	if drb != nil {
		mydrb = drb
	} else {
		mydrb = new(handlers.DetailReturnBlock)
	}

	for {
		appid := <-appidChan
		if appid == "exit" {
			break
		}

		url := detailURL + appid

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			panic(1)
		}

		for k, v := range p.headers {
			req.Header.Set(k, v)
		}

		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}

		client := &http.Client{Transport: transCfg}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			select {
			case appidChan <- appid:
			case <-time.After(500 * time.Millisecond):
				log.Printf("[Warning] waiting task queue is full, drop %s\n", appid)
			}
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			json.Unmarshal(body, mydrb)
			//unfinished

			if mydrb.AnalysisResult == -1 {
				select {
				case appidChan <- appid:
				case <-time.After(500 * time.Millisecond):
					log.Printf("[Warning] waiting task queue is full, drop %s\n", appid)
				}
			} else {
				atomic.AddUint64(&finishedCount, 1)
				atomic.AddUint64(&totalFinishedCount, 1)
			}

		} else {
			log.Printf("[%d]%s\n", resp.StatusCode, string(body[:]))
		}
	}

	appidChan <- "ok"

}

func makeUploadRequest(url string, params map[string]string, headers map[string]string, fileKey string, file string) (*http.Request, error) {
	body := &bytes.Buffer{}

	filePathName := strings.Split(file, "/")

	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(handlers.NFILE, filePathName[len(filePathName)-1])
	failOnError(err)

	reader, err := os.Open(file)
	failOnError(err)
	defer reader.Close()

	_, err = io.Copy(part, reader)
	failOnError(err)

	for k, v := range params {
		err := writer.WriteField(k, v)
		if err != nil {
			log.Println(err)
		}
	}

	err = writer.Close()
	failOnError(err)

	req, err := http.NewRequest("POST", url, body)
	failOnError(err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, err
}

func doUpload(url string, params map[string]string, headers map[string]string, fileKey string, file string) ([]byte, error) {
	req, err := makeUploadRequest(url, params, headers, handlers.NFILE, file)
	if err != nil {
		return nil, err
	}

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}

	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		atomic.AddUint64(&uploadSuccessCount, 1)
		atomic.AddUint64(&totalUploadSuccessCount, 1)

	} else {
		atomic.AddUint64(&uploadFailCount, 1)
		log.Printf("[%d]%s\n", resp.StatusCode, body)
	}

	return body, nil
}

func continueUpload(exit chan bool, appidChan chan string, url string, params map[string]string, headers map[string]string, fileKey string, file string) {

loop:
	for {
		select {
		case <-exit:
			break loop
		default:
			body, err := doUpload(url, params, headers, fileKey, file)
			if err != nil {
				atomic.AddUint64(&uploadFailCount, 1)
				log.Println(err)
			} else if appidChan != nil {
				rb := new(handlers.ReturnBlock)
				err := json.Unmarshal(body, rb)
				if err != nil {
					log.Println(string(body[:]))
					log.Println(err)
				} else {
					appidChan <- rb.FileID
				}
			}
		}

	}
	exit <- true
}

func printData(period int) {
	for {
		select {
		case <-time.After(time.Duration(period) * time.Second):
			sc := atomic.LoadUint64(&uploadSuccessCount)
			atomic.StoreUint64(&uploadSuccessCount, 0)
			fc := atomic.LoadUint64(&uploadFailCount)
			atomic.StoreUint64(&uploadFailCount, 0)
			fic := atomic.LoadUint64(&finishedCount)
			atomic.StoreUint64(&finishedCount, 0)
			tfic := atomic.LoadUint64(&totalFinishedCount)

			tsc := atomic.LoadUint64(&totalUploadSuccessCount)

			log.Printf("[upload_success]:%v, [upload_fail]:%v, [finised]:%v, [total_unfinished]:%v,[total_finished]:%v\n", sc, fc, fic, tsc-tfic, tfic)
		}
	}
}

func stressTest(p *Profile) {
	if p.file == "" && p.config == "" {
		log.Println("No specify file or config")
		flag.Usage()
		usage()
		return
	}

	uploadURL := p.url + v1basePath + "/upload"
	exitChans := make([]chan bool, p.concurrency)
	appidChan := make(chan string, 99999)

	go printData(1)
	go checkDetail(p, appidChan, nil)

	for i := 0; i < p.concurrency; i++ {
		exitChans[i] = make(chan bool)
		go continueUpload(exitChans[i], appidChan, uploadURL, p.params, p.headers, handlers.NFILE, p.file)
	}

	<-time.After(time.Duration(p.period) * time.Second)

}

func makeProfile() (*Profile, error) {

	p := new(Profile)

	//flag.BoolVar(&p.insecure, "k", false, "do not verify the authentication pem")
	flag.IntVar(&p.concurrency, "c", 1, "concurrency, number of thread used to stress")
	flag.IntVar(&p.period, "p", 300, "second to continuely send the task")
	flag.StringVar(&p.url, "h", "https://api-sh.emotibot.com", "url to send the task to.")
	flag.StringVar(&p.file, "f", "", "the file to upload to server")

	//flag.StringVar(&p.config, "g", "", "config file to use. Currently not supported")
	flag.StringVar(&p.fileType, "t", "wav", "file extension")
	now := time.Now()
	flag.Int64Var(&p.createT, "r", now.Unix(), "create_time of file")
	flag.IntVar(&p.duration, "d", 143, "period of voice in second")
	flag.StringVar(&p.mode, "m", "single", "stress test or test single file")
	flag.StringVar(&p.save, "s", "", "save the resutl in json format to the assigned file. Only used at single mode")
	flag.StringVar(&p.tag1, "t1", "", "tag1")
	flag.StringVar(&p.tag2, "t2", "", "tag2")
	kvHeader := flag.String("a", handlers.NAUTHORIZATION+":testappid", "authentication in header")
	flag.Parse()

	//body parameter
	p.params = make(map[string]string)
	p.params[handlers.NFILETYPE] = p.fileType
	p.params[handlers.NDURATION] = strconv.Itoa(p.duration)
	p.params[handlers.NFILET] = strconv.FormatInt(p.createT, 10)
	p.params[handlers.NTAG] = p.tag1
	p.params[handlers.NTAG2] = p.tag2

	//header parameter
	p.headers = make(map[string]string)
	kvs := strings.Split(*kvHeader, ":")
	if len(kvs) != 2 {
		return nil, errors.New("headers authentication wrong format")

	}
	p.headers[kvs[0]] = kvs[1]
	return p, nil
}

func correctnessTest(p *Profile) {

	appidChan := make(chan string, 1)
	drb := new(handlers.DetailReturnBlock)

	if p.file == "" {
		log.Println("No specified test file")
		flag.Usage()
		usage()
		return
	}
	//upload file
	uploadURL := p.url + v1basePath + "/upload"
	body, err := doUpload(uploadURL, p.params, p.headers, handlers.NFILE, p.file)
	if err != nil {
		log.Println(err)
		return
	}

	//parse return block
	rb := new(handlers.ReturnBlock)
	err = json.Unmarshal(body, rb)
	if err != nil {
		log.Println(string(body[:]))
		log.Println(err)
		return
	}

	go checkDetail(p, appidChan, drb)

	appidChan <- rb.FileID

	for {
		c := atomic.LoadUint64(&totalFinishedCount)
		if c == 1 {
			break
		}
		<-time.After(1 * time.Second)
		log.Printf("waiting...\n")
	}

	appidChan <- "exit"
	<-appidChan

	//log.Println("[result]:%d, [vad]:%d\n", len(drb.Channels))

	for i := 0; i < len(drb.Channels); i++ {
		log.Printf("channel:%d\n", drb.Channels[i].ChannelID)
		log.Println("\tresult:")
		for j := 0; j < len(drb.Channels[i].Result); j++ {
			log.Printf("\t\t[%s]:%v\n", drb.Channels[i].Result[j].Label, drb.Channels[i].Result[j].Score)
		}
		log.Println("\tvad_result:")
		for j := 0; j < len(drb.Channels[i].VadResults); j++ {
			log.Printf("\t\tduration:%v~%v\n", drb.Channels[i].VadResults[j].SegStartTime, drb.Channels[i].VadResults[j].SegEndTime)
			log.Printf("\t\tvad_status:%v\n", drb.Channels[i].VadResults[j].Status)
			log.Printf("\t\textra_info:%v\n", drb.Channels[i].VadResults[j].ExtraInfo)

			for k := 0; k < len(drb.Channels[i].VadResults[j].ScoreList); k++ {
				log.Printf("\t\t\t[%s]:%v\n", drb.Channels[i].VadResults[j].ScoreList[k].Label, drb.Channels[i].VadResults[j].ScoreList[k].Score)
			}

		}
	}

	if drb.AnalysisResult == 1 {
		log.Printf("[test_result]:%spass%s\n", CLR_G, CLR_N)
	} else {
		log.Printf("[test_result]:%serror%s\n", CLR_R, CLR_N)
	}

	if p.save != "" {
		f, err := os.Create(p.save)
		if err != nil {
			log.Println(err)
		} else {
			defer f.Close()
			d, err := json.Marshal(drb)
			if err != nil {
				log.Println(err)
			} else {
				f.WriteString(string(d[:]))
			}
		}

	}
}

func usage() {
	fmt.Printf("single test: ./tester -f xxx.wav -a Authorization:<appid> -h https://api-sh.emotibot.com\n")
	fmt.Printf("stress test: ./tester -c 10 -p 300 -f xxx.wav -a Authorization:<appid> -h https://api-sh.emotibot.com -m stress \n")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	p, err := makeProfile()
	failOnError(err)

	switch p.mode {
	case "single":
		correctnessTest(p)
	case "stress":
		stressTest(p)
	}

}
