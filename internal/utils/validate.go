package utils

import (
	"fmt"
	"strconv"
)

func ValidatePort(port string) (bool, error) {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false, err
	}

	if portNum < 1 || portNum > 65535 {
		return false, fmt.Errorf("PORT %s is out of range!\n", port)
	}

	if portNum < 1024 {
		return false, fmt.Errorf("PORT %s is a privileged port\n", port)
	}

	return true, nil
}
