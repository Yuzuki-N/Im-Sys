package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn
	
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User {
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,

		server: server,
	}

	go user.ListenMessage()

	return user
}

func (user *User) ListenMessage() {
	for {
		msg := <- user.C

		user.conn.Write([]byte(msg + "\n"))
	}
}

func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	user.server.Broadcast(user, "已上线")
}

func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	user.server.Broadcast(user, "下线了")
}


func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

func (user *User) DoMessage(msg string) {
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, usr := range user.server.OnlineMap {
			OnlineMsg := "[" + usr.Addr +"]" + usr.Name + ": " + "在线...\n"
			user.SendMsg(OnlineMsg)
		}
		user.server.mapLock.Unlock()
	}	else if len(msg) > 7 && msg[:7] == "rename|" {

		newName := strings.Split(msg, "|")[1]

		if _, ok := user.server.OnlineMap[newName]; ok {
			user.SendMsg("当前用户名被使用")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMsg("修改用户名为" + user.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]

		if remoteName == "" {
			user.SendMsg("消息格式不正确，请使用 \"to|用户名|消息内容 这样的格式\"\n")
			return 
		}
		
		remoteUser, ok := user.server.OnlineMap[remoteName]; 
		if !ok {
			user.SendMsg("该用户名不存在")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("无消息内容， 请重发")
			return
		}
		remoteUser.SendMsg(user.Name + "对您说" + content + "\n")

	} else {
		user.server.Broadcast(user, msg)
	}
	
}