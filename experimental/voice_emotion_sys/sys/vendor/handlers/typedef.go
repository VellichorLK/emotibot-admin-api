package handlers

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/streadway/amqp"
)

type RabbitmqConnection struct {
	Host  string
	Port  int
	User  string
	Pwd   string
	Conn  *amqp.Connection
	Lock  Mutex
	Count uint32
}

type ConnectionPool struct {
	Conns   []*RabbitmqConnection
	Counter uint32
}

const MAXCHANNEL = 30000
const MAXCONNECTIONS = 2000

const mutexLocked = 1 << iota

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked)
}

//TaskBlock used to send to the rabbitmq as a task
type TaskBlock struct {
	Path string `json:"path"`
	File string `json:"file"`
	//Extension string `json:"extension"`
	//FileID    string `json:"fileID"`
}

//FileInfo record the file information at the time user upload
type FileInfo struct {
	ReturnBlock
	Path      string
	Appid     string
	UFileName string
	UPTime    int64
	ID        string
}

type BasicInfo struct {
	FileID         string `json:"file_id"`
	FileName       string `json:"file_name"`
	FileType       string `json:"file_type"`
	Duration       uint32 `json:"duration"`
	Size           uint64 `json:"size"`
	CreateTime     uint64 `json:"created_time"`
	UploadTime     uint64 `json:"upload_time"`
	Checksum       string `json:"checksum"`
	Tag            string `json:"tag1"`
	Tag2           string `json:"tag2"`
	Priority       uint8  `json:"-"`
	AnalysisResult int    `json:"analysis_result"`
}

//ReturnBlock is used for api /emotion/files/<file_id>
type ReturnBlock struct {
	BasicInfo
	Channels  []*ChannelResult `json:"channels"`
	UsrColumn []*ColumnValue   `json:"user_column,omitempty"`
}

type ColumnValue struct {
	Field string `json:"col_name"`
	Value string `json:"col_value"`
	ColID string `json:"col_id"`
}

type ChannelResult struct {
	ChannelID int            `json:"channel_id"`
	Result    []*EmtionScore `json:"result"`
}

type DetailReturnBlock struct {
	BasicInfo
	Channels  []*DetailChannelResult `json:"channels"`
	UsrColumn []*ColumnValue         `json:"user_column,omitempty"`
}
type DetailChannelResult struct {
	ChannelResult
	VadResults []*VadResult `json:"vad_result"`
}

type VadResult struct {
	Status       int                    `json:"status"`
	SegStartTime float64                `json:"segment_start_time"`
	SegEndTime   float64                `json:"segment_end_time"`
	ScoreList    []*EmtionScore         `json:"scores_result"`
	ExtraInfo    map[string]interface{} `json:"extra_info"`
	Text         string                 `json:"text"`
}

type vadInfo struct {
	Text string
}

//EmotionMap mapping the emotion_type to its name
var EmotionMap map[int]string

var DefaultEmotion = map[int]string{
	0: "neutral",
	1: "anger",
}

var AngerType = "1"

//PAGELIMIT limit of row returned
const PAGELIMIT = "500"
const PAGELIMITINT = 500

//ResultPage page the query result if the count is over defined limit
type ResultPage struct {
	Total int `json:"total"`
	//Offset  int            `json:"offset"`
	Blocks []*ReturnBlock `json:"result"`
	//HasMore bool           `json:"has_more"`
	Cursor string `json:"cursor"`
}

//QCursor continue query by given cursor
type QCursor struct {
	Cursor string `json:"cursor"`
}

//IndexJoin used to index the CursorFieldName
//after IndexJoin number is for another table
const IndexJoin = 8

//CursorFieldName record the field value for later used by db query
var CursorFieldName = [...]string{
	"count",    //record the last query total count
	"offset",   //record the current point
	NFILET,     //created_time >=
	NFILET,     //created_time <=
	NFILENAME,  //file_name =
	NTAG,       //tag =
	NTAG2,      //tag2=
	NANARES,    //status, wait analysis_reseult =-1, done analysis_result > 0
	NSCOREANG1, //min_score for channel one, anger >=
	NSCOREANG2, //min score for channel two, anger >=
}

//QueryArgs is used for /emotion/files
type QueryArgs struct {
	T1       string
	T2       string
	FileName string
	Status   string
	Tag      string
	Tag2     string
	Ch1Anger string
	Ch2Anger string
}

//EmotionBlock is used for decode the data from voice module
type EmotionBlock struct {
	Segments      []VoiceSegment `json:"analysis_details"`
	Result        int            `json:"analysis_result"`
	AnalysisStart uint64         `json:"analysis_start_time"`
	AnalysisEnd   uint64         `json:"analysis_end_time"`
	ID            string         `json:"src_primary_key"`
	IDUint64      uint64         `json:"-"`
	RDuration     uint64         `json:"src_voice_length"`
	AsrIdentity   string         `json:"asr_missionIdentity"`
}

type VoiceSegment struct {
	Status          int                    `json:"status"`
	Channel         int                    `json:"channel"`
	SegStartTime    float64                `json:"segment_start_time"`
	SegEndTime      float64                `json:"segment_end_time"`
	ScoreList       []EmtionScore          `json:"scores_result"`
	ExtraInfo       map[string]interface{} `json:"extra_info"`
	SegmentID       int64                  `json:"segment_id"`
	ID              string                 `json:"src_primary_key"`
	AsrFileLocation string                 `json:"asr_vad_relative_file_location"`
}
type EmtionScore struct {
	Label interface{} `json:"label"`
	Score float64     `json:"score"`
}

type Task struct {
	PackagedTask string
	FileInfo     *FileInfo
	QueueN       string
}

type Queue struct {
	Name        string
	HasPriority bool
}

type Report struct {
	From               int64        `json:"from"`
	To                 int64        `json:"to"`
	SFrom              string       `json:"-"`
	STo                string       `json:"-"`
	Count              uint64       `json:"total_files"`
	CumulativeDuration float64      `json:"total_voice_duration"`
	Records            []*ReportRow `json:"records"`
}

type ReportRow struct {
	FileName   string  `json:"file_name"`
	Tag1       string  `json:"tag1"`
	Tag2       string  `json:"tag2"`
	Duration   int64   `json:"-"`
	FDuration  float64 `json:"duration"`
	UploadT    uint64  `json:"upload_time"`
	SUploadT   string  `json:"-"`
	ProcessST  uint64  `json:"-"`
	ProcessET  uint64  `json:"finished_time"`
	SProcessET string  `json:"-"`
}

var QUEUEMAP = map[string]Queue{
	"fakeappid":    {Name: "ecovacasQueue", HasPriority: true},
	"resultQueue":  {Name: "voiceResultQueue", HasPriority: false},
	"asrTaskQueue": {Name: "asr_tasks_queue", HasPriority: false},
}

//TaskQueue used for communicate between sender and new request
var TaskQueue = make(chan *Task)

//RelyQueue used for ack the sender that the task is pushed to queue successfully
var RelyQueue = make(chan bool)

//MainTable the main table of file information
const MainTable = "fileInformation"
const EmotionTable = "emotionInformation"
const AnalysisTable = "analysisInformation"
const ChannelTable = "channelScore"
const EmotionMapTable = "emotionMap"
const VadInfoTable = "asr_analysisInformation"

const DEFAULTPRIORITY = 0
const LIMITTAGLEN = 128
const NVADRESULT = "result"
const NVADFILEID = "fileInformation_id"

//column name of main table
const NID = "id"
const NFILE = "file"
const NFILEID = "file_id"
const NFILENAME = "file_name"
const NFILETYPE = "file_type"
const NFILEPATH = "path"
const NSIZE = "size"
const NDURATION = "duration"
const NFILET = "created_time"
const NCHECKSUM = "checksum"
const NTAG = "tag1"
const NTAG2 = "tag2"
const NPRIORITY = "priority"
const NAPPID = "appid"
const NUAPPID = "X-Appid"
const NASYST = "analysis_time"
const NSCOREANG1 = "ch1_anger_score"
const NSCOREANG2 = "ch2_anger_score"
const NSCOREMAX = "max_score"
const NSCOREMEAN = "mean_score"
const NANAST = "analysis_start_time"
const NANAET = "analysis_end_time"
const NANARES = "analysis_result"
const NUPT = "upload_time"
const NRDURATION = "real_duration"

//column name of analysis table
const NSEGID = "segment_id"
const NSEGST = "segment_start_time"
const NSEGET = "segment_end_time"
const NCHANNEL = "channel"
const NSTATUS = "status"
const NEXTAINFO = "extra_info"

//column name of emotion table
const NEMOID = "emotion_id"
const NEMOTYPE = "emotion_type"
const NSCORE = "score"

//column name of emotion map table
const NEMOTION = "emotion"
const NAUTHORIZATION = "Authorization"

//emotion query args
const NT1 = "t1"
const NT2 = "t2"

//report name
const NEXPORT = "export"
const NFINT = "finished_time"
const NFROM = "from"
const NTO = "to"
const NTOTFILES = "total_files"
const NSUMD = "total_voice_duration"
const NFIELD = "field"
const NDATE = "date"

const ReportLimitDay = 90 * 24 * 60 * 60
const TimeFormat = "2006/01/02 15:04:05"

const ContentTypeJSON = "application/json; charset=utf-8"
const ContentTypeCSV = "text/csv; charset=utf-8"

//Name of User defined column table
const (
	UsrColTable    = "userColumn"
	UsrColValTable = "userColumnValue"
	UsrSelValTable = "userSelectableValue"
)

//field name of user column table
const (
	NCOLID    = "col_id"
	NCOLTYPE  = "col_type"
	NCOLNAME  = "col_name"
	NDEDAULT  = "default_value"
	NCOLVALID = "col_val_id"
	NCOLVAL   = "col_value"
	NSELID    = "sel_id"
	NSELVAL   = "sel_value"
)

type DailyReport struct {
	Total       int           `json:"total"`
	AvgEmotions []*AvgEmotion `json:"date"`
}
type AvgEmotion struct {
	Date        string  `json:"date_str"`
	Total       uint64  `json:"total"`
	AvgDuration uint64  `json:"avg_duration"`
	AvgCh1Anger float64 `json:"avg_ch1_angry"`
	AvgCh2Anger float64 `json:"avg_ch2_angry"`
}

/*
type GroupReport struct {
	Total int                `json:"total"`
	Group []*GroupAvgEmotion `json:"group"`
}
*/

type GroupReport struct {
	Total int             `json:"total"`
	Group []*GroupAvgData `json:"group"`
}

type GroupDurationReport struct {
	Total int                 `json:"total"`
	Group []*GroupAvgDuration `json:"group"`
}
type GroupsEmotion []*GroupAvgData

func (s GroupsEmotion) Len() int {
	return len(s)
}
func (s GroupsEmotion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s GroupsEmotion) Less(i, j int) bool {
	return s[i].Tag < s[j].Tag
}

type GroupsDuration []*GroupAvgDuration

func (s GroupsDuration) Len() int {
	return len(s)
}
func (s GroupsDuration) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s GroupsDuration) Less(i, j int) bool {
	return s[i].Tag < s[j].Tag
}

type GroupAvgData struct {
	Tag         string              `json:"tag_value"`
	AvgCh1Anger float64             `json:"avg_ch1_angry"`
	Ch1AngerR   float64             `json:"ch1_anger_r"`
	AvgCh2Anger float64             `json:"avg_ch2_angry"`
	Ch2AngerR   float64             `json:"ch2_anger_r"`
	Data        []*SimpleAvgEmotion `json:"data"`
}
type GroupAvgDuration struct {
	Tag         string               `json:"tag_value"`
	AvgDuration int                  `json:"avg_duration"`
	AvgDurR     float64              `json:"duration_r"`
	Data        []*SimpleAvgDuration `json:"data"`
}

/*
type GroupAvgEmotion struct {
	Tag  string              `json:"tag_value"`
	Data []*SimpleAvgEmotion `json:"data"`
}
*/
type SimpleAvgEmotion struct {
	Date        string  `json:"date_str"`
	AvgCh1Anger float64 `json:"avg_ch1_angry"`
	AvgCh2Anger float64 `json:"avg_ch2_angry"`
}

type SimpleAvgDuration struct {
	Date        string `json:"date_str"`
	AvgDuration int    `json:"avg_duration"`
}

type ScoreCount struct {
	score1  float64
	count1  uint64
	scores1 []float64
	score2  float64
	count2  uint64
	scores2 []float64
}

type WhereStates struct {
	name    string
	compare string
}
