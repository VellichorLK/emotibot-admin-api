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
	FileID         string   `json:"file_id"`
	FileName       string   `json:"file_name"`
	FileType       string   `json:"file_type"`
	Duration       uint32   `json:"duration"`
	Size           uint64   `json:"size"`
	CreateTime     uint64   `json:"created_time"`
	UploadTime     int64    `json:"upload_time"`
	Checksum       string   `json:"checksum"`
	Priority       uint8    `json:"-"`
	AnalysisResult int      `json:"analysis_result"`
	Tags           []string `json:"tags"`
}

//ReturnBlock is used for api /emotion/files/<file_id>
type ReturnBlock struct {
	BasicInfo
	Channels  []*ChannelResult `json:"channels"`
	UsrColumn []*ColumnValue   `json:"user_column,omitempty"`
}

type ColumnValue struct {
	Field string `json:"col_name,omitempty"`
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
}

//EmotionMap mapping the emotion_type to its name
var EmotionMap map[int]string

//emotion string
const (
	ENEUTRAL = "neutral"
	EANGER   = "anger"
)

//emotion number
const (
	NNERTRAL = 0
	NANGER   = 1
)

var DefaultEmotion = map[int]string{
	0: ENEUTRAL,
	1: EANGER,
}

//emotion type in string, and threshold
var DefaultRevertEmotion = map[string][]string{
	ENEUTRAL: {"0", "80"},
	EANGER:   {"1", "80"},
}

//PAGELIMIT limit of row returned
const PAGELIMIT = "500"
const PAGELIMITINT = 500

//ResultPage page the query result if the count is over defined limit
type ResultPage struct {
	Total int64 `json:"total"`
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

type CursorArgs struct {
	Count  int64      `json:"count"`
	Offset int64      `json:"offset"`
	Qas    *QueryArgs `json:"qas"`
}

//QueryArgs is used for /emotion/files
type QueryArgs struct {
	T1       int64  `json:"t1"`
	T2       int64  `json:"t2"`
	FileName string `json:"file_name"`
	Status   string `json:"status"`
	//Ch1Emotion []string `json:"ch1_emotions"`
	//Ch2Emotion []string `json:"ch2_emotions"`
	Ch1Anger  int64          `json:"ch1_anger_score"`
	Ch2Anger  int64          `json:"ch2_anger_score"`
	Tags      []string       `json:"tags"`
	UsrColumn []*ColumnValue `json:"user_column"`
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
}

type VoiceSegment struct {
	Status       int                    `json:"status"`
	Channel      int                    `json:"channel"`
	SegStartTime float64                `json:"segment_start_time"`
	SegEndTime   float64                `json:"segment_end_time"`
	ScoreList    []EmtionScore          `json:"scores_result"`
	ExtraInfo    map[string]interface{} `json:"extra_info"`
}
type EmtionScore struct {
	Label interface{} `json:"label"`
	Score float64     `json:"score"`
}

type Task struct {
	PackagedTask []byte
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
	Duration   int64   `json:"-"`
	FDuration  float64 `json:"duration"`
	UploadT    uint64  `json:"upload_time"`
	SUploadT   string  `json:"-"`
	ProcessST  uint64  `json:"-"`
	ProcessET  uint64  `json:"finished_time"`
	SProcessET string  `json:"-"`
}

type Statistics struct {
	TimeUnit    string      `json:"time_unit"`
	From        int64       `json:"from"`
	To          int64       `json:"to"`
	Count       uint64      `json:"total"`
	Ch1AvgAnger float64     `json:"ch1_avg_anger"`
	Ch2AvgAnger float64     `json:"ch2_avg_anger"`
	Duration    uint64      `json:"duration"`
	Data        []*StatUnit `json:"data"`
}
type StatUnit struct {
	From        int64   `json:"from"`
	To          int64   `json:"to"`
	Ch1AvgAnger float64 `json:"ch1_avg_anger"`
	Ch2AvgAnger float64 `json:"ch2_avg_anger"`
	Count       uint64  `json:"total"`
	Duration    uint64  `json:"duration"`
}

type TagOp struct {
	FileID string `json:"file_id"`
	Tag    string `json:"tag"`
	NewTag string `json:"new_tag"`
}

var QUEUEMAP = map[string]Queue{
	"taskQueue":   {Name: "ecovacasQueue", HasPriority: true},
	"resultQueue": {Name: "voiceResultQueue", HasPriority: false},
	"cronQueue":   {Name: "cronQueue", HasPriority: false},
}

//TaskQueue used for communicate between sender and new request
var TaskQueue = make(chan *Task)

//RelyQueue used for ack the sender that the task is pushed to queue successfully
var RelyQueue = make(chan bool)

//CronQueue used to send the crontab information when user updated it
var CronQueue = make(chan *CronTask)

//--------------------------- Database relative name ----------------------------------

//DataBase is database name
const DataBase = "voice_emotion"

//table name of database
const (
	MainTable            = "fileInformation"
	EmotionTable         = "emotionInformation"
	AnalysisTable        = "analysisInformation"
	ChannelTable         = "channelScore"
	EmotionMapTable      = "emotionMap"
	UserDefinedTagsTable = "userDefinedTags"
	NotifyTable          = "reportTo"
)

//column name of MainTable
const NID = "id"
const NFILEID = "file_id"
const NFILEPATH = "path"
const NFILENAME = "file_name"
const NFILETYPE = "file_type"
const NSIZE = "size"
const NDURATION = "duration"
const NFILET = "created_time"
const NCHECKSUM = "checksum"
const NPRIORITY = "priority"
const NAPPID = "appid"
const NANAST = "analysis_start_time"
const NANAET = "analysis_end_time"
const NANARES = "analysis_result"
const NUPT = "upload_time"
const NRDURATION = "real_duration"

//column name of AnalysisTable
const NSEGID = "segment_id"
const NSEGST = "segment_start_time"
const NSEGET = "segment_end_time"
const NCHANNEL = "channel"
const NSTATUS = "status"
const NEXTAINFO = "extra_info"

//column name of EmotionTable
const NEMOID = "emotion_id"
const NEMOTYPE = "emotion_type"
const NSCORE = "score"

//column name of UserDefinedTagsTable
const NTAGID = "defined_id"
const NTAG = "tag"

//column name of EmotionMapTable
const NEMOTION = "emotion"

//column name of ReportTo
const (
	NREPORTID = "report_id"
	NCRONTAB  = "crontab"
	NEMAIL    = "email"
)

//--------------------------------- header parameters -------------------------------------

//header parameter
const HXAPPID = "X-Appid"

//-------------------------------- API parameters -----------------------------------------

//API - /upload parameter
const QFILETYPE = "file_type"
const QDURATION = "duration"
const QFILET = "created_time"
const QCHECKSUM = "checksum"
const QFILE = "file"
const QTAGS = "tags"

//API - /files parameter
const QT1 = "t1"
const QT2 = "t2"

//API - /report parameter
const QEXPORT = "export"

//API - /tags
const QTAG = "tag"

//---------------------- report parameters  -------------------------------

//report name
const NFINT = "finished_time"
const NFROM = "from"
const NTO = "to"
const NTOTFILES = "total_files"
const NSUMD = "total_voice_duration"

const ReportLimitDay = 90 * 24 * 60 * 60
const TimeFormat = "2006/01/02 15:04:05"

// ------------------------ other parameters -------------------------------å

const ContentTypeJSON = "application/json; charset=utf-8"
const ContentTypeCSV = "text/csv; charset=utf-8"

const DEFAULTPRIORITY = 0

const LIMITUSERTAGS = 5
const LIMITVALUELEN = 64

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

const NUSRCOL = "user_column"
const LIMITUSRCOL = 3
const LIMITUSRSEL = 5

//user colum type
const (
	UsrColumString = iota
	UsrColumSel
)
