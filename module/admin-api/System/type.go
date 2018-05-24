package System

import "encoding/json"

const (
	keyQPS       = "qps"
	keyMaxThread = "max_threads"
	minQPS       = 1
	maxQPS       = 200
	minThreadCnt = 10
	maxThreadCnt = 100
	dftQPS       = 200
	dftThreadCnt = 100
)

// ControllerSetting is the sturcture of value in consul
// at /idc/setting/controller
type ControllerSetting struct {
	qps          int
	maxThreadCnt int
}

// MarshalJSON will do marshal self
func (setting ControllerSetting) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		keyQPS:       setting.qps,
		keyMaxThread: setting.maxThreadCnt,
	})
}

// UnMarshalJSON will do unmarshal to itself
func (setting *ControllerSetting) UnMarshalJSON(data []byte) error {
	obj := map[string]int{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	setting.SetQPS(obj[keyQPS])
	setting.SetMaxThread(obj[keyMaxThread])
	return nil
}

// SetQPS is setter of qps in controller setting
func (setting *ControllerSetting) SetQPS(qps int) {
	setting.qps = qps
	if setting.qps < minQPS {
		setting.qps = minQPS
	} else if setting.qps > maxQPS {
		setting.qps = maxQPS
	}
}

// QPS is getter of qps in controller setting
func (setting ControllerSetting) QPS() int {
	return setting.qps
}

// SetMaxThread is setter of maxThreadCnt in controller setting
func (setting *ControllerSetting) SetMaxThread(threadCnt int) {
	setting.maxThreadCnt = threadCnt
	if setting.maxThreadCnt < minThreadCnt {
		setting.maxThreadCnt = minThreadCnt
	} else if setting.maxThreadCnt > maxThreadCnt {
		setting.maxThreadCnt = maxThreadCnt
	}
}

// MaxThread is getter of maxThreadCnt in controller setting
func (setting ControllerSetting) MaxThread() int {
	return setting.qps
}

func (setting *ControllerSetting) SetDefault() {
	setting.qps = dftQPS
	setting.maxThreadCnt = dftThreadCnt
}
