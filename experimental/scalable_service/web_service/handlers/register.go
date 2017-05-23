package handlers

import (
	"io/ioutil"
	"log"
	"net/http"

	"strings"

	"gopkg.in/yaml.v2"
)

var handlers ModuleHandleFunc
var modules YamlCfg
var funcSupport FuncSupportMethod

//RegisterAllHandlers register all usable module handler
func RegisterAllHandlers() {
	handlers = make(ModuleHandleFunc)
	funcSupport = make(FuncSupportMethod)

	handlers["default"] = DefaultHandler
	//add your handler here

	setHTTPPath()
}

//LoadCfg load service configure file
func LoadCfg(cfg string) {
	file, err := ioutil.ReadFile(cfg)
	if err != nil {
		log.Fatalln("Can't open service.cfg. ", err)
	}
	yaml.Unmarshal(file, &modules)
}

func isValidFunc(sd *SupportData, url string, handler string) bool {
	res := true

	if len(sd.Method) == 0 {
		log.Println(url, " has no supported method")
		res = false
	}

	if sd.Queue == "" {
		log.Println(url, " has no queue specified")
		res = false
	}

	if handler == "" {
		log.Println(url, " has no handler specified")
		res = false
	} else {
		_, ok := handlers[handler]
		if !ok {
			log.Printf("%s has no handler %q function registered\n", url, handler)
			res = false
		}
	}

	return res
}

func setHTTPPath() {

	var sd *SupportData
	var handler string

	for path, module := range modules.MethodData {

		sd = new(SupportData)
		handler = ""
		for method, value := range module {

			if method == "x-handler" {
				switch v := value.(type) {
				case string:
					handler = v
				}
			} else if method == "x-queue" {
				switch v := value.(type) {
				case string:
					sd.Queue = v
				}
			} else {
				if sd.Method == nil {
					sd.Method = make(map[string]string)
				}
				var produce string

				switch mapping := value.(type) {
				case map[interface{}]interface{}:
					produces, ok := mapping["produces"]
					if ok {
						produce = produces.([]interface{})[0].(string)
					} else {
						produce = "text/plain"
					}
				}
				sd.Method[strings.ToUpper(method)] = produce
			}
		}

		if isValidFunc(sd, path, handler) {
			funcSupport[path] = sd
			http.HandleFunc(path, handlers[handler])
		}
	}
}
