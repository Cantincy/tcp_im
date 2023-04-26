package main

import (
	"bufio"
	"im/util"
	"log"
	"net"
	"os"
)

type Client struct {
	Name string
	conn net.Conn
}

func NewClient(name string) (*Client, error) {
	serverAddr := util.GetServerAddr()
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	return &Client{Name: name, conn: conn}, nil
}

func main() {
	cli, err := NewClient("xiaoming")
	if err != nil {
		log.Fatal(err)
	}
	defer cli.conn.Close()

	go func(conn net.Conn) {
		defer conn.Close()
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(string(buf[:n+1]))
		}
	}(cli.conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		cmd, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		cli.conn.Write(append(cmd, ' '))
	}
}
