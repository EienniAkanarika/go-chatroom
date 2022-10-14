package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

// 实现目标
// 1.实现命令行般的聊天室，在客户端连接后，可在命令行进行昵称设置、发送广播消息、发送一对一消息
// 2.客户端连接后，所有已连接用户要收到广播。
// 3....
// goroutine通信结构体

type ChatMsg struct {
	From, To, Msg string
}

// 与客户端通信结构体

type ClientMsg struct {
	To      string  `json:"to"`
	Msg     string  `json:"msg"`
	Datalen uintptr `json:"datalen"`
}

// channel 消息中心使用
var chan_msgcenter chan ChatMsg
var mapName2CliAddr map[string]string
var mapCliaddr2Clients map[string]net.Conn

func logout(conn net.Conn, from string) {
	defer conn.Close()
	delete(mapCliaddr2Clients, from)
	msg := ChatMsg{from, "all", from + "-> logout"}
	chan_msgcenter <- msg
}
func handle_conn(conn net.Conn) {
	from := conn.RemoteAddr().String()
	mapCliaddr2Clients[from] = conn
	msg := ChatMsg{from, "all", from + "-> login"}
	chan_msgcenter <- msg
	defer logout(conn, from)
	// 分析消息
	buf := make([]byte, 256)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Failed to Read msg", err, from)
			break
		}
		if n > 0 {
			var climsg ClientMsg
			err = json.Unmarshal(buf[:n], &climsg)
			if err != nil {
				fmt.Println("Failed to Unmarshal msg", err, string(buf[:n]))
				continue
			}
			//if climsg.Datalen != unsafe.Sizeof(climsg) {
			//	fmt.Println("Msg format error", climsg)
			//	continue
			//}
			chatmsg := ChatMsg{from, "all", climsg.Msg}
			switch climsg.To {
			case "all":
			case "set":
				mapName2CliAddr[climsg.Msg] = from
				chatmsg.Msg = "You set your name is " + climsg.Msg + " success"
				chatmsg.From = "server"
				chatmsg.To = climsg.Msg
			default:
				chatmsg.To = climsg.To
			}
			chan_msgcenter <- chatmsg
		}
	}
}

// 消息中心
func msg_center() {
	for {
		msg := <-chan_msgcenter
		go send_msg(msg)
	}
}

// 消息中心处理操作
func send_msg(msg ChatMsg) {
	data, err := json.Marshal(msg)
	var content = string(data[:])
	content = strings.Replace(content, "\\u003e", ">", -1)
	fmt.Println(content)
	data = []byte(content)
	if err != nil {
		fmt.Println("Filed to Marshal	", err, msg)
		return
	}
	// 广播消息
	if msg.To == "all" {
		for _, v := range mapCliaddr2Clients {
			if msg.From == v.RemoteAddr().String() {
				_, err := v.Write(data)
				if err != nil {
					fmt.Println("Failed to send")
					return
				}
			}
		}
	} else {
		// 私聊
		from, ok := mapName2CliAddr[msg.To]
		if !ok {
			fmt.Println("User not exists", msg.To)
			conn, ok := mapCliaddr2Clients[msg.From]
			if !ok {
				fmt.Println("Client not exists", msg.From)
				return
			}
			response := []byte("user not exists")
			_, err := conn.Write(response)
			if err != nil {
				fmt.Println("Error writing")
				return
			}
		}
		fmt.Println("Find the user", msg.To)
		conn, ok := mapCliaddr2Clients[from]
		if !ok {
			fmt.Println("client not exists", from, msg.To)
			return
		}
		fmt.Println("Find the client")
		_, err := conn.Write(data)
		if err != nil {
			fmt.Println("Failed to send")
			return
		}
	}
}
func main() {
	mapCliaddr2Clients = make(map[string]net.Conn)
	mapName2CliAddr = make(map[string]string)
	chan_msgcenter = make(chan ChatMsg)
	// 绑定IP与监听端口
	listener, err := net.Listen("tcp", ":8989")
	if err != nil {
		log.Panic("Failed to listen", err)
	}
	defer listener.Close()
	go msg_center()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept", err)
			break
		}
		go handle_conn(conn)
	}
}
