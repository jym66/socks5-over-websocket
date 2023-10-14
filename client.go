package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

func createWsConnect(wsServerAddr string, header http.Header) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsServerAddr, header)
	//log.Println("创建websocket连接")
	return conn, err
}

func createSocks(socksProxyAddr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", socksProxyAddr)
	return ln, err
}

func main() {
	// 启动 WebSocket 服务器
	wsServerAddr := "ws://127.0.0.1:8080" //WebSocket 服务器地址
	// 启动 SOCKS 5 代理服务器并监听指定端口
	ln, err := createSocks("127.0.0.1:1080")
	if err != nil {
		log.Fatal("SOCKS 5 proxy server error:", err)
	}
	fmt.Printf("SOCKS 5 proxy server is listening on %s\n", "127.0.0.1:1080")
	for {
		socksConn, err := ln.Accept()
		if err != nil {
			log.Println("SOCKS 5 proxy accept error:", err)
			continue
		}
		//建立ws连接
		WsConn, err := createWsConnect(wsServerAddr, nil)
		if err != nil {
			log.Println("WebSocket connection error:", err)
			socksConn.Close()
			continue
		}

		go handleSocks5Connection(socksConn, WsConn)
	}
}

func handleSocks5Connection(socksConn net.Conn, WsConn *websocket.Conn) {
	// 从SOCKS客户端读取数据并转发到WebSocket服务器
	log.Println("进入handleSocks5Connection")
	go func() {
		for {
			buffer := make([]byte, 4*1024)
			//log.Println("==>从SOCKS客户端读取数据并转发到WebSocket服务器")
			n, err := socksConn.Read(buffer)
			if err != nil {
				log.Println("==>从SOCKS客户端读取数据错误:", err)
				//WsConn.Close()
				return
			}
			err = WsConn.WriteMessage(websocket.BinaryMessage, buffer[:n])
			if err != nil {
				log.Println("==>从SOCKS客户端读取数据并转发到WebSocket服务器错误:", err)
				return
			}
		}
	}()

	// 从WebSocket服务器读取数据并转发到SOCKS客户端
	go func() {
		for {
			//log.Println("==>从WebSocket服务器读取数据并转发到SOCKS客户端")
			_, p, err := WsConn.ReadMessage()
			if err != nil {
				log.Println("==>从WebSocket服务器读取数据错误:", err)
				return
			}
			_, err = socksConn.Write(p)
			if err != nil {
				log.Println("==>从WebSocket服务器读取数据并转发到SOCKS客户端错误:", err)
				return
			}
		}
	}()
}
