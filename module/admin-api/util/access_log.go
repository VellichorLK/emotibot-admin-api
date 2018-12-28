package util

import "emotibot.com/emotigo/pkg/logger"

type AccessLog struct {
	Path         string
	UserID       string
	UserIP       string
	AppID        string
	EnterpriseID string
	Time         float64
	Input        string
	Output       string
	StatusCode   int
}

func InitAccessLog(channel chan AccessLog) {
	go func() {
		for {
			log := <-channel
			logger.Info.Printf("REQ: [%s][%d] [%.3fs][%s]@[%s][%s] [%s]",
				log.Path, log.StatusCode, log.Time, log.UserID, log.EnterpriseID, log.AppID, log.UserIP)
		}
	}()
}
