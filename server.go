package main

import (
	"net"
	"fmt"
	"sync"
	"io"
	"time"
)

type Server struct {
	Ip string
	Port int

	OnlineMap map[string]*User
	mapLock sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}

	return server
}


func (server *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg

	server.Message <- sendMsg
}

func (server *Server) ListenMessager() {
	for {
		msg := <- server.Message

		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) Handler(conn net.Conn) {
	user := NewUser(conn, server)

	user.Online()

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return 
			}
			
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return 
			}

			msg := string(buf[:n-1])
			user.DoMessage(msg)

			isLive <- true
		}
	}()
	for {
		select {
		case <- isLive:
		case <- time.After(time.Second * 600): 
			user.SendMsg("你超时被T了")
			close(user.C)
			conn.Close()
			return //runtime.Goexit()
		}

	}
}

func (server *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listener err: ", err)
	}
	defer listener.Close()
	go server.ListenMessager()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener accept err: ", err)
			continue
		}

		go server.Handler(conn)
	}
}