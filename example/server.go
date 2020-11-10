package main

import (
	"fmt"
	"github.com/gitteamer/tcp"
	"net"
	"os"
	"os/signal"
)

func main() {
	// Receive system interrupt signal
	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt, os.Kill)

	// new tcp server
	server := tcp.NewServer()

	// interrupt server
	go func() {
		<-stopCh
		fmt.Println("receive interrupt command, now stopping...")
		server.Close()
	}()

	server.OnConnect(func(conn *net.TCPConn, addr *tcp.Addr) {
		fmt.Println(fmt.Sprintf("one client connect, remote address=%s.", conn.RemoteAddr().String()))
	})

	server.OnRecv(func(addr *tcp.Addr, req *tcp.Message) {
		fmt.Println(fmt.Sprintf("req.Type=%d, req.Data=%s.", req.Type, string(req.Data)))

		_ = server.Send(*addr, &tcp.Message{
			Type: 1,
			Data: []byte(fmt.Sprintf("hello: %s.", addr.GetAddress())),
		})
	})

	server.OnDisconnect(func(addr *tcp.Addr) {

	})

	server.Run(":8080")
}
