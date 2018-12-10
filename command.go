package sckv

import (
	"bufio"
	"errors"
	"io"
)

const (
	MAX_REQ_READ_LEN = 4096 //请求命令每次读取最大长度
)

var (
	ErrProtocolParse  = errors.New("redis protocol parse error")
	ErrProtocolFormat = errors.New("redis protocol format error")
	ErrProtocolEnd    = errors.New("redis protocol end error")
	ErrReaderNull     = errors.New("redis null error")
)

type RedisCmd struct {
	Args [][]byte
}

//命令对象
type RequestCmd struct {
	reader io.Reader
	buf    []byte
	head   int
	tail   int
}

//创建请求命令
func NewRequestCmd(reader io.Reader) *RequestCmd {
	return &RequestCmd{
		reader: bufio.NewReader(reader),
		buf:    make([]byte, MAX_REQ_READ_LEN),
	}
}

//命令解析
func (req *RequestCmd) ParseCommand() (redisCmds []RedisCmd, err error) {

	buf := req.buf[req.head:req.tail] //读取命令行数据

	if (req.head == req.tail) && len(req.buf) > MAX_REQ_READ_LEN {
		req.buf = req.buf[:MAX_REQ_READ_LEN]
		req.head = 0
		req.tail = 0
	}
	//有数据的情况
	if len(buf) > 0 {
	FORWARD:
		flag := buf[0] //获取第一位标识符
		if flag == '*' {
			var cmd RedisCmd
		INNER:
			for i := 1; i < len(buf); i++ {
				//处理命令行
				if buf[i] == '\n' {
					if buf[i-1] != '\r' {
						return nil, ErrProtocolFormat
					}
					//根据redis协议获取 * 后的表述命令数的数值
					argscount, err := btoi(buf[1 : i-1])
					if err != nil || argscount <= 0 {
						return nil, ErrProtocolParse
					}
					i++
					//截获参数
					for j := 0; j < argscount; j++ {
						if i < len(buf) {
							if buf[i] != '$' {
								//正常情况必须跟着'$参数长度'的标识
								return nil, ErrProtocolFormat
							}
							//边界开始位置
							argsStart := i
							for ; i < len(buf); i++ {
								//到达下一个边界前位置
								if buf[i] == '\n' {
									if buf[i-1] != '\r' {
										return nil, ErrProtocolFormat
									}
									//根据两个边界,截取参数长度
									blen := buf[argsStart+1 : i-1]
									argLen, err := btoi(blen)
									if err != nil || argLen <= 0 {
										//非法长度,返回错误
										return nil, ErrProtocolParse
									}

									if i+argLen+2 >= len(buf) {
										break INNER
									}

									if buf[i+argLen+2] != '\n' || buf[i+argLen+1] != '\r' {
										return nil, ErrProtocolEnd
									}
									field := buf[i+1 : i+1+argLen]
									cmd.Args = append(cmd.Args, field)
									i = i + argLen + 3 //指向下一个$符号位置
									break
								}
							}
						}
					}
					//调整b,获取剩余的部分
					redisCmds = append(redisCmds, cmd)
					buf = buf[i:]
					//还有命令还没处理完,例如以下的多命令情况
					//"*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n*2\r\n$3\r\nGET\r\n$5\r\nmyCmd\r\n"
					if len(buf) > 0 {
						goto FORWARD
					}
				}
			}
		}
		req.head = req.tail - len(buf)
	}

	if len(redisCmds) > 0 {
		return redisCmds, nil
	}

	if req.reader == nil {
		return nil, ErrReaderNull
	}
	//尝试申请空间
	req.tryGrow()
	//把缓冲区写满
	n, err := req.reader.Read(req.buf[req.tail:])
	if err != nil {
		return nil, err
	}
	req.tail += n
	return req.ParseCommand()
}

//请求缓冲区增长
func (req *RequestCmd) tryGrow() {
	if req.tail >= len(req.buf) {
		if req.tail-req.head == 0 {
			req.head = 0
			req.tail = 0
		} else {
			req.grow()
		}
	}
}

//扩展缓冲区
func (req *RequestCmd) grow() {
	newBuff := make([]byte, len(req.buf)*2)
	copy(newBuff, req.buf)
	req.buf = newBuff
}

func btoi(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	i := 0
	sign := 1
	if data[0] == '-' {
		i++
		sign *= -1
	}
	if i >= len(data) {
		return 0, ErrProtocolParse
	}
	var l int
	for ; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, ErrProtocolParse
		}
		l = l*10 + int(c-'0')
	}
	return sign * l, nil
}
