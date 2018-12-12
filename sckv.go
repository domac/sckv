package sckv

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

var CONN_ERROR error = errors.New("STOP LISTENING")

//会话对象
type Session struct {
	conn   *net.TCPConn
	reader *RequestCmdReader
	writer *ResponseCmdWriter
}

func NewSession(conn *net.TCPConn) *Session {
	return &Session{
		conn:   conn,
		reader: NewRequestCmdReader(conn),
		writer: NewResponseCmdWriter(conn),
	}
}

//获取远程连接地址
func (sess *Session) RemoteAddr() string {
	return sess.conn.RemoteAddr().String()
}

func (sess *Session) GetReader() *RequestCmdReader {
	return sess.reader
}

func (sess *Session) GetWriter() *ResponseCmdWriter {
	return sess.writer
}

type ReqHandler interface {
	HandleSession(session *Session)
}

type HandlerFunc func(session *Session)

func (hf HandlerFunc) HandleSession(session *Session) {
	hf(session)
}

//请求服务
type Server struct {
	mu         sync.Mutex
	handler    ReqHandler
	hostport   string
	keepalive  time.Duration
	stopChan   chan bool
	isShutdown bool
}

func NewServer(addr string, handler ReqHandler) *Server {
	return &Server{
		hostport:  addr,
		handler:   handler,
		stopChan:  make(chan bool, 1),
		keepalive: 5 * time.Minute,
	}
}

//请求服务开启
func (svr *Server) ListenAndServe() error {
	addr, err := net.ResolveTCPAddr("tcp4", svr.hostport)
	if nil != err {
		return err
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if nil != err {
		return err
	}

	tcpListener := &TcpListener{listener, svr.stopChan, svr.keepalive}

	svr.serve(tcpListener)
	return nil
}

func (svr *Server) serve(ln *TcpListener) {
	for !svr.isShutdown {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("server accept error: %s\n", err.Error())
			continue
		}

		sess := NewSession(conn)
		go svr.handler.HandleSession(sess)
	}
}

//服务关闭退出
func (svr *Server) Shutdown() {
	svr.isShutdown = true
	close(svr.stopChan)
	log.Println("sckv tcp server shutdown...")
}

type TcpListener struct {
	*net.TCPListener
	stop      chan bool
	keepalive time.Duration
}

func (ln *TcpListener) Accept() (*net.TCPConn, error) {
	for {
		conn, err := ln.AcceptTCP()
		select {
		case <-ln.stop:
			return nil, CONN_ERROR
		default:
			//do nothing
		}
		if nil == err {
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(ln.keepalive)
		} else {
			return nil, err
		}
		return conn, err
	}
}
