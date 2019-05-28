package lqcheck

import (
	"emotibot.com/emotigo/module/admin-api/Robot/config.v1"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// 健康度检查报告生成主流程
// 返回冲突检查id
func createReport(appid string) (*ConflictCheckReturn, AdminErrors.AdminError) {
	dacRet := SsmDacRet{}
	healthReport := HealthReport{}
	ConflictCheckRequest := ConflictCheckRequest{}

	// 获取所有语料
	dacRet.getSqLqFromDac(appid)

	// 发送冲突检查请求
	_, ConflictCheckReturn := ConflictCheckRequest.convertSqLq(appid, &dacRet).requestConflictCheck()

	// 异步计算标准问语料数量
	dacRet.countSqLq(appid, ConflictCheckReturn.TaskId, &healthReport)

	// 返回冲突检查id
	return ConflictCheckReturn, nil
}

func getDacUrl() string {
	url := "http://" + getENV("DAC_URL")
	return url
}

func getDacUpdateCheckUrl() string {
	url := getDacUrl() + "/ssm/dac/openapi/ischecked?appId="
	return url
}

func (d *SsmDacRet) getDacApi() string {
	url := getDacUrl() + "/ssm/dac/openapi/sq/lqs?appId="
	return url
}

func (d *SsmDacRet) getSqLqFromDac(appid string) AdminErrors.AdminError {
	// 实际地址从环境变量读取
	res, err := http.Get(d.getDacApi() + appid)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoRequestError, err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoTypeConvert, err.Error())
	}

	err = json.Unmarshal([]byte(body), d)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoJSONParse, err.Error())
	}

	return nil
}

// 获取所有标准问语料
func getSqLq(appid string) (interface{}, AdminErrors.AdminError) {
	d := SsmDacRet{}

	// 获取所有标准问语料
	d.getSqLqFromDac(appid)

	sqLq := make(map[int64]*ReportSq)

	for _, v := range d.ActualResults {
		if _, ok := sqLq[v.SqId]; !ok {
			tmpSq := &ReportSq{}
			tmpSq.SqId = v.SqId
			tmpSq.Sq = v.SqContent
			tmpSq.LqCount = 1
			tmpSq.Lq = []ReportLq{}
			tmpLq := ReportLq{}
			tmpLq.LqId = v.LqId
			tmpLq.Lq = v.LqContent
			tmpSq.Lq = append(tmpSq.Lq, tmpLq)

			sqLq[v.SqId] = tmpSq
		} else {
			sqLq[v.SqId].LqCount += 1

			tmpLq := ReportLq{}
			tmpLq.LqId = v.LqId
			tmpLq.Lq = v.LqContent
			sqLq[v.SqId].Lq = append(sqLq[v.SqId].Lq, tmpLq)
		}
	}

	return sqLq, nil
}

func (d *SsmDacRet) countSqLq(appid string, taskid string, hp *HealthReport) AdminErrors.AdminError {
	res := getSqLqUpdateTime(appid, 0)
	hp.LqLatestUpdateTime = res.ActualResults.Time

	// 语料总数
	hp.LqQuality.LqCount = len(d.ActualResults)
	hp.LqSqRate.LqCount = len(d.ActualResults)
	// 各标准问包含的语料数
	sqLq := make(map[int64]int)
	for _, v := range d.ActualResults {
		if _, ok := sqLq[v.SqId]; ok {
			sqLq[v.SqId] = 1
		} else {
			sqLq[v.SqId]++
		}
	}
	// 标准问总数
	hp.LqSqRate.SqCount = len(sqLq)
	// 标准问语料比例
	hp.LqSqRate.SqLqRate = fmt.Sprintf("%.2f", float32(hp.LqSqRate.SqCount)/float32(hp.LqSqRate.LqCount)*100)

	// 需要读取模板
	healthCheckConfigs := getBFOPconfig("")
	_ = json.Unmarshal([]byte(healthCheckConfigs["score_standard"]), &hp.HealthScore.Standard)
	_ = json.Unmarshal([]byte(healthCheckConfigs["lq_sq_rate_remark"]), &hp.LqSqRate.LqSqRateRemark)
	_ = json.Unmarshal([]byte(healthCheckConfigs["lq_distribution_recommended"]), &hp.LqDistribution.Recommended)

	//lq_sq_rate_remark
	fmt.Println(healthCheckConfigs)
	fmt.Println(healthCheckConfigs["score_standard"])
	fmt.Println(hp.HealthScore.Standard)

	// 标准问语料数量分布计算
	LqDist := hp.LqDistribution.Recommended

	// 计算语料分布计数
	for _, v := range sqLq {
		for kk, vv := range LqDist {
			if vv.To == 0 {
				if v >= vv.From {
					LqDist[kk].SqNum += 1
				}
			} else {
				if v >= vv.From && v <= vv.To {
					LqDist[kk].SqNum += 1
				}
			}
		}
	}

	// 计算语料分布占比
	for k, v := range LqDist {
		rate := float32(v.SqNum) / float32(hp.LqQuality.LqCount) * 100
		LqDist[k].SqRate, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", rate), 3)
	}
	hp.LqDistribution.Current = LqDist

	// 计算健康度分数
	hp.HealthScore.Score = 20

	// 查询冲突检查结果

	// 保存检查结果
	hpStr, _ := json.Marshal(hp)
	fmt.Println(string(hpStr))

	reportRec := make([]interface{}, 3)
	reportRec[0] = taskid
	reportRec[1] = appid
	reportRec[2] = string(hpStr)
	saveReportRecord(reportRec)

	return AdminErrors.New(AdminErrors.ErrnoJSONParse, "err")
}

// 获取健康检查状态
func getHealthCheckStatus(taskid string) (interface{}, AdminErrors.AdminError) {
	res, err := getHealthCheck(taskid)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return res[0], nil
}

// 获取健康检查报告
func getHealCheckReport(appid string) (interface{}, AdminErrors.AdminError) {
	res, err := getLatestHealthCheckReport(appid)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	//var ret map[string]interface{}
	var hr HealthReport
	if _, ok := res[0]["report"]; ok {
		json.Unmarshal([]byte(res[0]["report"]), &hr)
		checkSqLqUpdateStatus(appid, &hr)
	}

	return hr, nil
}

// 查询标准问语料是否发生变更
func checkSqLqUpdateStatus(appid string, report *HealthReport) {
	// 查询语料最后更新时间
	res := getSqLqUpdateTime(appid, report.LqLatestUpdateTime)
	report.Recheck = res.ActualResults.Needcheck
}

func getSqLqUpdateTime(appid string, lastCheckTime int64) *SsmDacCheckRet {
	var d SsmDacCheckRet

	url := getDacUpdateCheckUrl() + appid
	if lastCheckTime != 0 {
		url += "&lasttime=" + strconv.FormatInt(lastCheckTime, 10)
	}
	res, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil
	}

	return &d
}

// 标准问语料格式转换
func (c *ConflictCheckRequest) convertSqLq(appid string, d *SsmDacRet) *ConflictCheckRequest {
	c.AppId = appid
	for _, v := range d.ActualResults {
		tmp := ConflictCheckSqLq{}
		tmp.Lq = v.LqContent
		tmp.Sq = v.SqContent
		c.Data = append(c.Data, tmp)
	}
	return c
}

// 发送冲突检查请求
func (c *ConflictCheckRequest) requestConflictCheck() (*ConflictCheckResponse, *ConflictCheckReturn) {
	d, _ := json.Marshal(c)
	res, _ := http.Post(c.getConflictCheckApi(), "application/json", strings.NewReader(string(d)))

	body, _ := ioutil.ReadAll(res.Body)

	resp := ConflictCheckResponse{}
	_ = json.Unmarshal(body, &resp)

	ret := ConflictCheckReturn{}
	_ = json.Unmarshal(body, &ret)

	return &resp, &ret
}

func (c *ConflictCheckRequest) getConflictCheckApi() string {
	url := "http://" + getENV("CONFLICT_CHECK_URL") + "/data_health_check"

	return url
}

func getENV(key string) string {
	moduleName := "lqcheck"
	envs := util.GetModuleEnvironments(moduleName)
	env, _ := envs[key]

	return env
}

func getBFOPconfig(module string) map[string]string {
	if len(module) == 0 {
		module = "health_check"
	}

	confs, _ := config.GetDefaultConfigs()

	hcConfig := make(map[string]string)
	for _, v := range confs {
		if v.Module == module {
			hcConfig[v.Code] = v.Value
		}
	}

	return hcConfig
}
