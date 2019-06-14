package lqcheck

// 冲突检测标准问语料对
type ConflictCheckSqLq struct {
	// 语料
	Lq string `json:"lq"`
	// 标准问
	Sq string `json:"sq"`
}

// 冲突检测请求
type ConflictCheckRequest struct {
	AppId  string              `json:"appid"`
	Locale string              `json:"locale"`
	Data   []ConflictCheckSqLq `json:"data"`
}

// 冲突检测任务
type ConflictCheckResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	TaskId string `json:"task_id"`
	QSize  int    `json:"q_size"`
}

// 冲突检测返回
type ConflictCheckReturn struct {
	TaskId string `json:"task_id"`
}

// ssm3 dac 返回结构
type SsmDacRet struct {
	ActualResults []SsmDacSqLq `json:"actualResults"`
	Errno         string       `json:"errno"`
}

// ssm3 dac 语料返回
type SsmDacSqLq struct {
	AppId     string `json:"app_id"`
	LqContent string `json:"lq_content"`
	LqId      int64  `json:"lq_id"`
	SqContent string `json:"sq_content"`
	SqId      int64  `json:"sq_id"`
}

type SsmDacCheckRet struct {
	ActualResults SsmDacCheckData `json:"actualResults"`
	Errno         string          `json:"errno"`
}

type SsmDacCheckData struct {
	Needcheck bool  `json:"needcheck"`
	Time      int64 `json:"time"`
}

// 返回标准问语料列表
type ReportSq struct {
	SqId    int64      `json:"sq_id"`
	Sq      string     `json:"sq"`
	LqCount int        `json:"lq_count"`
	Lq      []ReportLq `json:"lq"`
}

type ReportLq struct {
	LqId int64  `json:"lq_id"`
	Lq   string `json:"lq"`
}

// 整体健康度
type HealthScore struct {
	Score    string              `json:"score"`
	Standard HealthScoreStandard `json:"standard"`
}

type HealthScoreStandard struct {
	Label []string `json:"label"`
	Score []int    `json:"score"`
}

// 语料质量
type LqQuality struct {
	LqCount        int    `json:"lq_count"`
	Bad            int    `json:"bad"`
	Recommended    int    `json:"recommended"`
	ReportFilePath string `json:"report_file_path"`
}

// 推荐标准问语料比
type LqSqRateRecommended struct {
	LqCount  int     `json:"lq_count"`
	SqCount  int     `json:"sq_count"`
	SqLqRate float32 `json:"sq_lq_rate"`
}

// 标准问语料比
type LqSqRate struct {
	LqCount        int      `json:"lq_count"`
	SqCount        int      `json:"sq_count"`
	LqRate         float64  `json:"lq_rate"`
	SqLqRate       string   `json:"sq_lq_rate"`
	LqSqRateRemark []string `json:"remark"`
}

// 语料数量分布
type LqDistributionTemplate struct {
	Label       string  `json:"label"`
	From        int     `json:"from,omitempty"`
	To          int     `json:"to,omitempty"`
	SqNum       int     `json:"sq_num"`
	SqRate      float64 `json:"sq_rate"`
	SqRateScore float64 `json:"sq_rate_score"`
}

type LqDistribution struct {
	Current     []LqDistributionTemplate `json:"current"`
	Recommended []LqDistributionTemplate `json:"recommended"`
}

// 报告
type HealthReport struct {
	Recheck            bool           `json:"recheck"`
	LastCheckTime      string         `json:"last_check_time"`
	HealthScore        HealthScore    `json:"health_score"`
	LqQuality          LqQuality      `json:"lq_quality"`
	LqSqRate           LqSqRate       `json:"lq_sq_rate"`
	LqDistribution     LqDistribution `json:"lq_distribution"`
	LqLatestUpdateTime int64          `json:"lq_latest_update_time"`
	ReportStatus       bool           `json:"report_status"`
}

// 报告记录
type HealthReportRecord struct {
	TaskId     string
	AppId      string
	Report     string
	UpdateTime int
}

type LqSqRateRange struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Score float64 `json:"score"`
}

type LqConflictScoreRange struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Score float64 `json:"score"`
}

type HealthReportScoreWeight struct {
	LqConflictScore     *HealthReportScoreWeightTemplate `json:"lq_conflict_score"`
	LqSqRateScore       *HealthReportScoreWeightTemplate `json:"lq_sq_rate_score"`
	LqDistributionScore *HealthReportScoreWeightTemplate `json:"lq_distribution_score"`
}

type HealthReportScoreWeightTemplate struct {
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}
