package util

import (
	"fmt"
	"strings"
)

func ParseRemoteIP(remoteAddr string) (string, error) {
	ipPort := strings.Split(remoteAddr, ":")
	if len(ipPort) != 2 {
		return "", fmt.Errorf("Invalid ip:port format of: %s", remoteAddr)
	}

	return ipPort[0], nil
}
