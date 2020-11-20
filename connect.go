/**
 * Copyright gitteamer 2020
 * @date: 2020/11/10
 * @note: server connect manager
 */
package tcp

import (
	"io"
	"net"
	"sync"
	"time"
)

/**
 * author gitteamer 2020/11/10
 * new server connect obj
 */
func NewConnect(server *Server, conn *net.TCPConn, addr *Addr) *Connect {
	c := new(Connect)
	c.server = server
	c.conn = conn
	c.addr = addr
	c.recvCh = make(chan *Message, 10)
	c.sendCh = make(chan *Message, 10)
	c.running = true
	return c
}

/**
 * author gitteamer 2020/11/10
 * client connect info, receive chan, send chan
 */
type Connect struct {
	name    string
	server  *Server
	addr    *Addr
	conn    *net.TCPConn
	recvCh  chan *Message
	sendCh  chan *Message
	running bool
	wg      sync.WaitGroup
	lock    sync.RWMutex
}

/**
 * author gitteamer 2020/11/10
 * receive, handle, send in goroutine
 */
func (c *Connect) worker() {
	go c.recv()
	go c.handle()
	go c.send()
}

/**
 * author gitteamer 2020/11/10
 * receive data form client
 */
func (c *Connect) recv() {
	c.wg.Add(1)
	defer c.wg.Done()

	for ; c.running; {
		time.Sleep(time.Second)
		msg := Message{}

		//c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		err := read(c.conn, &msg)
		if err != nil {
			switch err {
			case io.EOF /*errRecvEOF, errRemoteForceDisconnect*/ :
				svrLog.Error("this client connect is close: %s.", err.Error())
				c.server.closeConnect(c.addr)
				c.Close()
				return
			default:
				svrLog.Error("recv msg err: %s.", err.Error())
			}

			continue
		}

		// skip heart beat package
		if msg.Type == heartbeat {
			continue
		}

		c.recvCh <- &msg
	}
}

/**
 * author gitteamer 2020/11/10
 * handle a request.
 * receive data, handle request data, then send response data to client
 */
func (c *Connect) handle() {
	c.wg.Add(1)
	defer c.wg.Done()

	for ; c.running; {
		select {
		case msg := <-c.recvCh:
			go func() {
				c.server.callOnRecv(c.addr, msg)
			}()
		}
	}
}

/**
 * author gitteamer 2020/11/10
 * send data to client
 */
func (c *Connect) send() {
	c.wg.Add(1)
	defer c.wg.Done()

	for ; c.running; {
		select {
		case msg := <-c.sendCh:
			data, err := pack(msg)
			if err != nil {
				svrLog.Error("pack data address(%s) error:%s.", c.addr.GetAddress(), err.Error())
				continue
			}

			//c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			_, err = c.conn.Write(data.Bytes())
			if err != nil {
				// broken pipe, use of closed network connection, or other write error
				svrLog.Error("send data to client address(%s) error:%s.", c.addr.GetAddress(), err.Error())
				c.server.closeConnect(c.addr)
				c.Close()
				return
			}
		}
	}
}

/**
 * author gitteamer 2020/11/10
 * close connect
 */
func (c *Connect) Close() {
	c.lock.Lock()
	c.lock.Unlock()
	if c.running {
		c.running = false
		c.conn.Close()
		c.wg.Wait()
	}
}
