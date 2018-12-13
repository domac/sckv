package sckv

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

var CONN_ERROR error = errors.New("STOP LISTENING")

const (
	default_server_max_connects = 1024
)

//会话对象
type Session struct {
	conn   *net.TCPConn
	reader *RequestCmdReader
	writer *ResponseCmdWriter

	//连接释放回调
	releaseOnce sync.Once
	release     func()
}

func NewSession(conn *net.TCPConn, release func()) *Session {
	return &Session{
		conn:    conn,
		reader:  NewRequestCmdReader(conn),
		writer:  NewResponseCmdWriter(conn),
		release: release,
	}
}

//获取远程连接地址
func (sess *Session) RemoteAddr() string {
	return sess.conn.RemoteAddr().String()
}

func (sess *Session) Receive() ([]RedisCmd, error) {
	cmds, err := sess.reader.ParseCommand()
	if err != nil {
		sess.Close()
	}
	return cmds, err
}

func (sess *Session) WriteOK() error {
	err := sess.writer.WriteOK()
	if err != nil {
		sess.Close()
	}
	return err
}

//会话关闭
func (sess *Session) Close() error {
	err := sess.conn.Close()
	sess.releaseOnce.Do(sess.release)
	return err
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
	mu          sync.Mutex
	handler     ReqHandler
	hostport    string
	keepalive   time.Duration
	stopChan    chan bool
	isShutdown  bool
	maxconnects int
}

func NewServer(addr string, handler ReqHandler, maxconnects int) *Server {

	if maxconnects <= 0 {
		maxconnects = default_server_max_connects
	}

	return &Server{
		hostport:    addr,
		handler:     handler,
		stopChan:    make(chan bool, 1),
		keepalive:   5 * time.Minute,
		maxconnects: maxconnects,
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

	tcpListener := NewTcpListener(listener, svr.stopChan, svr.keepalive, svr.maxconnects)

	svr.serve(tcpListener)
	return nil
}

func (svr *Server) serve(ln *TcpListener) {
	for !svr.isShutdown {
		sess, err := ln.Accept()
		if err != nil {
			log.Printf("server accept error: %s\n", err.Error())
			continue
		}
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
	sem       chan struct{}
}

func NewTcpListener(l *net.TCPListener, stopchan chan bool, keepalive time.Duration, n int) *TcpListener {
	return &TcpListener{l, stopchan, keepalive, make(chan struct{}, n)}
}

func (ln *TcpListener) acquire() {
	ln.sem <- struct{}{}
}

func (ln *TcpListener) release() {
	<-ln.sem
}

//accept 处理
func (ln *TcpListener) Accept() (*Session, error) {
	ln.acquire()
	for {
		conn, err := ln.AcceptTCP()
		select {
		case <-ln.stop:
			ln.release()
			return nil, CONN_ERROR
		default:
			//do nothing
		}
		if nil == err {
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(ln.keepalive)
		} else {
			ln.release()
			return nil, err
		}
		return NewSession(conn, ln.release), err
	}
}
