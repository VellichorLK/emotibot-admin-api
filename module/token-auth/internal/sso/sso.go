package sso

import (
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/sso/huawei"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

type SSOHandler interface {
	LoadConfig(*util.SSOConfig) error
	ValidateRequest(*http.Request) (string, string, error)
	ValidateDebug(*http.Request) string
}

func GetHandler(config *util.SSOConfig) SSOHandler {
	if config == nil {
		util.LogTrace.Println("Nil config, nil handler")
		return nil
	}

	matchType := strings.ToLower(config.SSOType)
	switch matchType {
	case "huawei":
		handler := &huawei.HuaweiSSO{}
		handler.LoadConfig(config)
		return handler
	}
	return nil
}
