/**
 * Copyright gitteamer 2020
 * @date: 2020/11/10
 * @note: tcp client
 */

package tcp

import (
	"fmt"
	"github.com/gitteamer/log"
	"io"
	"net"
	"sync"
	"time"
)

var (
	cliLog = log.NewLogger()
)

// init log, only Console
func init() {
	// init log
	cliLog.SetLogger("tcp", log.Console, log.LevelInfo)
}

/**
 * author gitteamer 2020/11/10
 * receive event callback function
 */
type ClientRecvHandle = func(recv *Message)

/**
 * author gitteamer 2020/11/10
 * new client by address, ip:port
 */
func NewClient(address string) *Client {
	client := new(Client)
	client.connectServer(address)
	client.running = true
	go client.reconnect()

	return client
}

/**
 * author gitteamer 2020/11/10
 * tcp client struct
 */
type Client struct {
	remoteAddr   *Addr            // remote address
	localAddr    *Addr            // local address
	conn         net.Conn         // connect server obj, receive chan, send chan
	onRecv       ClientRecvHandle // receive event callback function
	onDisconnect DisconnectHandle // disconnect event callback function
	running      bool             // is running flag
	connected    bool             // is connect flag
	wg           sync.WaitGroup   // wait event obj, use save exit
	connectLock  sync.Mutex       // connect lock, Avoid concurrent connections
}

/**
 * author gitteamer 2020/11/10
 * connect server. if not connected, will Reconnect Every 3 seconds.
 * @param string address：	server addr
 */
func (c *Client) connectServer(address string) error {
	c.connectLock.Lock()
	defer c.connectLock.Unlock()

	if c.connected {
		return nil
	}
	c.remoteAddr = NewAddr(address)

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return fmt.Errorf("get address info error:%s.", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("connect server error:%s.", err.Error())
	}

	if err = conn.SetKeepAlive(true); err != nil {
		cliLog.Error("set keep alive err: %s.", err.Error())
	}
	if err = conn.SetKeepAlivePeriod(5 * time.Second); err != nil {
		cliLog.Error("set keep alive period err: %s.", err.Error())
	}

	c.conn = conn
	c.connected = true
	c.localAddr = NewAddr(conn.LocalAddr().String())

	cliLog.Info("connect server success, address=%s.", address)
	go c.recv()
	//go c.heartbeat()

	return nil
}

/**
 * author gitteamer 2020/11/10
 * register receive event callback function
 * @param ClientRecvHandle handle:	callback function
 */
func (c *Client) OnRecv(handle ClientRecvHandle) {
	c.onRecv = handle
}

/**
 * author gitteamer 2020/11/10
 * call receive event function
 * @param *Message recv:	the receive data package
 */
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
 * author gitteamer 2020/11/10
 * Reconnecting the server every 3 seconds When is disconnected.
 */
func (c *Client) reconnect() {
	c.wg.Add(1)
	defer c.wg.Done()

	var currentConnectStatus error
	for ; c.running; {
		time.Sleep(time.Second * 3)

		err := c.connectServer(c.remoteAddr.GetAddress())
		// filtering duplicate logs
		if err != nil {
			if currentConnectStatus == nil || err.Error() != currentConnectStatus.Error() {
				currentConnectStatus = err
				cliLog.Error(err.Error())
			}
		}
	}
}

/**
 * author gitteamer 2020/11/10
 * send data to server
 * @param *Message msg:		send message data
 */
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

/**
 * author gitteamer 2020/11/10
 * read data until not running or disconnect
 */
func (c *Client) recv() {
	c.wg.Add(1)
	defer c.wg.Done()

	for ; c.running && c.connected; {
		msg := Message{}
		//c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		if err := read(c.conn, &msg); err != nil {
			switch err {
			case io.EOF/*errRecvEOF, errRemoteForceDisconnect*/:
				cliLog.Error("this client connect disconnect: %s.", err.Error())
				c.connected = false
				break
			default:
				cliLog.Error("recv msg err: %s.", err.Error())
			}
			continue
		}

		c.callOnRecv(&msg)
	}
}

/**
 * author gitteamer 2020/11/10
 * close connect
 */
func (c *Client) Close() {
	c.running = false
	if c.conn != nil { // fix: connect fail, then c.conn is nil.
		c.conn.Close()
	}
	c.wg.Wait()
}
