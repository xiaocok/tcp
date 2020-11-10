package main

import (
	"fmt"
	"github.com/gitteamer/tcp"
	"os"
	"os/signal"
)

func main() {
	client := tcp.NewClient("127.0.0.1:8080")

	// Receive system interrupt signal
	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt, os.Kill)

	// interrupt server
	go func() {
		<-stopCh
		fmt.Println("receive interrupt command, now stopping...")
		client.Close()
	}()

	err := client.Send(&tcp.Message{
		Type: 1,
		Data: []byte("hello server"),
	})
	if err != nil {
		fmt.Println("send data error:", err.Error())
	}

	client.OnRecv(func(recv *tcp.Message) {
		fmt.Println(fmt.Sprintf("recv data, recv.Type=%d, recv.Data=%s.", recv.Type, string(recv.Data)))
	})

	ch := make(chan struct{})
	<-ch
}
