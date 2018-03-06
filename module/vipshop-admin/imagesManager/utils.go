package imagesManager

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
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
				return nil, errors.New("Unknown parameter value " + v)
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

func GetFullCategory(categories map[int]*Category, categoryID int) ([]string, error) {

	if categories == nil {
		return nil, errors.New("categories map is nil")
	}

	const MAXLEVEL = 5
	levels := make([]string, 0, MAXLEVEL)
	for i := 0; i < MAXLEVEL; i++ {
		c, ok := categories[categoryID]
		if !ok {
			return nil, errors.New("No categoryID " + strconv.Itoa(categoryID))
		}
		levels = append(levels, c.Name)
		if c.ParentID == 0 {
			break
		}
		categoryID = c.ParentID
	}

	levels = levels[:len(levels)]
	reverseSlice(levels)
	return levels, nil
}

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GetUniqueString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
