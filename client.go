package tcp

import (
	"fmt"
	"github.com/gitteamer/log"
	"net"
	"sync"
	"time"
)

type ClientRecvHandle = func(recv *Message)

func NewClient(address string) *Client {
	client := new(Client)
	client.connectServer(address)
	client.running = true
	go client.reconnect()

	return client
}

type Client struct {
	remoteAddr   *Addr            // remote address
	localAddr    *Addr            // local address
	conn         net.Conn         // connect server obj, receive chan, send chan
	onRecv       ClientRecvHandle // receive event callback function
	onDisconnect DisconnectHandle // disconnect event callback function
	running      bool
	connected    bool
	wg           sync.WaitGroup
	connectLock  sync.Mutex
}

func (c *Client) connectServer(address string) error {
	c.connectLock.Lock()
	defer c.connectLock.Unlock()

	if c.connected {
		return nil
	}
	c.remoteAddr = NewAddr(address)

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Error("get address info error:%s.", err.Error())
		return fmt.Errorf("get address info error:%s.", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Error("connect server error:%s.", err.Error())
		return fmt.Errorf("connect server error:%s.", err.Error())
	}

	if err = conn.SetKeepAlive(true); err != nil {
		log.Error("set keep alive err: %s.", err.Error())
	}
	if err = conn.SetKeepAlivePeriod(5 * time.Second); err != nil {
		log.Error("set keep alive period err: %s.", err.Error())
	}

	c.conn = conn
	c.connected = true
	c.localAddr = NewAddr(conn.LocalAddr().String())

	log.Info("connect server success, address=%s.", address)
	go c.recv()
	//go c.heartbeat()

	return nil
}

func (c *Client) OnRecv(handle ClientRecvHandle) {
	c.onRecv = handle
}

func (c *Client) callOnRecv(recv *Message) {
	go func() {
		if c.onRecv != nil {
			c.onRecv(recv)
		}
	}()
}

/*
func (c *Client) heartbeat() {
	ticker := time.NewTicker(time.Second * 5)
	c.wg.Add(1)

	for ; c.running && c.connected; {
		select {
		case <-ticker.C:
			if c.conn != nil {
				_ = c.Send(&Message{Type: heartbeat, Data: nil,})
			}
		}
	}

	ticker.Stop()
	c.wg.Done()
}
*/

/**
 * Reconnecting the server every 3 seconds When is disconnected.
 */
func (c *Client) reconnect() {
	c.wg.Add(1)
	defer c.wg.Done()

	for ; c.running; {
		time.Sleep(time.Second * 3)

		c.connectServer(c.remoteAddr.GetAddress())
	}

	fmt.Println("reconnect")
}

func (c *Client) Send(msg *Message) error {
	if c.conn != nil {
		// pack data
		buf, err := pack(msg)
		if err != nil {
			return fmt.Errorf("pack data error:%s.", err.Error())
		}

		// send data
		//c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		_, err = c.conn.Write(buf.Bytes())
		if err != nil {
			// broken pipe, use of closed network connection, or other write error.
			// if connect is close, then reconnect function will connect to server later.
			c.connected = false
			return fmt.Errorf("send data error:%s.", err.Error())
		}
	}
	return nil
}

func (c *Client) recv() {
	c.wg.Add(1)
	defer c.wg.Done()

	for ; c.running && c.connected; {
		msg := Message{}
		//c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		if err := read(c.conn, &msg); err != nil {
			switch err {
			case errRecvEOF, errRemoteForceDisconnect:
				log.Error("this client connect disconnect: %s.", err.Error())
				c.connected = false
				break
			default:
				log.Error("recv msg err: %s.", err.Error())
			}
			continue
		}

		c.callOnRecv(&msg)
	}

	fmt.Println("recv")
}

func (c *Client) Close() {
	c.running = false
	c.conn.Close()
	c.wg.Wait()
}
