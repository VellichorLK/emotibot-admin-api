package imagesManager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestSaveTo(t *testing.T) {
	var location = "./testdata/test2.jpg"
	data, _ := ioutil.ReadFile("./testdata/golang.jpg")
	//Tear Down after test
	defer tearDownFiles([]string{location})
	err := saveTo(location, data)
	if err != nil {
		t.Fatal(err)
	}
	err = saveTo(location, data)
	if err != os.ErrExist {
		t.Fatal("Should have raised error if call to same location twice")
	}

}

func TestStore5Time(t *testing.T) {
	var image = Image{
		ID:       1,
		FileName: "test.jpg",
	}
	var cleanUpList = make([]string, 5)
	root := "./testdata"
	defer tearDownFiles(cleanUpList)
	expectedData, _ := ioutil.ReadFile("./testdata/golang.jpg")
	for i := 0; i < 5; i++ {
		fileName, err := image.Store(expectedData, root)
		if err != nil {
			t.Fatal(err)
		}
		cleanUpList[i] = root + "/" + fileName
		if i == 0 {
			if fileName != "test.jpg" {
				t.Fatalf("fileName should be test.jpg, but got %s", fileName)
			}
		} else if expectedFileName := fmt.Sprintf("test_%d.jpg", i); fileName != expectedFileName {
			t.Fatalf("fileName should be %s, but got %s", expectedFileName, fileName)
		}
	}

	for _, e := range cleanUpList {
		data, err := ioutil.ReadFile(e)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(expectedData, data) != 0 {
			t.Fatal("Content are differnet after read and write")
		}
	}
}

func TestGetURL(t *testing.T) {
	var image = Image{
		ID:       1,
		FileName: "test.jpg",
	}
	var webAddress = "http://127.0.0.1:8080"
	place, err := image.GetURL(webAddress)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal(place)
}

func tearDownFiles(locations []string) {
	fmt.Println(locations)
	for _, e := range locations {
		if !strings.HasPrefix(e, "./testdata/") {
			// do not delete anything out of testdata
			continue
		}
		os.Remove(e)
	}
}
