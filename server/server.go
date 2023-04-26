package server

import (
	"fmt"
	"im/entity"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip               string
	Port             int
	OnlineMap        map[string]*entity.User
	MapLock          sync.RWMutex
	BroadCastChannel chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:               ip,
		Port:             port,
		OnlineMap:        make(map[string]*entity.User),
		BroadCastChannel: make(chan string),
	}
}

// BroadCastMsg 广播消息（将消息放入广播通道）
func (this *Server) BroadCastMsg(msg string) {
	this.BroadCastChannel <- msg
}

// ListenBroadCastMsg 广播消息（从广播通道中读取消息并发给所有用户）
func (this *Server) ListenBroadCastMsg() {
	for {
		msg := <-this.BroadCastChannel
		this.MapLock.Lock()
		for _, user := range this.OnlineMap { // 将广播信息发送给所有用户
			user.C <- msg
		}
		this.MapLock.Unlock()
	}
}

// ServerHandler Server每建立一个Conn都要执行Handler
func (this *Server) ServerHandler(conn net.Conn) {
	user := entity.NewUser(conn)

	this.UserOnline(user)

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := user.Conn.Read(buf)
			if n == 0 { // ctrl + c
				this.UserOffline(user)
				return
			}

			if err != nil && err != io.EOF { // 未知错误
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			log.Println(msg)

			isLive <- true

			if msg == "allOnlineUser" { // 获取所有在线用户信息
				user.C <- this.AllOnlineUser()
			} else if strings.HasPrefix(msg, "search ") { //查询在线用户
				user.C <- this.SearchUser(msg[7:])
			} else if strings.HasPrefix(msg, "rename ") { // 修改用户名
				user.C <- this.Rename(user.Addr, msg[7:])
			} else if strings.HasPrefix(msg, "sendTo ") { // 私聊
				parts := strings.Split(msg, " ")
				if len(parts) < 3 {
					user.C <- "cmd format error"
				} else {
					user.C <- this.SendTo(parts[1], fmt.Sprintf("[%s]:%s", user.Addr, msg[len(parts[0])+len(parts[1])+2:]))
				}
			} else if msg == "exit" { // 下线
				this.UserOffline(user)
				return
			} else { // 公聊
				this.BroadCastMsg(user.Name + ": " + msg)
			}
		}
	}()

	for {
		select {
		case <-isLive:

		case <-time.After(time.Second * 180):
			// 超时
			user.C <- "You are forced offline because no operation has been performed for a long time..."
			time.Sleep(time.Second)

			this.MapLock.Lock()
			delete(this.OnlineMap, user.Addr)
			this.MapLock.Unlock()

			close(user.C)
			conn.Close()

			return
		}
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go this.ListenBroadCastMsg() // 将广播信息发给所有用户

	log.Println("Server is Running...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go this.ServerHandler(conn)
	}
}

func (this *Server) UserOnline(user *entity.User) {
	this.MapLock.Lock()
	this.OnlineMap[user.Addr] = user // 将user写入map
	this.MapLock.Unlock()

	this.BroadCastMsg(user.Name + ": Online...") // 将用户上线信息广播给所有用户
}

func (this *Server) UserOffline(user *entity.User) {
	this.MapLock.Lock()
	delete(this.OnlineMap, user.Addr)
	this.MapLock.Unlock()

	close(user.C)
	user.Conn.Close()

	this.BroadCastMsg(user.Name + ": Offline...") // 将用户下线信息广播给所有用户
}

func (this *Server) SearchUser(userName string) string {
	_, ok := this.OnlineMap[userName]
	if ok {
		return userName + " is Online"
	}
	return userName + " is Offline"
}

func (this *Server) AllOnlineUser() string {
	res := strings.Builder{}
	for _, user := range this.OnlineMap {
		res.WriteString(fmt.Sprintf("Name:%s Addr:%s\n", user.Name, user.Addr))
	}
	return res.String()
}

func (this *Server) Rename(userAddr, newName string) string {
	user, ok := this.OnlineMap[userAddr]
	if !ok {
		return "user is not online"
	}
	user.Name = newName
	this.MapLock.Lock()
	this.OnlineMap[userAddr] = user
	this.MapLock.Unlock()
	return "success"
}

func (this *Server) SendTo(userAddr, msg string) string {
	user, ok := this.OnlineMap[userAddr]
	if !ok {
		return "user is not online"
	} else {
		user.C <- msg
		return "success"
	}
}
