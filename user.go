package main

import (
	"net"
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

func (user *User) DoMessage(msg string) {
	user.server.Broadcast(user, msg)
}