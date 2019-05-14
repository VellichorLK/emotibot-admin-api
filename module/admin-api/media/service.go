package media

import (
	"fmt"
	"io"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"

	"emotibot.com/emotigo/pkg/services/fileservice"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

const MinioNamespace = "media"
const IDLength = 16

func AddFile(appid string, input io.Reader) (string, AdminErrors.AdminError) {
	now := time.Now()
	id := fmt.Sprintf("%s-%s", now.Format("20060102150405"), util.GenRandomString(IDLength))
	path := fmt.Sprintf("%s/%s", appid, id)
	err := fileservice.AddFile(MinioNamespace, path, input)
	if err != nil {
		return "", AdminErrors.New(AdminErrors.ErrnoAPIError, err.Error())
	}
	return id, nil
}
