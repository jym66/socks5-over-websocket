package main

import (
	"fmt"
	"github.com/armon/go-socks5"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   4 * 1024,
	WriteBufferSize:  4 * 1024,
	HandshakeTimeout: time.Second * 4,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func createSocks5() {
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	socketPath := "./socks5.sock"

	// Remove the socket file if it already exists
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		os.Remove(socketPath)
	}

	// Create SOCKS5 proxy using Unix socket
	if err := server.ListenAndServe("unix", socketPath); err != nil {
		panic(err)
	}
}
func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("进入handleWebSocketConnection")
	remoteConn, err := net.Dial("unix", "./socks5.sock")
	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket connection failed:", err)
		return
	}

	if err != nil {
		log.Fatal("Failed to connect to remote server:", err)
	}
	defer remoteConn.Close()

	// 从 WebSocket 读取数据并转发到远程服务器
	go func() {
		for {
			//log.Println("==> 从WebSocket读取数据并转发到远程服务器")
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("==> 从WebSocket读取数据错误", err)
				return
			}
			_, err = remoteConn.Write(p)
			if err != nil {
				log.Println("==> 从WebSocket读取数据并转发到远程服务器错误", err)
				return
			}
		}
	}()

	// 从远程服务器读取数据并转发到 WebSocket
	for {
		//log.Println("==> 从远程服务器读取数据并转发到WebSocket")
		buffer := make([]byte, 4*1024)
		n, err := remoteConn.Read(buffer)
		if err != nil {
			log.Println("==> 从远程服务器读取数据错误", err)
			return
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			log.Println("==> 从远程服务器读取数据转发到WebSocket错误", err)
			return
		}
	}
}

func main() {
	go createSocks5()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocketConnection(w, r)
	})
	// 启动 WebSocket 服务器并监听指定端口
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Println("WebSocket server error:", err)
	}
}
