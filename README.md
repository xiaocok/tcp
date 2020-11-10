# tcp
A golang language library for tcp server and client

---
### server
```go
package main

import (
	"fmt"
	"github.com/gitteamer/tcp"
	"net"
)

func main()  {
 	// new tcp server
 	server := tcp.NewServer()
 
 	// on connect event
 	server.OnConnect(func(conn *net.TCPConn, addr *tcp.Addr) {
 		fmt.Println(fmt.Sprintf("one client connect, remote address=%s.", conn.RemoteAddr().String()))
 	})
 
 	// on receive data event
 	server.OnRecv(func(addr *tcp.Addr, req *tcp.Message) {
 		fmt.Println(fmt.Sprintf("req.Type=%d, req.Data=%s.", req.Type, string(req.Data)))
 
 		_ = server.Send(*addr, &tcp.Message{
 			Type: 1,
 			Data: []byte(fmt.Sprintf("hello: %s.", addr.GetAddress())),
 		})
 	})
 
 	// on disconnect event
 	server.OnDisconnect(func(addr *tcp.Addr) {
 
 	})
 
 	server.Run(":8080")
}
```

### client
```go
package main

import (
	"fmt"
	"github.com/gitteamer/tcp"
	"os"
	"os/signal"
)

func main() {
	client := tcp.NewClient("127.0.0.1:8080")

	// send data
	err := client.Send(&tcp.Message{
		Type: 1,
		Data: []byte("hello server"),
	})
	if err != nil {
		fmt.Println("send data error:", err.Error())
	}

	// on receive data event
	client.OnRecv(func(recv *tcp.Message) {
		fmt.Println(fmt.Sprintf("recv data, recv.Type=%d, recv.Data=%s.", recv.Type, string(recv.Data)))
	})
}
```

---
### message type
```go
type Message struct {
	Type uint
	Data []byte
}
```
* Type<br/>
Because of the heartbeat is define 0. (heartbeat = 0 // heartbeat package)<br/>
You can define other flag. 
* Data<br/>
the data that you want to transfer. It could be JSON, gob, protobuf and so on.

