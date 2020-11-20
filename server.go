/**
 * Copyright gitteamer 2020
 * @date: 2020/11/10
 * @note: tcp server
 */
package tcp

import (
	"fmt"
	"github.com/gitteamer/log"
	"net"
	"sync"
)

var (
	svrLog = log.NewLogger()
)

// init log, only Console
func init() {
	// init log
	svrLog.SetLogger("tcp", log.Console, log.LevelInfo)
}

/**
 * author gitteamer 2020/11/10
 * the callback function for event
 * @param *Message req:	the message recv from client connect
 * @return *Message:	the message response to client. if the message is nil, then do nothing
 */
type ServerRecvHandle = func(addr *Addr, req *Message)
type ConnectHandle = func(conn *net.TCPConn, addr *Addr)
type DisconnectHandle = func(addr *Addr)

func NewServer() *Server {
	s := new(Server)
	s.connMgr = make(map[Addr]*Connect)
	s.running = true
	return s
}

/**
 * author gitteamer 2020/11/10
 * tcp server.
 * Please use NewServer() to create server obj.
 */
type Server struct {
	listener     *net.TCPListener  // the tcp server Listener
	connMgr      map[Addr]*Connect // client connect map
	onRecv       ServerRecvHandle  // receive event callback function
	onConnect    ConnectHandle     // connect event callback function
	onDisconnect DisconnectHandle  // disconnect event callback function
	lock         sync.RWMutex
	running      bool
}

/**
 * author gitteamer 2020/11/10
 * register callback function when receive data
 * @param RecvHandle handle:	callback function
 */
func (s *Server) OnRecv(handle ServerRecvHandle) {
	s.onRecv = handle
}

/**
 * author gitteamer 2020/11/10
 * call OnRecv callback function
 * @param *Addr addr:	client addr
 * @param *Message req:	request message
 */
func (s *Server) callOnRecv(addr *Addr, req *Message) {
	go func() {
		if s.onRecv != nil {
			s.onRecv(addr, req)
		}
	}()
}

/**
 * author gitteamer 2020/11/10
 * send a message to tcp client
 * @param Addr addr:	the client address struct
 * @param *Message msg:	the message sent to client
 */
func (s *Server) Send(addr Addr, msg *Message) error {
	if connect, ok := s.connMgr[addr]; !ok {
		return fmt.Errorf("the address[%s] of connect is not exist.", addr.GetAddress())
	} else {
		go func() {
			connect.sendCh <- msg
		}()
	}
	return nil
}

/**
 * author gitteamer 2020/11/10
 * register callback function when a client connect
 * @param ConnectHandle handle:	callback function
 */
func (s *Server) OnConnect(handle ConnectHandle) {
	s.onConnect = handle
}

/**
 * author gitteamer 2020/11/10
 * call OnConnect callback function
 * @param *net.TCPConn conn:the client connect
 * @param *Addr addr:		client addr
 */
func (s *Server) callOnConnect(conn *net.TCPConn, addr *Addr) {
	go func() {
		if s.onConnect != nil {
			s.onConnect(conn, addr)
		}
	}()
}

/**
 * author gitteamer 2020/11/10
 * register callback function when a client disconnect
 * @param DisconnectHandle handle:	callback function
 */
func (s *Server) OnDisconnect(handle DisconnectHandle) {
	s.onDisconnect = handle
}

/**
 * author gitteamer 2020/11/10
 * call OnDisconnect callback function
 * @param *Addr addr:		client addr
 */
func (s *Server) callOnDisconnect(addr *Addr) {
	go func() {
		if s.onDisconnect != nil {
			s.onDisconnect(addr)
		}
	}()
}

/**
 * author gitteamer 2020/11/10
 * close server and recycling resources
 */
func (s *Server) Close() {
	s.lock.Lock()
	defer s.lock.Lock()

	s.running = false

	// close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// close all connect
	var wg sync.WaitGroup
	for addr, connect := range s.connMgr {
		go func() {
			wg.Add(1)
			connect.Close()
			delete(s.connMgr, addr)
			wg.Done()
		}()
	}

	wg.Wait()
}

/**
 * author gitteamer 2020/11/10
 * close specified connect
 * @param *Addr addr:	the client address
 */
func (s *Server) closeConnect(addr *Addr) {
	s.callOnDisconnect(addr)

	if connect, ok := s.connMgr[*addr]; ok {
		connect.Close()
		delete(s.connMgr, *addr)
	}
}

/**
 * author gitteamer 2020/11/10
 * run tcp server.
 * If you want to execute in the background, Please run as:  go server.Run(addr)
 * @param string addr:	Listen address, as ip:port or :port
 */
func (s *Server) Run(addr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		svrLog.Error("the server listen address err:%s.", err.Error())
		return
	}

	// server listen
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		svrLog.Error("tcp server listen addr(%s) error:%s.", addr, err.Error())
		return
	}

	// save listener
	s.listener = listener

	// accept client
	for ; s.running; {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			svrLog.Error("tcp server accept one client error:%s", err.Error())
			continue
		}

		// get tcp addr
		addr := NewAddr(conn.RemoteAddr().String())

		// new connect obj for recv and send message later.
		connect := NewConnect(s, conn, addr)

		// save client connect
		if s.connMgr == nil {
			s.connMgr = make(map[Addr]*Connect)
		}
		s.connMgr[*addr] = connect

		// trigger user registered connect event
		s.callOnConnect(conn, addr)

		// use connect for send data, handle request,  receive data
		connect.worker()
	}
}
