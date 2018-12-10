package sckv

import (
	"net"
	"sync"
)

//请求服务
type Server struct {
	mu sync.Mutex
	ln net.Listener
	db StoreKV
}

func NewServer(cfg StoreConfig) *Server {
	return &Server{}
}

//请求服务开启
func (svr *Server) Start(addr string) error {
	return nil
}

//处理连接请求
func handleRequestConn(conn net.Conn) {

}

//服务关闭退出
func (svr *Server) Shutdown() error {
	return nil
}
