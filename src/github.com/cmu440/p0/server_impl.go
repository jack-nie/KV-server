// Implementation of a KeyValueServer. Students should write their code in this file.

package p0

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const bufSize = 1024

var buff = make([]byte, bufSize)

type channelMap struct {
	writeMap map[string]chan []byte
}

func (c *channelMap) Put(key string) chan []byte {
	result := make(chan []byte, 500)
	c.writeMap[key] = result
	return result
}

func (c *channelMap) Get(key string) chan []byte {
	return c.writeMap[key]
}

type keyValueServer struct {
	listen          net.Listener
	clientCount     int
	clients         map[string]net.Conn
	clientMsgs      map[string]chan []byte
	addClientChan   chan net.Conn
	closeClientChan chan string
	serverWriteMap  channelMap
	serverWriteChan chan string
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	return &keyValueServer{
		listen:          nil,
		clientCount:     0,
		clients:         make(map[string]net.Conn),
		clientMsgs:      make(map[string]chan []byte),
		addClientChan:   make(chan net.Conn),
		closeClientChan: make(chan string),
		serverWriteMap:  channelMap{writeMap: make(map[string]chan []byte)},
		serverWriteChan: make(chan string),
	}
}

func (kvs *keyValueServer) Close() {
	for _, conn := range kvs.clients {
		conn.Close()
		kvs.clientCount--
	}
	kvs.listen.Close()
}

func (kvs *keyValueServer) Start(port int) error {
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Failed to listen on port: ", port)
	}
	kvs.listen = listen
	go kvs.handleEvent()
	go kvs.handleAccept()
	return err
}

func (kvs *keyValueServer) handleAccept() {
	for {
		conn, err := kvs.listen.Accept()
		fmt.Printf("客户端：%s 已经连接!\n", conn.RemoteAddr().String())
		if err != nil {
			fmt.Printf("错误： %s\n", err.Error())
		}
		kvs.addClientChan <- conn
	}
}

func (kvs *keyValueServer) Count() int {
	return kvs.clientCount
}

func (kvs *keyValueServer) handleConn(conn net.Conn) {
	if conn == nil {
		return
	}

	for {
		n, err := conn.Read(buff)
		if err == io.EOF {
			fmt.Printf("远程主机:%s 已经关闭!\n", conn.RemoteAddr().String())
			kvs.closeClientChan <- conn.RemoteAddr().String()
			return
		}
		if err != nil {
			fmt.Printf("错误： %s", err.Error())
			kvs.closeClientChan <- conn.RemoteAddr().String()
			return
		}
		if string(buff[:n]) == "exit\r\n" {
			fmt.Printf("客户端: %s 已经退出!\r\n", conn.RemoteAddr().String())
			kvs.closeClientChan <- conn.RemoteAddr().String()
			return
		}
		if n > 0 {
			fmt.Printf("接收数据: %s", string(buff[:n]))
		}
		kvs.handleAction(string(buff[:n]), conn)
	}
}

func (kvs *keyValueServer) handleAction(input string, conn net.Conn) {
	fmt.Println(input)
	command := strings.Fields(input)
	addr := conn.RemoteAddr().String()
	switch command[0] {
	case "get":
		kvs.serverWriteMap.Put(addr) <- []byte(get(command[1]))
	case "put":
		put(command[1], []byte(command[2]))
		kvs.serverWriteMap.Put(addr) <- []byte("ok")
	default:
		kvs.serverWriteMap.Put(addr) <- []byte("输入的命令不合法！")
	}
	kvs.serverWriteChan <- conn.RemoteAddr().String()
}

func (kvs *keyValueServer) handleEvent() {
	for {
		select {
		case conn := <-kvs.addClientChan:
			kvs.addClient(conn)
			go kvs.handleConn(conn)
		case clientAddr := <-kvs.closeClientChan:
			kvs.removeClient(clientAddr)
		case clientAddr := <-kvs.serverWriteChan:
			kvs.handleWrite(clientAddr)
		}
	}
}

func (kvs *keyValueServer) handleWrite(addr string) {
	content := <-kvs.serverWriteMap.Get(addr)
	kvs.clients[addr].Write(content)
}

func (kvs *keyValueServer) addClient(conn net.Conn) {
	if _, ok := kvs.clientMsgs[conn.RemoteAddr().String()]; !ok {
		kvs.clientMsgs[conn.RemoteAddr().String()] = make(chan []byte, 100)
	}
	kvs.clients[conn.RemoteAddr().String()] = conn
	kvs.clientCount++
}

func (kvs *keyValueServer) removeClient(addr string) {
	if conn, ok := kvs.clients[addr]; ok {
		conn.Close()
		delete(kvs.clients, addr)
		kvs.clientCount--
	}
}
