package entity

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	Conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{userAddr, userAddr, make(chan string), conn}

	go user.SendMsgToClient()

	return user
}

// SendMsgToClient 不断从用户通道读取消息发往客户端
func (this *User) SendMsgToClient() {
	for {
		msg := <-this.C
		this.Conn.Write([]byte(msg + "\n")) //TCP-Conn 支持双工
	}
}
