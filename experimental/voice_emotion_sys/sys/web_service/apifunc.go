package main

import (
	"net/http"

	"handlers"
)

const v1basePath = "/voice/emotion/v1"

type apiFunc map[string]func(w http.ResponseWriter, r *http.Request)

//add here if you hava v2 version api or something else
//new the apiFunc and add it to services
var v1api = apiFunc{
	v1basePath + "/upload":            handlers.Upload,
	v1basePath + "/files":             handlers.QueryEmotions,
	v1basePath + "/files_continue":    handlers.QueryContinue,
	v1basePath + "/files/":            handlers.QueryEmotionDetail,
	v1basePath + "/report":            handlers.GenerateReport,
	v1basePath + "/user_column":       handlers.GetUserColumn,
	v1basePath + "/user_column_value": handlers.UpdateColumnVal,
}

var services = [...]apiFunc{
	v1api,
}

type golabalHandler func(next http.Handler) http.Handler

//MiddleServices function would be called before mux
// so you can put some check authentication or blablabla before enter into the api service
var MiddleServices = [...]golabalHandler{
//handlers.CheckAuth,
}

func fakeEnv() {
	envs["RABBITMQ_HOST"] = "127.0.0.1"
	envs["RABBITMQ_PORT"] = "5672"
	//envs["DB_HOST"] = "192.168.3.208"
	envs["DB_HOST"] = "127.0.0.1"
	envs["DB_PORT"] = "3306"
	envs["DB_USER"] = "root"
	envs["DB_PWD"] = "password"
	envs["FILE_PREFIX"] = "/Users/public/go/src/emotibot.com/emotigo/experimental/voice_emotion_sys/sys/web_service/upload_file"
	envs["LISTEN_PORT"] = ":8080"
	envs["RABBITMQ_USER"] = "guest"
	envs["RABBITMQ_PWD"] = "guest"
}
func FakeHandlers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("fakehandler", "fakeHandler")
		next.ServeHTTP(w, r)

	})
}

func FakeAPi(w http.ResponseWriter, r *http.Request) {
	fakeHeader := r.Header.Get("fakehandler")
	appid := r.Header.Get(handlers.NAPPID)

	res := "fakeHeader:" + fakeHeader + ", appid:" + appid

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))

}
