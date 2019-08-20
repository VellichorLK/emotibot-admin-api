package data

import "emotibot.com/emotigo/module/token-auth/internal/lang"

type AppPropV5 struct {
	ID     interface{} `json:"id"`
	PKey   string      `json:"p_key"`
	PValue string      `json:"p_value"`
	PName  string      `json:"p_name"`
}

type AppPropsV5 []*AppPropV5

type AppPropRelV5 struct {
	AppPropV5
	AppId string `json:"app_id"`
}

type AppPropRelsV5 []*AppPropRelV5

func (props AppPropsV5) GetAppPropsName(locale string) error {
	if len(props) == 0 {
		return nil
	}
	for _, v := range props {
		v.PName = lang.Get(locale, v.PKey+"_"+v.PValue)
	}
	return nil
}

func (props AppPropRelsV5) GetAppPropsName(locale string) error {
	if len(props) == 0 {
		return nil
	}
	for _, v := range props {
		v.PName = lang.Get(locale, v.PKey+"_"+v.PValue)
	}
	return nil
}
