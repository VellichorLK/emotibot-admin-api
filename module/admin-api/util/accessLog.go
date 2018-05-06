package util

type AccessLog struct {
	Path       string
	UserID     string
	AppID      string
	Time       float64
	Input      string
	Output     string
	StatusCode int
}

func InitAccessLog(channel chan AccessLog) {
	go func() {
		for {
			log := <-channel
			LogInfo.Printf("REQ: [%s] [%.3fs][%s@%s]",
				log.Path, log.Time, log.UserID, log.AppID)
		}
	}()
}
