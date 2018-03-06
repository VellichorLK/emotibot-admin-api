package imagesManager

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
)

//saveTo will rewrite the data to filePath location, return os.ErrExist if filePath already occupied.
func saveTo(filePath string, data []byte) error {
	file, err := os.Open(filePath)
	if err == nil {
		return os.ErrExist
	}
	if os.IsNotExist(err) {
		file, err = os.Create(filePath)
	}
	if err != nil {
		return fmt.Errorf("Create file on %s failed, %v", filePath, err)
	}

	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Writing to file %s failed, %v", filePath, err)
	}

	return nil
}

//Store will try to save image with it's fileName and write it's actully location
//After Encounter same fileName, store function will try to add a tailing name in it fileName
func (image Image) Store(data []byte, root string) (string, error) {
	var i = 1
	var originFileBase = path.Base(image.FileName)
	var extension = path.Ext(originFileBase)
	var fileName = originFileBase[0 : len(originFileBase)-len(extension)]
	root = strings.TrimRight(root, "/")

	var err = saveTo(root+"/"+image.FileName, data)
	for ; err == os.ErrExist; err = saveTo(root+"/"+image.FileName, data) {
		image.FileName = fmt.Sprintf("%s_%d%s", fileName, i, extension)
		i++
	}
	return image.FileName, err
}

//GetURL will try to assemble image URL by webAddress and location
// return error if webAddress is not a valid web location
func (image Image) GetURL(webAddress string) (string, error) {
	u, err := url.Parse(webAddress)
	if err != nil {
		return "", fmt.Errorf("url parsing failed, %+v", err)
	}
	return fmt.Sprintf("http://%s/images/%s", u.Host, image.FileName), nil
}
