package main

import (
	"im/config"
	"im/server"
)

func main() {
	s := server.NewServer(config.ServerIP, config.ServerPort)
	s.Start()
}
