package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"unsafe"
)

type ClientMsg struct {
	To      string  `json:"to"`
	Msg     string  `json:"msg"`
	Datalen uintptr `json:"datalen"`
}

func Help() {
	fmt.Println("1. set:your name --设置用户名")
	fmt.Println("2. all:your msg --broadcast广播")
	fmt.Println("3. anyone:your msg --private msg私聊")
	fmt.Println("4. quit --quit退出")

}
func handle_conn(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Panic("Failed to Read", err)
		}
		fmt.Println(" < this is response msg >>> " + string(buf[:n]))
		fmt.Printf("I want to say >")
	}
}
func main() {

	// 连接服务器
	conn, err := net.Dial("tcp", "10.5.253.119:8989")
	if err != nil {
		log.Panic("Failed to Dial", err)
	}
	defer conn.Close()
	go handle_conn(conn)
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to this chatroom >")
	Help()
	for {
		fmt.Printf("I want to say >")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Panic("Failed to ReadString", err)
		}
		msg = strings.Trim(msg, "\r\n")
		if msg == "quit" {
			fmt.Println("byte bye")
			break
		}
		if msg == "help" {
			Help()
			continue
		}
		msgs := strings.Split(msg, ":")
		if len(msgs) == 2 {
			var climsg ClientMsg
			climsg.To = msgs[0]
			climsg.Msg = msgs[1]
			climsg.Datalen = unsafe.Sizeof(climsg)
			data, err := json.Marshal(climsg)
			if err != nil {
				fmt.Println("Failed to Marshal", err, climsg)
				continue
			}
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println("Failed to write", err, climsg)
				break
			} else {
				fmt.Println("OK")
			}
		}
	}
}
