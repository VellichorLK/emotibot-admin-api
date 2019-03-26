package qi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

type RealtimeCallResp struct {
	CallID     int64  `json:"call_id"`
	CallUUID   string `json:"call_uuid"`
	RemoteFile string `json:"remote_file"`
}

func RealtimeCallWorkflow(output []byte) error {
	logger.Trace.Println("Realtime call workflow started")

	var callResp RealtimeCallResp
	var isDone bool

	err := json.Unmarshal(output, &callResp)
	if err != nil {
		return fmt.Errorf("Unmarshal realtime call response failed %s, body: %s",
			err, output)
	}

	c, err := Call(callResp.CallUUID, "")
	if err == ErrNotFound {
		return fmt.Errorf("Call '%s' no such call exist", callResp.CallUUID)
	} else if err != nil {
		return fmt.Errorf("Fetch call failed, %v", err)
	}

	defer func() {
		if isDone {
			return
		}

		c.Status = model.CallStatusFailed
		updateErr := UpdateCall(&c)
		if updateErr != nil {
			logger.Error.Println("update call critical failed, ", updateErr)
		}
	}()

	if volume == "" {
		return fmt.Errorf("Volume does not exist, please contact ops and check init log for volume init error")
	}

	logger.Trace.Printf("Start downloading realtime call: %s",
		callResp.RemoteFile)

	resp, err := http.Get(callResp.RemoteFile)
	if err != nil {
		return fmt.Errorf("Fail to download remote file: %s, error: %s",
			callResp.RemoteFile, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fail to download remote file: %s, status: %d",
			callResp.RemoteFile, resp.StatusCode)
	}

	filename := fmt.Sprint(c.ID, ".wav")
	fp := fmt.Sprint(volume, "/", filename)
	outFile, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("Create file failed, %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("Error while copying remote file: %s, error: %s",
			callResp.RemoteFile, err.Error())
	}

	logger.Trace.Printf("Download realtime call: %s completed\n",
		callResp.RemoteFile)

	c.FilePath = &filename
	err = ConfirmCall(&c)
	if err != nil {
		return fmt.Errorf("Confirm call failed, %v", err)
	}

    isDone = true
	return nil
}
