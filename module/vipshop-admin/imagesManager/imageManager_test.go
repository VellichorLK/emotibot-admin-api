package imagesManager

import (
	"fmt"
	"os"
	"strings"
)

func tearDownFiles(locations ...string) {
	for _, e := range locations {
		if !strings.HasPrefix(e, "./testdata/") && !strings.HasPrefix(e, "testdata/") {
			fmt.Println("Try to delete " + e + ", stopped.")
			continue
		}
		err := os.RemoveAll(e)
		if err != nil {
			fmt.Println("Tear down Delete failed, " + err.Error())
		}
	}
}
