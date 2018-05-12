package main

import (
	"flag"
	"fmt"
	"os"

	"emotibot.com/emotigo/module/admin-api/util"
)

func main() {
	runEncryptFlag := flag.Bool("e", false, "If -e is set, run encrypt. If not, run decrypt")
	textFlag := flag.String("t", "", "Text for encrypt/decrypt")

	flag.Parse()
	if textFlag == nil {
		fmt.Fprintln(os.Stderr, "Parse flag error")
		os.Exit(1)
	}

	text := *textFlag
	runEncrypt := false
	if runEncryptFlag != nil {
		runEncrypt = *runEncryptFlag
	}

	var result string
	var err error

	if runEncrypt {
		result, err = util.DesEncrypt([]byte(text), []byte(util.DesEncryptKey))
	} else {
		result, err = util.DesDecrypt(text, []byte(util.DesEncryptKey))
	}
	action := "Decrypt"
	if runEncrypt {
		action = "Encrypt"
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s fail: %s\n", action, err.Error())
		os.Exit(2)
	}
	fmt.Printf(result)
}
