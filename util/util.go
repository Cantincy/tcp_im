package util

import (
	"fmt"
	"im/config"
)

func GetServerAddr() string {
	return fmt.Sprintf("%s:%d", config.ServerIP, config.ServerPort)
}
