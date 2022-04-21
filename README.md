# codec

用于网络通信的数据包管理 类似于 netty 的 FixedLengthFrameDecoder

``` golang

import (
	"errors"
	"fmt"
	"github.com/BlueStorm001/codec"
	"log"
	"net"
	"sync"
	"time"
)

type Bootstrap struct {
	conn          net.Conn
	Addr          string
	codec         *codec.Packer
	Receive       func(bootstrap *Bootstrap, body []byte)
	MaxIdleTime   time.Time
	mu            sync.Mutex
	ConnectStatus bool
}

func New(addr string) *Bootstrap {
    return &Bootstrap{
        Addr: addr, 
        codec: codec.NewPacketFieldLength(4)，// netty 解包
    }
}

func (bootstrap *Bootstrap) Connect() (err error) {
	if bootstrap.Receive == nil {
		return errors.New("没有回调方法")
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp", bootstrap.Addr)
	if bootstrap.conn, err = net.Dial("tcp", tcpAddr.String()); err != nil {
		return err
	}
	bootstrap.MaxIdleTime = time.Now()
	go bootstrap.Read()
	return
}

// InspectStatus 状态检查
// 主要是利用服务器向客户端发起心跳包或数据记载时间
func (bootstrap *Bootstrap) InspectStatus() bool {
	if time.Now().Sub(bootstrap.MaxIdleTime).Seconds() > 30 {
		return false
	}
	return true
}
func (bootstrap *Bootstrap) Read() {
	for {
		var data = make([]byte, 1024*10)
		n, err := bootstrap.conn.Read(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		if bootstrap.codec.Receiver == nil {
			bootstrap.codec.Receiver = bootstrap.receive
		}
		bootstrap.MaxIdleTime = time.Now()
		bootstrap.codec.PacketFieldLengthDecode(data[:n])
	}
}

func (bootstrap *Bootstrap) receive(data []byte) {
	go bootstrap.Receive(bootstrap, data)
}

func (bootstrap *Bootstrap) Send(data []byte) (err error) {
	if bootstrap.conn == nil {
		return errors.New("链接失败")
	}
	bootstrap.mu.Lock()
	//netty 组包
	if _, err = bootstrap.conn.Write(bootstrap.codec.PacketFieldLengthEncode(data)); err != nil {
		bootstrap.ConnectStatus = false
	}
	bootstrap.mu.Unlock()
	return
}

func (bootstrap *Bootstrap) Close() {
	bootstrap.MaxIdleTime = time.Now().Add(time.Hour * -1)
	if bootstrap.conn == nil {
		return
	}
	bootstrap.conn.Close()
}



func main() {
    //以下是客户端链接服务器的测试
    bootstrap := New("127.0.0.1:8088")
    if err := bootstrap.Connect(); err == nil {
        bootstrap.Receive = func(bootstrap *Bootstrap, body []byte) {
            fmt.Println(body)
            if err = bootstrap.Send(body); err != nil {
                fmt.Println(err)
            }
        }
        bootstrap.Send([]byte("ok"))
    }
}
```