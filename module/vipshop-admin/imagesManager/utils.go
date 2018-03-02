package imagesManager

import (
	"errors"
	"strconv"
)

func reverseSlice(name []string) {
	for i := 0; i < len(name)/2; i++ {
		j := len(name) - i - 1
		name[i], name[j] = name[j], name[i]
	}
}

func getImagesParams(params map[string]string) (*getImagesArg, error) {
	listArgs := &getImagesArg{Order: attrID, Page: 0, Limit: 12}

	for k, v := range params {
		switch k {
		case ORDER:
			switch v {
			case valID:
				listArgs.Order = attrID
			case valName:
				listArgs.Order = attrFileName
			case valTime:
				listArgs.Order = attrLatestUpdate
			default:
				return nil, errors.New("Unknown parameter " + k)
			}
		case PAGE:
			page, err := strconv.ParseInt(v, 10, 64)
			if err != nil || page < 0 {
				return nil, errors.New(k + "(" + v + ")" + " is not numeric or negative")
			}
			listArgs.Page = page
		case LIMIT:
			limit, err := strconv.ParseInt(v, 10, 64)
			if err != nil || limit <= 0 {
				return nil, errors.New(k + "(" + v + ")" + " is not numeric or <=0")
			}
			listArgs.Limit = limit
		case KEYWORD:
			listArgs.Keyword = v
		default:
			return nil, errors.New("Unknown parameter " + k)
		}
	}

	return listArgs, nil
}
